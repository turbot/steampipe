# Test Generation Guide

Guidelines for writing effective tests.

## Focus on Value

Prioritize tests that:
- Catch real bugs
- Verify complex logic and edge cases
- Test error handling and concurrency
- Cover critical functionality

Avoid simple tests of getters, setters, or trivial constructors.

## Test Generation Process

### 1. Understand the Code

Before writing tests:
- Read the source code thoroughly
- Identify complex logic paths
- Look for error handling code
- Check for concurrency patterns
- Review TODOs and FIXMEs

### 2. Focus Areas

Look for:
- **Nil pointer dereferences** - Missing nil checks
- **Race conditions** - Concurrent access to shared state
- **Resource leaks** - Goroutines, connections, files not cleaned up
- **Edge cases** - Empty strings, zero values, boundary conditions
- **Error handling** - Incorrect error propagation
- **Concurrency issues** - Deadlocks, goroutine leaks
- **Complex logic paths** - Multiple branches, state machines

### 3. Test Structure

```go
func TestFunctionName_Scenario(t *testing.T) {
    // ARRANGE: Set up test conditions

    // ACT: Execute the code under test

    // ASSERT: Verify results

    // CLEANUP: Defer cleanup if needed
}
```

### 4. When You Find a Bug

1. Mark the test with `t.Skip()`
2. Add skip message: `"Demonstrates bug #XXXX - description. Remove skip in bug fix PR."`
3. Create a GitHub issue (see [bug-workflow.md](bug-workflow.md))
4. Continue testing

Example:
```go
func TestResetPools_NilPools(t *testing.T) {
    t.Skip("Demonstrates bug #4698 - ResetPools panics with nil pools. Remove skip in bug fix PR.")

    client := &DbClient{}
    client.ResetPools(context.Background()) // Should not panic
}
```

### 5. Test Organization

#### File Naming
- `*_test.go` in same package as code under test
- Use `<package>_test` for black-box testing

#### Test Naming
- `Test<FunctionName>_<Scenario>`
- Examples:
  - `TestValidateSnapshotTags_EdgeCases`
  - `TestSpinner_ConcurrentShowHide`
  - `TestGetDbClient_WithConnectionString`

#### Subtests
Use `t.Run()` for multiple related scenarios:
```go
func TestValidation_EdgeCases(t *testing.T) {
    tests := []struct {
        name      string
        input     string
        shouldErr bool
    }{
        {"empty_string", "", true},
        {"valid_input", "test", false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := Validate(tt.input)
            if (err != nil) != tt.shouldErr {
                t.Errorf("Validate() error = %v, shouldErr %v", err, tt.shouldErr)
            }
        })
    }
}
```

### 6. Testing Best Practices

#### Concurrency Testing
```go
func TestConcurrent_Operation(t *testing.T) {
    var wg sync.WaitGroup
    errors := make(chan error, 100)

    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            if err := Operation(); err != nil {
                errors <- err
            }
        }()
    }

    wg.Wait()
    close(errors)

    for err := range errors {
        t.Error(err)
    }
}
```

**IMPORTANT**: Don't call `t.Errorf()` from goroutines - it's not thread-safe. Use channels instead.

#### Resource Cleanup
```go
func TestWithResources(t *testing.T) {
    resource := setupResource(t)
    defer resource.Cleanup()

    // ... test code ...
}
```

#### Table-Driven Tests
For multiple similar scenarios:
```go
tests := []struct {
    name     string
    input    string
    expected string
    wantErr  bool
}{
    {"scenario1", "input1", "output1", false},
    {"scenario2", "input2", "output2", false},
    {"error_case", "bad", "", true},
}
```

### 7. What NOT to Test

Avoid LOW-value tests:
- ❌ Simple getters/setters
- ❌ Trivial constructors
- ❌ Tests that just call the function
- ❌ Tests of external libraries
- ❌ Tests that duplicate each other

### 8. Test Output Quality

Tests should provide clear diagnostics on failure:
```go
// Good
t.Errorf("Expected tag validation to fail for %q, but got nil error", invalidTag)

// Bad
t.Error("validation failed")
```

### 9. Performance Considerations

- Use `testing.Short()` for slow tests
- Skip expensive tests in short mode
- Document expected execution time

```go
func TestLargeDataset(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping large dataset test in short mode")
    }
    // ... test code ...
}
```

### 10. Bug Documentation

When a test demonstrates a bug:
- Add clear comments explaining the bug
- Reference the GitHub issue number
- Show expected vs actual behavior
- Include reproduction steps

```go
// BUG: GetDbClient returns non-nil client even when error occurs
// This violates Go conventions and causes nil pointer panics
func TestGetDbClient_ErrorHandling(t *testing.T) {
    t.Skip("Demonstrates bug #4767. Remove skip in fix PR.")

    client, err := GetDbClient("invalid://connection")

    if err != nil {
        // BUG: Client should be nil when error occurs
        if client != nil {
            t.Error("Client should be nil when error is returned")
        }
    }
}
```

## Tools

- `go test -race` - Always run concurrency tests with race detector
- `go test -v` - Verbose output for debugging
- `go test -short` - Skip slow tests
- `go test -run TestName` - Run specific test

## Next Steps

When tests are complete:
1. Create GitHub issues for bugs found
2. Follow [bug-workflow.md](bug-workflow.md) for PR workflow
