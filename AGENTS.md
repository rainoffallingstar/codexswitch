# Repository Guidelines

## Project Structure & Module Organization
This repository is a small Go CLI for switching Codex provider configs.
- `main.go`: program entrypoint.
- `cmd/`: CLI command routing (`Execute`, `add`, `list`, default interactive flow).
- `internal/ui/`: terminal UI and prompt handling.
- `internal/store/`: filesystem persistence and provider activation logic.
- `internal/config/`: config file generation (`auth.json`, `config.toml`).
- `internal/types/`: shared domain types.

Keep new functionality inside `internal/` unless it must be part of the public CLI surface.

## Build, Test, and Development Commands
- `go run .` runs the CLI locally.
- `go build ./...` compiles all packages to catch build errors.
- `go build -o codexswitch .` builds the distributable binary.
- `go test ./...` runs all unit tests (add tests as features are added).
- `go fmt ./...` formats Go source files.
- `go vet ./...` checks for common correctness issues.

Run `go fmt`, `go vet`, and `go test` before opening a PR.

## Coding Style & Naming Conventions
- Use standard Go formatting (`gofmt`) and tabs for indentation.
- Keep package names short and lowercase (`ui`, `store`, `config`).
- Exported identifiers use `PascalCase`; internal helpers use `camelCase`.
- Prefer explicit, user-facing error messages (for example: `"load providers: %v"`).
- Keep CLI behavior deterministic: avoid hidden side effects outside `~/.codexswitch` and `~/.codex`.

## Testing Guidelines
Use Go’s `testing` package with table-driven tests where possible.
- Place tests alongside code as `*_test.go` (for example: `internal/store/store_test.go`).
- Name tests `TestXxx` and focus on observable behavior.
- Prioritize coverage for parsing (`parseTOML`, `parseKV`), file operations (`SaveProvider`, `Activate`), and CLI edge cases.

## Commit & Pull Request Guidelines
Git history is not available in this workspace, so follow this convention:
- Commit message format: imperative, concise subject (for example: `Add provider validation for empty model`).
- Keep commits focused to one logical change.
- PRs should include: summary, rationale, test evidence (`go test ./...` output), and any CLI UX changes with sample terminal output.

## Security & Configuration Tips
- Never commit real API keys, `auth.json`, or user-specific files from `~/.codex*`.
- Treat provider slugs as filesystem inputs; validate/sanitize new inputs before writing paths.
