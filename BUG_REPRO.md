# Bug Reproduction

- Baseline: `origin/base_bug_008`
- Category: concurrency
- Failure: workers treat a closed queue receive as an empty job.
- Reproduce: `go test ./internal/infrastructure -count=1 -run '^TestBug008ClosedQueueDoesNotDispatchEmptyJobs$'`
- Expected repair: stop each worker when the task channel reports that it is closed.
