# GitHub Issue Guidelines

Guidelines for creating bug reports and issues.

## Bug Issue Format

**Title:**
```
BUG: Brief description of the problem
```

For security issues, use `[SECURITY]` prefix.

**Labels:** Add `bug` label

**Body Template:**

```markdown
## Description
[Clear description of the bug]

## Severity
**[HIGH/MEDIUM/LOW]** - [Impact statement]

## Reproduction
1. [Step 1]
2. [Step 2]
3. [Observed result]

## Expected Behavior
[What should happen]

## Current Behavior
[What actually happens]

## Test Reference
See `TestName` in `path/file_test.go:line` (currently skipped)

## Suggested Fix
[Optional: proposed solution]

## Related Code
- `path/file.go:line` - [description]
```

## Example

```markdown
## Description
The `GetDbClient` function returns a non-nil client even when an error
occurs during connection, causing nil pointer panics when callers
attempt to call `Close()` on the returned client.

## Severity
**HIGH** - Nil pointer panic crashes the application

## Reproduction
1. Call `GetDbClient()` with an invalid connection string
2. Function returns both an error AND a non-nil client
3. Caller attempts to defer `client.Close()` which panics

## Expected Behavior
When an error occurs, `GetDbClient` should return `(nil, error)`
following Go conventions.

## Current Behavior
Returns `(non-nil-but-invalid-client, error)` leading to panics.

## Test Reference
See `TestGetDbClient_WithConnectionString` in
`pkg/initialisation/init_data_test.go:322` (currently skipped)

## Suggested Fix
Ensure all error paths return `nil` for the client value.

## Related Code
- `pkg/initialisation/init_data.go:45-60` - GetDbClient function
```

## When You Find a Bug

1. **Create the GitHub issue** using the template above
2. **Skip the test** with reference to the issue:
   ```go
   t.Skip("Demonstrates bug #XXXX - description. Remove skip in bug fix PR.")
   ```
3. **Continue your work** - don't stop to fix immediately

## Bug Fix Workflow

See [bug-fix-prs.md](bug-fix-prs.md) for the bug fix PR workflow (2-commit pattern).

## Best Practices

- Include specific reproduction steps
- Reference exact code locations with line numbers
- Explain the impact clearly
- Link to the test that demonstrates the bug
- For security issues: assess severity carefully and consider private disclosure
