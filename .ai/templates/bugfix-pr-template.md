# Bug Fix PR Template

Use this template when creating PRs to fix bugs discovered during testing.

## PR Title

```
Fix #<issue-number>: Brief description of what was fixed
```

Examples:
- `Fix #4767: GetDbClient error handling`
- `Fix #4748: SQL injection vulnerability in GetSetConnectionStateSql`
- `Fix #4743: Race condition on StatusSpinner.visible field`

## PR Description

```markdown
Closes #<issue-number>

## Summary

[1-2 sentence description of the bug and how it was fixed]

## Problem

[Brief description of what was wrong]

## Solution

[Brief description of how you fixed it]

## Changes

### Commit 1: Demonstrate Bug
- Unskipped/added test `TestName` in `pkg/path/file_test.go`
- Test demonstrates [specific behavior] by [what the test does]
- Test **FAILS** with [error/panic/wrong result]

### Commit 2: Fix Bug
- Modified `pkg/path/file.go` to [what changed]
- Added [nil check / mutex / validation / etc.]
- Test now **PASSES**

## Test Results

### Before Fix (Commit 1)
```bash
$ go test -v -run TestName ./pkg/path
=== RUN   TestName
[Error output showing failure]
--- FAIL: TestName (0.00s)
FAIL
```

### After Fix (Commit 2)
```bash
$ go test -v -run TestName ./pkg/path
=== RUN   TestName
--- PASS: TestName (0.00s)
PASS
```

## Files Changed

- `pkg/path/file.go` - [description of changes]
- `pkg/path/file_test.go` - [unskipped test or added new test]

## Related Issues

Closes #<issue-number>

## Checklist

- [ ] Test added/unskipped in commit 1
- [ ] Test fails in commit 1
- [ ] Pushed commit 1 separately (CI should fail)
- [ ] Fix implemented in commit 2
- [ ] Test passes in commit 2
- [ ] Pushed commit 2 separately (CI should pass)
- [ ] No unrelated changes included
- [ ] Exactly 2 commits in PR
- [ ] PR description starts with "Closes #XXXX"
- [ ] CI history shows: failed run (commit 1) → passed run (commit 2)
```

## Examples

### Example 1: Nil Pointer Fix

```markdown
Closes #4750

## Summary

Fixed a nil pointer panic in `RegisterExporters` by adding nil checks before calling methods on exporter interfaces.

## Problem

The `RegisterExporters` method accepted variadic exporter arguments but didn't validate they were non-nil before passing them to `ExportManager.Register()`, which caused segmentation violations when `Register()` called `exporter.Name()` on a nil interface.

## Solution

Added nil check in the `RegisterExporters` loop to skip nil exporters before they cause panics.

## Changes

### Commit 1: Demonstrate Bug
- Unskipped test `TestInitData_NilExporter` in `pkg/initialisation/init_data_test.go`
- Test passes `nil` to `RegisterExporters()`
- Test **FAILS** with segmentation violation

### Commit 2: Fix Bug
- Modified `pkg/initialisation/init_data.go`
- Added `if e == nil { continue }` check in RegisterExporters loop
- Test now **PASSES**

## Test Results

### Before Fix (Commit 1)
```bash
$ go test -v -run TestInitData_NilExporter ./pkg/initialisation
=== RUN   TestInitData_NilExporter
panic: runtime error: invalid memory address or nil pointer dereference
```

### After Fix (Commit 2)
```bash
$ go test -v -run TestInitData_NilExporter ./pkg/initialisation
=== RUN   TestInitData_NilExporter
--- PASS: TestInitData_NilExporter (0.00s)
PASS
```

## Files Changed

- `pkg/initialisation/init_data.go` - Added nil check in RegisterExporters
- `pkg/initialisation/init_data_test.go` - Unskipped test

## Related Issues

Closes #4750

## Checklist

- [x] Test added/unskipped in commit 1
- [x] Test fails in commit 1
- [x] Pushed commit 1 separately (CI should fail)
- [x] Fix implemented in commit 2
- [x] Test passes in commit 2
- [x] Pushed commit 2 separately (CI should pass)
- [x] No unrelated changes included
- [x] Exactly 2 commits in PR
- [x] PR description starts with "Closes #XXXX"
- [x] CI history shows: failed run (commit 1) → passed run (commit 2)
```

### Example 2: Race Condition Fix

```markdown
Closes #4743

## Summary

Fixed race condition on `StatusSpinner.visible` field by adding mutex protection around all reads and writes.

## Problem

The `visible` field in `StatusSpinner` was accessed concurrently by multiple goroutines without synchronization, causing data races detected by Go's race detector.

## Solution

Added `sync.Mutex` to `StatusSpinner` struct and protected all accesses to the `visible` field with `Lock()/Unlock()`.

## Changes

### Commit 1: Demonstrate Bug
- Unskipped test `TestSpinnerConcurrentShowHide` in `pkg/statushooks/statushooks_test.go`
- Test calls `Show()` and `Hide()` from 10 concurrent goroutines
- Test **FAILS** with race detector warnings

### Commit 2: Fix Bug
- Modified `pkg/statushooks/spinner.go`
- Added `mu sync.Mutex` field to `StatusSpinner` struct
- Protected `visible` field access in `Show()`, `Hide()`, and `UpdateSpinnerMessage()` methods
- Test now **PASSES** with no race conditions

## Test Results

### Before Fix (Commit 1)
```bash
$ go test -race -v -run TestSpinnerConcurrentShowHide ./pkg/statushooks
==================
WARNING: DATA RACE
Write at 0x00c00013d118 by goroutine 10:
  github.com/turbot/steampipe/v2/pkg/statushooks.(*StatusSpinner).Hide()
--- FAIL: TestSpinnerConcurrentShowHide (0.01s)
```

### After Fix (Commit 2)
```bash
$ go test -race -v -run TestSpinnerConcurrentShowHide ./pkg/statushooks
=== RUN   TestSpinnerConcurrentShowHide
--- PASS: TestSpinnerConcurrentShowHide (0.02s)
PASS
```

## Files Changed

- `pkg/statushooks/spinner.go` - Added mutex protection
- `pkg/statushooks/statushooks_test.go` - Unskipped test

## Related Issues

Closes #4743

## Checklist

- [x] Test added/unskipped in commit 1
- [x] Test fails in commit 1 (with -race flag)
- [x] Pushed commit 1 separately (CI should fail)
- [x] Fix implemented in commit 2
- [x] Test passes in commit 2
- [x] Pushed commit 2 separately (CI should pass)
- [x] No unrelated changes included
- [x] Exactly 2 commits in PR
- [x] PR description starts with "Closes #XXXX"
- [x] CI history shows: failed run (commit 1) → passed run (commit 2)
```

## Commit Messages

### Commit 1 Message Format

```
Unskip test demonstrating bug #<issue>: <brief description>
```

or for new tests:

```
Add test for #<issue>: <brief description>
```

### Commit 2 Message Format

```
Fix #<issue>: <brief description of fix>
```

## Tips

1. **Be specific**: Explain exactly what the bug was and how you fixed it
2. **Show test results**: Include actual output from before/after
3. **Keep it minimal**: Don't include unrelated changes
4. **Link the issue**: Always start with "Closes #XXXX"
5. **Verify structure**: Exactly 2 commits, test fails then passes
