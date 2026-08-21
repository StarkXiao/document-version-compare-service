# Bug Reproduction

- Baseline: `origin/base_bug_006`
- Category: defer
- Failure: deferred loop closures observe the reused final paragraph value.
- Reproduce: `go test ./internal/domain -count=1 -run '^TestBug006DeferredParagraphCapture$'`
- Expected repair: capture each paragraph immediately without deferring work in the loop.
