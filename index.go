// Package toolindex provides a global registry and search layer for tools.
// It ingests toolmodel.Tool and toolmodel.ToolBackend and provides progressive
// discovery (summaries + namespaces) and canonical lookup by tool ID.
package toolindex

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/jonwraymond/toolmodel"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// MaxShortDescriptionLen is the maximum length of the ShortDescription field in Summary.
const MaxShortDescriptionLen = 120

// Error values for consistent error handling by callers.
var (
	ErrNotFound       = errors.New("tool not found")
	ErrInvalidTool    = errors.New("invalid tool")
	ErrInvalidBackend = errors.New("invalid backend")
	ErrInvalidCursor  = errors.New("invalid cursor")
)

// Summary represents a lightweight view of a tool for search results.
// It contains only the essential information for display and discovery,
// without the full schema payloads.
type Summary struct {
	ID               string   `json:"id"`
	Name             string   `json:"name"`
	Namespace        string   `json:"namespace,omitempty"`
	ShortDescription string   `json:"shortDescription,omitempty"`
	Tags             []string `json:"tags,omitempty"`
}

// SearchDoc is the internal/exported struct used by Searcher implementations.
// It contains precomputed search data for efficient querying.
type SearchDoc struct {
	ID      string  // Canonical tool ID
	DocText string  // Lowercased concatenation of name/namespace/description/tags
	Summary Summary // Prebuilt summary for fast return
}

// Index defines the interface for a tool registry.
type Index interface {
	// Registration
	RegisterTool(tool toolmodel.Tool, backend toolmodel.ToolBackend) error
	RegisterTools(regs []ToolRegistration) error
	RegisterToolsFromMCP(serverName string, tools []toolmodel.Tool) error

	// Unregistration
	UnregisterBackend(toolID string, kind toolmodel.BackendKind, backendID string) error

	// Lookup
	GetTool(id string) (toolmodel.Tool, toolmodel.ToolBackend, error)
	GetAllBackends(id string) ([]toolmodel.ToolBackend, error)

	// Discovery
	Search(query string, limit int) ([]Summary, error)
	SearchPage(query string, limit int, cursor string) ([]Summary, string, error)
	ListNamespaces() ([]string, error)
	ListNamespacesPage(limit int, cursor string) ([]string, string, error)
}

// ToolRegistration pairs a tool with its backend for batch registration.
type ToolRegistration struct {
	Tool    toolmodel.Tool
	Backend toolmodel.ToolBackend
}

// BackendSelector is a function that selects the default backend from a list.
type BackendSelector func([]toolmodel.ToolBackend) toolmodel.ToolBackend

// Searcher is the interface for search implementations.
type Searcher interface {
	Search(query string, limit int, docs []SearchDoc) ([]Summary, error)
}

// ChangeType describes a mutation event in the index.
type ChangeType string

const (
	ChangeRegistered     ChangeType = "registered"
	ChangeUpdated        ChangeType = "updated"
	ChangeBackendRemoved ChangeType = "backend_removed"
	ChangeToolRemoved    ChangeType = "tool_removed"
	ChangeRefreshed      ChangeType = "refreshed"
)

// ChangeEvent captures a mutation in the index for reactive integration.
type ChangeEvent struct {
	Type    ChangeType
	ToolID  string
	Backend toolmodel.ToolBackend
	Version uint64
}

// ChangeListener receives change events from an Index implementation.
type ChangeListener func(ChangeEvent)

// ChangeNotifier is an optional interface for receiving change events.
type ChangeNotifier interface {
	OnChange(listener ChangeListener) (unsubscribe func())
}

// Refresher is an optional interface for forcing a refresh of cached search docs.
type Refresher interface {
	Refresh() uint64
}

// IndexOptions configures the behavior of an Index implementation.
type IndexOptions struct {
	BackendSelector BackendSelector
	Searcher        Searcher
}

// toolRecord holds all data for a single registered tool.
type toolRecord struct {
	tool           toolmodel.Tool
	backends       []toolmodel.ToolBackend
	backendKeys    map[string]int // maps backend identity key to index in backends slice
	normalizedTags []string       // normalized tags for search
	docText        string         // cached search doc text
	summary        Summary        // cached summary
}

// InMemoryIndex is the default in-memory implementation of Index.
type InMemoryIndex struct {
	mu              sync.RWMutex
	tools           map[string]*toolRecord // keyed by tool ID
	namespaces      map[string]struct{}    // set of namespaces
	namespaceCounts map[string]int         // number of tools per namespace
	backendSelector BackendSelector
	searcher        Searcher
	listeners       []listenerEntry
	nextListenerID  uint64

	// Search doc cache
	searchDocs        []SearchDoc
	searchDocsDirty   bool
	searchDocsVersion uint64
	indexVersion      uint64
	searchDocsBuilds  int // for test visibility
}

type listenerEntry struct {
	id uint64
	fn ChangeListener
}

// NewInMemoryIndex creates a new in-memory tool index.
func NewInMemoryIndex(opts ...IndexOptions) *InMemoryIndex {
	idx := &InMemoryIndex{
		tools:           make(map[string]*toolRecord),
		namespaces:      make(map[string]struct{}),
		namespaceCounts: make(map[string]int),
		backendSelector: DefaultBackendSelector,
		searcher:        &lexicalSearcher{},
	}

	if len(opts) > 0 {
		opt := opts[0]
		if opt.BackendSelector != nil {
			idx.backendSelector = opt.BackendSelector
		}
		if opt.Searcher != nil {
			idx.searcher = opt.Searcher
		}
	}

	return idx
}

// OnChange registers a listener for index mutations.
// Returns an unsubscribe function.
func (idx *InMemoryIndex) OnChange(listener ChangeListener) func() {
	if listener == nil {
		return func() {}
	}
	idx.mu.Lock()
	idx.nextListenerID++
	entry := listenerEntry{id: idx.nextListenerID, fn: listener}
	idx.listeners = append(idx.listeners, entry)
	idx.mu.Unlock()

	return func() {
		idx.removeListener(entry.id)
	}
}

func (idx *InMemoryIndex) removeListener(id uint64) {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	for i, entry := range idx.listeners {
		if entry.id == id {
			idx.listeners = append(idx.listeners[:i], idx.listeners[i+1:]...)
			return
		}
	}
}

// Refresh rebuilds the search docs cache and emits a refresh event.
func (idx *InMemoryIndex) Refresh() uint64 {
	idx.mu.Lock()
	idx.markSearchDocsDirtyLocked()
	idx.rebuildSearchDocsLocked()
	version := idx.indexVersion
	listeners := idx.snapshotListenersLocked()
	idx.mu.Unlock()

	notifyListeners(listeners, ChangeEvent{Type: ChangeRefreshed, Version: version})
	return version
}

// DefaultBackendSelector implements the default priority: local > provider > mcp.
// Exported so other modules (for example, toolrun) can match the same policy.
func DefaultBackendSelector(backends []toolmodel.ToolBackend) toolmodel.ToolBackend {
	if len(backends) == 0 {
		return toolmodel.ToolBackend{}
	}

	// Priority order: local > provider > mcp
	for _, b := range backends {
		if b.Kind == toolmodel.BackendKindLocal {
			return b
		}
	}
	for _, b := range backends {
		if b.Kind == toolmodel.BackendKindProvider {
			return b
		}
	}
	for _, b := range backends {
		if b.Kind == toolmodel.BackendKindMCP {
			return b
		}
	}

	return backends[0]
}

// backendIdentity returns a unique key for a backend based on its kind and identity fields.
func backendIdentity(backend toolmodel.ToolBackend) string {
	switch backend.Kind {
	case toolmodel.BackendKindMCP:
		if backend.MCP != nil {
			return encodeIdentity(string(backend.Kind), backend.MCP.ServerName)
		}
	case toolmodel.BackendKindProvider:
		if backend.Provider != nil {
			return encodeIdentity(string(backend.Kind), backend.Provider.ProviderID, backend.Provider.ToolID)
		}
	case toolmodel.BackendKindLocal:
		if backend.Local != nil {
			return encodeIdentity(string(backend.Kind), backend.Local.Name)
		}
	}
	return ""
}

// encodeIdentity builds an unambiguous identity string using length-prefixed parts.
// This prevents collisions when fields include separators like ":".
func encodeIdentity(parts ...string) string {
	var b strings.Builder
	for i, part := range parts {
		if i > 0 {
			b.WriteByte('|')
		}
		b.WriteString(strconv.Itoa(len(part)))
		b.WriteByte(':')
		b.WriteString(part)
	}
	return b.String()
}

// validateBackend checks if a backend is valid.
func validateBackend(backend toolmodel.ToolBackend) error {
	switch backend.Kind {
	case toolmodel.BackendKindMCP:
		if backend.MCP == nil || backend.MCP.ServerName == "" {
			return fmt.Errorf("%w: MCP backend requires ServerName", ErrInvalidBackend)
		}
	case toolmodel.BackendKindProvider:
		if backend.Provider == nil {
			return fmt.Errorf("%w: Provider backend requires Provider details", ErrInvalidBackend)
		}
		if backend.Provider.ProviderID == "" {
			return fmt.Errorf("%w: Provider backend requires ProviderID", ErrInvalidBackend)
		}
		if backend.Provider.ToolID == "" {
			return fmt.Errorf("%w: Provider backend requires ToolID", ErrInvalidBackend)
		}
	case toolmodel.BackendKindLocal:
		if backend.Local == nil || backend.Local.Name == "" {
			return fmt.Errorf("%w: Local backend requires Name", ErrInvalidBackend)
		}
	default:
		return fmt.Errorf("%w: unknown backend kind %q", ErrInvalidBackend, backend.Kind)
	}
	return nil
}

// toolMCPFieldsEqual compares the MCP-spec fields of two tools for equivalence.
// It compares all MCP Tool fields:
// - Name, Title, Description (string fields)
// - InputSchema, OutputSchema (schema fields, compared via JSON normalization)
// - Annotations (ToolAnnotations pointer)
// - Icons (slice of Icon)
// - Meta (map[string]any for additional metadata)
//
// toolmodel extensions (Namespace, Version, Tags) are intentionally ignored
// as they are toolindex-specific and may legitimately differ across backends.
func toolMCPFieldsEqual(a, b toolmodel.Tool) bool {
	// Compare string fields
	if a.Name != b.Name {
		return false
	}
	if a.Title != b.Title {
		return false
	}
	if a.Description != b.Description {
		return false
	}

	// Compare InputSchema (deep equality via JSON comparison)
	if !jsonEqual(a.InputSchema, b.InputSchema) {
		return false
	}

	// Compare OutputSchema if present
	if !jsonEqual(a.OutputSchema, b.OutputSchema) {
		return false
	}

	// Compare Annotations if present
	if !annotationsEqual(a.Annotations, b.Annotations) {
		return false
	}

	// Compare Icons
	if !iconsEqual(a.Icons, b.Icons) {
		return false
	}

	// Compare Meta (additional metadata)
	if !metaEqual(a.Meta, b.Meta) {
		return false
	}

	return true
}

// jsonEqual compares two interface{} values for JSON-structural equality.
// Handles json.RawMessage, []byte, maps, slices, and primitive types.
func jsonEqual(a, b any) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	// First, check if b is bytes/RawMessage and a is not - handle this case
	switch b.(type) {
	case json.RawMessage, []byte:
		// If b is bytes but a is not, swap and use jsonEqualBytes
		switch av := a.(type) {
		case json.RawMessage:
			return jsonEqualBytes([]byte(av), b)
		case []byte:
			return jsonEqualBytes(av, b)
		default:
			// a is not bytes, but b is - swap the comparison
			return jsonEqual(b, a)
		}
	}

	// Type switch for common cases
	switch av := a.(type) {
	case json.RawMessage:
		return jsonEqualBytes([]byte(av), b)
	case []byte:
		return jsonEqualBytes(av, b)
	case map[string]any:
		bv, ok := b.(map[string]any)
		if !ok {
			return false
		}
		if len(av) != len(bv) {
			return false
		}
		for k, va := range av {
			vb, exists := bv[k]
			if !exists || !jsonEqual(va, vb) {
				return false
			}
		}
		return true
	case []any:
		bv, ok := b.([]any)
		if !ok {
			return false
		}
		if len(av) != len(bv) {
			return false
		}
		for i := range av {
			if !jsonEqual(av[i], bv[i]) {
				return false
			}
		}
		return true
	case string:
		bv, ok := b.(string)
		return ok && av == bv
	case float64:
		bv, ok := b.(float64)
		return ok && av == bv
	case bool:
		bv, ok := b.(bool)
		return ok && av == bv
	case int:
		// Handle int (Go may use int in some contexts)
		switch bv := b.(type) {
		case int:
			return av == bv
		case float64:
			return float64(av) == bv
		}
		return false
	default:
		// Fallback: direct comparison
		return a == b
	}
}

// jsonEqualBytes compares a byte slice (JSON) against another value.
// It unmarshals the bytes to a normalized form and compares.
func jsonEqualBytes(aBytes []byte, b any) bool {
	// Handle b also being bytes/RawMessage
	var bBytes []byte
	switch bv := b.(type) {
	case json.RawMessage:
		bBytes = []byte(bv)
	case []byte:
		bBytes = bv
	default:
		// b is not bytes, unmarshal a and compare
		var aVal any
		if err := json.Unmarshal(aBytes, &aVal); err != nil {
			return false
		}
		return jsonEqual(aVal, b)
	}

	// Both are bytes - unmarshal both and compare
	var aVal, bVal any
	if err := json.Unmarshal(aBytes, &aVal); err != nil {
		return false
	}
	if err := json.Unmarshal(bBytes, &bVal); err != nil {
		return false
	}
	return jsonEqual(aVal, bVal)
}

// iconsEqual compares two slices of mcp.Icon for equality.
func iconsEqual(a, b []mcp.Icon) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !iconEqual(a[i], b[i]) {
			return false
		}
	}
	return true
}

// iconEqual compares two mcp.Icon structs for equality.
func iconEqual(a, b mcp.Icon) bool {
	if a.Source != b.Source {
		return false
	}
	if a.MIMEType != b.MIMEType {
		return false
	}
	if a.Theme != b.Theme {
		return false
	}
	// Compare Sizes slices
	if len(a.Sizes) != len(b.Sizes) {
		return false
	}
	for i := range a.Sizes {
		if a.Sizes[i] != b.Sizes[i] {
			return false
		}
	}
	return true
}

// metaEqual compares two mcp.Meta (map[string]any) for equality.
func metaEqual(a, b mcp.Meta) bool {
	if len(a) != len(b) {
		return false
	}
	for k, va := range a {
		vb, exists := b[k]
		if !exists || !jsonEqual(va, vb) {
			return false
		}
	}
	return true
}

// annotationsEqual compares tool annotations for equality.
func annotationsEqual(a, b *mcp.ToolAnnotations) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	// Compare ToolAnnotations fields
	if a.Title != b.Title {
		return false
	}
	if a.ReadOnlyHint != b.ReadOnlyHint {
		return false
	}
	if a.IdempotentHint != b.IdempotentHint {
		return false
	}
	if !boolPtrEqual(a.DestructiveHint, b.DestructiveHint) {
		return false
	}
	if !boolPtrEqual(a.OpenWorldHint, b.OpenWorldHint) {
		return false
	}
	return true
}

// boolPtrEqual compares two *bool values for equality.
func boolPtrEqual(a, b *bool) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func (idx *InMemoryIndex) addNamespaceLocked(namespace string) {
	if idx.namespaceCounts == nil {
		idx.namespaceCounts = make(map[string]int)
	}
	idx.namespaceCounts[namespace]++
	idx.namespaces[namespace] = struct{}{}
}

func (idx *InMemoryIndex) removeNamespaceLocked(namespace string) {
	if idx.namespaceCounts == nil {
		return
	}
	count, ok := idx.namespaceCounts[namespace]
	if !ok {
		return
	}
	if count <= 1 {
		delete(idx.namespaceCounts, namespace)
		delete(idx.namespaces, namespace)
		return
	}
	idx.namespaceCounts[namespace] = count - 1
}

// RegisterTool registers a single tool with its backend.
func (idx *InMemoryIndex) RegisterTool(tool toolmodel.Tool, backend toolmodel.ToolBackend) error {
	// Validate tool
	if err := tool.Validate(); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidTool, err)
	}

	// Validate backend
	if err := validateBackend(backend); err != nil {
		return err
	}

	toolID := tool.ToolID()
	backendKey := backendIdentity(backend)
	normalizedTags := toolmodel.NormalizeTags(tool.Tags)

	idx.mu.Lock()

	record, exists := idx.tools[toolID]
	changeType := ChangeRegistered
	if !exists {
		record = &toolRecord{
			tool:           tool,
			backends:       []toolmodel.ToolBackend{backend},
			backendKeys:    map[string]int{backendKey: 0},
			normalizedTags: normalizedTags,
		}
		refreshRecordDerived(record)
		idx.tools[toolID] = record
		idx.addNamespaceLocked(tool.Namespace)
	} else {
		changeType = ChangeUpdated
		// Check MCP field consistency: new tool's MCP fields must match existing
		if !toolMCPFieldsEqual(record.tool, tool) {
			idx.mu.Unlock()
			return fmt.Errorf("%w: tool %q MCP fields differ from existing registration", ErrInvalidTool, toolID)
		}

		// Track namespace changes if tool is re-registered under a new namespace.
		if record.tool.Namespace != tool.Namespace {
			idx.removeNamespaceLocked(record.tool.Namespace)
			idx.addNamespaceLocked(tool.Namespace)
		}

		// Update toolmodel extensions (Tags) - these are allowed to differ
		record.tool = tool
		record.normalizedTags = normalizedTags
		refreshRecordDerived(record)

		// Check if backend already exists
		if existingIdx, ok := record.backendKeys[backendKey]; ok {
			// Replace existing backend
			record.backends[existingIdx] = backend
		} else {
			// Add new backend
			record.backendKeys[backendKey] = len(record.backends)
			record.backends = append(record.backends, backend)
		}
	}

	idx.markSearchDocsDirtyLocked()
	version := idx.indexVersion
	listeners := idx.snapshotListenersLocked()
	idx.mu.Unlock()

	notifyListeners(listeners, ChangeEvent{
		Type:    changeType,
		ToolID:  toolID,
		Backend: backend,
		Version: version,
	})
	return nil
}

// RegisterTools registers multiple tools in batch.
func (idx *InMemoryIndex) RegisterTools(regs []ToolRegistration) error {
	for _, reg := range regs {
		if err := idx.RegisterTool(reg.Tool, reg.Backend); err != nil {
			return err
		}
	}
	return nil
}

// RegisterToolsFromMCP is a convenience method for registering tools from an MCP server.
func (idx *InMemoryIndex) RegisterToolsFromMCP(serverName string, tools []toolmodel.Tool) error {
	backend := toolmodel.ToolBackend{
		Kind: toolmodel.BackendKindMCP,
		MCP:  &toolmodel.MCPBackend{ServerName: serverName},
	}

	for _, tool := range tools {
		if err := idx.RegisterTool(tool, backend); err != nil {
			return err
		}
	}
	return nil
}

// UnregisterBackend removes a specific backend from a tool.
// If the last backend is removed, the tool is also removed.
//
// For provider backends, backendID must be in the format "providerID:toolID".
// For MCP backends, backendID is the server name.
// For local backends, backendID is the handler name.
func (idx *InMemoryIndex) UnregisterBackend(toolID string, kind toolmodel.BackendKind, backendID string) error {
	var providerID string
	var providerToolID string

	// Validate backendID format for provider backends
	if kind == toolmodel.BackendKindProvider {
		if !strings.Contains(backendID, ":") {
			return fmt.Errorf("%w: provider backendID must be in format 'providerID:toolID'", ErrInvalidBackend)
		}
		parts := strings.SplitN(backendID, ":", 2)
		if parts[0] == "" || parts[1] == "" {
			return fmt.Errorf("%w: provider backendID must have non-empty providerID and toolID", ErrInvalidBackend)
		}
		providerID = parts[0]
		providerToolID = parts[1]
	}

	idx.mu.Lock()

	record, exists := idx.tools[toolID]
	if !exists {
		idx.mu.Unlock()
		return fmt.Errorf("%w: %s", ErrNotFound, toolID)
	}

	// Build the backend key to find
	var searchKey string
	switch kind {
	case toolmodel.BackendKindMCP:
		searchKey = encodeIdentity(string(kind), backendID)
	case toolmodel.BackendKindProvider:
		searchKey = encodeIdentity(string(kind), providerID, providerToolID)
	case toolmodel.BackendKindLocal:
		searchKey = encodeIdentity(string(kind), backendID)
	}

	// Find and remove the backend
	foundIdx := -1
	for key, idx := range record.backendKeys {
		if key == searchKey {
			foundIdx = idx
			delete(record.backendKeys, key)
			break
		}
	}

	if foundIdx == -1 {
		idx.mu.Unlock()
		return fmt.Errorf("%w: backend not found", ErrNotFound)
	}

	removedBackend := record.backends[foundIdx]

	// Remove from slice
	record.backends = append(record.backends[:foundIdx], record.backends[foundIdx+1:]...)

	// Update indices in backendKeys for backends after the removed one
	for key, idx := range record.backendKeys {
		if idx > foundIdx {
			record.backendKeys[key] = idx - 1
		}
	}

	// If no backends left, remove the tool entirely
	changeType := ChangeBackendRemoved
	if len(record.backends) == 0 {
		namespace := record.tool.Namespace
		delete(idx.tools, toolID)
		idx.removeNamespaceLocked(namespace)
		changeType = ChangeToolRemoved
	}

	idx.markSearchDocsDirtyLocked()
	version := idx.indexVersion
	listeners := idx.snapshotListenersLocked()
	idx.mu.Unlock()

	notifyListeners(listeners, ChangeEvent{
		Type:    changeType,
		ToolID:  toolID,
		Backend: removedBackend,
		Version: version,
	})
	return nil
}

// GetTool returns the full tool and its default backend.
func (idx *InMemoryIndex) GetTool(id string) (toolmodel.Tool, toolmodel.ToolBackend, error) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	record, exists := idx.tools[id]
	if !exists {
		return toolmodel.Tool{}, toolmodel.ToolBackend{}, fmt.Errorf("%w: %s", ErrNotFound, id)
	}

	defaultBackend := idx.backendSelector(record.backends)
	return record.tool, defaultBackend, nil
}

// GetAllBackends returns all backends for a tool.
func (idx *InMemoryIndex) GetAllBackends(id string) ([]toolmodel.ToolBackend, error) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	record, exists := idx.tools[id]
	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrNotFound, id)
	}

	// Return a copy to prevent external modification
	result := make([]toolmodel.ToolBackend, len(record.backends))
	copy(result, record.backends)
	return result, nil
}

// Search performs a search over the indexed tools.
func (idx *InMemoryIndex) Search(query string, limit int) ([]Summary, error) {
	docs, _ := idx.snapshotSearchDocs()
	return idx.searcher.Search(query, limit, docs)
}

// SearchPage performs a search over the indexed tools with cursor pagination.
func (idx *InMemoryIndex) SearchPage(query string, limit int, cursor string) ([]Summary, string, error) {
	if limit <= 0 {
		return nil, "", fmt.Errorf("limit must be positive")
	}

	docs, version := idx.snapshotSearchDocs()
	results, err := idx.searcher.Search(query, len(docs), docs)
	if err != nil {
		return nil, "", err
	}

	page, nextCursor, err := paginateResults(results, limit, cursor, version)
	if err != nil {
		return nil, "", err
	}
	return page, nextCursor, nil
}

// ensureSearchDocsLocked rebuilds the search docs cache if dirty.
// Must be called with idx.mu held.
func (idx *InMemoryIndex) ensureSearchDocsLocked() {
	if !idx.searchDocsDirty && idx.searchDocs != nil && idx.searchDocsVersion == idx.indexVersion {
		return
	}
	idx.rebuildSearchDocsLocked()
}

func (idx *InMemoryIndex) snapshotSearchDocs() ([]SearchDoc, uint64) {
	// Fast path: serve a cached snapshot under a read lock.
	idx.mu.RLock()
	if !idx.searchDocsDirty && idx.searchDocs != nil && idx.searchDocsVersion == idx.indexVersion {
		docs := make([]SearchDoc, len(idx.searchDocs))
		copy(docs, idx.searchDocs)
		version := idx.searchDocsVersion
		idx.mu.RUnlock()
		return docs, version
	}
	idx.mu.RUnlock()

	// Slow path: rebuild the cache under an exclusive lock.
	idx.mu.Lock()
	idx.ensureSearchDocsLocked()
	docs := make([]SearchDoc, len(idx.searchDocs))
	copy(docs, idx.searchDocs)
	version := idx.searchDocsVersion
	idx.mu.Unlock()

	return docs, version
}

// rebuildSearchDocsLocked rebuilds the search docs from scratch.
// Must be called with idx.mu held.
func (idx *InMemoryIndex) rebuildSearchDocsLocked() {
	docs := make([]SearchDoc, 0, len(idx.tools))
	for id, record := range idx.tools {
		docs = append(docs, SearchDoc{
			ID:      id,
			DocText: record.docText,
			Summary: record.summary,
		})
	}
	// Sort by ID for deterministic order
	sort.Slice(docs, func(i, j int) bool {
		return docs[i].ID < docs[j].ID
	})
	idx.searchDocs = docs
	idx.searchDocsDirty = false
	idx.searchDocsVersion = idx.indexVersion
	idx.searchDocsBuilds++
}

// markSearchDocsDirtyLocked marks the search docs cache as stale.
// Must be called with idx.mu held.
func (idx *InMemoryIndex) markSearchDocsDirtyLocked() {
	idx.searchDocsDirty = true
	idx.indexVersion++
}

func (idx *InMemoryIndex) snapshotListenersLocked() []ChangeListener {
	if len(idx.listeners) == 0 {
		return nil
	}
	out := make([]ChangeListener, len(idx.listeners))
	for i, entry := range idx.listeners {
		out[i] = entry.fn
	}
	return out
}

func notifyListeners(listeners []ChangeListener, event ChangeEvent) {
	for _, listener := range listeners {
		listener(event)
	}
}

// ListNamespaces returns all namespaces in alphabetical order.
func (idx *InMemoryIndex) ListNamespaces() ([]string, error) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	result := make([]string, 0, len(idx.namespaces))
	for ns := range idx.namespaces {
		result = append(result, ns)
	}
	sort.Strings(result)
	return result, nil
}

// ListNamespacesPage returns namespaces with cursor pagination.
func (idx *InMemoryIndex) ListNamespacesPage(limit int, cursor string) ([]string, string, error) {
	if limit <= 0 {
		return nil, "", fmt.Errorf("limit must be positive")
	}

	idx.mu.RLock()
	version := idx.indexVersion
	result := make([]string, 0, len(idx.namespaces))
	for ns := range idx.namespaces {
		result = append(result, ns)
	}
	idx.mu.RUnlock()

	sort.Strings(result)
	page, nextCursor, err := paginateResults(result, limit, cursor, version)
	if err != nil {
		return nil, "", err
	}
	return page, nextCursor, nil
}

// refreshRecordDerived recomputes cached derived fields for a tool record.
func refreshRecordDerived(record *toolRecord) {
	record.docText = buildDocText(record.tool, record.normalizedTags)
	record.summary = buildSummary(record.tool, record.normalizedTags)
}

// buildDocText creates the lowercased search text for a tool.
func buildDocText(tool toolmodel.Tool, normalizedTags []string) string {
	parts := []string{
		strings.ToLower(tool.Name),
		strings.ToLower(tool.Namespace),
		strings.ToLower(tool.Description),
	}
	parts = append(parts, normalizedTags...) // already normalized/lowercased
	return strings.Join(parts, " ")
}

// buildSummary creates a Summary from tool fields and normalized tags.
func buildSummary(tool toolmodel.Tool, normalizedTags []string) Summary {
	shortDesc := tool.Description
	if len(shortDesc) > MaxShortDescriptionLen {
		shortDesc = shortDesc[:MaxShortDescriptionLen]
	}

	return Summary{
		ID:               tool.ToolID(),
		Name:             tool.Name,
		Namespace:        tool.Namespace,
		ShortDescription: shortDesc,
		Tags:             normalizedTags,
	}
}

// lexicalSearcher is the default search implementation using simple lexical matching.
type lexicalSearcher struct{}

// scoredResult holds a result with its score for ranking.
type scoredResult struct {
	summary Summary
	score   int
}

func (s *lexicalSearcher) Search(query string, limit int, docs []SearchDoc) ([]Summary, error) {
	query = strings.ToLower(strings.TrimSpace(query))

	// Empty query returns all results (up to limit)
	if query == "" {
		results := make([]Summary, 0, limit)
		for i, doc := range docs {
			if i >= limit {
				break
			}
			results = append(results, doc.Summary)
		}
		return results, nil
	}

	// Score and collect matching results
	var scored []scoredResult
	for _, doc := range docs {
		score := 0

		// Name match (highest priority)
		nameLower := strings.ToLower(doc.Summary.Name)
		if strings.Contains(nameLower, query) {
			score += 100
			if nameLower == query {
				score += 50 // Exact match bonus
			}
		}

		// Namespace match
		nsLower := strings.ToLower(doc.Summary.Namespace)
		if strings.Contains(nsLower, query) {
			score += 50
		}

		// Description/tags match (via DocText)
		if score == 0 && strings.Contains(doc.DocText, query) {
			score += 10
		}

		if score > 0 {
			scored = append(scored, scoredResult{summary: doc.Summary, score: score})
		}
	}

	// Sort by score descending
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	// Apply limit
	if len(scored) > limit {
		scored = scored[:limit]
	}

	// Extract summaries
	results := make([]Summary, len(scored))
	for i, sr := range scored {
		results[i] = sr.summary
	}

	return results, nil
}
