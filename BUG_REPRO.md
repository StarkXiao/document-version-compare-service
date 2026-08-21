# Bug Reproduction

- Baseline: `origin/base_bug_007`
- Category: other
- Failure: UTF-8 line offsets use a different unit from downstream character positions.
- Reproduce: `go test ./internal/domain -count=1 -run '^TestBug007LineOffsetsUseBytePositions$'`
- Expected repair: define and use one consistent byte or rune offset contract.
