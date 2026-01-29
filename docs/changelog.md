# Changelog

## Unreleased

### Added
- Cursor-based pagination (`SearchPage`, `ListNamespacesPage`) with opaque tokens and `ErrInvalidCursor`.
- Deterministic tie-breaker for equal-score search results to keep pagination stable.

### Changed
- Search pagination now validates cursors against index versioning.

