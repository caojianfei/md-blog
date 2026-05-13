#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"

cd "$ROOT_DIR/web/admin"
npm install
npm run build

cd "$ROOT_DIR"
if command -v go >/dev/null 2>&1; then
  mkdir -p dist
  go build -o dist/md-blog ./cmd/server
else
  echo "go command not found, skip go build"
fi
