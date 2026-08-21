# Bug Reproduction

- Baseline: `origin/base_bug_004`
- Category: error
- Failure: converting a not-found error to text prevents HTTP classification as 404.
- Reproduce: `go test ./internal/transport/http -count=1 -run '^TestBug004WrappedNotFoundLosesClassification$'`
- Expected repair: preserve sentinel identity with wrapped errors and errors.Is.
