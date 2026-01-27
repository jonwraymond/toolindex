#!/usr/bin/env bash
set -euo pipefail

min_version="v1.2.0"

sdk_version=$(awk '/github.com\/modelcontextprotocol\/go-sdk/ {print $2; exit}' go.mod || true)
if [[ -z "${sdk_version}" ]]; then
  echo "error: github.com/modelcontextprotocol/go-sdk not found in go.mod" >&2
  exit 1
fi

normalize() {
  echo "$1" | sed -E 's/^v//'
}

if [[ "$(printf '%s\n' "$(normalize "$min_version")" "$(normalize "$sdk_version")" | sort -V | head -n1)" != "$(normalize "$min_version")" ]]; then
  echo "error: go-sdk version ${sdk_version} is less than required ${min_version}" >&2
  exit 1
fi

echo "ok: go-sdk ${sdk_version} >= ${min_version}"
