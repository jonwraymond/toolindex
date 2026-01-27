# toolindex

Global registry + search layer for tools. toolindex ingests `toolmodel.Tool` and `toolmodel.ToolBackend` (MCP 2025-11-25 via toolmodel) and provides canonical lookup by tool ID plus progressive discovery (summaries + namespaces).

## Project Layout
- Flat module at repo root (library-only).
- No `cmd/` or `pkg/` in v1.
- `internal/` only if private subpackages are needed later.

## Key Features
- Register tools with per-tool backend bindings.
- Canonical IDs via `Tool.ToolID()` (namespace:name).
- Deterministic backend selection (default: local > provider > mcp).
- Search summaries only (token-cheap) with normalized tags.
- Thread-safe in-memory index.

## Tag Normalization
Tags should be normalized on ingest via toolmodel:

```go
reg.Tool.Tags = toolmodel.NormalizeTags(reg.Tool.Tags)
```

## Usage (Sketch)

```go
idx := toolindex.NewInMemoryIndex()
err := idx.RegisterTool(tool, backend)
if err != nil {
    // handle
}

results, _ := idx.Search("repo", 10)
for _, r := range results {
    fmt.Println(r.ID, r.ShortDescription)
}
```

## CI
GitHub Actions runs:
- MCP SDK version guard (`tools/verify-mcp-sdk-version.sh`)
- `go test ./...`

## Status
- In-memory index with pluggable searcher
- MCP/toolmodel-aligned backend bindings
