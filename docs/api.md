# API Reference

## Index interface

```go
type Index interface {
  RegisterTool(tool toolmodel.Tool, backend toolmodel.ToolBackend) error
  RegisterTools(regs []ToolRegistration) error
  RegisterToolsFromMCP(serverName string, tools []toolmodel.Tool) error

  UnregisterBackend(toolID string, kind toolmodel.BackendKind, backendID string) error

  GetTool(id string) (toolmodel.Tool, toolmodel.ToolBackend, error)
  GetAllBackends(id string) ([]toolmodel.ToolBackend, error)

  Search(query string, limit int) ([]Summary, error)
  SearchPage(query string, limit int, cursor string) ([]Summary, string, error)
  ListNamespaces() ([]string, error)
  ListNamespacesPage(limit int, cursor string) ([]string, string, error)
}
```

### Index contract

- Concurrency: implementations are safe for concurrent use.
- Errors: use `errors.Is` with `ErrInvalidTool`, `ErrInvalidBackend`, `ErrNotFound`,
  `ErrInvalidCursor`, and `ErrNonDeterministicSearcher`.
- Ownership: returned slices are caller-owned; elements are read-only snapshots.
- Determinism: search and namespace listings must return stable ordering.
- Nil/zero: `SearchPage` requires `limit > 0`; empty inputs are treated as no-ops.

## Change notifications (optional)

```go
type ChangeType string

const (
  ChangeRegistered     ChangeType = "registered"
  ChangeUpdated        ChangeType = "updated"
  ChangeBackendRemoved ChangeType = "backend_removed"
  ChangeToolRemoved    ChangeType = "tool_removed"
  ChangeRefreshed      ChangeType = "refreshed"
)

type ChangeEvent struct {
  Type    ChangeType
  ToolID  string
  Backend toolmodel.ToolBackend
  Version uint64
}

type ChangeListener func(ChangeEvent)

type ChangeNotifier interface {
  OnChange(listener ChangeListener) (unsubscribe func())
}

type Refresher interface {
  Refresh() uint64
}
```

### ChangeNotifier/Refresher contract

- `OnChange` returns a non-nil unsubscribe func; it is safe to call multiple times.
- `Refresh` returns a monotonic version and is safe for concurrent use.

## Summary

```go
type Summary struct {
  ID               string
  Name             string
  Namespace        string
  ShortDescription string
  Tags             []string
}
```

## Registration

```go
type ToolRegistration struct {
  Tool    toolmodel.Tool
  Backend toolmodel.ToolBackend
}
```

## Searcher

```go
type Searcher interface {
  Search(query string, limit int, docs []SearchDoc) ([]Summary, error)
}

// Optional deterministic marker for cursor pagination
type DeterministicSearcher interface {
  Searcher
  Deterministic() bool
}

type SearchDoc struct {
  ID      string
  DocText string
  Summary Summary
}
```

### Searcher contract

- Concurrency: implementations should be safe for concurrent use or documented otherwise.
- Determinism: stable ordering required for cursor pagination; use deterministic tie-breaks.
- Nil/zero: `limit <= 0` should return an empty result set.

## Options

```go
type IndexOptions struct {
  BackendSelector BackendSelector
  Searcher        Searcher
  RequireDeterministicSearcher *bool
}

type BackendSelector func([]toolmodel.ToolBackend) toolmodel.ToolBackend
```

## Errors

- `ErrNotFound`
- `ErrInvalidTool`
- `ErrInvalidBackend`
- `ErrInvalidCursor`
