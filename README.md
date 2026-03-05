# codexswitch

`codexswitch` is a small CLI tool for managing multiple Codex provider configs and quickly switching the active one.

## Features

- Add provider configs (`add`)
- Edit existing provider configs (`edit`)
- Remove provider configs (`remove`)
- List all configured providers (`list`)
- Interactive switch for active provider (default command)

Provider files are stored in `~/.codexswitch/<slug>/`, and activation copies files to `~/.codex/`.

## Build

### Build local binary

```bash
go build -o codexswitch .
```

### Build static binary (low runtime dependency)

```bash
CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o codexswitch-static .
```

## Usage

### Add provider

```bash
codexswitch add \
  --slug openai \
  --name OpenAI \
  --api-key sk-xxx \
  --model gpt-4.1 \
  --base-url https://api.openai.com/v1 \
  --wire-api responses
```

If flags are missing in a TTY, prompts will ask for missing fields.

### List providers

```bash
codexswitch list
```

### Switch active provider

```bash
codexswitch
```

### Edit provider

```bash
codexswitch edit --slug openai --model gpt-4.1-mini
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
codexswitch edit --help
codexswitch remove --help
```
