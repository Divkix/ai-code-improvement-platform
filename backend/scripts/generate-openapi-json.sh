#!/usr/bin/env bash
set -euo pipefail
# Generate JSON version of OpenAPI spec from YAML using yq
REPO_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
YAML_FILE="$REPO_ROOT/backend/api/openapi.yaml"
JSON_FILE="$REPO_ROOT/backend/api/openapi.json"

if ! command -v yq >/dev/null 2>&1; then
  echo "\e[31mError: \`yq\` command not found. Install from https://github.com/mikefarah/yq\e[0m" >&2
  exit 1
fi

echo "Generating $JSON_FILE from $YAML_FILE"
mkdir -p "$(dirname "$JSON_FILE")"
# Convert YAML to JSON
yq -o=json "$YAML_FILE" > "$JSON_FILE"

echo "✅ OpenAPI JSON generated." 