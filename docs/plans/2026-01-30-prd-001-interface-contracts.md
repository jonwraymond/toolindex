# PRD-001 Interface Contracts — toolindex

**Status:** Done
**Date:** 2026-01-30


## Overview
Define explicit interface contracts (GoDoc + documented semantics) for all interfaces in this repo. Contracts must state concurrency guarantees, error semantics, ownership of inputs/outputs, and context handling.


## Goals
- Every interface has explicit GoDoc describing behavioral contract.
- Contract behavior is codified in tests (contract tests).
- Docs/README updated where behavior is user-facing.


## Non-Goals
- No API shape changes unless required to satisfy the contract tests.
- No new features beyond contract clarity and tests.


## Interface Inventory
| Interface | File | Methods |
| --- | --- | --- |
| `Index` | `toolindex/index.go:51` | RegisterTool(tool toolmodel.Tool, backend toolmodel.ToolBackend) error<br/>RegisterTools(regs []ToolRegistration) error<br/>RegisterToolsFromMCP(serverName string, tools []toolmodel.Tool) error<br/>UnregisterBackend(toolID string, kind toolmodel.BackendKind, backendID string) error<br/>GetTool(id string) (toolmodel.Tool, toolmodel.ToolBackend, error)<br/>GetAllBackends(id string) ([]toolmodel.ToolBackend, error)<br/>Search(query string, limit int) ([]Summary, error)<br/>SearchPage(query string, limit int, cursor string) ([]Summary, string, error)<br/>ListNamespaces() ([]string, error)<br/>ListNamespacesPage(limit int, cursor string) ([]string, string, error) |
| `Searcher` | `toolindex/index.go:81` | Search(query string, limit int, docs []SearchDoc) ([]Summary, error) |
| `DeterministicSearcher` | `toolindex/index.go:90` | Searcher<br/>Deterministic() bool |
| `ChangeNotifier` | `toolindex/index.go:118` | OnChange(listener ChangeListener) (unsubscribe func()) |
| `Refresher` | `toolindex/index.go:123` | Refresh() uint64 |

## Contract Template (apply per interface)
- **Thread-safety:** explicitly state if safe for concurrent use.
- **Context:** cancellation/deadline handling (if context is a parameter).
- **Errors:** classification, retryability, and wrapping expectations.
- **Ownership:** who owns/allocates inputs/outputs; mutation expectations.
- **Determinism/order:** ordering guarantees for returned slices/maps/streams.
- **Nil/zero handling:** behavior for nil inputs or empty values.


## Acceptance Criteria
- All interfaces have GoDoc with explicit behavioral contract.
- Contract tests exist and pass.
- No interface contract contradictions across repos.
