# Test Suite PR Template

## PR Title

```
Add tests for pkg/{package1,package2}
```

## PR Description

```markdown
## Summary
Added tests for [packages], focusing on [areas: edge cases, concurrency, error handling, etc.].

## Tests Added
- **pkg/package1** - [brief description of what's tested]
- **pkg/package2** - [brief description of what's tested]

## Bugs Found
[If bugs were discovered:]
- #<issue>: [brief description]
- #<issue>: [brief description]

[Tests demonstrating bugs are marked with `t.Skip()` and issue references]

## Execution
```bash
go test ./pkg/package1 ./pkg/package2
go test -race ./pkg/package1  # if concurrency tests included
```
```

## Branch

```
feature/tests-<packages>
```

Example: `feature/tests-snapshot-task`

## Notes

- Base branch: `develop`
- Single commit with all tests
- Bug-demonstrating tests should be skipped with issue references
- Bugs will be fixed in separate PRs
