# Bench directory

This folder contains benchmarking scripts. They rely on a Linux system with
NFS support. To run the benchmarks:

1. Ensure the Go binaries are built (e.g., using `go build ./cmd/...`).
2. Use the provided `start-*.sh` scripts to start the server or environment.
3. Run the appropriate `run-*.sh` script.

These scripts may require `sudo` and a kernel with NFS support. No Go code
lives here so formatting checks are unnecessary.
