# Changelog

## [0.3.0](https://github.com/jonwraymond/toolindex/compare/toolindex-v0.2.0...toolindex-v0.3.0) (2026-01-29)


### Features

* enforce deterministic searcher for pagination ([962da40](https://github.com/jonwraymond/toolindex/commit/962da40e72df6a5ad07d993f44824e965a1108b2))
* **mcp:** add cursor-based pagination ([5f0f95c](https://github.com/jonwraymond/toolindex/commit/5f0f95ca326643b5a789251e516e874af3e9dc7c))
* **mcp:** add cursor-based pagination ([88d0a81](https://github.com/jonwraymond/toolindex/commit/88d0a8123a2eab3ed71720d616715999f21e7c89))

## [Unreleased]

### Added
- Cursor-based pagination (`SearchPage`, `ListNamespacesPage`) with opaque tokens and `ErrInvalidCursor`.
- Deterministic tie-breaker for equal-score search results to keep pagination stable.
- Deterministic searcher enforcement for cursor pagination (`ErrNonDeterministicSearcher`).

### Changed
- Search pagination now validates cursors against index versioning.

## [0.2.0](https://github.com/jonwraymond/toolindex/compare/toolindex-v0.1.9...toolindex-v0.2.0) (2026-01-28)


### Features

* add change notifications and safe backend identity ([8873191](https://github.com/jonwraymond/toolindex/commit/88731917b25bb07afca8c92a67d8f0b0ce5c68ff))
* export default backend selector ([c7c805c](https://github.com/jonwraymond/toolindex/commit/c7c805ca3510dcd5239b976f31732982dbc82e1d))


### Bug Fixes

* correct release-please step id ([c45789e](https://github.com/jonwraymond/toolindex/commit/c45789e59d599056d7c7288f48b36196e21f6d3b))
* simplify release-please token handling ([2e8e098](https://github.com/jonwraymond/toolindex/commit/2e8e098db7888282cca1e5b7c9f8b8b7d579b503))
* use app token for release-please ([ac75a67](https://github.com/jonwraymond/toolindex/commit/ac75a679fcd15e0d8c56c58aef65f7e95ca33388))
* use PAT for release-please ([9849270](https://github.com/jonwraymond/toolindex/commit/984927009cb44eb06a8894a1f079450f44d63589))


### Performance Improvements

* cache search docs and stabilize ordering ([ebbf3fc](https://github.com/jonwraymond/toolindex/commit/ebbf3fc03eb9d83ef12873251ed1052ed9537887))
* track namespace counts ([f07fea4](https://github.com/jonwraymond/toolindex/commit/f07fea4aa3bed1230debd3dd0fad59d64c1dc597))
