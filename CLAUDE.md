# grocy-scanner

Terminal UI barcode scanner for Grocy inventory management, built with Go and Bubble Tea.

## Build & Install

```
go build ./...                                     # verify compilation
go build -o /home/kevin/bin/grocy-scanner .        # build and install
```

## Issue Processing

When an issue arrives from `./issues/`, follow this workflow in order:

1. **Implement** the requested change.
2. **Verify** compilation: `go build ./...`
3. **Commit** using the project's commit style (see `git log --oneline`).
4. **Code review** the changes: invoke the `code-review` skill at medium effort. Fix any real bugs found before proceeding.
5. **Build and install**: `go build -o /home/kevin/bin/grocy-scanner .`

Do not move or delete the issue file — the caller handles that.
