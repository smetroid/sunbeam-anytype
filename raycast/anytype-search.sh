#!/bin/bash

# Required parameters:
# @raycast.schemaVersion 1
# @raycast.title Anytype Commands
# @raycast.description Search command snippets from Anytype
# @raycast.mode filter
# @raycast.argument1 "query" {title = "Search tags", placeholder = "shell, commands"}

# Optional parameters:
# @raycast.icon 📋
# @raycast.author smetroid
# @raycast.packageName com.smetroid.sunbeam-anytype

set -e

BINARY_PATH="$HOME/projects/sunbeam-memos/sunbeam-anytype/sunbeam-anytype"

QUERY="${1:-}"

if [ -n "$QUERY" ]; then
    RESULT=$($BINARY_PATH --tags "$QUERY" 2>/dev/null)
else
    RESULT=$($BINARY_PATH 2>/dev/null)
fi

echo "$RESULT" | jq -n '
def escape: . | gsub("\""; "\\\"") | gsub("\n"; "\\n");
[
    inputs | select(.cmd != "") | {
        title: (.cmd[0:80] + if (.cmd | length) > 80 then "..." else "" end),
        subtitle: .tags,
        type: "default",
        action: {
            type: "copy",
            text: .cmd,
            title: "Copy Command"
        }
    }
] | {items: .}
' 2>/dev/null || echo '{"items": [{"title": "No results", "subtitle": "Run with --clipboard or --shellCommand to add new entries"}]}'
