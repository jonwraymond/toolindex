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
  ListNamespaces() ([]string, error)
}
```

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

type SearchDoc struct {
  ID      string
  DocText string
  Summary Summary
}
```

## Options

```go
type IndexOptions struct {
  BackendSelector BackendSelector
  Searcher        Searcher
}

type BackendSelector func([]toolmodel.ToolBackend) toolmodel.ToolBackend
```

## Errors

- `ErrNotFound`
- `ErrInvalidTool`
- `ErrInvalidBackend`
