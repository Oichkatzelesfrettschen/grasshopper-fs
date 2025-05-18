# Command binaries

This directory holds small programs built from the NFS server library. Each
subdirectory contains a `main.go` entry point.

Typical development cycle:
1. `go build ./cmd/...` to build all commands.
2. Run commands from their directory or via `go run`.
3. Ensure `go vet ./cmd/...` and `go test ./...` pass at the repository root.

