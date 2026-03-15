#!/bin/bash

# Required parameters:
# @raycast.schemaVersion 1
# @raycast.title Anytype Commands
# @raycast.description Search command snippets from Anytype
# @raycast.mode fullOutput
# @raycast.argument1 { "type": "text", "placeholder": "shell, commands" }

# Optional parameters:
# @raycast.icon 📋

# Documentation:
# @raycast.author smetroid

set -e

BINARY_PATH="$HOME/projects/sunbeam-anytype/sunbeam-anytype"

QUERY="${1:-}"

if [ -n "$QUERY" ]; then
    RESULT=$($BINARY_PATH --tags "$QUERY" 2>/dev/null || echo "[]")
else
    RESULT=$($BINARY_PATH 2>/dev/null || echo "[]")
fi

echo "$RESULT" | jq -r '.[] | select(.cmd != "") | .cmd' 2>/dev/null || echo "No commands found"
