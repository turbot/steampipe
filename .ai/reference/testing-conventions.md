# Testing Conventions

## File Naming
- Test files: `*_test.go` (alongside source)
- Test data: `testdata/` directory
- Mocks: `pkg/test/mocks/`
- Helpers: `pkg/test/helpers/`

## Test Function Naming

### Unit Tests
```go
func TestFunctionName(t *testing.T)                    // Basic test
func TestFunctionName_EdgeCase(t *testing.T)           // Specific scenario
func TestFunctionName_ErrorCondition(t *testing.T)     // Error case
```

### Integration Tests
```go
//go:build integration
func TestIntegration_FeatureName(t *testing.T)
```

### Benchmarks
```go
func BenchmarkFunctionName(b *testing.B)
```

## Test Structure (Arrange-Act-Assert)

```go
func TestExample(t *testing.T) {
    // Arrange - setup
    input := "test"
    expected := "expected"

    // Act - execute
    result := FunctionToTest(input)

    // Assert - verify
    helpers.AssertEqual(t, expected, result)
}
```

## Table-Driven Tests (Preferred)

```go
func TestExample(t *testing.T) {
    tests := map[string]struct {
        input       string
        expected    string
        wantError   bool
    }{
        "happy path": {
            input:     "input",
            expected:  "output",
            wantError: false,
        },
        "error case": {
            input:     "bad",
            expected:  "",
            wantError: true,
        },
    }

    for name, tc := range tests {
        t.Run(name, func(t *testing.T) {
            result, err := FunctionToTest(tc.input)

            if tc.wantError {
                helpers.AssertError(t, err, "")
            } else {
                helpers.AssertNoError(t, err)
                helpers.AssertEqual(t, tc.expected, result)
            }
        })
    }
}
```

## Test Helpers

### Use t.Helper()
```go
func assertValid(t *testing.T, value string) {
    t.Helper()  // Makes failures point to caller
    if !isValid(value) {
        t.Fatalf("invalid: %s", value)
    }
}
```

### Use t.Cleanup()
```go
func TestExample(t *testing.T) {
    tempDir := createTempDir()
    t.Cleanup(func() {
        os.RemoveAll(tempDir)
    })

    // Test uses tempDir
}
```

## Mocking

### Interface-Based Mocks
```go
type MockClient struct {
    DoSomethingFunc func(string) error
    CallCount       int
}

func (m *MockClient) DoSomething(s string) error {
    m.CallCount++
    if m.DoSomethingFunc != nil {
        return m.DoSomethingFunc(s)
    }
    return nil  // Default behavior
}
```

### Usage
```go
func TestWithMock(t *testing.T) {
    mock := &MockClient{
        DoSomethingFunc: func(s string) error {
            return errors.New("test error")
        },
    }

    // Use mock in test
}
```

## Test Data

### Golden Files
```go
func TestOutput(t *testing.T) {
    result := generateOutput()

    // Compare with golden file
    goldenPath := filepath.Join("testdata", "output.golden.json")
    if *update {
        os.WriteFile(goldenPath, result, 0644)
    }

    expected, _ := os.ReadFile(goldenPath)
    helpers.AssertEqual(t, expected, result)
}
```

### Test Fixtures
```
testdata/
├── config/
│   ├── valid.spc
│   └── invalid.spc
├── golden/
│   ├── output.json
│   └── output.csv
└── fixtures/
    └── test_data.json
```

## Assertions

### Prefer Helpers
```go
// Good
helpers.AssertNoError(t, err)
helpers.AssertEqual(t, expected, actual)

// Avoid
if err != nil {
    t.Fatal(err)
}
if expected != actual {
    t.Fatalf("expected %v, got %v", expected, actual)
}
```

## Test Coverage

### Run with Coverage
```bash
go test -cover ./pkg/example/
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Aim for Quality, Not Just Quantity
- Cover critical paths thoroughly
- Don't test trivial getters/setters
- Focus on logic and error paths

## Test Performance

### Keep Tests Fast
- Unit tests: <1ms each
- Integration tests: <100ms each
- Use mocks to avoid I/O
- Run expensive tests in parallel

### Parallel Tests
```go
func TestParallel(t *testing.T) {
    t.Parallel()  // Run in parallel with other parallel tests

    // Test code
}
```

## Common Patterns

### Testing Errors
```go
func TestError(t *testing.T) {
    _, err := FunctionThatErrors()

    helpers.AssertError(t, err, "expected error message")
}
```

### Testing Panics
```go
func TestPanic(t *testing.T) {
    defer func() {
        if r := recover(); r == nil {
            t.Error("Expected panic")
        }
    }()

    FunctionThatPanics()
}
```

### Testing Interfaces
```go
func TestInterface(t *testing.T) {
    var _ InterfaceName = (*ImplementationType)(nil)  // Compile-time check
}
```

## What NOT to Test

- Third-party library code
- Generated code
- Trivial getters/setters
- Simple constructors
- Constants

## Code Review Checklist

For test code reviews:
- [ ] Tests have clear names
- [ ] Tests are independent
- [ ] Tests are deterministic (no flaky tests)
- [ ] Tests clean up resources
- [ ] Tests use table-driven pattern where appropriate
- [ ] Tests check both success and error paths
- [ ] Tests are fast
- [ ] Mocks are simple and focused
