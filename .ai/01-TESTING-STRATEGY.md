# Steampipe Testing Strategy

## Overview
This document defines the comprehensive testing strategy for adding high-value tests to Steampipe.

## Testing Philosophy

### Guiding Principles
1. **Safety First** - Never break existing functionality
2. **Value-Driven** - Test critical paths before edge cases
3. **Maintainable** - Write clear, simple tests
4. **Fast Feedback** - Unit tests run in milliseconds
5. **Comprehensive** - Cover units, integration, and E2E

### Test Pyramid Strategy

```
Current State:              Target State:
    /\                          /\
   /E2\  Good (160+)           /E2\  Maintain
  /____\                      /____\
 /      \  Weak               /  I  \  Add
/________\                   /______\
/  Unit  \  CRITICAL GAP    /  Unit  \  BUILD!
/__4%____\                 /__60%+___\
```

## Testing Layers

### 1. Unit Tests (PRIMARY FOCUS)
**Current:** ~4% coverage (9 test files)
**Target:** 60%+ coverage

**What to Test:**
- Individual functions and methods
- Business logic isolation
- Error handling paths
- Edge cases and boundaries
- State transitions

**How to Test:**
- Table-driven tests
- Mock external dependencies
- Fast execution (<1ms per test)
- No network/filesystem I/O
- Isolated test setup

**Example Pattern:**
```go
func TestConnectionStateTransition(t *testing.T) {
    tests := map[string]struct {
        initialState string
        event        string
        expectedState string
        expectError  bool
    }{
        "pending to ready": {
            initialState: "pending",
            event: "initialized",
            expectedState: "ready",
            expectError: false,
        },
        // ... more cases
    }

    for name, tc := range tests {
        t.Run(name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### 2. Integration Tests (SELECTIVE)
**Current:** Minimal
**Target:** Key integration points

**What to Test:**
- Component interactions
- Database operations (with test DB)
- File system operations (with temp dirs)
- Plugin communication
- gRPC services

**How to Test:**
- Real dependencies in isolated environment
- Test databases/temp directories
- Cleanup after each test
- Moderate execution time (<100ms)

### 3. Acceptance Tests (MAINTAIN)
**Current:** 160+ BATS tests (Excellent!)
**Target:** Maintain + add targeted tests

**What to Test:**
- End-to-end workflows
- Real-world scenarios
- CLI commands
- User-facing features

**How to Test:**
- BATS framework (existing)
- Real Steampipe installation
- Test plugins (chaos)
- Full execution time (seconds)

### 4. Performance Benchmarks (NEW)
**Current:** None
**Target:** Benchmarks for critical paths

**What to Benchmark:**
- Query execution time
- Service startup time
- Plugin loading time
- Connection initialization
- Large result set handling

**How to Benchmark:**
```go
func BenchmarkQueryExecution(b *testing.B) {
    // Setup
    for i := 0; i < b.N; i++ {
        // Benchmark code
    }
}
```

## Priority Matrix

### High Priority (Wave 1 & 2)
Files that are both **critical** AND **change frequently**:

| File | Changes | Risk | Priority |
|------|---------|------|----------|
| `pkg/query/queryexecute/execute.go` | 31 | Critical | 🔴 P0 |
| `pkg/db/db_local/start_services.go` | 26 | Critical | 🔴 P0 |
| `pkg/pluginmanager_service/plugin_manager.go` | 24 | Critical | 🔴 P0 |
| `pkg/connection/refresh_connections_state.go` | 20 | Critical | 🔴 P0 |
| `pkg/db/db_client/db_client_execute.go` | 19 | Critical | 🔴 P0 |
| `cmd/query.go` | 46 | High | 🟠 P1 |
| `cmd/service.go` | 22 | High | 🟠 P1 |
| `cmd/plugin.go` | 31 | High | 🟠 P1 |
| `pkg/steampipeconfig/load_config.go` | 33 | High | 🟠 P1 |
| `pkg/interactive/interactive_client.go` | 22 | Medium | 🟡 P2 |

### Medium Priority (Wave 3)
Important but less frequently changed:
- Database client operations
- Configuration management
- Plugin installation
- Error handling utilities
- Display/formatting

### Lower Priority (Wave 4)
Utilities and less critical paths:
- File path management
- Version checking
- Display helpers
- Task management

## Test Coverage Strategy

### Phase 1: Foundation (15-20%)
**Focus:** Critical paths that absolutely cannot break

Packages to test:
1. **pkg/db/db_local/** - Service lifecycle
   - `start_services.go` - Service startup
   - `stop_services.go` - Service shutdown
   - `install.go` - Database installation
   - Target: 70% of critical functions

2. **pkg/query/queryexecute/** - Query execution
   - `execute.go` - Query execution logic
   - Target: 60% of core paths

3. **pkg/connection/** - Connection management
   - `refresh_connections_state.go` - State refresh
   - Target: 60% of state transitions

4. **pkg/pluginmanager_service/** - Plugin manager
   - `plugin_manager.go` - Core manager
   - Target: 50% of main workflows

5. **pkg/db/db_client/** - Database client
   - `db_client_connect.go` - Connection
   - `db_client_execute.go` - Execution
   - Target: 60% of critical paths

### Phase 2: Core Features (35-40%)
**Focus:** Main user-facing features

Additional packages:
1. **cmd/** - All CLI commands
   - `query.go`, `service.go`, `plugin.go`
   - Target: 50% coverage

2. **pkg/steampipeconfig/** - Configuration
   - `load_config.go`, `connection_updates.go`
   - Target: 60% coverage

3. **pkg/plugin/** - Plugin operations
   - `install.go`, `plugin_remove.go`
   - Target: 60% coverage

4. **pkg/interactive/** - Interactive console
   - `interactive_client.go`
   - Target: 40% coverage (complex UI)

### Phase 3: Integration & Edge Cases (50-55%)
**Focus:** Integration points and error scenarios

1. Integration test framework
2. Error handling paths
3. Concurrent operations
4. Resource limits
5. Edge cases

### Phase 4: Polish (60%+)
**Focus:** Remaining gaps

1. Utility packages
2. Helper functions
3. Display formatting
4. Documentation

## Test Infrastructure

### Test Helpers (To Build)

#### 1. Mock Database Client
```go
type MockDBClient struct {
    ExecuteFunc func(context.Context, string) error
    ConnectFunc func() error
}
```

#### 2. Mock Plugin Manager
```go
type MockPluginManager struct {
    GetPluginFunc func(string) (*Plugin, error)
    StartFunc func() error
}
```

#### 3. Test Fixtures
```go
func NewTestConfig() *SteampipeConfig
func NewTestConnection(name string) *Connection
func CreateTempDatabase(t *testing.T) string
```

#### 4. Assertion Helpers
```go
func AssertNoError(t *testing.T, err error)
func AssertEqual(t *testing.T, expected, actual interface{})
func AssertContains(t *testing.T, haystack, needle string)
```

### Test Data Management

#### Golden Files Pattern
For complex output comparison:
```
testdata/
├── query_output.golden.json
├── plugin_list.golden.txt
└── connection_state.golden.json
```

#### Test Database
For integration tests:
- Use Docker or embedded PostgreSQL test instance
- Fixtures for schema setup
- Cleanup after each test

#### Test Config Files
```
testdata/
├── config/
│   ├── valid_config.spc
│   ├── invalid_config.spc
│   └── complex_aggregator.spc
```

## Mocking Strategy

### What to Mock

#### External Dependencies
1. **Database connections** - Use mock client
2. **File system** - Use afero or temp directories
3. **Network calls** - Mock HTTP clients
4. **Plugin processes** - Mock plugin manager
5. **gRPC services** - Mock gRPC clients

#### Internal Dependencies
1. **Configuration** - Inject test configs
2. **State** - Inject test state
3. **Timers** - Use fake time
4. **Random** - Use seeded random

### What NOT to Mock

#### Keep Real
1. **Pure functions** - Test actual logic
2. **Data structures** - Use real structs
3. **Simple utilities** - Test real implementation
4. **Validation** - Test actual validators

## Test Organization

### File Structure
```
pkg/
├── db/
│   ├── db_local/
│   │   ├── start_services.go
│   │   ├── start_services_test.go      # Unit tests
│   │   ├── stop_services.go
│   │   ├── stop_services_test.go
│   │   └── testdata/                    # Test fixtures
│   │       └── test_config.json
```

### Test Naming Convention
```go
// Unit tests
func TestFunctionName(t *testing.T) {}
func TestFunctionName_ErrorCase(t *testing.T) {}

// Integration tests
func TestIntegration_FeatureName(t *testing.T) {}

// Benchmarks
func BenchmarkFunctionName(b *testing.B) {}
```

### Test Tags (Build Tags)
```go
//go:build integration
// +build integration

package db_test
```

Usage:
```bash
# Run only unit tests (default)
go test ./...

# Run integration tests
go test -tags=integration ./...
```

## CI/CD Integration

### Current State
- Go unit tests run on PR (quick)
- BATS acceptance tests run in parallel (21 jobs)
- No coverage reporting

### Target State

#### 1. Coverage Reporting
```yaml
- name: Test with coverage
  run: |
    go test -coverprofile=coverage.out -covermode=atomic ./...
    go tool cover -html=coverage.out -o coverage.html

- name: Upload coverage
  uses: codecov/codecov-action@v3
  with:
    files: ./coverage.out
```

#### 2. Coverage Gates
```yaml
- name: Check coverage
  run: |
    coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
    if (( $(echo "$coverage < 60" | bc -l) )); then
      echo "Coverage $coverage% is below 60%"
      exit 1
    fi
```

#### 3. Test Categorization
```yaml
- name: Unit tests
  run: go test -short ./...

- name: Integration tests
  run: go test -tags=integration ./...

- name: Benchmarks
  run: go test -bench=. -run=^$ ./...
```

## Quality Standards

### Test Quality Checklist
- [ ] Tests are deterministic (no flaky tests)
- [ ] Tests are independent (can run in any order)
- [ ] Tests are fast (<1ms for unit, <100ms for integration)
- [ ] Tests have clear names describing what they test
- [ ] Tests use table-driven pattern where applicable
- [ ] Tests clean up resources (defer cleanup)
- [ ] Tests check both success and error paths
- [ ] Tests have meaningful assertions
- [ ] Tests don't rely on external state

### Code Review Standards
- New code must have tests
- Changes to critical paths require new tests
- Test coverage should not decrease
- All tests must pass before merge

## Success Metrics

### Per Wave
- Coverage increase by target %
- All new tests pass
- All existing tests still pass
- No increase in test execution time (unit tests)

### Overall Project
- 60%+ overall coverage
- 80%+ coverage on critical paths
- <5s for all unit tests
- Benchmarks for key operations
- Zero skipped tests

## Risk Mitigation

### Testing the Tests
1. **Mutation Testing** - Verify tests catch bugs
2. **Coverage Analysis** - Identify untested code
3. **Flaky Test Detection** - Run tests 10x to find flakes
4. **Performance Monitoring** - Track test execution time

### Rollback Strategy
1. Git commit after each passing wave
2. Keep old tests until new tests proven
3. Feature flags for new test infrastructure
4. Gradual rollout of mocking

## Next Steps
1. Set up test helpers and mocks (Wave 1)
2. Create example tests demonstrating patterns
3. Build test infrastructure
4. Begin Wave 1 testing

See `milestones/wave-1-foundation/` for detailed Wave 1 plan.
