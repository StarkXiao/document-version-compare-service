# Bug Reproduction

- Baseline: `origin/base_bug_002`
- Category: nil
- Failure: the first comment write on a fresh store can write through an uninitialized index.
- Reproduce: `go test ./internal/infrastructure -count=1 -run '^TestBug002CommentIndexNilMap$'`
- Expected repair: initialize all comment indexes before accepting writes.
