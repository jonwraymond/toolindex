# toolindex

`toolindex` is the global registry and progressive discovery layer for tools.
It stores `toolmodel.Tool` + `toolmodel.ToolBackend`, provides search, and
returns token-cheap summaries.

[![Docs](https://img.shields.io/badge/docs-ai--tools--stack-blue)](https://jonwraymond.github.io/ai-tools-stack/)

## Deep dives
- Design Notes: `design-notes.md`
- User Journey: `user-journey.md`

## Motivation

- **Progressive disclosure**: search returns summaries, not schemas
- **Deterministic lookup**: canonical tool IDs resolve to stable backends
- **Pluggable search**: swap lexical search for BM25 or semantic ranking

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

## Usability notes

- Summaries are token-cheap and safe to display in discovery
- Namespaces group tools for easy filtering
- Backends can be replaced without changing the tool ID

## Next

- Architecture and data flow: `architecture.md`
- Usage patterns and options: `usage.md`
- Searcher examples: `examples.md`
- Design Notes: `design-notes.md`
- User Journey: `user-journey.md`

