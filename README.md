# sunbeam-anytype

CLI tool to manage command snippets in Anytype via Sunbeam.

## Why

- Save rarely used commands, useful when working from multiple systems (work and personal computer)
- Save links/bookmarks or information to read later
- Access all your snippets directly from Sunbeam

## Features

- Add entries from clipboard
- Add entries from the command line (last shell command)
- Filter entries by tags
- Execute saved commands directly from Sunbeam

## Pre-requisites

1. [Anytype](https://anytype.io) - Local-first note-taking app
2. [Sunbeam](https://sunbeam.pomdtr.me) - Terminal launcher
3. [Go](https://go.dev) - To build the binary

## Installation

### Option 1: Download Binary

Pre-built binaries are available on the [Releases](https://github.com/smetroid/sunbeam-anytype/releases) page.

```bash
# macOS Apple Silicon
curl -L -o sunbeam-anytype https://github.com/smetroid/sunbeam-anytype/releases/latest/download/sunbeam-anytype-darwin-arm64

# macOS Intel
curl -L -o sunbeam-anytype https://github.com/smetroid/sunbeam-anytype/releases/latest/download/sunbeam-anytype-darwin-amd64

# Linux
curl -L -o sunbeam-anytype https://github.com/smetroid/sunbeam-anytype/releases/latest/download/sunbeam-anytype-linux-amd64

chmod +x sunbeam-anytype
```

### Option 2: Build from Source

1. Clone the repository:
```bash
git clone https://github.com/smetroid/sunbeam-anytype.git
cd sunbeam-anytype
```

2. Build the binary:
```bash
make build
```

3. Install Sunbeam extension:
```bash
make sunbeam-install
```

4. Configure the Sunbeam extension by adding to your `~/.config/sunbeam/sunbeam.json`:
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

Or set via environment variables:
```bash
export ANYTYPE_APP_KEY="your-app-key"
export ANYTYPE_SPACE_ID="your-space-id"
```

## Usage

### CLI

Save last shell command:
```bash
./sunbeam-anytype --shellCommand
```

Save clipboard content:
```bash
./sunbeam-anytype --clipboard
```

Get all objects (filtered by tags):
```bash
./sunbeam-anytype --tags "shell,commands"
```

### Raycast

To launch Sunbeam with Alacritty from Raycast:

1. Copy `raycast/raycast-alacritty.sh` to your Raycast Scripts directory:
```bash
cp raycast/raycast-alacritty.sh ~/Library/Application\ Support/Raycast/Scripts/
```

2. Run `raycast reload` or restart Raycast

3. Type "sunbeam" in Raycast to launch Sunbeam via Alacritty

### Sunbeam

Open Sunbeam and search for:
- `anytype-cmds` - Filter command blocks
- `anytype-snippets` - Filter snippet content
- `anytype-all` - View all objects

Commands available:
- **Run Command** - Execute the saved command
- **View Command** - View details with copy to clipboard
- **Edit Object** - Edit the object in Vim

## License

Apache 2.0
