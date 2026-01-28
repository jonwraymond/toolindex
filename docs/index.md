# toolindex

`toolindex` is the global registry and progressive discovery layer for tools.
It stores `toolmodel.Tool` + `toolmodel.ToolBackend`, provides search, and
returns token-cheap summaries.

## What this library provides

- In-memory index with thread-safe lookup
- Search over name/namespace/description/tags
- Namespace listing
- Default backend selection (local > provider > mcp)
- Optional searcher injection (e.g., BM25 via `toolsearch`)

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

- Architecture and internal data model: `architecture.md`
- Usage patterns and options: `usage.md`
- Examples and custom search: `examples.md`
