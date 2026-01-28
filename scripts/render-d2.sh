#!/usr/bin/env bash
set -euo pipefail

if [ ! -d "docs/diagrams" ]; then
  echo "docs/diagrams not found; skipping"
  exit 0
fi

mkdir -p docs/assets/diagrams

shopt -s nullglob
files=(docs/diagrams/*.d2)
if [ ${#files[@]} -eq 0 ]; then
  echo "No .d2 files found; skipping"
  exit 0
fi

for f in "${files[@]}"; do
  out="docs/assets/diagrams/$(basename "${f%.d2}").svg"
  d2 --layout=elk --theme=104 --dark-theme=201 --pad=24 --center "$f" "$out"
done
