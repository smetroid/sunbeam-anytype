#!/bin/bash

# Required parameters:
# @raycast.schemaVersion 1
# @raycast.title Add from Clipboard
# @raycast.description Save clipboard content to Anytype
# @raycast.mode silent
# @raycast.icon 📋
# @raycast.author smetroid
# @raycast.packageName com.smetroid.sunbeam-anytype

set -e

BINARY_PATH="$HOME/projects/sunbeam-memos/sunbeam-anytype/sunbeam-anytype"

# Check for optional tag argument
TAGS="${1:-}"

if [ -n "$TAGS" ]; then
    $BINARY_PATH --clipboard --tags "$TAGS"
else
    $BINARY_PATH --clipboard
fi

echo "Added to Anytype!"
