#!/bin/bash

# Required parameters:
# @raycast.schemaVersion 1
# @raycast.title Anytype Commands
# @raycast.description Search and run command snippets from Anytype
# @raycast.mode fullOutput
# @raycast.argument1 mode (optional) The mode: search, clipboard, shellCommand

# Optional parameters:
# @raycast.icon 📋

# Documentation:
# @raycast.author smetroid

set -e

BINARY_PATH="$HOME/projects/sunbeam-memos/sunbeam-anytype/sunbeam-anytype"
MODE="${1:-search}"

if [ "$MODE" = "clipboard" ]; then
    $BINARY_PATH --clipboard
elif [ "$MODE" = "shellCommand" ]; then
    $BINARY_PATH --shellCommand
else
    # Default: search with optional tags
    TAGS="${2:-}"
    if [ -n "$TAGS" ]; then
        $BINARY_PATH --tags "$TAGS"
    else
        $BINARY_PATH
    fi
fi
