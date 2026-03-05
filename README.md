# codexswitch

`codexswitch` is a small CLI tool for managing multiple Codex provider configs and quickly switching the active one.

## Features

- Add provider configs (`add`)
- Edit existing provider configs (`edit`)
- Remove provider configs (`remove`)
- List all configured providers (`list`)
- Switch active provider from CLI (`switch --slug`) or interactive menu (default command)
- Configurable `reasoning-effort` (`none|minimal|low|medium|high|xhigh`, default `medium`)

Provider files are stored in `~/.codexswitch/<slug>/`, and activation copies files to `~/.codex/`.
Provider slug must match `[a-z0-9_-]` (1-64 chars, cannot start with `.`).

## Build

### Build local binary

```bash
go build -o codexswitch .
```

### Build static binary (low runtime dependency)

```bash
CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o codexswitch-static .
```

## Release

Run the `Release` workflow manually (`workflow_dispatch`) in GitHub Actions.
The workflow builds static binaries for Linux, macOS, and Windows, then publishes them to a GitHub Release.
The release tag is automatically set to the current UTC date (format: `YYYY-MM-DD`).

Release assets include:
- `codexswitch-linux-amd64`
- `codexswitch-macos-amd64`
- `codexswitch-macos-arm64`
- `codexswitch-windows-amd64.exe`
- `SHA256SUMS`

## Usage

### Add provider

```bash
codexswitch add \
  --slug openai \
  --name OpenAI \
  --api-key sk-xxx \
  --model gpt-4.1 \
  --base-url https://api.openai.com/v1 \
  --wire-api responses \
  --reasoning-effort medium
```

If flags are missing in a TTY, prompts will ask for missing fields (including `reasoning-effort`).

### List providers

```bash
codexswitch list
```

### Switch active provider

```bash
codexswitch
codexswitch switch --slug openai
```

`codexswitch switch` also supports interactive provider selection in a TTY.

### Edit provider

```bash
codexswitch edit --slug openai --model gpt-4.1-mini
codexswitch edit --slug openai --reasoning-effort high
```

Without `--slug` in interactive mode, you can select a provider from a menu.

### Remove provider

```bash
codexswitch remove --slug openai
codexswitch remove --slug openai --yes
```

The active provider cannot be removed directly.

## Development

```bash
go test ./...
```

Show help:

```bash
codexswitch --help
codexswitch switch --help
codexswitch add --help
codexswitch edit --help
codexswitch remove --help
```
