# Bug Reproduction

- Baseline: `origin/base_bug_003`
- Category: slice
- Failure: callers can mutate the store-owned paragraph slice through a read result.
- Reproduce: `go test ./internal/application -count=1 -run '^TestBug003VersionReadMutatesStoredParagraph$'`
- Expected repair: return an independent paragraph slice from the repository.
