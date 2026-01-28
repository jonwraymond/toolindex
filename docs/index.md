# toolindex

`toolindex` is the global registry and progressive discovery layer for tools.
It stores `toolmodel.Tool` + `toolmodel.ToolBackend`, provides search, and
returns token-cheap summaries.

## Key APIs

- `Index` interface
- `InMemoryIndex` implementation
- `Search` + `ListNamespaces`
- `RegisterTool(s)` + `GetTool`
- `Searcher` interface for pluggable ranking

## Quickstart

```go
idx := toolindex.NewInMemoryIndex()

_ = idx.RegisterTool(tool, backend)

summaries, _ := idx.Search("repo", 5)
for _, s := range summaries {
  fmt.Println(s.ID, s.ShortDescription)
}
```

## Next

- Architecture and data flow: `architecture.md`
- Usage patterns and options: `usage.md`
- Searcher examples: `examples.md`
