# Migration Guide: toolindex to tooldiscovery/index

This guide covers migrating from `github.com/jonwraymond/toolindex` to `github.com/jonwraymond/tooldiscovery/index`.

## Import Path Changes

| Old Import | New Import |
|------------|------------|
| `github.com/jonwraymond/toolindex` | `github.com/jonwraymond/tooldiscovery/index` |

## Quick Migration

### 1. Update go.mod

```bash
# Remove old dependency
go mod edit -droprequire github.com/jonwraymond/toolindex

# Add new dependency
go get github.com/jonwraymond/tooldiscovery/index@latest
```

### 2. Update imports

**Before:**
```go
import (
    "github.com/jonwraymond/toolindex"
)

idx := toolindex.NewInMemoryIndex()
```

**After:**
```go
import (
    "github.com/jonwraymond/tooldiscovery/index"
)

idx := index.NewInMemoryIndex()
```

### 3. Find and replace

Run this command to update all import statements:

```bash
find . -name "*.go" -exec sed -i '' \
    's|github.com/jonwraymond/toolindex|github.com/jonwraymond/tooldiscovery/index|g' {} +
```

Or use `goimports`:

```bash
goimports -w .
```

## API Compatibility

The `tooldiscovery/index` package maintains full API compatibility with `toolindex`. All exported types, functions, and constants have been preserved:

### Types
- `Index` (interface)
- `InMemoryIndex`
- `IndexOptions`
- `BackendSelector`
- `Searcher`

### Functions
- `NewInMemoryIndex(opts ...IndexOptions) *InMemoryIndex`
- `DefaultBackendSelector`

### Methods on Index
- `RegisterTool(tool toolmodel.Tool, backend toolmodel.ToolBackend) error`
- `GetTool(id string) (toolmodel.Tool, toolmodel.ToolBackend, error)`
- `GetAllBackends(id string) ([]toolmodel.ToolBackend, error)`
- `Search(query string, limit int) ([]toolmodel.ToolSummary, error)`
- `SearchPage(query string, limit int, cursor string) ([]toolmodel.ToolSummary, string, error)`

## Troubleshooting

### Build errors after migration

If you see import cycle errors, ensure you've updated all files:

```bash
grep -r "jonwraymond/toolindex" --include="*.go" .
```

### Version conflicts

If you have transitive dependencies still using `toolindex`, you may need to update those dependencies first or use a replace directive temporarily:

```go
// go.mod
replace github.com/jonwraymond/toolindex => github.com/jonwraymond/tooldiscovery/index v0.x.x
```

## Getting Help

- File issues at [tooldiscovery](https://github.com/jonwraymond/tooldiscovery/issues)
- See [ApertureStack docs](https://jonwraymond.github.io/ai-tools-stack/) for architecture overview
