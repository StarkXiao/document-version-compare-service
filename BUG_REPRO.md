# Bug Reproduction

- Baseline: `origin/base_bug_001`
- Category: concurrency
- Failure: concurrent comment persistence can race with snapshot serialization.
- Reproduce: `go test ./internal/infrastructure -race -count=1 -run '^TestBug001ConcurrentCommentWrites$'`
- Expected repair: keep shared-state access and persistence consistent under concurrent writes.
