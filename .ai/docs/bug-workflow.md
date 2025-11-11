# Bug Workflow Guide

## Overview

When you discover a bug through testing, follow this workflow to properly document and fix it.

## Step 1: Create GitHub Issue

### Issue Title Format

```
[Severity] Brief description of the bug
```

Examples:
- `BUG: GetDbClient returns non-nil client when error occurs, causing nil pointer panic on Close`
- `[SECURITY] SQL Injection vulnerability in GetSetConnectionStateSql`
- `BUG: Goroutine leak when snapshot rows are not fully consumed`

### Issue Labels

- Add the `bug` label to all bug-reporting issues
- No other labels are required at this stage

### Issue Template

```markdown
## Description
[Clear description of the bug]

## Severity
**[HIGH/MEDIUM/LOW]** - [Brief impact statement]

## Steps to Reproduce
1. [Step 1]
2. [Step 2]
3. [Result]

## Expected Behavior
[What should happen]

## Current Behavior
[What actually happens]

## Test Demonstrating Bug
See `[TestName]` in `[file path]:[line]` (currently skipped)

## Suggested Fix
[Optional: How to fix it]

## Related Code
- `[file path]:[line]` - [Description]

## Impact
[Who/what is affected by this bug]
```

### Example Issue

```markdown
## Description
The `GetDbClient` function in `pkg/initialisation/init_data.go` returns a non-nil client even when an error occurs during connection. This causes a nil pointer panic when callers attempt to call `Close()` on the returned client.

## Severity
**HIGH** - nil pointer panic

## Steps to Reproduce
1. Call `GetDbClient()` with an invalid connection string
2. The function returns both an error AND a non-nil client
3. Caller attempts to defer `client.Close()` which panics

## Expected Behavior
When an error occurs, `GetDbClient` should return `(nil, error)` following Go conventions.

## Current Behavior
Returns `(non-nil-but-invalid-client, error)` leading to panics in calling code.

## Test Demonstrating Bug
See `TestGetDbClient_WithConnectionString` in `pkg/initialisation/init_data_test.go:322` (currently skipped)

## Suggested Fix
Ensure all error paths return `nil` for the client value:
```go
if err != nil {
    return nil, err
}
```

## Related Code
- `pkg/initialisation/init_data.go:45-60` - GetDbClient function
- `pkg/initialisation/init_data_test.go:322` - Test demonstrating the bug

## Impact
Any code path that encounters a connection error will panic when attempting to close the client, potentially crashing the application.
```

## Step 2: Update Test Skip Message

After creating the issue, update the test's skip message:

```go
func TestGetDbClient_WithConnectionString(t *testing.T) {
    t.Skip("Demonstrates bug #4767 - GetDbClient returns non-nil client when error occurs. Remove this skip in bug fix PR commit 1, then fix in commit 2.")

    // ... test code ...
}
```

## Step 3: Continue Test Generation

- Don't stop to fix bugs immediately
- Document all bugs as you find them
- Complete the test generation phase first
- Bugs will be fixed in separate PRs later

## Step 4: Test Suite PR

After generating all tests:

1. **Commit Structure**: Single commit with all tests
2. **Branch**: `feature/tests-for-<packages>`
3. **Base**: `develop`
4. **Title**: `Add comprehensive tests for pkg/{package1,package2,...}`
5. **Description**: Include test count, bug discoveries, value metrics

### Test Suite PR Template

See [templates/test-pr-template.md](../templates/test-pr-template.md)

## Step 5: Create Bug Fix PRs

For each bug, create a separate PR with **exactly 2 commits**:

### Commit 1: Demonstrate the Bug

```bash
# In a git worktree for this fix
git commit -m "Unskip test demonstrating bug #4767: GetDbClient error handling"
```

Changes:
- Remove or comment out the `t.Skip()` line
- NO other changes
- Test should FAIL when run

### Commit 2: Fix the Bug

```bash
git commit -m "Fix #4767: GetDbClient returns (nil, error) on failure"
```

Changes:
- Implement the fix in production code
- NO changes to test code
- Test should PASS when run

### Bug Fix PR Details

- **Branch**: `fix/<issue-number>-brief-description`
- **Base**: `develop`
- **Title**: `Fix #<issue>: Brief description`
- **Body**: Start with `Closes #<issue>` (for automatic linking)

### Bug Fix PR Template

See [templates/bugfix-pr-template.md](../templates/bugfix-pr-template.md)

## Step 6: Push to GitHub (Two-Phase Push)

**IMPORTANT**: Push commits separately to trigger separate CI runs. This gives reviewers visual proof that the test fails before the fix and passes after.

### Phase 1: Push Test Commit

```bash
# Verify test FAILS locally
go test -v -run TestName ./pkg/path

# Push ONLY the first commit
git push -u origin fix/<issue>-description
```

**Result**: GitHub Actions runs and should **FAIL** (proves test catches bug)

### Phase 2: Push Fix Commit

```bash
# Verify test PASSES locally
go test -v -run TestName ./pkg/path

# Push the second commit
git push
```

**Result**: GitHub Actions runs again and should **PASS** (proves fix works)

### Why Two Pushes?

Reviewers can see in the PR's CI history:
1. ❌ First CI run fails (test demonstrates bug)
2. ✅ Second CI run passes (fix resolves bug)

No manual verification needed - the CI runs provide visual proof.

**See**: [PR Commit Structure Guide](./pr-commit-structure.md#pushing-to-github-two-phase-push) for detailed workflow.

## Workflow Diagram

```
┌─────────────────┐
│ Find Bug        │
│ During Testing  │
└────────┬────────┘
         │
         v
┌─────────────────┐
│ Create GitHub   │
│ Issue           │
└────────┬────────┘
         │
         v
┌─────────────────┐
│ Mark Test as    │
│ Skipped         │
└────────┬────────┘
         │
         v
┌─────────────────┐
│ Continue        │
│ Testing         │
└────────┬────────┘
         │
         v
┌─────────────────┐
│ Test Suite PR   │
│ (all tests)     │
└────────┬────────┘
         │
         v
┌─────────────────┐
│ Bug Fix PRs     │
│ (one per bug)   │
│ 2 commits each  │
└─────────────────┘
```

## Parallel Execution

Use git worktree to fix multiple bugs in parallel:

```bash
# Create worktree for each bug fix
git worktree add /tmp/fix-4767 develop
git worktree add /tmp/fix-4768 develop

# Work on each in parallel (or use Task tool)
```

See [git-worktree-guide.md](git-worktree-guide.md) for details.

## GitHub Issue Best Practices

### DO:
- ✅ Include clear reproduction steps
- ✅ Reference specific code locations with line numbers
- ✅ Explain the impact
- ✅ Link to the test that demonstrates the bug
- ✅ Suggest a fix if you know it
- ✅ Use appropriate severity labels

### DON'T:
- ❌ Create issues for intentional behavior
- ❌ Report bugs without reproduction steps
- ❌ Use vague descriptions
- ❌ Skip the test reference
- ❌ Forget to assess severity/impact

## Security Bugs

For security vulnerabilities:
1. Use `[SECURITY]` prefix in title
2. Add detailed impact assessment
3. Mark as `critical` severity
4. Include CVSS score if applicable
5. Suggest mitigation steps
6. Consider private disclosure if needed

## Next Steps

After creating issues and PRs:
1. Link PRs to issues (use `Closes #XXXX` in PR body)
2. Request reviews
3. Address feedback
4. Merge test suite PR first
5. Merge bug fix PRs independently

## Examples

See [examples/pr-workflow-example.md](../examples/pr-workflow-example.md) for a complete walkthrough.
