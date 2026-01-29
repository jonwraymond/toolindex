# Usage

## Register tools

```go
idx := toolindex.NewInMemoryIndex()

reg := toolindex.ToolRegistration{Tool: tool, Backend: backend}
if err := idx.RegisterTools([]toolindex.ToolRegistration{reg}); err != nil {
  // handle error
}
```

## Search

```go
summaries, err := idx.Search("repo metadata", 5)
if err != nil {
  // handle error
}
```

## Search with cursor pagination

```go
summaries, nextCursor, err := idx.SearchPage("repo metadata", 20, "")
if err != nil {
  // handle error
}
if nextCursor != "" {
  nextPage, _, _ := idx.SearchPage("repo metadata", 20, nextCursor)
  _ = nextPage
}
```

## List namespaces with cursor pagination

```go
namespaces, nextCursor, err := idx.ListNamespacesPage(50, "")
if err != nil {
  // handle error
}
if nextCursor != "" {
  nextPage, _, _ := idx.ListNamespacesPage(50, nextCursor)
  _ = nextPage
}
```

## Lookup and backends

```go
t, defaultBackend, err := idx.GetTool("github:get_repo")
if err != nil {
  // handle not found
}

allBackends, _ := idx.GetAllBackends(t.ToolID())
```

## Register from MCP

```go
_ = idx.RegisterToolsFromMCP("github", []toolmodel.Tool{toolA, toolB})
```

## Options

```go
idx := toolindex.NewInMemoryIndex(toolindex.IndexOptions{
  BackendSelector: mySelector,
  Searcher:        mySearcher,
  // RequireDeterministicSearcher: boolPtr(true), // enforce stable pagination
})
```
