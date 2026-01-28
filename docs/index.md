# toolindex

Global registry and search layer for tools.

## What this repo provides

- In-memory index
- Search and namespace listing
- Backend resolution by tool ID

## Example

```go
idx := toolindex.NewInMemoryIndex()
_ = idx.RegisterTools([]toolmodel.Tool{tool}, backend)
```
