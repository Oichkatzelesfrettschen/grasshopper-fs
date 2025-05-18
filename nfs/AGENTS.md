# NFS Package

The `nfs` package implements the server's main logic. Tests live in this
package as well (`*_test.go`).

When modifying code here:
- Run `go fmt ./nfs` and `go vet ./nfs`.
- Run `go test ./nfs` when dependencies are available.
- Keep exported symbols documented with comments.

