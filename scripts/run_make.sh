#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

if command -v make >/dev/null 2>&1; then
  exec make "$@"
fi

if command -v gmake >/dev/null 2>&1; then
  exec gmake "$@"
fi

echo "GNU Make is required but neither 'make' nor 'gmake' was found in PATH" >&2
exit 127
