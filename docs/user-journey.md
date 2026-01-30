# User Journey

This journey shows how `toolindex` supports end-to-end agent workflows by powering discovery and canonical lookup.

## End-to-end flow (stack view)

![Diagram](assets/diagrams/user-journey.svg)

```mermaid
%%{init: {'theme': 'base', 'themeVariables': {'primaryColor': '#3182ce', 'primaryTextColor': '#fff'}}}%%
flowchart TB
    subgraph ingest["Tool Ingestion"]
        MCP["📡 MCP Servers"]
        Local["🏠 Local Tools"]
        Provider["🔌 Provider Tools"]
        RegMCP["RegisterToolsFromMCP()"]
        RegTool["RegisterTool()"]
    end

    subgraph index["toolindex.Index"]
        Registry["📇 In-Memory Registry<br/><small>Thread-safe, RWMutex</small>"]
        Backends["⚙️ Backend Map<br/><small>toolID → []ToolBackend</small>"]
    end

    subgraph search["Search"]
        Query["🔍 Search(query, limit)"]
        Searcher["🎯 Searcher Interface<br/><small>lexical | BM25 | semantic</small>"]
        Results["📋 Summary[]<br/><small>No schemas (token-cheap)</small>"]
    end

    subgraph lookup["Lookup"]
        Get["📖 GetTool(id)"]
        GetBE["⚙️ GetAllBackends(id)"]
        Tool["🧱 toolmodel.Tool"]
    end

    MCP --> RegMCP --> Registry
    Local --> RegTool --> Registry
    Provider --> RegTool
    Registry --> Backends

    Registry --> Query --> Searcher --> Results
    Registry --> Get --> Tool
    Backends --> GetBE

    style ingest fill:#718096,stroke:#4a5568
    style index fill:#3182ce,stroke:#2c5282,stroke-width:2px
    style search fill:#d69e2e,stroke:#b7791f
    style lookup fill:#38a169,stroke:#276749
```

## Step-by-step

1. **Ingest tools** from MCP servers or local registries:
   - `RegisterToolsFromMCP(serverName, tools)`
   - `RegisterTool(tool, backend)` for local/provider tools
2. **Agent discovers tools** via `search_tools` (summary only).
3. **Agent selects a tool ID** and requests schema/docs (`describe_tool`).
4. **Agent executes** via `run_tool` or `run_chain`.

## Example: register and search

```go
idx := toolindex.NewInMemoryIndex()

// MCP-backed tools
_ = idx.RegisterToolsFromMCP("github", []toolmodel.Tool{repoTool})

// Local tool
_ = idx.RegisterTool(localTool, toolmodel.ToolBackend{
  Kind: toolmodel.BackendKindLocal,
  Local: &toolmodel.LocalBackend{Name: "local_handler"},
})

summaries, _ := idx.Search("repo", 5)
for _, s := range summaries {
  fmt.Println(s.ID, s.ShortDescription)
}
```

## Expected outcomes

- Fast, deterministic search results.
- Stable tool IDs across backends.
- Safe summaries for progressive disclosure.

## Common failure modes

- `ErrInvalidTool` when tool validation fails.
- `ErrInvalidBackend` when backend metadata is incomplete.
- `ErrNotFound` for missing tool IDs or backends.

## Why this matters

`toolindex` is the “front door” for tool discovery. It keeps discovery cheap and consistent while letting execution and documentation happen elsewhere.
