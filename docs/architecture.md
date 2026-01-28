# Architecture

`toolindex` maintains a canonical map of tools and a cached search document set.
It is optimized for frequent reads and infrequent writes.

## Registration + search flow


![Diagram](assets/diagrams/registration-search-flow.svg)


## Search sequence


![Diagram](assets/diagrams/registration-search-flow.svg)


## Progressive disclosure contract

- `Search` returns summaries only
- Schema and examples are retrieved later via `tooldocs`

## Default backend policy

The default backend selector prefers:

1. local
2. provider
3. mcp

Exported as `toolindex.DefaultBackendSelector`.
