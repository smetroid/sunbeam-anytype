# sunbeam-anytype

CLI tool to manage command snippets in Anytype via Sunbeam.

## Why

- Save rarely used commands, useful when working from multiple systems (work and personal computer)
- Save links/bookmarks or information to read later
- Access all your snippets directly from Sunbeam

## Features

- Single binary handles both CLI and Sunbeam extension modes
- YAML-driven templates (hot-reload without recompiling)
- Add entries from clipboard
- Add entries from the command line (last shell command)
- Filter entries by tags
- Execute saved commands directly from Sunbeam

## Pre-requisites

1. [Anytype](https://anytype.io) - Local-first note-taking app
2. [Sunbeam](https://sunbeam.pomdtr.me) - Terminal launcher
3. [Go](https://go.dev) - To build the binary (optional, pre-built binaries available)

## Installation

### Option 1: Download Binary

Pre-built binaries are available on the [Releases](https://github.com/smetroid/sunbeam-anytype/releases) page.

```bash
# macOS Apple Silicon
curl -L -o anytype https://github.com/smetroid/sunbeam-anytype/releases/latest/download/anytype-darwin-arm64

# macOS Intel
curl -L -o anytype https://github.com/smetroid/sunbeam-anytype/releases/latest/download/anytype-darwin-amd64

# Linux
curl -L -o anytype https://github.com/smetroid/sunbeam-anytype/releases/latest/download/anytype-linux-amd64

chmod +x anytype
```

### Option 2: Build from Source

```bash
git clone https://github.com/smetroid/sunbeam-anytype.git
cd sunbeam-anytype
make build
```

### Option 3: Install via Sunbeam

```bash
# After building
sunbeam extension install ./anytype
```

### Configuration

The extension will use Anytype credentials from:
1. Sunbeam preferences (in `sunbeam.json`)
2. Environment variables

Add to your `~/.config/sunbeam/sunbeam.json`:
```json
{
  "extensions": {
    "anytype": {
      "preferences": {
        "anytype_app_key": "your-app-key",
        "anytype_space_id": "your-space-id"
      }
    }
  }
}
```

Or set via environment variables:
```bash
export ANYTYPE_APP_KEY="your-app-key"
export ANYTYPE_SPACE_ID="your-space-id"
```

### Getting Anytype Credentials

**App Key**: You need to create a bot account in Anytype:
```bash
# Run Anytype server (if not already running)
anytype serve

# Create a bot account
anytype auth create <name>

# Join a space via invite link
anytype space join <invite-link>
```

The API key will be displayed after authentication. The Space ID can be found in Anytype settings or via:
```bash
anytype-cli spaces list
```

## Usage

### CLI Mode

Search objects by tag:
```bash
./anytype -tags "cmd"
```

Save last shell command:
```bash
./anytype --shellCommand
```

Save clipboard content:
```bash
./anytype --clipboard
```

Update an object:
```bash
./anytype -update -name "object-id"
```

### Sunbeam Mode

Open Sunbeam and search for:
- `anytype-cmd` - Filter command blocks (tagged with "cmd")
- `anytype-snippet` - Filter snippet content (tagged with "snippet")
- `anytype-all` - View all objects

**Actions:**
- **Copy to clipboard** - Copy the command
- **Run Command** - Execute in terminal (konsole)
- **View Command** - View details with copy/edit options
- **Edit Object** - Edit in Vim and sync back to Anytype

## Customization

Edit `anytype.yaml` to customize templates without recompiling:

```yaml
templates:
  cmd:
    title: "{{.cmd}}"
    accessories: "{{.tags}}"
    actions:
      - type: copy
        title: "Copy to clipboard"
        text: "{{.cmd}}"
        exit: true
      - type: run
        title: "Run Command"
        command: run-command
        params:
          codeblock: "{{.cmd}}"
```

Changes are hot-reloaded on next invocation - no recompile needed.

## Raycast

To launch Sunbeam with Alacritty from Raycast:

1. Copy `raycast/raycast-alacritty.sh` to your Raycast Scripts directory:
```bash
cp raycast/raycast-alacritty.sh ~/Library/Application\ Support/Raycast/Scripts/
```

2. Run `raycast reload` or restart Raycast

3. Type "sunbeam" in Raycast to launch Sunbeam via Alacritty

## License

Apache 2.0