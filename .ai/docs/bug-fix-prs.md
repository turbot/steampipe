# Bug Fix PR Guide

## Two-Commit Pattern

Every bug fix PR must have **exactly 2 commits**:
1. **Commit 1**: Demonstrate the bug (test fails)
2. **Commit 2**: Fix the bug (test passes)

This pattern provides:
- Clear demonstration that the bug exists
- Proof that the fix resolves the issue
- Easy code review (reviewers can see the test fail, then pass)
- Test-driven development (TDD) workflow

## Commit 1: Unskip/Add Test

### Purpose
Demonstrate that the bug exists by having a failing test.

### Changes
- If test exists in test suite: Remove `t.Skip()` line
- If test doesn't exist: Add the test
- **NO OTHER CHANGES**

### Commit Message Format
```
Unskip test demonstrating bug #<issue>: <brief description>
```

or

```
Add test for #<issue>: <brief description>
```

### Examples
```
Unskip test demonstrating bug #4767: GetDbClient error handling
```

```
Add test for #4717: Target.Export() should handle nil exporter gracefully
```

### Verification
```bash
# Test should FAIL
go test -v -run TestName ./pkg/path
# Exit code: 1
```

## Commit 2: Implement Fix

### Purpose
Fix the bug with minimal changes.

### Changes
- Implement the fix in production code
- **NO changes to test code**
- Keep changes minimal and focused

### Commit Message Format
```
Fix #<issue>: <brief description of fix>
```

### Examples
```
Fix #4767: GetDbClient returns (nil, error) on failure
```

```
Fix #4717: Add nil check to Target.Export()
```

### Verification
```bash
# Test should PASS
go test -v -run TestName ./pkg/path
# Exit code: 0
```

## Creating the Two Commits

### Method 1: Interactive Rebase (Recommended)

If you have more commits, squash them:

```bash
# View commit history
git log --oneline -5

# Interactive rebase to squash
git rebase -i HEAD~3

# Mark commits:
# pick <hash> Unskip test...
# squash <hash> Additional test changes
# pick <hash> Fix bug
# squash <hash> Address review comments
```

### Method 2: Cherry-Pick

If rebasing from another branch:

```bash
# In your fix branch based on develop
git cherry-pick <test-commit-hash>
git cherry-pick <fix-commit-hash>
```

### Method 3: Build Commits Correctly

```bash
# Start from develop
git checkout -b fix/1234-description develop

# Commit 1: Unskip test
# Edit test file to remove t.Skip()
git add pkg/path/file_test.go
git commit -m "Unskip test demonstrating bug #1234: Description"

# Verify it fails
go test -v -run TestName ./pkg/path

# Commit 2: Fix bug
# Edit production code
git add pkg/path/file.go
git commit -m "Fix #1234: Description of fix"

# Verify it passes
go test -v -run TestName ./pkg/path
```

## Pushing to GitHub: Two-Phase Push

**IMPORTANT**: Push commits separately to trigger CI runs for each commit. This provides clear visual evidence in the PR that the test fails before the fix and passes after.

### Phase 1: Push Test Commit (Should Fail CI)

```bash
# Create and switch to your branch
git checkout -b fix/1234-description develop

# Make commit 1 (unskip test)
git add pkg/path/file_test.go
git commit -m "Unskip test demonstrating bug #1234: Description"

# Verify test fails locally
go test -v -run TestName ./pkg/path

# Push ONLY the first commit
git push -u origin fix/1234-description
```

At this point:
- GitHub Actions will run tests
- CI should **FAIL** on the test you unskipped
- This proves the test catches the bug

### Phase 2: Push Fix Commit (Should Pass CI)

```bash
# Make commit 2 (fix bug)
git add pkg/path/file.go
git commit -m "Fix #1234: Description of fix"

# Verify test passes locally
go test -v -run TestName ./pkg/path

# Push the second commit
git push
```

At this point:
- GitHub Actions will run tests again
- CI should **PASS** with the fix
- This proves the fix works

### Creating the PR

Create the PR after the first push (before the fix):

```bash
# After phase 1 push
gh pr create --base develop \
  --title "Brief description closes #1234" \
  --body "## Summary
[Description]

## Changes
- Commit 1: Unskipped test demonstrating the bug
- Commit 2: Implemented fix (coming in next push)

## Test Results
Will be visible in CI runs:
- First CI run should FAIL (demonstrating bug)
- Second CI run should PASS (proving fix works)
"
```

Or create it after both commits are pushed - either way works.

### Why This Matters for Reviewers

This two-phase push gives reviewers:
1. **Visual proof** the test fails without the fix (failed CI run)
2. **Visual proof** the test passes with the fix (passed CI run)
3. **No manual verification needed** - just look at the CI history in the PR
4. **Clear diff** between what fails and what fixes it

### Example PR Timeline

```
✅ PR opened
❌ CI run #1: Test failure (commit 1)
   "FAIL: TestName - expected nil, got non-nil client"
⏱️ Commit 2 pushed
✅ CI run #2: All tests pass (commit 2)
   "PASS: TestName"
```

Reviewers can click through the CI runs to see the exact failure and success.

## PR Structure

### Branch Naming

```
fix/<issue-number>-brief-kebab-case-description
```

Examples:
- `fix/4767-getdbclient-error-handling`
- `fix/4743-status-spinner-visible-race`
- `fix/4717-nil-exporter-check`

### PR Title

```
Brief description closes #<issue>
```

Examples:
- `GetDbClient error handling closes #4767`
- `Race condition on StatusSpinner.visible field closes #4743`

### PR Description

```markdown
## Summary
[Brief description of the bug and fix]

## Changes
- Commit 1: Unskipped test demonstrating the bug
- Commit 2: Implemented fix by [description]

## Test Results
- Before fix: [Describe failure - panic, wrong result, etc.]
- After fix: Test passes

## Verification
\`\`\`bash
# Commit 1 (test only)
go test -v -run TestName ./pkg/path
# FAIL: [error message]

# Commit 2 (with fix)
go test -v -run TestName ./pkg/path
# PASS
\`\`\`
```

### Labels

Add appropriate labels:
- `bug`
- Severity: `critical`, `high-priority` (if available)
- Type: `security`, `race-condition`, `nil-pointer`, etc.

## What NOT to Include

### ❌ Don't Add to Commits
- Unrelated formatting changes
- Refactoring not directly related to the bug
- go.mod changes (unless required by new imports)
- Documentation updates (separate PR)
- Multiple bug fixes in one PR

### ❌ Don't Combine Commits
- Keep test and fix as separate commits
- Don't squash them together
- Don't add "fix review comments" commits (amend instead)

## Handling Review Feedback

### If Test Needs Changes
```bash
# Amend commit 1
git checkout HEAD~1
# Make test changes
git add file_test.go
git commit --amend
git rebase --continue
```

### If Fix Needs Changes
```bash
# Amend commit 2
# Make fix changes
git add file.go
git commit --amend
```

### Force Push After Amendments
```bash
git push --force-with-lease
```

## Multiple Related Bugs

If fixing multiple related bugs:
- Create separate issues for each
- Create separate PRs for each
- Don't combine into one PR
- Each PR: 2 commits

## Test Suite PRs (Different Pattern)

Test suite PRs follow a different pattern:
- **Single commit** with all tests
- Branch: `feature/tests-for-<packages>`
- Base: `develop`
- Include bug-demonstrating tests (marked as skipped)

See [templates/test-pr-template.md](../templates/test-pr-template.md)

## Verifying Commit Structure

Before pushing:

```bash
# Check commit count
git log --oneline origin/develop..HEAD
# Should show exactly 2 commits

# Check first commit (test only)
git show HEAD~1 --stat
# Should only modify test file(s)

# Check second commit (fix only)
git show HEAD --stat
# Should only modify production code file(s)

# Verify test behavior
git checkout HEAD~1 && go test -v -run TestName ./pkg/path  # Should FAIL
git checkout HEAD && go test -v -run TestName ./pkg/path    # Should PASS
```

## Common Mistakes

### ❌ Mistake 1: Combined Commit
```
Fix #1234: Add test and fix bug
```
**Problem**: Can't verify test catches the bug

**Solution**: Split into 2 commits

### ❌ Mistake 2: Modified Test in Fix Commit
```
Commit 1: Add test
Commit 2: Fix bug and adjust test
```
**Problem**: Test changes hide whether original test would pass

**Solution**: Only modify test in commit 1

### ❌ Mistake 3: Multiple Bugs in One PR
```
Fix #1234 and #1235: Multiple fixes
```
**Problem**: Hard to review, test, and merge independently

**Solution**: Create separate PRs

### ❌ Mistake 4: Extra Commits
```
Commit 1: Add test
Commit 2: Fix bug
Commit 3: Address review
Commit 4: Fix typo
```
**Problem**: Cluttered history

**Solution**: Squash into 2 commits

## Examples

Real examples from our codebase:
- PR #4769: [Fix #4750: Nil pointer panic in RegisterExporters](https://github.com/turbot/steampipe/pull/4769)
- PR #4773: [Fix #4748: SQL injection vulnerability](https://github.com/turbot/steampipe/pull/4773)

## Next Steps

- [GitHub Issues](bug-workflow.md) - Creating bug reports
- [Parallel Coordination](parallel-coordination.md) - Working on multiple bugs in parallel
- [Templates](../templates/) - PR templates
