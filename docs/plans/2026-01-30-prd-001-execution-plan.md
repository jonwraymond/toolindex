# PRD-001 Execution Plan — toolindex (TDD)

**Status:** Ready
**Date:** 2026-01-30
**PRD:** `2026-01-30-prd-001-interface-contracts.md`


## TDD Workflow (required)
1. Red — write failing contract tests
2. Red verification — run tests
3. Green — minimal code/doc changes
4. Green verification — run tests
5. Commit — one commit per task


## Tasks
### Task 0 — Inventory + contract outline
- Confirm interface list and method signatures.
- Draft explicit contract bullets for each interface.
- Update docs/plans/README.md with this PRD + plan.
### Task 1 — Contract tests (Red/Green)
- Add `*_contract_test.go` with tests for each interface listed below.
- Use stub implementations where needed.
### Task 2 — GoDoc contracts
- Add/expand GoDoc on each interface with explicit contract clauses (thread-safety, errors, context, ownership).
- Update README/design-notes if user-facing.
### Task 3 — Verification
- Run `go test ./...`
- Run linters if configured (golangci-lint / gosec).


## Test Skeletons (contract_test.go)
### Index
```go
func TestIndex_Contract(t *testing.T) {
    // Methods:
    // - RegisterTool(tool toolmodel.Tool, backend toolmodel.ToolBackend) error
    // - RegisterTools(regs []ToolRegistration) error
    // - RegisterToolsFromMCP(serverName string, tools []toolmodel.Tool) error
    // - UnregisterBackend(toolID string, kind toolmodel.BackendKind, backendID string) error
    // - GetTool(id string) (toolmodel.Tool, toolmodel.ToolBackend, error)
    // - GetAllBackends(id string) ([]toolmodel.ToolBackend, error)
    // - Search(query string, limit int) ([]Summary, error)
    // - SearchPage(query string, limit int, cursor string) ([]Summary, string, error)
    // - ListNamespaces() ([]string, error)
    // - ListNamespacesPage(limit int, cursor string) ([]string, string, error)
    // Contract assertions:
    // - Concurrency guarantees documented and enforced
    // - Error semantics (types/wrapping) validated
    // - Context cancellation respected (if applicable)
    // - Deterministic ordering where required
    // - Nil/zero input handling specified
}
```
### Searcher
```go
func TestSearcher_Contract(t *testing.T) {
    // Methods:
    // - Search(query string, limit int, docs []SearchDoc) ([]Summary, error)
    // Contract assertions:
    // - Concurrency guarantees documented and enforced
    // - Error semantics (types/wrapping) validated
    // - Context cancellation respected (if applicable)
    // - Deterministic ordering where required
    // - Nil/zero input handling specified
}
```
### DeterministicSearcher
```go
func TestDeterministicSearcher_Contract(t *testing.T) {
    // Methods:
    // - Searcher
    // - Deterministic() bool
    // Contract assertions:
    // - Concurrency guarantees documented and enforced
    // - Error semantics (types/wrapping) validated
    // - Context cancellation respected (if applicable)
    // - Deterministic ordering where required
    // - Nil/zero input handling specified
}
```
### ChangeNotifier
```go
func TestChangeNotifier_Contract(t *testing.T) {
    // Methods:
    // - OnChange(listener ChangeListener) (unsubscribe func())
    // Contract assertions:
    // - Concurrency guarantees documented and enforced
    // - Error semantics (types/wrapping) validated
    // - Context cancellation respected (if applicable)
    // - Deterministic ordering where required
    // - Nil/zero input handling specified
}
```
### Refresher
```go
func TestRefresher_Contract(t *testing.T) {
    // Methods:
    // - Refresh() uint64
    // Contract assertions:
    // - Concurrency guarantees documented and enforced
    // - Error semantics (types/wrapping) validated
    // - Context cancellation respected (if applicable)
    // - Deterministic ordering where required
    // - Nil/zero input handling specified
}
```
