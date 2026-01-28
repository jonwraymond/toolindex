# Architecture

`toolindex` maintains a canonical map of tools and a cached search document
set. It is designed for fast reads and infrequent writes.

```mermaid
flowchart LR
  A[RegisterTool(s)] --> B[toolRecord map]
  B --> C[SearchDoc cache]
  C --> D[Searcher]
  D --> E[Summary list]

  B --> F[GetTool]
  B --> G[ListNamespaces]
```

## Default backend policy

The default backend selector prefers:

1. local
2. provider
3. mcp

This policy is exported as `toolindex.DefaultBackendSelector` so downstream
layers can stay consistent.
