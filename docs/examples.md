# Examples

## Custom backend selection

```go
selector := func(backends []toolmodel.ToolBackend) toolmodel.ToolBackend {
  // Prefer MCP for this app
  for _, b := range backends {
    if b.Kind == toolmodel.BackendKindMCP {
      return b
    }
  }
  return toolindex.DefaultBackendSelector(backends)
}

idx := toolindex.NewInMemoryIndex(toolindex.IndexOptions{BackendSelector: selector})
```

## Inject BM25 searcher

```go
searcher := toolsearch.NewBM25Searcher(toolsearch.BM25Config{K1: 1.4, B: 0.75})
idx := toolindex.NewInMemoryIndex(toolindex.IndexOptions{Searcher: searcher})
```

## Multiple backends

```go
_ = idx.RegisterTool(tool, toolmodel.ToolBackend{
  Kind: toolmodel.BackendKindMCP,
  MCP:  &toolmodel.MCPBackend{ServerName: "github"},
})

_ = idx.RegisterTool(tool, toolmodel.ToolBackend{
  Kind:  toolmodel.BackendKindLocal,
  Local: &toolmodel.LocalBackend{Name: "get_repo"},
})
```
