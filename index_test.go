package toolindex

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/jonwraymond/toolmodel"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Helper to create a valid tool for testing
func makeTestTool(name, namespace, description string, tags []string) toolmodel.Tool {
	return toolmodel.Tool{
		Tool: mcp.Tool{
			Name:        name,
			Description: description,
			InputSchema: map[string]any{
				"type":       "object",
				"properties": map[string]any{},
			},
		},
		Namespace: namespace,
		Tags:      tags,
	}
}

// Helper to create an MCP backend
func makeMCPBackend(serverName string) toolmodel.ToolBackend {
	return toolmodel.ToolBackend{
		Kind: toolmodel.BackendKindMCP,
		MCP:  &toolmodel.MCPBackend{ServerName: serverName},
	}
}

// Helper to create a provider backend
func makeProviderBackend(providerID, toolID string) toolmodel.ToolBackend {
	return toolmodel.ToolBackend{
		Kind:     toolmodel.BackendKindProvider,
		Provider: &toolmodel.ProviderBackend{ProviderID: providerID, ToolID: toolID},
	}
}

// Helper to create a local backend
func makeLocalBackend(name string) toolmodel.ToolBackend {
	return toolmodel.ToolBackend{
		Kind:  toolmodel.BackendKindLocal,
		Local: &toolmodel.LocalBackend{Name: name},
	}
}

func mustRegister(t *testing.T, idx *InMemoryIndex, tool toolmodel.Tool, backend toolmodel.ToolBackend) {
	t.Helper()
	if err := idx.RegisterTool(tool, backend); err != nil {
		t.Fatalf("RegisterTool failed: %v", err)
	}
}

// ============================================================
// Tests for Summary and Index interface definition
// ============================================================

func TestSummaryStruct(t *testing.T) {
	// Summary struct should have all required fields
	s := Summary{
		ID:               "ns:toolname",
		Name:             "toolname",
		Namespace:        "ns",
		ShortDescription: "A short description",
		Tags:             []string{"tag1", "tag2"},
	}

	if s.ID != "ns:toolname" {
		t.Errorf("expected ID 'ns:toolname', got %q", s.ID)
	}
	if s.Name != "toolname" {
		t.Errorf("expected Name 'toolname', got %q", s.Name)
	}
	if s.Namespace != "ns" {
		t.Errorf("expected Namespace 'ns', got %q", s.Namespace)
	}
	if s.ShortDescription != "A short description" {
		t.Errorf("expected ShortDescription 'A short description', got %q", s.ShortDescription)
	}
	if len(s.Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(s.Tags))
	}
}

// ============================================================
// Tests for Tool Registration
// ============================================================

func TestRegisterTool_Basic(t *testing.T) {
	idx := NewInMemoryIndex()

	tool := makeTestTool("mytool", "myns", "A test tool", nil)
	backend := makeMCPBackend("server1")

	err := idx.RegisterTool(tool, backend)
	if err != nil {
		t.Fatalf("RegisterTool failed: %v", err)
	}

	// Should be able to get the tool back
	gotTool, gotBackend, err := idx.GetTool("myns:mytool")
	if err != nil {
		t.Fatalf("GetTool failed: %v", err)
	}
	if gotTool.Name != "mytool" {
		t.Errorf("expected tool name 'mytool', got %q", gotTool.Name)
	}
	if gotBackend.Kind != toolmodel.BackendKindMCP {
		t.Errorf("expected backend kind MCP, got %v", gotBackend.Kind)
	}
}

func TestRegisterTool_NoNamespace(t *testing.T) {
	idx := NewInMemoryIndex()

	tool := makeTestTool("simpletool", "", "A simple tool", nil)
	backend := makeMCPBackend("server1")

	err := idx.RegisterTool(tool, backend)
	if err != nil {
		t.Fatalf("RegisterTool failed: %v", err)
	}

	// Tool ID should just be the name when no namespace
	gotTool, _, err := idx.GetTool("simpletool")
	if err != nil {
		t.Fatalf("GetTool failed: %v", err)
	}
	if gotTool.Name != "simpletool" {
		t.Errorf("expected tool name 'simpletool', got %q", gotTool.Name)
	}
}

func TestRegisterTool_InvalidTool(t *testing.T) {
	idx := NewInMemoryIndex()

	// Tool with no name should fail validation
	tool := toolmodel.Tool{
		Tool: mcp.Tool{
			Name:        "",
			Description: "description",
			InputSchema: map[string]any{
				"type":       "object",
				"properties": map[string]any{},
			},
		},
	}
	backend := makeMCPBackend("server1")

	err := idx.RegisterTool(tool, backend)
	if err == nil {
		t.Fatal("expected error for invalid tool, got nil")
	}
	if !errors.Is(err, ErrInvalidTool) {
		t.Errorf("expected ErrInvalidTool, got %v", err)
	}
}

func TestRegisterTool_InvalidBackend(t *testing.T) {
	idx := NewInMemoryIndex()

	tool := makeTestTool("mytool", "myns", "A test tool", nil)
	// Backend with no kind or details
	backend := toolmodel.ToolBackend{}

	err := idx.RegisterTool(tool, backend)
	if err == nil {
		t.Fatal("expected error for invalid backend, got nil")
	}
	if !errors.Is(err, ErrInvalidBackend) {
		t.Errorf("expected ErrInvalidBackend, got %v", err)
	}
}

func TestRegisterTool_ProviderBackendRequiresToolID(t *testing.T) {
	idx := NewInMemoryIndex()

	tool := makeTestTool("mytool", "ns", "A test tool", nil)
	// Provider backend without ToolID
	backend := toolmodel.ToolBackend{
		Kind:     toolmodel.BackendKindProvider,
		Provider: &toolmodel.ProviderBackend{ProviderID: "provider1", ToolID: ""},
	}

	err := idx.RegisterTool(tool, backend)
	if err == nil {
		t.Fatal("expected error for provider backend without ToolID, got nil")
	}
	if !errors.Is(err, ErrInvalidBackend) {
		t.Errorf("expected ErrInvalidBackend, got %v", err)
	}
}

func TestRegisterTool_ProviderBackendRequiresProviderID(t *testing.T) {
	idx := NewInMemoryIndex()

	tool := makeTestTool("mytool", "ns", "A test tool", nil)
	// Provider backend without ProviderID
	backend := toolmodel.ToolBackend{
		Kind:     toolmodel.BackendKindProvider,
		Provider: &toolmodel.ProviderBackend{ProviderID: "", ToolID: "tool-id"},
	}

	err := idx.RegisterTool(tool, backend)
	if err == nil {
		t.Fatal("expected error for provider backend without ProviderID, got nil")
	}
	if !errors.Is(err, ErrInvalidBackend) {
		t.Errorf("expected ErrInvalidBackend, got %v", err)
	}
}

func TestRegisterTools_Batch(t *testing.T) {
	idx := NewInMemoryIndex()

	regs := []ToolRegistration{
		{
			Tool:    makeTestTool("tool1", "ns", "Tool 1", nil),
			Backend: makeMCPBackend("server1"),
		},
		{
			Tool:    makeTestTool("tool2", "ns", "Tool 2", nil),
			Backend: makeMCPBackend("server1"),
		},
	}

	err := idx.RegisterTools(regs)
	if err != nil {
		t.Fatalf("RegisterTools failed: %v", err)
	}

	// Both tools should be retrievable
	_, _, err = idx.GetTool("ns:tool1")
	if err != nil {
		t.Errorf("GetTool(ns:tool1) failed: %v", err)
	}
	_, _, err = idx.GetTool("ns:tool2")
	if err != nil {
		t.Errorf("GetTool(ns:tool2) failed: %v", err)
	}
}

func TestRegisterToolsFromMCP(t *testing.T) {
	idx := NewInMemoryIndex()

	tools := []toolmodel.Tool{
		makeTestTool("tool1", "mcp", "Tool 1", nil),
		makeTestTool("tool2", "mcp", "Tool 2", nil),
	}

	err := idx.RegisterToolsFromMCP("my-mcp-server", tools)
	if err != nil {
		t.Fatalf("RegisterToolsFromMCP failed: %v", err)
	}

	// Tools should be retrievable with MCP backend
	tool, backend, err := idx.GetTool("mcp:tool1")
	if err != nil {
		t.Fatalf("GetTool failed: %v", err)
	}
	if backend.Kind != toolmodel.BackendKindMCP {
		t.Errorf("expected MCP backend, got %v", backend.Kind)
	}
	if backend.MCP.ServerName != "my-mcp-server" {
		t.Errorf("expected server name 'my-mcp-server', got %q", backend.MCP.ServerName)
	}
	if tool.Name != "tool1" {
		t.Errorf("expected tool name 'tool1', got %q", tool.Name)
	}
}

// ============================================================
// Tests for Backend Identity and Replacement
// ============================================================

func TestRegisterTool_ReplacesSameBackend(t *testing.T) {
	idx := NewInMemoryIndex()

	tool := makeTestTool("mytool", "ns", "A description", []string{"tag1"})
	backend := makeMCPBackend("server1")

	err := idx.RegisterTool(tool, backend)
	if err != nil {
		t.Fatalf("first RegisterTool failed: %v", err)
	}

	// Register same tool+backend with different tags (allowed - tags are toolmodel extension)
	tool2 := makeTestTool("mytool", "ns", "A description", []string{"tag2", "tag3"})
	err = idx.RegisterTool(tool2, backend)
	if err != nil {
		t.Fatalf("second RegisterTool failed: %v", err)
	}

	// Should only have one backend
	backends, err := idx.GetAllBackends("ns:mytool")
	if err != nil {
		t.Fatalf("GetAllBackends failed: %v", err)
	}
	if len(backends) != 1 {
		t.Errorf("expected 1 backend, got %d", len(backends))
	}

	// Tags should be updated
	results, _ := idx.Search("mytool", 10)
	if len(results) == 0 {
		t.Fatal("expected search result")
	}
	if len(results[0].Tags) != 2 {
		t.Errorf("expected 2 tags after update, got %d", len(results[0].Tags))
	}
}

func TestRegisterTool_BackendIdentityNoColonCollision(t *testing.T) {
	idx := NewInMemoryIndex()
	tool := makeTestTool("mytool", "ns", "desc", nil)

	backendA := makeProviderBackend("a:b", "c")
	backendB := makeProviderBackend("a", "b:c")

	if err := idx.RegisterTool(tool, backendA); err != nil {
		t.Fatalf("RegisterTool backendA failed: %v", err)
	}
	if err := idx.RegisterTool(tool, backendB); err != nil {
		t.Fatalf("RegisterTool backendB failed: %v", err)
	}

	backends, err := idx.GetAllBackends("ns:mytool")
	if err != nil {
		t.Fatalf("GetAllBackends failed: %v", err)
	}
	if len(backends) != 2 {
		t.Fatalf("expected 2 backends, got %d", len(backends))
	}
}

func TestInMemoryIndex_OnChange_EmitsEvents(t *testing.T) {
	idx := NewInMemoryIndex()
	var events []ChangeEvent
	idx.OnChange(func(ev ChangeEvent) {
		events = append(events, ev)
	})

	tool := makeTestTool("mytool", "ns", "desc", nil)
	backend := makeMCPBackend("server1")

	if err := idx.RegisterTool(tool, backend); err != nil {
		t.Fatalf("RegisterTool failed: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Type != ChangeRegistered {
		t.Fatalf("expected ChangeRegistered, got %v", events[0].Type)
	}
	if events[0].ToolID != "ns:mytool" {
		t.Fatalf("expected tool ID ns:mytool, got %q", events[0].ToolID)
	}
	if events[0].Version == 0 {
		t.Fatalf("expected version > 0")
	}

	if err := idx.UnregisterBackend(tool.ToolID(), backend.Kind, backend.MCP.ServerName); err != nil {
		t.Fatalf("UnregisterBackend failed: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}
	if events[1].Type != ChangeToolRemoved {
		t.Fatalf("expected ChangeToolRemoved, got %v", events[1].Type)
	}
}

func TestRegisterTool_MCPFieldMismatchRejected(t *testing.T) {
	idx := NewInMemoryIndex()

	tool := makeTestTool("mytool", "ns", "Original description", nil)
	backend := makeMCPBackend("server1")

	err := idx.RegisterTool(tool, backend)
	if err != nil {
		t.Fatalf("first RegisterTool failed: %v", err)
	}

	// Try to register same tool ID with different MCP fields (description changed)
	tool2 := makeTestTool("mytool", "ns", "Different description", nil)
	err = idx.RegisterTool(tool2, makeMCPBackend("server2"))
	if err == nil {
		t.Fatal("expected error for MCP field mismatch, got nil")
	}
	if !errors.Is(err, ErrInvalidTool) {
		t.Errorf("expected ErrInvalidTool, got %v", err)
	}
}

func TestRegisterTool_MCPFieldMismatchSchema(t *testing.T) {
	idx := NewInMemoryIndex()

	tool := makeTestTool("mytool", "ns", "desc", nil)
	err := idx.RegisterTool(tool, makeMCPBackend("server1"))
	if err != nil {
		t.Fatalf("first RegisterTool failed: %v", err)
	}

	// Try to register with different InputSchema
	tool2 := toolmodel.Tool{
		Tool: mcp.Tool{
			Name:        "mytool",
			Description: "desc",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"newField": map[string]any{"type": "string"},
				},
			},
		},
		Namespace: "ns",
	}
	err = idx.RegisterTool(tool2, makeMCPBackend("server2"))
	if err == nil {
		t.Fatal("expected error for schema mismatch, got nil")
	}
	if !errors.Is(err, ErrInvalidTool) {
		t.Errorf("expected ErrInvalidTool, got %v", err)
	}
}

func TestRegisterTool_MCPFieldMismatchTitle(t *testing.T) {
	idx := NewInMemoryIndex()

	tool := toolmodel.Tool{
		Tool: mcp.Tool{
			Name:        "mytool",
			Title:       "Original Title",
			Description: "desc",
			InputSchema: map[string]any{
				"type":       "object",
				"properties": map[string]any{},
			},
		},
		Namespace: "ns",
	}
	err := idx.RegisterTool(tool, makeMCPBackend("server1"))
	if err != nil {
		t.Fatalf("first RegisterTool failed: %v", err)
	}

	// Try to register with different Title
	tool2 := toolmodel.Tool{
		Tool: mcp.Tool{
			Name:        "mytool",
			Title:       "Different Title",
			Description: "desc",
			InputSchema: map[string]any{
				"type":       "object",
				"properties": map[string]any{},
			},
		},
		Namespace: "ns",
	}
	err = idx.RegisterTool(tool2, makeMCPBackend("server2"))
	if err == nil {
		t.Fatal("expected error for Title mismatch, got nil")
	}
	if !errors.Is(err, ErrInvalidTool) {
		t.Errorf("expected ErrInvalidTool, got %v", err)
	}
}

func TestRegisterTool_MCPFieldMismatchIcons(t *testing.T) {
	idx := NewInMemoryIndex()

	tool := toolmodel.Tool{
		Tool: mcp.Tool{
			Name:        "mytool",
			Description: "desc",
			InputSchema: map[string]any{
				"type":       "object",
				"properties": map[string]any{},
			},
			Icons: []mcp.Icon{
				{Source: "https://example.com/icon1.png"},
			},
		},
		Namespace: "ns",
	}
	err := idx.RegisterTool(tool, makeMCPBackend("server1"))
	if err != nil {
		t.Fatalf("first RegisterTool failed: %v", err)
	}

	// Try to register with different Icons
	tool2 := toolmodel.Tool{
		Tool: mcp.Tool{
			Name:        "mytool",
			Description: "desc",
			InputSchema: map[string]any{
				"type":       "object",
				"properties": map[string]any{},
			},
			Icons: []mcp.Icon{
				{Source: "https://example.com/different-icon.png"},
			},
		},
		Namespace: "ns",
	}
	err = idx.RegisterTool(tool2, makeMCPBackend("server2"))
	if err == nil {
		t.Fatal("expected error for Icons mismatch, got nil")
	}
	if !errors.Is(err, ErrInvalidTool) {
		t.Errorf("expected ErrInvalidTool, got %v", err)
	}
}

func TestRegisterTool_RawMessageSchemaEquality(t *testing.T) {
	idx := NewInMemoryIndex()

	// Register with map[string]any schema
	tool := toolmodel.Tool{
		Tool: mcp.Tool{
			Name:        "mytool",
			Description: "desc",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name": map[string]any{"type": "string"},
				},
			},
		},
		Namespace: "ns",
	}
	err := idx.RegisterTool(tool, makeMCPBackend("server1"))
	if err != nil {
		t.Fatalf("first RegisterTool failed: %v", err)
	}

	// Register same schema as json.RawMessage - should succeed
	tool2 := toolmodel.Tool{
		Tool: mcp.Tool{
			Name:        "mytool",
			Description: "desc",
			InputSchema: json.RawMessage(`{"type":"object","properties":{"name":{"type":"string"}}}`),
		},
		Namespace: "ns",
	}
	err = idx.RegisterTool(tool2, makeMCPBackend("server2"))
	if err != nil {
		t.Fatalf("RegisterTool with equivalent RawMessage schema should succeed: %v", err)
	}

	// Should have 2 backends now
	backends, err := idx.GetAllBackends("ns:mytool")
	if err != nil {
		t.Fatalf("GetAllBackends failed: %v", err)
	}
	if len(backends) != 2 {
		t.Errorf("expected 2 backends, got %d", len(backends))
	}
}

func TestRegisterTool_RawMessageSchemaMismatch(t *testing.T) {
	idx := NewInMemoryIndex()

	// Register with map[string]any schema
	tool := toolmodel.Tool{
		Tool: mcp.Tool{
			Name:        "mytool",
			Description: "desc",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name": map[string]any{"type": "string"},
				},
			},
		},
		Namespace: "ns",
	}
	err := idx.RegisterTool(tool, makeMCPBackend("server1"))
	if err != nil {
		t.Fatalf("first RegisterTool failed: %v", err)
	}

	// Register different schema as json.RawMessage - should fail
	tool2 := toolmodel.Tool{
		Tool: mcp.Tool{
			Name:        "mytool",
			Description: "desc",
			InputSchema: json.RawMessage(`{"type":"object","properties":{"age":{"type":"number"}}}`),
		},
		Namespace: "ns",
	}
	err = idx.RegisterTool(tool2, makeMCPBackend("server2"))
	if err == nil {
		t.Fatal("expected error for RawMessage schema mismatch, got nil")
	}
	if !errors.Is(err, ErrInvalidTool) {
		t.Errorf("expected ErrInvalidTool, got %v", err)
	}
}

func TestRegisterTool_ByteSliceSchemaEquality(t *testing.T) {
	idx := NewInMemoryIndex()

	// Register with []byte schema
	tool := toolmodel.Tool{
		Tool: mcp.Tool{
			Name:        "mytool",
			Description: "desc",
			InputSchema: []byte(`{"type":"object","properties":{"x":{"type":"integer"}}}`),
		},
		Namespace: "ns",
	}
	err := idx.RegisterTool(tool, makeMCPBackend("server1"))
	if err != nil {
		t.Fatalf("first RegisterTool failed: %v", err)
	}

	// Register same schema as map[string]any - should succeed
	tool2 := toolmodel.Tool{
		Tool: mcp.Tool{
			Name:        "mytool",
			Description: "desc",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"x": map[string]any{"type": "integer"},
				},
			},
		},
		Namespace: "ns",
	}
	err = idx.RegisterTool(tool2, makeMCPBackend("server2"))
	if err != nil {
		t.Fatalf("RegisterTool with equivalent schema should succeed: %v", err)
	}
}

func TestRegisterTool_MultipleBackends(t *testing.T) {
	idx := NewInMemoryIndex()

	tool := makeTestTool("mytool", "ns", "A tool", nil)

	// Register with MCP backend
	err := idx.RegisterTool(tool, makeMCPBackend("server1"))
	if err != nil {
		t.Fatalf("RegisterTool with MCP failed: %v", err)
	}

	// Register same tool with Provider backend
	err = idx.RegisterTool(tool, makeProviderBackend("provider1", "external-id"))
	if err != nil {
		t.Fatalf("RegisterTool with Provider failed: %v", err)
	}

	// Register same tool with Local backend
	err = idx.RegisterTool(tool, makeLocalBackend("local-handler"))
	if err != nil {
		t.Fatalf("RegisterTool with Local failed: %v", err)
	}

	backends, err := idx.GetAllBackends("ns:mytool")
	if err != nil {
		t.Fatalf("GetAllBackends failed: %v", err)
	}
	if len(backends) != 3 {
		t.Errorf("expected 3 backends, got %d", len(backends))
	}
}

func TestUnregisterBackend(t *testing.T) {
	idx := NewInMemoryIndex()

	tool := makeTestTool("mytool", "ns", "A tool", nil)

	// Register with multiple backends
	mustRegister(t, idx, tool, makeMCPBackend("server1"))
	mustRegister(t, idx, tool, makeLocalBackend("local-handler"))

	// Unregister the MCP backend
	err := idx.UnregisterBackend("ns:mytool", toolmodel.BackendKindMCP, "server1")
	if err != nil {
		t.Fatalf("UnregisterBackend failed: %v", err)
	}

	backends, err := idx.GetAllBackends("ns:mytool")
	if err != nil {
		t.Fatalf("GetAllBackends failed: %v", err)
	}
	if len(backends) != 1 {
		t.Errorf("expected 1 backend after unregister, got %d", len(backends))
	}
	if backends[0].Kind != toolmodel.BackendKindLocal {
		t.Errorf("expected local backend to remain, got %v", backends[0].Kind)
	}
}

func TestUnregisterBackend_LastBackendRemovesTool(t *testing.T) {
	idx := NewInMemoryIndex()

	tool := makeTestTool("mytool", "ns", "A tool", nil)
	mustRegister(t, idx, tool, makeMCPBackend("server1"))

	// Unregister the only backend
	err := idx.UnregisterBackend("ns:mytool", toolmodel.BackendKindMCP, "server1")
	if err != nil {
		t.Fatalf("UnregisterBackend failed: %v", err)
	}

	// Tool should no longer exist
	_, _, err = idx.GetTool("ns:mytool")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound after removing last backend, got %v", err)
	}

	namespaces, err := idx.ListNamespaces()
	if err != nil {
		t.Fatalf("ListNamespaces failed: %v", err)
	}
	if len(namespaces) != 0 {
		t.Errorf("expected 0 namespaces after removing last tool, got %d", len(namespaces))
	}
}

func TestUnregisterBackend_NotFound(t *testing.T) {
	idx := NewInMemoryIndex()

	err := idx.UnregisterBackend("nonexistent", toolmodel.BackendKindMCP, "server1")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestUnregisterBackend_ProviderRequiresFormat(t *testing.T) {
	idx := NewInMemoryIndex()

	tool := makeTestTool("mytool", "ns", "A tool", nil)
	mustRegister(t, idx, tool, makeProviderBackend("provider1", "tool-a"))

	// Unregister with proper format "providerID:toolID"
	err := idx.UnregisterBackend("ns:mytool", toolmodel.BackendKindProvider, "provider1:tool-a")
	if err != nil {
		t.Fatalf("UnregisterBackend with proper format failed: %v", err)
	}

	// Tool should be removed
	_, _, err = idx.GetTool("ns:mytool")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected tool to be removed, got %v", err)
	}
}

func TestUnregisterBackend_ProviderInvalidFormat(t *testing.T) {
	idx := NewInMemoryIndex()

	tool := makeTestTool("mytool", "ns", "A tool", nil)
	mustRegister(t, idx, tool, makeProviderBackend("provider1", "tool-a"))

	// Try to unregister without colon - should fail
	err := idx.UnregisterBackend("ns:mytool", toolmodel.BackendKindProvider, "provider1")
	if err == nil {
		t.Fatal("expected error for invalid provider backendID format, got nil")
	}
	if !errors.Is(err, ErrInvalidBackend) {
		t.Errorf("expected ErrInvalidBackend, got %v", err)
	}
}

func TestUnregisterBackend_ProviderEmptyParts(t *testing.T) {
	idx := NewInMemoryIndex()

	tool := makeTestTool("mytool", "ns", "A tool", nil)
	mustRegister(t, idx, tool, makeProviderBackend("provider1", "tool-a"))

	// Try with empty providerID
	err := idx.UnregisterBackend("ns:mytool", toolmodel.BackendKindProvider, ":tool-a")
	if err == nil {
		t.Fatal("expected error for empty providerID, got nil")
	}
	if !errors.Is(err, ErrInvalidBackend) {
		t.Errorf("expected ErrInvalidBackend, got %v", err)
	}

	// Try with empty toolID
	err = idx.UnregisterBackend("ns:mytool", toolmodel.BackendKindProvider, "provider1:")
	if err == nil {
		t.Fatal("expected error for empty toolID, got nil")
	}
	if !errors.Is(err, ErrInvalidBackend) {
		t.Errorf("expected ErrInvalidBackend, got %v", err)
	}
}

// ============================================================
// Tests for Backend Selection Policy
// ============================================================

func TestDefaultBackendSelection_LocalFirst(t *testing.T) {
	idx := NewInMemoryIndex()

	tool := makeTestTool("mytool", "ns", "A tool", nil)

	// Register in reverse priority order
	mustRegister(t, idx, tool, makeMCPBackend("server1"))
	mustRegister(t, idx, tool, makeProviderBackend("provider1", "id1"))
	mustRegister(t, idx, tool, makeLocalBackend("local1"))

	_, backend, err := idx.GetTool("ns:mytool")
	if err != nil {
		t.Fatalf("GetTool failed: %v", err)
	}

	// Default policy: local > provider > mcp
	if backend.Kind != toolmodel.BackendKindLocal {
		t.Errorf("expected Local backend (highest priority), got %v", backend.Kind)
	}
}

func TestDefaultBackendSelection_ProviderOverMCP(t *testing.T) {
	idx := NewInMemoryIndex()

	tool := makeTestTool("mytool", "ns", "A tool", nil)

	mustRegister(t, idx, tool, makeMCPBackend("server1"))
	mustRegister(t, idx, tool, makeProviderBackend("provider1", "id1"))

	_, backend, err := idx.GetTool("ns:mytool")
	if err != nil {
		t.Fatalf("GetTool failed: %v", err)
	}

	if backend.Kind != toolmodel.BackendKindProvider {
		t.Errorf("expected Provider backend over MCP, got %v", backend.Kind)
	}
}

func TestCustomBackendSelector(t *testing.T) {
	// Custom selector that always prefers MCP
	customSelector := func(backends []toolmodel.ToolBackend) toolmodel.ToolBackend {
		for _, b := range backends {
			if b.Kind == toolmodel.BackendKindMCP {
				return b
			}
		}
		return backends[0]
	}

	idx := NewInMemoryIndex(IndexOptions{BackendSelector: customSelector})

	tool := makeTestTool("mytool", "ns", "A tool", nil)
	mustRegister(t, idx, tool, makeLocalBackend("local1"))
	mustRegister(t, idx, tool, makeMCPBackend("server1"))

	_, backend, err := idx.GetTool("ns:mytool")
	if err != nil {
		t.Fatalf("GetTool failed: %v", err)
	}

	if backend.Kind != toolmodel.BackendKindMCP {
		t.Errorf("expected MCP backend from custom selector, got %v", backend.Kind)
	}
}

// ============================================================
// Tests for Tool Lookup
// ============================================================

func TestGetTool_NotFound(t *testing.T) {
	idx := NewInMemoryIndex()

	_, _, err := idx.GetTool("nonexistent")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestGetAllBackends_NotFound(t *testing.T) {
	idx := NewInMemoryIndex()

	_, err := idx.GetAllBackends("nonexistent")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

// ============================================================
// Tests for Namespaces
// ============================================================

func TestListNamespaces_Empty(t *testing.T) {
	idx := NewInMemoryIndex()

	namespaces, err := idx.ListNamespaces()
	if err != nil {
		t.Fatalf("ListNamespaces failed: %v", err)
	}
	if len(namespaces) != 0 {
		t.Errorf("expected 0 namespaces, got %d", len(namespaces))
	}
}

func TestListNamespaces_MultipleNamespaces(t *testing.T) {
	idx := NewInMemoryIndex()

	mustRegister(t, idx, makeTestTool("tool1", "beta", "desc", nil), makeMCPBackend("s"))
	mustRegister(t, idx, makeTestTool("tool2", "alpha", "desc", nil), makeMCPBackend("s"))
	mustRegister(t, idx, makeTestTool("tool3", "gamma", "desc", nil), makeMCPBackend("s"))

	namespaces, err := idx.ListNamespaces()
	if err != nil {
		t.Fatalf("ListNamespaces failed: %v", err)
	}

	// Should be alphabetically sorted
	expected := []string{"alpha", "beta", "gamma"}
	if len(namespaces) != len(expected) {
		t.Fatalf("expected %d namespaces, got %d", len(expected), len(namespaces))
	}
	for i, ns := range expected {
		if namespaces[i] != ns {
			t.Errorf("expected namespace[%d]=%q, got %q", i, ns, namespaces[i])
		}
	}
}

func TestListNamespaces_IncludesEmptyNamespace(t *testing.T) {
	idx := NewInMemoryIndex()

	mustRegister(t, idx, makeTestTool("tool1", "", "desc", nil), makeMCPBackend("s"))
	mustRegister(t, idx, makeTestTool("tool2", "ns", "desc", nil), makeMCPBackend("s"))

	namespaces, err := idx.ListNamespaces()
	if err != nil {
		t.Fatalf("ListNamespaces failed: %v", err)
	}

	// Empty namespace should be included (first alphabetically)
	if len(namespaces) != 2 {
		t.Fatalf("expected 2 namespaces, got %d", len(namespaces))
	}
	if namespaces[0] != "" {
		t.Errorf("expected empty namespace first, got %q", namespaces[0])
	}
	if namespaces[1] != "ns" {
		t.Errorf("expected 'ns' second, got %q", namespaces[1])
	}
}

func TestListNamespaces_Deterministic(t *testing.T) {
	idx := NewInMemoryIndex()

	mustRegister(t, idx, makeTestTool("tool1", "zebra", "desc", nil), makeMCPBackend("s"))
	mustRegister(t, idx, makeTestTool("tool2", "apple", "desc", nil), makeMCPBackend("s"))
	mustRegister(t, idx, makeTestTool("tool3", "middle", "desc", nil), makeMCPBackend("s"))

	// Call multiple times, should always be same order
	for i := 0; i < 3; i++ {
		namespaces, err := idx.ListNamespaces()
		if err != nil {
			t.Fatalf("ListNamespaces failed: %v", err)
		}
		expected := []string{"apple", "middle", "zebra"}
		for j, ns := range expected {
			if namespaces[j] != ns {
				t.Errorf("iteration %d: expected namespace[%d]=%q, got %q", i, j, ns, namespaces[j])
			}
		}
	}
}

// ============================================================
// Tests for Search
// ============================================================

func TestSearch_ByName(t *testing.T) {
	idx := NewInMemoryIndex()

	mustRegister(t, idx, makeTestTool("calculator", "math", "A calculator tool", nil), makeMCPBackend("s"))
	mustRegister(t, idx, makeTestTool("weather", "api", "Weather information", nil), makeMCPBackend("s"))

	results, err := idx.Search("calculator", 10)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) == 0 {
		t.Fatal("expected at least one result")
	}
	if results[0].Name != "calculator" {
		t.Errorf("expected first result to be calculator, got %q", results[0].Name)
	}
}

func TestSearch_ByNamespace(t *testing.T) {
	idx := NewInMemoryIndex()

	mustRegister(t, idx, makeTestTool("add", "math", "Addition", nil), makeMCPBackend("s"))
	mustRegister(t, idx, makeTestTool("sub", "math", "Subtraction", nil), makeMCPBackend("s"))
	mustRegister(t, idx, makeTestTool("weather", "api", "Weather", nil), makeMCPBackend("s"))

	results, err := idx.Search("math", 10)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	// Should find both math tools
	if len(results) < 2 {
		t.Fatalf("expected at least 2 results, got %d", len(results))
	}
}

func TestSearch_ByDescription(t *testing.T) {
	idx := NewInMemoryIndex()

	mustRegister(t, idx, makeTestTool("tool1", "ns", "Performs authentication", nil), makeMCPBackend("s"))
	mustRegister(t, idx, makeTestTool("tool2", "ns", "Sends emails", nil), makeMCPBackend("s"))

	results, err := idx.Search("authentication", 10)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) == 0 {
		t.Fatal("expected at least one result")
	}
	if results[0].Name != "tool1" {
		t.Errorf("expected first result to be tool1, got %q", results[0].Name)
	}
}

func TestSearch_ByTags(t *testing.T) {
	idx := NewInMemoryIndex()

	mustRegister(t, idx, makeTestTool("tool1", "ns", "desc", []string{"security", "auth"}), makeMCPBackend("s"))
	mustRegister(t, idx, makeTestTool("tool2", "ns", "desc", []string{"networking"}), makeMCPBackend("s"))

	results, err := idx.Search("security", 10)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) == 0 {
		t.Fatal("expected at least one result")
	}
	if results[0].Name != "tool1" {
		t.Errorf("expected first result to be tool1, got %q", results[0].Name)
	}
}

func TestSearch_CaseInsensitive(t *testing.T) {
	idx := NewInMemoryIndex()

	mustRegister(t, idx, makeTestTool("Calculator", "Math", "A CALCULATOR tool", nil), makeMCPBackend("s"))

	results, err := idx.Search("calculator", 10)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) == 0 {
		t.Fatal("expected at least one result for case-insensitive search")
	}
}

func TestSearch_Limit(t *testing.T) {
	idx := NewInMemoryIndex()

	// Register many tools with different names
	for i := 0; i < 20; i++ {
		name := "searchtool" + string(rune('a'+i))
		mustRegister(t, idx, makeTestTool(name, "ns", "description with searchtool", nil), makeMCPBackend("s"))
	}

	results, err := idx.Search("searchtool", 5)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) > 5 {
		t.Errorf("expected at most 5 results, got %d", len(results))
	}
}

func TestSearch_RankingNameHigherThanDescription(t *testing.T) {
	idx := NewInMemoryIndex()

	// Tool with 'calculator' in description
	mustRegister(t, idx, makeTestTool("mathhelper", "ns", "A calculator for math", nil), makeMCPBackend("s1"))
	// Tool with 'calculator' in name
	mustRegister(t, idx, makeTestTool("calculator", "ns", "Does math operations", nil), makeMCPBackend("s2"))

	results, err := idx.Search("calculator", 10)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) < 2 {
		t.Fatal("expected at least 2 results")
	}
	// Name match should rank higher
	if results[0].Name != "calculator" {
		t.Errorf("expected name match 'calculator' first, got %q", results[0].Name)
	}
}

func TestSearch_NoResults(t *testing.T) {
	idx := NewInMemoryIndex()

	mustRegister(t, idx, makeTestTool("tool1", "ns", "desc", nil), makeMCPBackend("s"))

	results, err := idx.Search("nonexistent", 10)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestSearch_EmptyQuery(t *testing.T) {
	idx := NewInMemoryIndex()

	mustRegister(t, idx, makeTestTool("tool1", "ns", "desc", nil), makeMCPBackend("s"))
	mustRegister(t, idx, makeTestTool("tool2", "ns", "desc", nil), makeMCPBackend("s"))

	results, err := idx.Search("", 10)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	// Empty query should return all tools (up to limit)
	if len(results) != 2 {
		t.Errorf("expected 2 results for empty query, got %d", len(results))
	}
}

// ============================================================
// Tests for Summary Results
// ============================================================

func TestSearch_ReturnsSummaryOnly(t *testing.T) {
	idx := NewInMemoryIndex()

	tool := makeTestTool("mytool", "ns", "A test tool description", []string{"tag1", "tag2"})
	mustRegister(t, idx, tool, makeMCPBackend("s"))

	results, err := idx.Search("mytool", 10)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) == 0 {
		t.Fatal("expected at least one result")
	}

	summary := results[0]
	if summary.ID != "ns:mytool" {
		t.Errorf("expected ID 'ns:mytool', got %q", summary.ID)
	}
	if summary.Name != "mytool" {
		t.Errorf("expected Name 'mytool', got %q", summary.Name)
	}
	if summary.Namespace != "ns" {
		t.Errorf("expected Namespace 'ns', got %q", summary.Namespace)
	}
}

func TestSearch_TruncatesLongDescription(t *testing.T) {
	idx := NewInMemoryIndex()

	longDesc := strings.Repeat("x", 200) // 200 chars
	tool := makeTestTool("mytool", "ns", longDesc, nil)
	mustRegister(t, idx, tool, makeMCPBackend("s"))

	results, err := idx.Search("mytool", 10)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) == 0 {
		t.Fatal("expected at least one result")
	}

	if len(results[0].ShortDescription) > MaxShortDescriptionLen {
		t.Errorf("expected ShortDescription <= %d chars, got %d", MaxShortDescriptionLen, len(results[0].ShortDescription))
	}
}

func TestSearch_NormalizedTagsInSummary(t *testing.T) {
	idx := NewInMemoryIndex()

	// Register with unnormalized tags
	tool := makeTestTool("mytool", "ns", "desc", []string{"  UPPER CASE  ", "with spaces"})
	mustRegister(t, idx, tool, makeMCPBackend("s"))

	results, err := idx.Search("mytool", 10)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) == 0 {
		t.Fatal("expected at least one result")
	}

	// Tags in summary should be normalized
	tags := results[0].Tags
	if len(tags) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(tags))
	}

	// Check normalization was applied
	for _, tag := range tags {
		if tag != strings.ToLower(tag) {
			t.Errorf("expected lowercase tag, got %q", tag)
		}
		if strings.Contains(tag, " ") {
			t.Errorf("expected no spaces in tag, got %q", tag)
		}
	}
}

// ============================================================
// Tests for Tag Normalization on Ingest
// ============================================================

func TestTagNormalization_OnIngest(t *testing.T) {
	idx := NewInMemoryIndex()

	tool := makeTestTool("mytool", "ns", "desc", []string{"  TAG ONE  ", "TAG-TWO"})
	mustRegister(t, idx, tool, makeMCPBackend("s"))

	// Search should work with normalized tags
	results, err := idx.Search("tag-one", 10)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) == 0 {
		t.Fatal("expected search to find tool by normalized tag")
	}
}

// ============================================================
// Tests for Custom Searcher
// ============================================================

func TestCustomSearcher(t *testing.T) {
	// Custom searcher that always returns empty results
	customSearcher := &mockSearcher{
		searchFunc: func(_ string, _ int, _ []SearchDoc) ([]Summary, error) {
			return []Summary{}, nil
		},
	}

	idx := NewInMemoryIndex(IndexOptions{Searcher: customSearcher})

	mustRegister(t, idx, makeTestTool("mytool", "ns", "desc", nil), makeMCPBackend("s"))

	results, err := idx.Search("mytool", 10)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	// Custom searcher returns empty
	if len(results) != 0 {
		t.Errorf("expected 0 results from custom searcher, got %d", len(results))
	}
}

type mockSearcher struct {
	searchFunc func(query string, limit int, docs []SearchDoc) ([]Summary, error)
}

func (m *mockSearcher) Search(query string, limit int, docs []SearchDoc) ([]Summary, error) {
	return m.searchFunc(query, limit, docs)
}

// ============================================================
// Tests for Thread Safety
// ============================================================

func TestConcurrentAccess(t *testing.T) {
	idx := NewInMemoryIndex()

	var wg sync.WaitGroup
	errCh := make(chan error, 400)

	// Writer goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			tool := makeTestTool("tool"+string(rune('a'+i%26)), "ns", "desc", nil)
			if err := idx.RegisterTool(tool, makeMCPBackend("s")); err != nil {
				errCh <- fmt.Errorf("register: %w", err)
			}
		}
	}()

	// Reader goroutines
	for r := 0; r < 3; r++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				if _, err := idx.Search("tool", 10); err != nil {
					errCh <- fmt.Errorf("search: %w", err)
				}
				if _, err := idx.ListNamespaces(); err != nil {
					errCh <- fmt.Errorf("list namespaces: %w", err)
				}
				if _, _, err := idx.GetTool("ns:toola"); err != nil && !errors.Is(err, ErrNotFound) {
					errCh <- fmt.Errorf("get tool: %w", err)
				}
			}
		}()
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Errorf("concurrent access error: %v", err)
	}
}

// ============================================================
// Tests for SearchDoc struct (exported for custom searchers)
// ============================================================

func TestSearchDoc_ExposedToSearcher(t *testing.T) {
	var receivedDocs []SearchDoc

	customSearcher := &mockSearcher{
		searchFunc: func(_ string, _ int, docs []SearchDoc) ([]Summary, error) {
			receivedDocs = docs
			return []Summary{}, nil
		},
	}

	idx := NewInMemoryIndex(IndexOptions{Searcher: customSearcher})
	mustRegister(t, idx, makeTestTool("mytool", "ns", "My description", []string{"tag1"}), makeMCPBackend("s"))

	if _, err := idx.Search("test", 10); err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(receivedDocs) != 1 {
		t.Fatalf("expected 1 doc passed to searcher, got %d", len(receivedDocs))
	}

	doc := receivedDocs[0]
	if doc.ID != "ns:mytool" {
		t.Errorf("expected ID 'ns:mytool', got %q", doc.ID)
	}
	if doc.DocText == "" {
		t.Error("expected DocText to be populated")
	}
	if doc.Summary.Name != "mytool" {
		t.Errorf("expected Summary.Name 'mytool', got %q", doc.Summary.Name)
	}
}

// ============================================================
// Tests for Error Values
// ============================================================

func TestErrorValues(t *testing.T) {
	// Verify error values are defined and distinct
	if ErrNotFound == nil {
		t.Error("ErrNotFound should be defined")
	}
	if ErrInvalidTool == nil {
		t.Error("ErrInvalidTool should be defined")
	}
	if ErrInvalidBackend == nil {
		t.Error("ErrInvalidBackend should be defined")
	}
	if ErrInvalidCursor == nil {
		t.Error("ErrInvalidCursor should be defined")
	}

	// Should be distinct
	if errors.Is(ErrNotFound, ErrInvalidTool) {
		t.Error("ErrNotFound and ErrInvalidTool should be distinct")
	}
	if errors.Is(ErrNotFound, ErrInvalidBackend) {
		t.Error("ErrNotFound and ErrInvalidBackend should be distinct")
	}
	if errors.Is(ErrInvalidTool, ErrInvalidBackend) {
		t.Error("ErrInvalidTool and ErrInvalidBackend should be distinct")
	}
	if errors.Is(ErrInvalidCursor, ErrInvalidTool) || errors.Is(ErrInvalidCursor, ErrInvalidBackend) || errors.Is(ErrInvalidCursor, ErrNotFound) {
		t.Error("ErrInvalidCursor should be distinct")
	}
}

// ============================================================
// Provider Backend Identity Tests
// ============================================================

func TestProviderBackendIdentity(t *testing.T) {
	idx := NewInMemoryIndex()

	tool := makeTestTool("mytool", "ns", "desc", nil)

	// Register same provider ID but different tool IDs - should be separate backends
	mustRegister(t, idx, tool, makeProviderBackend("provider1", "tool-a"))
	mustRegister(t, idx, tool, makeProviderBackend("provider1", "tool-b"))

	backends, err := idx.GetAllBackends("ns:mytool")
	if err != nil {
		t.Fatalf("GetAllBackends failed: %v", err)
	}

	if len(backends) != 2 {
		t.Errorf("expected 2 backends (different tool IDs), got %d", len(backends))
	}
}

func TestProviderBackendReplacementByIdentity(t *testing.T) {
	idx := NewInMemoryIndex()

	tool := makeTestTool("mytool", "ns", "original", nil)
	mustRegister(t, idx, tool, makeProviderBackend("provider1", "tool-a"))

	// Re-register with same provider ID + tool ID - should replace
	// MCP fields must remain identical; toolmodel extensions (e.g., tags) may change.
	tool2 := makeTestTool("mytool", "ns", "original", []string{"tag"})
	mustRegister(t, idx, tool2, makeProviderBackend("provider1", "tool-a"))

	backends, err := idx.GetAllBackends("ns:mytool")
	if err != nil {
		t.Fatalf("GetAllBackends failed: %v", err)
	}

	if len(backends) != 1 {
		t.Errorf("expected 1 backend after replacement, got %d", len(backends))
	}
}

// ============================================================
// Tests for Search Doc Caching and Deterministic Order
// ============================================================

func TestSearchDocs_SortedByID(t *testing.T) {
	var receivedDocs []SearchDoc
	mockSearcher := &mockSearcher{
		searchFunc: func(_ string, _ int, docs []SearchDoc) ([]Summary, error) {
			receivedDocs = docs
			return nil, nil
		},
	}

	idx := NewInMemoryIndex(IndexOptions{Searcher: mockSearcher})

	// Register tools in non-alphabetical order
	mustRegister(t, idx, makeTestTool("zebra", "ns", "desc", nil), makeMCPBackend("s"))
	mustRegister(t, idx, makeTestTool("alpha", "ns", "desc", nil), makeMCPBackend("s"))
	mustRegister(t, idx, makeTestTool("middle", "ns", "desc", nil), makeMCPBackend("s"))

	if _, err := idx.Search("test", 10); err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	// Assert docs are sorted by ID ascending
	if len(receivedDocs) != 3 {
		t.Fatalf("expected 3 docs, got %d", len(receivedDocs))
	}
	expectedOrder := []string{"ns:alpha", "ns:middle", "ns:zebra"}
	for i, expected := range expectedOrder {
		if receivedDocs[i].ID != expected {
			t.Errorf("doc[%d]: expected ID %q, got %q", i, expected, receivedDocs[i].ID)
		}
	}
}

func TestSearchDocs_CachedBetweenSearches(t *testing.T) {
	idx := NewInMemoryIndex()
	mustRegister(t, idx, makeTestTool("tool", "ns", "desc", nil), makeMCPBackend("s"))

	if _, err := idx.Search("test", 10); err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if _, err := idx.Search("another", 10); err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	// Assert searchDocsBuilds == 1 (only built once)
	if idx.searchDocsBuilds != 1 {
		t.Errorf("expected 1 doc build, got %d", idx.searchDocsBuilds)
	}
}

func TestSearchDocs_RebuildsAfterMutation(t *testing.T) {
	idx := NewInMemoryIndex()
	mustRegister(t, idx, makeTestTool("tool1", "ns", "desc", nil), makeMCPBackend("s"))
	if _, err := idx.Search("test", 10); err != nil {
		t.Fatalf("Search failed: %v", err)
	} // builds=1

	mustRegister(t, idx, makeTestTool("tool2", "ns", "desc", nil), makeMCPBackend("s"))
	if _, err := idx.Search("test", 10); err != nil {
		t.Fatalf("Search failed: %v", err)
	} // builds=2

	if idx.searchDocsBuilds != 2 {
		t.Errorf("expected 2 doc builds after mutation, got %d", idx.searchDocsBuilds)
	}
}

func TestSearchDocs_ConcurrentDirtyCacheBuildsOnce(t *testing.T) {
	idx := NewInMemoryIndex()
	mustRegister(t, idx, makeTestTool("tool", "ns", "desc", nil), makeMCPBackend("s"))

	const workers = 32
	var wg sync.WaitGroup
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			if _, err := idx.Search("tool", 10); err != nil {
				t.Errorf("Search failed: %v", err)
			}
		}()
	}

	wg.Wait()

	if idx.searchDocsBuilds != 1 {
		t.Errorf("expected 1 doc build under concurrent dirty cache, got %d", idx.searchDocsBuilds)
	}
}

func TestSearchDocs_RebuildsAfterUnregister(t *testing.T) {
	idx := NewInMemoryIndex()
	mustRegister(t, idx, makeTestTool("tool1", "ns", "desc", nil), makeMCPBackend("s1"))
	mustRegister(t, idx, makeTestTool("tool2", "ns", "desc", nil), makeMCPBackend("s2"))
	if _, err := idx.Search("test", 10); err != nil {
		t.Fatalf("Search failed: %v", err)
	} // builds=1

	if err := idx.UnregisterBackend("ns:tool1", toolmodel.BackendKindMCP, "s1"); err != nil {
		t.Fatalf("UnregisterBackend failed: %v", err)
	}
	if _, err := idx.Search("test", 10); err != nil {
		t.Fatalf("Search failed: %v", err)
	} // builds=2

	if idx.searchDocsBuilds != 2 {
		t.Errorf("expected 2 doc builds after unregister, got %d", idx.searchDocsBuilds)
	}
}

func TestSearchDocs_DerivedFieldsRefreshOnUpdate(t *testing.T) {
	var receivedDocs []SearchDoc
	mockSearcher := &mockSearcher{
		searchFunc: func(_ string, _ int, docs []SearchDoc) ([]Summary, error) {
			receivedDocs = docs
			return nil, nil
		},
	}

	idx := NewInMemoryIndex(IndexOptions{Searcher: mockSearcher})

	// Register with description A
	mustRegister(t, idx, makeTestTool("tool", "ns", "description A", nil), makeMCPBackend("s"))
	if _, err := idx.Search("test", 10); err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	firstDocText := receivedDocs[0].DocText
	if !strings.Contains(firstDocText, "description a") {
		t.Errorf("expected docText to contain 'description a', got %q", firstDocText)
	}

	// Re-register same tool with different tags (MCP fields same, tags can change)
	// Note: Description is MCP field, so use tags which are toolmodel extension
	tool := makeTestTool("tool", "ns", "description A", []string{"newtag"})
	mustRegister(t, idx, tool, makeMCPBackend("s"))
	if _, err := idx.Search("test", 10); err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	// DocText should now include "newtag"
	if !strings.Contains(receivedDocs[0].DocText, "newtag") {
		t.Errorf("expected docText to contain 'newtag' after update, got %q", receivedDocs[0].DocText)
	}
}

// ============================================================
// Tests for Cursor Pagination
// ============================================================

func TestSearchPage_PaginatesWithCursor(t *testing.T) {
	idx := NewInMemoryIndex()

	mustRegister(t, idx, makeTestTool("alpha", "ns1", "alpha tool", nil), makeLocalBackend("alpha"))
	mustRegister(t, idx, makeTestTool("beta", "ns1", "beta tool", nil), makeLocalBackend("beta"))
	mustRegister(t, idx, makeTestTool("gamma", "ns2", "gamma tool", nil), makeLocalBackend("gamma"))

	results, cursor, err := idx.SearchPage("", 2, "")
	if err != nil {
		t.Fatalf("SearchPage failed: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if cursor == "" {
		t.Fatal("expected next cursor")
	}
	if results[0].ID != "ns1:alpha" || results[1].ID != "ns1:beta" {
		t.Fatalf("unexpected ordering: %q, %q", results[0].ID, results[1].ID)
	}

	nextResults, nextCursor, err := idx.SearchPage("", 2, cursor)
	if err != nil {
		t.Fatalf("SearchPage with cursor failed: %v", err)
	}
	if len(nextResults) != 1 {
		t.Fatalf("expected 1 result, got %d", len(nextResults))
	}
	if nextCursor != "" {
		t.Fatalf("expected empty cursor, got %q", nextCursor)
	}
	if nextResults[0].ID != "ns2:gamma" {
		t.Fatalf("expected ns2:gamma, got %q", nextResults[0].ID)
	}
}

func TestSearchPage_InvalidCursor(t *testing.T) {
	idx := NewInMemoryIndex()
	mustRegister(t, idx, makeTestTool("alpha", "ns1", "alpha tool", nil), makeLocalBackend("alpha"))

	_, _, err := idx.SearchPage("", 1, "not-base64")
	if !errors.Is(err, ErrInvalidCursor) {
		t.Fatalf("expected ErrInvalidCursor, got %v", err)
	}
}

func TestSearchPage_StaleCursor(t *testing.T) {
	idx := NewInMemoryIndex()
	mustRegister(t, idx, makeTestTool("alpha", "ns1", "alpha tool", nil), makeLocalBackend("alpha"))
	mustRegister(t, idx, makeTestTool("beta", "ns1", "beta tool", nil), makeLocalBackend("beta"))

	_, cursor, err := idx.SearchPage("", 1, "")
	if err != nil {
		t.Fatalf("SearchPage failed: %v", err)
	}

	mustRegister(t, idx, makeTestTool("gamma", "ns2", "gamma tool", nil), makeLocalBackend("gamma"))

	_, _, err = idx.SearchPage("", 1, cursor)
	if !errors.Is(err, ErrInvalidCursor) {
		t.Fatalf("expected ErrInvalidCursor, got %v", err)
	}
}

func TestListNamespacesPage_PaginatesWithCursor(t *testing.T) {
	idx := NewInMemoryIndex()
	mustRegister(t, idx, makeTestTool("alpha", "ns1", "alpha tool", nil), makeLocalBackend("alpha"))
	mustRegister(t, idx, makeTestTool("beta", "ns2", "beta tool", nil), makeLocalBackend("beta"))
	mustRegister(t, idx, makeTestTool("gamma", "ns3", "gamma tool", nil), makeLocalBackend("gamma"))

	namespaces, cursor, err := idx.ListNamespacesPage(2, "")
	if err != nil {
		t.Fatalf("ListNamespacesPage failed: %v", err)
	}
	if len(namespaces) != 2 {
		t.Fatalf("expected 2 namespaces, got %d", len(namespaces))
	}
	if cursor == "" {
		t.Fatal("expected next cursor")
	}
	if namespaces[0] != "ns1" || namespaces[1] != "ns2" {
		t.Fatalf("unexpected namespaces: %v", namespaces)
	}

	nextNamespaces, nextCursor, err := idx.ListNamespacesPage(2, cursor)
	if err != nil {
		t.Fatalf("ListNamespacesPage with cursor failed: %v", err)
	}
	if len(nextNamespaces) != 1 {
		t.Fatalf("expected 1 namespace, got %d", len(nextNamespaces))
	}
	if nextCursor != "" {
		t.Fatalf("expected empty cursor, got %q", nextCursor)
	}
	if nextNamespaces[0] != "ns3" {
		t.Fatalf("expected ns3, got %q", nextNamespaces[0])
	}
}
