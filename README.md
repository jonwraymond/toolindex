# toolindex

> **DEPRECATED**: This package has been superseded by [`github.com/jonwraymond/tooldiscovery/index`](https://github.com/jonwraymond/tooldiscovery).
>
> All new development should use `tooldiscovery/index`. This repository is maintained for backward compatibility only and will receive no new features.
>
> See [MIGRATION.md](./MIGRATION.md) for migration instructions.

---

[![Docs](https://img.shields.io/badge/docs-ai--tools--stack-blue)](https://jonwraymond.github.io/ai-tools-stack/)

## Migration

The `toolindex` package has been consolidated into the `tooldiscovery` repository as part of the ApertureStack reorganization. The new package provides the same functionality with improved integration into the discovery pipeline.

```bash
# Old import (deprecated)
go get github.com/jonwraymond/toolindex

# New import
go get github.com/jonwraymond/tooldiscovery/index
```

## Why the change?

- **Unified discovery**: Index and discovery logic now live together
- **Simplified dependencies**: Fewer modules to manage
- **Better integration**: Seamless handoff between indexing and discovery phases

## Timeline

- **Now**: Both packages work; `toolindex` issues deprecation warnings at build time
- **Next major release**: `toolindex` will be archived
- **Future**: `toolindex` may be removed entirely

For the full migration guide, see [MIGRATION.md](./MIGRATION.md).
