# codexswitch

`codexswitch` is a small CLI tool for managing multiple Codex provider configs and quickly switching the active one.

## Features

- Add provider configs (`add`)
- Edit existing provider configs (`edit`)
- Remove provider configs (`remove`)
- Copy provider configs with auto suffix (`copy` / `replicate`)
- List all configured providers (`list`)
- Switch active provider from CLI (`switch --slug`) or interactive menu (default command)
- Sync provider store via WebDAV (`sync push|pull`)
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

### Copy provider

```bash
codexswitch copy --slug openai
codexswitch replicate --slug openai
```

Creates a new provider using `<slug>-copyN` and `<DisplayName> copyN` (for example `openai-copy1`) without switching the active provider.

### WebDAV sync

Upload your local `~/.codexswitch` to a WebDAV file URL:

```bash
codexswitch sync push --webdav-url https://example.com/dav/codexswitch.tar.gz
```

Download from WebDAV and replace your local store (backs up existing `~/.codexswitch` by default):

```bash
codexswitch sync pull --webdav-url https://example.com/dav/codexswitch.tar.gz
```

You can also set credentials via environment variables:

```bash
export CODEXSWITCH_WEBDAV_URL="https://example.com/dav/codexswitch.tar.gz"
export CODEXSWITCH_WEBDAV_USER="alice"
export CODEXSWITCH_WEBDAV_PASS="secret"
codexswitch sync push
```

Or persist WebDAV settings in `~/.codexswitch/.webdav.json`. If `sync push/pull` runs in a TTY and no `--webdav-url`/`CODEXSWITCH_WEBDAV_URL` is provided, it will prompt you for the missing fields and save this file automatically.

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
codexswitch copy --help
```
