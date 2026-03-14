#!/bin/bash

# Required parameters:
# @raycast.schemaVersion 1
# @raycast.title Add from Shell Command
# @raycast.description Save last shell command to Anytype
# @raycast.mode silent
# @raycast.icon 📋
# @raycast.author smetroid
# @raycast.packageName com.smetroid.sunbeam-anytype

set -e

BINARY_PATH="$HOME/projects/sunbeam-anytype/sunbeam-anytype"

# Check for optional tag argument
TAGS="${1:-}"

if [ -n "$TAGS" ]; then
    $BINARY_PATH --shellCommand --tags "$TAGS"
else
    $BINARY_PATH --shellCommand
fi

echo "Added to Anytype!"
