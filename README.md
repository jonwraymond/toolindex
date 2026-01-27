# toolindex

`toolindex` is the global registry + progressive discovery layer for tools.
It ingests `toolmodel.Tool` plus `toolmodel.ToolBackend` bindings and provides:
- canonical lookup by tool ID, and
- token-cheap discovery via summaries and namespaces.

This module sits directly on top of `toolmodel` and is used by `tooldocs`,
`toolrun`, and `toolcode`.

## Install

```bash
go get github.com/jonwraymond/toolindex
```

## Core behaviors

- Canonical IDs come from `Tool.ToolID()` (`namespace:name`)
- Tools can have multiple backends
- Default backend selection is deterministic:
  - `local > provider > mcp`
- Search returns summaries only (no schemas)
- The in-memory index is thread-safe

The default backend policy is exported as:
- `toolindex.DefaultBackendSelector`

## Quick start

Register tools and search:

```go
import (
  "fmt"
  "log"

  "github.com/jonwraymond/toolindex"
  "github.com/jonwraymond/toolmodel"
  "github.com/modelcontextprotocol/go-sdk/mcp"
)

idx := toolindex.NewInMemoryIndex()

tool := toolmodel.Tool{
  Namespace: "github",
  Tool: mcp.Tool{
    Name:        "get_repo",
    Description: "Get repository metadata",
    InputSchema: map[string]any{
      "type": "object",
      "properties": map[string]any{
        "owner": {"type": "string"},
        "repo":  {"type": "string"},
      },
      "required": []string{"owner", "repo"},
    },
  },
  Tags: toolmodel.NormalizeTags([]string{"GitHub", "repos"}),
}

backend := toolmodel.ToolBackend{
  Kind: toolmodel.BackendKindMCP,
  MCP:  &toolmodel.MCPBackend{ServerName: "github"},
}

if err := idx.RegisterTool(tool, backend); err != nil {
  log.Fatal(err)
}

summaries, _ := idx.Search("repo metadata", 5)
for _, s := range summaries {
  fmt.Println(s.ID, s.ShortDescription)
}
```

Resolve a tool for execution:

```go
import "log"

t, defaultBackend, err := idx.GetTool("github:get_repo")
if err != nil {
  log.Fatal(err)
}

allBackends, _ := idx.GetAllBackends(t.ToolID())
fmt.Println(defaultBackend.Kind, len(allBackends))
```

## Configuration and extension points

`NewInMemoryIndex` supports optional overrides:
- `BackendSelector` (choose a different default backend policy)
- `Searcher` (replace lexical search with another strategy)

These are provided via `toolindex.IndexOptions`.

## Version compatibility (current tags)

- `toolmodel`: `v0.1.0`
- `toolindex`: `v0.1.1`
- `tooldocs`: `v0.1.1`
- `toolrun`: `v0.1.0`
- `toolcode`: `v0.1.0`

Downstream libraries should import tagged versions to keep the stack aligned.
