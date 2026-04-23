# AGENTS.md

## Build & Test
- Build: `go build ./...`
- Lint: `go vet ./...`
- No test suite yet. When added, run all: `go test ./...`, single: `go test ./pkg/workflowy -run TestXxx`

## Architecture
- `pkg/workflowy/` — Public SDK. Entry point: `NewClient()` with functional options. Sub-services (`Nodes`, `Targets`) use fluent builder pattern (e.g. `client.Nodes.Create("x").ParentID(...).Do(ctx)`).
- `internal/cli/` — CLI layer (cobra). Pure protocol translation: parse flags → call SDK → format output. `app` struct holds shared state; `PersistentPreRunE` handles client init.
- `cmd/wf/` — CLI entrypoint, delegates to `cli.Execute()`.
- `tree.go` — Stateless `BuildTree()` reconstructs tree from flat export.

## Code Style
- Go 1.25+. No generics used. Standard library preferred; only deps: cobra, x/term.
- Errors: wrap with `fmt.Errorf("workflowy: context: %w", err)`. API errors use `*workflowy.Error` type; check with `IsNotFound`/`IsRateLimited`. Error-to-wire conversion at CLI boundary only.
- Naming: `XxxCall` for fluent builders, `XxxService` for sub-services. Unexported `service` struct holds shared HTTP infra.
- Fluent builders must always be called as `call = call.Method(...)` (assign return value).
- JSON tags use snake_case. Pointer fields (`*string`, `*LayoutMode`) for optional/nullable API fields.
- CLI output: `printSuccess`/`printError` for `--json`; `fmt.Print` for human output. Use `truncateOutput` (rune-safe) for `--max-output`.
- Config/cache paths follow XDG conventions; `configDir()`/`cacheDir()` return `(string, error)`.
