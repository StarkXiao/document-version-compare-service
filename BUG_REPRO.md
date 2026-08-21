# Bug Reproduction

- Baseline: `origin/base_bug_009`
- Category: other
- Failure: migration versions are ordered lexically instead of numerically.
- Reproduce: `go test ./internal/infrastructure -count=1 -run '^TestBug009MigrationVersionsUseNumericOrder$'`
- Expected repair: parse numeric prefixes and apply migrations in numeric order.
