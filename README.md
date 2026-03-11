# PromptRails CLI

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Command-line interface for [PromptRails](https://promptrails.ai) — the AI agent orchestration platform.

## Installation

### Homebrew (macOS / Linux)

```bash
brew tap promptrails/tap
brew install promptrails
```

### Go Install

```bash
go install github.com/promptrails/cli/cmd/promptrails@latest
```

### Binary Download

Download the latest release from the [releases page](https://github.com/promptrails/cli/releases).

## Quick Start

```bash
# Interactive setup
promptrails init

# Or non-interactive (CI/CD)
promptrails init --api-key pr_key_...

# Check current context
promptrails status
```

## Commands

### Setup

```bash
promptrails init              # Set up CLI with API key
promptrails status            # Show current auth & workspace context
promptrails version           # Show CLI version
```

### Agents

```bash
promptrails agent list                           # List agents
promptrails agent list --type simple             # Filter by type
promptrails agent get <id>                       # Get details
promptrails agent create --name "My Agent"       # Create
promptrails agent update <id> --name "New Name"  # Update
promptrails agent delete <id>                    # Delete
promptrails agent execute <id> --input '{"q": "hello"}'
promptrails agent versions <id>                  # List versions
promptrails agent promote <id> <version-id>      # Set current version
```

### Prompts

```bash
promptrails prompt list
promptrails prompt get <id>
promptrails prompt create --name "My Prompt"
promptrails prompt run <id> --user-prompt "Hello"
promptrails prompt versions <id>
promptrails prompt promote <id> <version-id>
```

### Executions

```bash
promptrails execution list                # List executions
promptrails execution list --agent <id>   # Filter by agent
promptrails execution get <id>            # Get details
```

### Credentials

```bash
promptrails credential list
promptrails credential create --provider openai --name "Production"
promptrails credential delete <id>
promptrails credential check <id>
```

### API Keys

```bash
promptrails apikey list
promptrails apikey create --name "CI"     # Displays the key once
promptrails apikey delete <id>
```

### Media Studio

```bash
promptrails media generate --provider stability --media-type image_gen --model sd3.5-large --prompt "A sunset"
promptrails media generate --provider elevenlabs --media-type tts --model eleven_multilingual_v2 --prompt "Hello" --config voice_id=21m00Tcm4TlvDq8ikWAM
```

### Assets

```bash
promptrails assets list                          # List all assets
promptrails assets list --type image             # Filter by type
promptrails assets get <id>                      # Get details
promptrails assets signed-url <id>               # Get download URL
promptrails assets delete <id>                   # Delete
```

### Media Models

```bash
promptrails media-models list                    # List all media models
promptrails media-models list --provider fal     # Filter by provider
promptrails media-models list --media-type tts   # Filter by media type
```

### Webhook Triggers

```bash
promptrails webhook-trigger list
promptrails wt get <trigger-id>
promptrails wt create --name "GitHub" --agent-id <id>
promptrails wt update <trigger-id> --active
promptrails wt delete <trigger-id>
```

### Shell Completion

```bash
promptrails completion bash > /etc/bash_completion.d/promptrails
promptrails completion zsh  > "${fpath[1]}/_promptrails"
promptrails completion fish > ~/.config/fish/completions/promptrails.fish
```

## Global Flags

| Flag           | Description                                |
| -------------- | ------------------------------------------ |
| `-o, --output` | Output format: `table` (default) or `json` |
| `--workspace`  | Override workspace ID for this command      |
| `--api-url`    | Override API base URL                       |
| `--no-color`   | Disable color output                        |

## Environment Variables

| Variable                   | Description                            |
| -------------------------- | -------------------------------------- |
| `PROMPTRAILS_API_KEY`      | API key (overrides stored credentials) |
| `PROMPTRAILS_WORKSPACE_ID` | Workspace ID override                  |
| `PROMPTRAILS_API_URL`      | API base URL override                  |

## Config Files

Stored in `~/.promptrails/`:

| File               | Permissions | Contents                                 |
| ------------------ | ----------- | ---------------------------------------- |
| `config.json`      | `0644`      | API URL, active workspace, output format |
| `credentials.json` | `0600`      | API key                                  |

## License

MIT
