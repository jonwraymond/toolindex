# Design Notes

This page captures the tradeoffs and error semantics behind `toolindex`.

## Design tradeoffs

- **In-memory first.** `InMemoryIndex` favors low latency and zero dependencies. It trades persistence for speed and simplicity (persistence can be added later behind the same `Index` interface).
- **Progressive disclosure.** Search returns summaries only; full schemas stay out of the discovery path to keep token costs low.
- **Deterministic behavior.** Search docs are cached and sorted by tool ID to keep results reproducible across runs; cursor pagination validates against index versioning.
- **Protocol-agnostic backends.** Backends are stored as metadata only; the index does not execute tools or depend on transport details.
- **MCP-field consistency check.** If multiple backends register the same tool ID, the MCP tool fields must match. This prevents silent divergence across backends.
- **Pluggable search.** `Searcher` allows swapping lexical search with BM25 or semantic search without changing the index API.

## Error semantics

`toolindex` exposes sentinel errors for predictable failure handling:

- `ErrNotFound` – tool or backend not present.
- `ErrInvalidTool` – tool validation failed (delegates to `toolmodel.Tool.Validate`).
- `ErrInvalidBackend` – backend is missing required fields for its kind.
- `ErrInvalidCursor` – cursor tokens are malformed or stale.

### Registration failures

- If a tool with the same ID is registered and its MCP fields differ, registration fails with `ErrInvalidTool`.
- Invalid backend structures return `ErrInvalidBackend` with a descriptive message.

### Lookup failures

- `GetTool` / `GetAllBackends` return `ErrNotFound` when the tool ID is missing.
- `UnregisterBackend` returns `ErrNotFound` if the tool or backend is not present.

## Search behavior

- **Lexical default:** substring matching with scoring (name > namespace > description/tags).
- **Empty queries:** return the first N tools (deterministic order).
- **Cursor pagination:** `SearchPage` and `ListNamespacesPage` return opaque cursor tokens validated against index version.
- **Tags:** normalized via `toolmodel.NormalizeTags` and included in the search corpus.

## Extension points

- **Custom backend selector:** inject a policy (e.g., “prefer MCP over local”).
- **Custom searcher:** replace lexical search with `toolsearch` BM25 or a semantic engine.
- **External indexing:** implement `Index` with a DB-backed or remote store in the future.

## Operational guidance

- Prefer `RegisterToolsFromMCP` for ingesting MCP server tools.
- Normalize tags at ingestion to keep search results consistent.
- Keep namespaces stable so tool IDs remain durable across deployments.
