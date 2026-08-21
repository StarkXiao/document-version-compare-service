# Bug Reproduction

- Baseline: `origin/base_bug_010`
- Category: nil
- Failure: a nil enqueue callback is invoked during a valid version write.
- Reproduce: `go test ./internal/application -count=1 -run '^TestBug010NilEnqueueCallbackPanicsOnVersionCreate$'`
- Expected repair: make asynchronous enqueue optional without breaking version persistence.
