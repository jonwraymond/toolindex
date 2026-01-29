# Changelog

## Unreleased

### Added
- Cursor-based pagination (`SearchPage`, `ListNamespacesPage`) with opaque tokens and `ErrInvalidCursor`.
- Deterministic tie-breaker for equal-score search results to keep pagination stable.
- Deterministic searcher enforcement for cursor pagination (`ErrNonDeterministicSearcher`).

### Changed
- Search pagination now validates cursors against index versioning.
