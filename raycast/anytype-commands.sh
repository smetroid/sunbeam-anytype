#!/bin/bash

# Required parameters:
# @raycast.schemaVersion 1
# @raycast.title Anytype Commands
# @raycast.description Search command snippets from Anytype
# @raycast.mode filter
# @raycast.argument1 "query" {title = "Search tags", placeholder = "shell, commands"}

# Optional parameters:
# @raycast.icon 📋

# Documentation:
# @raycast.author smetroid

set -e

BINARY_PATH="$HOME/projects/sunbeam-memos/sunbeam-anytype/sunbeam-anytype"

# Get the search query from the argument
QUERY="${1:-}"

# Run the binary and get JSON output
if [ -n "$QUERY" ]; then
    RESULT=$($BINARY_PATH --tags "$QUERY" 2>/dev/null || echo "[]")
else
    RESULT=$($BINARY_PATH 2>/dev/null || echo "[]")
fi

# Convert the JSON output to Raycast format
echo "$RESULT" | jq -r '.[] | {
    title: .cmd // .content[0:100],
    subtitle: .tags,
    actions: [
        {
            type: "copy",
            title: "Copy Command",
            text: .cmd,
            "exit": true
        },
        {
            type: "run",
            title: "Copy Content",
            command: "raycast-default",
            params: {
                "terminal-command": "echo " + (.content | @sh) + " | pbcopy"
            },
            "exit": true
        }
    ]
} | { items: [.] }' 2>/dev/null || echo '{"items": []}'
