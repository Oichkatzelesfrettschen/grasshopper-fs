# Repository Guidelines

This repository hosts the GoNFS NFSv3 server code. Development uses Go modules
and requires Go 1.23 or newer.

## General tasks
- Format all Go sources with `go fmt ./...` before committing.
- Run `go vet ./...` to check for common issues.
- Run `go test ./...` and `go build ./...` when possible. Tests require the
  modules to be downloaded; if the network is disabled, they may fail.
- Use `go mod tidy` and `go mod vendor` to manage dependencies.
- The `Makefile` contains a `check` target which runs `gofmt` and `go vet`.

## Style
- Go's standard formatting (tabs for indentation) is required.
- Keep lines under 100 characters when practical.

## Subdirectories
Some directories contain their own `AGENTS.md` with specific notes.
Check for those files when working in a subdirectory.

## Go setup
The Go runtime is already present in the container. If needed, verify with
`go version`. Modules can be fetched with `go mod download` when network
access is available.

