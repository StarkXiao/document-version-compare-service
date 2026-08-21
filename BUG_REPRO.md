# Bug Reproduction

- Baseline: `origin/base_bug_005`
- Category: context
- Failure: a cancelled comparison continues and commits a result.
- Reproduce: `go test ./internal/application -count=1 -run '^TestBug005CancelledComparisonStillWritesResult$'`
- Expected repair: check cancellation before downstream work and before result persistence.
