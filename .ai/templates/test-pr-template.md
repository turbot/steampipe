# Test Suite PR Template

Use this template when creating PRs for comprehensive test suites.

## PR Title

```
Add comprehensive tests for pkg/{package1,package2,package3}
```

Examples:
- `Add comprehensive tests for pkg/{task,snapshot,cmdconfig}`
- `Add quality tests for pkg/export`

## PR Description

```markdown
## Summary

Adds [number] tests covering [list packages], focusing on [quality areas: complex logic, error paths, edge cases, concurrency, etc.].

## Test Overview

### Packages Tested

**pkg/package1 ([X] tests):**
- [Test category 1]: [description]
- [Test category 2]: [description]
- Bug found: #[issue] - [brief description]

**pkg/package2 ([Y] tests):**
- [Test category 1]: [description]
- [Test category 2]: [description]

### Test Categories

- **Bug Hunting**: [X] tests demonstrate bugs (see Issues Found)
- **Concurrency**: [Y] tests for race conditions and goroutine safety
- **Error Handling**: [Z] tests for error paths and edge cases
- **Complex Logic**: [W] tests for multi-branch logic and state machines

## Quality Metrics

- **Tests Added**: [number] total
- **Bug-Demonstrating Tests**: [number] (marked as skipped with issue references)
- **Test Value Score**: [score] (HIGH: [X], MEDIUM: [Y], LOW: [Z])
- **Execution Time**: ~[X]s total
- **Race Conditions Tested**: [X] tests with `-race` flag

## Issues Found

During test generation, the following bugs were discovered and documented:

- #[issue1]: [Brief description] - **[SEVERITY]**
- #[issue2]: [Brief description] - **[SEVERITY]**
- #[issue3]: [Brief description] - **[SEVERITY]**

All bug-demonstrating tests are marked with `t.Skip()` and include issue references. These will be fixed in separate PRs following the 2-commit pattern (unskip test → fix bug).

## Test Structure

### Skipped Tests

[X] tests are skipped because they demonstrate bugs:

```go
t.Skip("Demonstrates bug #[issue] - [description]. Remove this skip in bug fix PR commit 1, then fix in commit 2.")
```

These tests will be unskipped and fixed in individual bug fix PRs.

### Passing Tests

[Y] tests pass and provide coverage for:
- Core functionality
- Edge cases
- Error handling
- Concurrency patterns
- Resource cleanup

## Files Changed

- `pkg/package1/file_test.go` (+[X] lines)
- `pkg/package2/file_test.go` (+[Y] lines)
- `go.mod` (added testify for assertions)

## Test Examples

### High-Value Test Example

[Describe a particularly valuable test that demonstrates complex behavior or catches a subtle bug]

### Concurrency Test Example

[Describe a race condition test if applicable]

## Execution

```bash
# Run all tests
go test ./pkg/package1 ./pkg/package2

# Run with race detector
go test -race ./pkg/package1 ./pkg/package2

# Run in short mode (skips slow tests)
go test -short ./pkg/package1 ./pkg/package2
```

## Next Steps

After this PR merges:
1. Create bug fix PRs for issues #[X], #[Y], #[Z]
2. Each bug fix PR will unskip the relevant test and implement the fix
3. Bug fix PRs can be reviewed and merged independently

## Checklist

- [ ] All tests follow naming conventions
- [ ] Test value score > 2.0 (quality over coverage)
- [ ] Bug-demonstrating tests marked with t.Skip() and issue references
- [ ] GitHub issues created for all discovered bugs
- [ ] Tests organized into logical groups
- [ ] Concurrency tests use error channels (not t.Errorf in goroutines)
- [ ] Resource cleanup with defer where applicable
- [ ] Tests are documented with clear comments
```

## Example

```markdown
## Summary

Adds 205 tests covering pkg/{task,snapshot,cmdconfig,statushooks,introspection,initialisation,ociinstaller}, focusing on complex logic, error paths, edge cases, and concurrency.

## Test Overview

### Packages Tested

**pkg/cmdconfig (18 tests):**
- Snapshot validation (tags, location, args)
- Viper global state thread safety
- Configuration edge cases
- Bugs found: #4756, #4757 (Viper race conditions)

**pkg/statushooks (15 tests):**
- Spinner concurrency and lifecycle
- Progress reporting
- Message deferral and restart logic
- Bugs found: #4743, #4744 (StatusSpinner race conditions)

**pkg/introspection (24 tests):**
- SQL generation for all introspection queries
- Special character handling in SQL
- Parameter validation
- Bug found: #4748 (CRITICAL - SQL injection)

**pkg/initialisation (13 tests):**
- Pipes metadata parsing
- Init data lifecycle
- Database client creation
- Bugs found: #4750, #4767 (nil pointer panics)

**pkg/ociinstaller (135 tests):**
- Database and FDW installation flows
- Image data download and validation
- File move operations and cleanup
- All tests pass

### Test Categories

- **Bug Hunting**: 14 tests demonstrate bugs (see Issues Found)
- **Concurrency**: 12 tests for race conditions
- **Error Handling**: 45 tests for error paths
- **Complex Logic**: 134 tests for multi-branch logic

## Quality Metrics

- **Tests Added**: 205 total
- **Bug-Demonstrating Tests**: 14 (marked as skipped with issue references)
- **Test Value Score**: 2.46 (HIGH: 150, MEDIUM: 55, LOW: 0)
- **Execution Time**: ~5s total
- **Race Conditions Tested**: 12 tests with `-race` flag

## Issues Found

During test generation, the following bugs were discovered and documented:

- #4748: SQL injection vulnerability in GetSetConnectionStateSql - **CRITICAL**
- #4750: Nil pointer panic when registering nil exporter - **HIGH**
- #4767: GetDbClient returns non-nil client on error - **HIGH**
- #4743: Race condition on StatusSpinner.visible field - **MEDIUM**
- #4744: Race condition on spinner.Suffix field - **MEDIUM**
- #4756: Race condition in Viper global state - **MEDIUM**
- #4757: Race condition in exemplarSchemaMap write - **MEDIUM**
- #4768: Goroutine leak when snapshot rows not consumed - **MEDIUM**

All bug-demonstrating tests are marked with `t.Skip()` and include issue references.

## Test Structure

### Skipped Tests

14 tests are skipped because they demonstrate bugs:

```go
t.Skip("Demonstrates bug #4748 - CRITICAL SQL injection vulnerability. Remove this skip in bug fix PR commit 1, then fix in commit 2.")
```

### Passing Tests

191 tests pass and provide coverage for:
- Core functionality across 7 packages
- Edge cases (empty strings, nil values, special characters)
- Error handling (network failures, invalid input)
- Concurrency patterns (safe concurrent access)
- Resource cleanup (goroutines, connections, files)

## Files Changed

- `pkg/cmdconfig/validate_test.go` (+361 lines)
- `pkg/cmdconfig/viper_test.go` (+654 lines)
- `pkg/initialisation/init_data_test.go` (+382 lines)
- `pkg/introspection/introspection_test.go` (+706 lines)
- `pkg/ociinstaller/db_test.go` (+157 lines)
- `pkg/ociinstaller/fdw_test.go` (+115 lines)
- `pkg/snapshot/snapshot_test.go` (+549 lines)
- `pkg/statushooks/statushooks_test.go` (+364 lines)
- `pkg/task/runner_test.go` (+392 lines)
- `pkg/task/version_checker_test.go` (+317 lines)
- `go.mod` (added testify v1.10.0)

Total: +3,997 lines of test code

## Execution

```bash
# Run all new tests
go test ./pkg/cmdconfig ./pkg/statushooks ./pkg/introspection ./pkg/initialisation ./pkg/ociinstaller ./pkg/snapshot ./pkg/task

# Run with race detector
go test -race ./pkg/cmdconfig ./pkg/statushooks

# Run in short mode
go test -short ./pkg/snapshot
```

## Next Steps

After this PR merges:
1. Create bug fix PRs for issues #4743, #4744, #4748, #4750, #4756, #4757, #4767, #4768
2. Each bug fix PR will unskip the relevant test and implement the fix
3. Bug fix PRs can be reviewed and merged independently

## Checklist

- [x] All tests follow naming conventions
- [x] Test value score > 2.0 (2.46 achieved)
- [x] Bug-demonstrating tests marked with t.Skip() and issue references
- [x] GitHub issues created for all discovered bugs
- [x] Tests organized into logical groups
- [x] Concurrency tests use error channels
- [x] Resource cleanup with defer
- [x] Tests documented with clear comments
```
