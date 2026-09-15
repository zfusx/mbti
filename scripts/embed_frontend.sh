#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
REPO_ROOT="$(dirname "$SCRIPT_DIR")"
SOURCE="$REPO_ROOT/frontend/dist/frontend/"
DESTINATION="$REPO_ROOT/backend/web/static/"

if [[ ! -d "$SOURCE" ]]; then
  echo "Frontend build not found. Run 'npm run build' in frontend first." >&2
  exit 1
fi

rsync -a --delete "$SOURCE" "$DESTINATION"
echo "Updated embedded frontend assets."
