#!/bin/bash

OUTPUT="/tmp/anytype.json"

if [ $# -eq 0 ]; then
    jq -n '{
        title: "Anytype",
        description: "Search Anytype objects",
        preferences: [
            {
                name: "anytype_app_key",
                title: "Anytype App Key",
                type: "string"
            },
            {
                name: "anytype_space_id",
                title: "Anytype Space ID",
                type: "string"
            }
        ],
        commands: [
            {
                name: "anytype-cmds",
                title: "Anytype cmd blocks",
                mode: "filter"
            },
            {
                name: "anytype-snippets",
                title: "Anytype snippets blocks",
                mode: "filter"
            },
            {
                name: "anytype-all",
                title: "ALL objects",
                mode: "filter"
            },
            {
                name: "run-command",
                title: "execute command",
                mode: "tty",
                exit: "true"
            },
            {
                name: "view-command",
                title: "view command",
                mode: "detail",
                exit: "false"
            },
            {
                name: "edit-object",
                title: "edit object",
                mode: "tty",
                exit: "false"
            }
        ]
    }'
    exit 0
fi

COMMAND=$(echo "$1" | jq -r '.command')
FILTER=$(echo "$1" | jq -r '.command | split("-")[1]')

if [ "$COMMAND" = "anytype-cmds" ]; then
    OBJECTS=$(~/projects/sunbeam-anytype/sunbeam-anytype -tags "${FILTER}")
    echo "$OBJECTS" | jq '{
        "items": map({
            "title": .cmd,
            "accessories": [.tags],
            "actions": [{
                "type": "run",
                "title": "Run Command",
                "command": "run-command",
                "params": {
                    "codeblock": .cmd,
                }
                },{
                "type": "run",
                "title": "View Command",
                "command": "view-command",
                "params": {
                    "content": .content,
                    "codeblock": .cmd,
                },
            }]
        }),
        "actions": [{
          "title": "Refresh items",
          "type": "reload",
          "exit": "true"
      }]
    }' 2> >(tee /dev/stderr) || {
        echo "Error: jq failed to process JSON" >&2
        exit 2
    }
    exit 0
fi

if [ "$COMMAND" = "anytype-snippets" ]; then
    OBJECTS=$(~/projects/sunbeam-anytype/sunbeam-anytype -tags "${FILTER}")
    echo "$OBJECTS" | jq '{
        "items": map({
            "title": .content,
            "accessories": [.tags],
            "actions": [{
                "type": "run",
                "title": "view cmd ",
                "command": "view-command",
                "params": {
                    "content": .content,
                    "codeblock": .cmd,
                },
            }]
        }),
        "actions": [{
          "title": "Refresh items",
          "type": "reload",
          "exit": "true"
      }]
    }' 2> >(tee /dev/stderr) || {
        echo "Error: jq failed to process JSON" >&2
        exit 2
    }
    exit 0
fi

if [ "$COMMAND" = "anytype-all" ]; then
    OBJECTS=$(~/projects/sunbeam-anytype/sunbeam-anytype)
    echo "$OBJECTS" | jq '{
        "items": map({
            "title": .content,
            "subtitle": .cmd,
            "accessories": [.tags],
            "actions": [{
                "type": "run",
                "title": "view object",
                "command": "view-command",
                    "params": {
                        "content": .content,
                        "codeblock": .cmd,
                        "name": .name
                    },
                },
                {
                "type": "run",
                "title": "edit object",
                "command": "edit-object",
                "params": {
                    "content": .content,
                    "codeblock": .cmd,
                    "name": .name
                }
            }]
        }),
        "actions": [{
          "title": "Refresh items",
          "type": "reload",
          "exit": "true"
        }]
  }' 2> >(tee /dev/stderr) || {
        echo "Error: jq failed to process JSON" >&2
        exit 2
    }
    exit 0
fi

if [ "$COMMAND" = "run-command" ]; then
    CMD=$(echo "$1" | jq -r '.params.codeblock')
    konsole -e bash -c "$CMD; exec bash"
elif [ "$COMMAND" = "view-command" ]; then
    content=$(echo "$1" | jq -r '.params.content')
    codeblock=$(echo "$1" | jq -r '.params.codeblock')
    name=$(echo "$1" | jq -r '.params.name')

    jq -n \
        --arg content "$content" \
        --arg codeblock "$codeblock" \
        --arg name "$name" '{
        "markdown": $content,
        "actions": [{
            type: "copy",
            title: "Copy to clipboard",
            text: $codeblock,
            exit: false
        },
        {
            type: "run",
            title: "Edit",
            command: "edit-object",
            params: {
                "content": $content,
                "codeblock": $codeblock,
                "name": $name
            }
        }],
    }' 2> >(tee /dev/stderr) || {
        echo "Error: jq failed to process JSON" >&2
        exit 2
    }
fi

if [ "$COMMAND" = "edit-object" ]; then
    if [ ! -d "/tmp/anytype" ]; then
        mkdir /tmp/anytype
    fi
    name=$(echo "$1" | jq -r '.params.name')
    TMP_FILE="/tmp/$name.md"
    content=$(echo "$1" | jq -r '.params.content')
    codeblock=$(echo "$1" | jq -r '.params.codeblock')
    name=$(echo "$1" | jq -r '.params.name')
    echo "$1" | jq -r '.params.content' >$TMP_FILE
    before_edit=$(stat -c %Y "$TMP_FILE" 2>/dev/null || stat -f %m "$TMP_FILE")
    vim $TMP_FILE
    after_edit=$(stat -c %Y "$TMP_FILE" 2>/dev/null || stat -f %m "$TMP_FILE")
    if [ "$before_edit" -ne "$after_edit" ]; then
        STATUS=$(~/projects/sunbeam-anytype/sunbeam-anytype -update -name ${name})
        if [ $? -eq 0 ]; then
            echo "Success: updated object"
            echo $STATUS
            read -r -p "Press enter to exit, and go back to the app"
            exit 0
        else
            echo "Error: failed to update object"
            echo $STATUS
            read -r -p "Press enter to exit, and go back to the app"
            exit 1
        fi
    else
        echo "No changes made to the file."
        read -r -p "Press enter to exit, and go back to the app"
        exit 0
    fi
fi
