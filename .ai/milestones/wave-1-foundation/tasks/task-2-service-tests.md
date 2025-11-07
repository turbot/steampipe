# Task 2: Service Lifecycle Tests

## Context
You are testing the Steampipe service lifecycle - the critical code that starts, stops, and manages the Steampipe database service. This is P0 critical code that has changed 26 times. Service failures prevent ALL Steampipe operations.

## Goal
Create comprehensive unit tests for service lifecycle management in `pkg/db/db_local/`.

## Why This Matters
- Service start/stop is used by every Steampipe operation
- 26 commits to start_services.go = high change frequency = high risk
- Bugs here prevent Steampipe from running at all
- 676 lines in start_services.go need testing

## Files to Test

### Primary Files
1. **`pkg/db/db_local/start_services.go`** (676 lines, 26 changes)
   - StartServices() - Main entry point
   - startPluginManager() - Plugin manager startup
   - Port binding logic
   - SSL certificate management
   - Database initialization

2. **`pkg/db/db_local/stop_services.go`** (363 lines)
   - StopServices() - Graceful shutdown
   - ForceStopServices() - Force shutdown
   - Client connection checking
   - Process cleanup

3. **`pkg/db/db_local/install.go`** (542 lines, 17 changes)
   - InstallDB() - Database installation
   - Version checking
   - Binary extraction
   - Database initialization

## Test Files to Create

### 1. start_services_test.go

```go
package db_local

import (
    "testing"
    "context"
    "github.com/turbot/steampipe/pkg/test/helpers"
)

func TestStartServices_Success(t *testing.T) {
    tests := map[string]struct {
        config map[string]interface{}
        setupFunc func(t *testing.T)
        wantError bool
    }{
        "standard start": {
            config: map[string]interface{}{
                "port": 9193,
            },
            wantError: false,
        },
        "custom port": {
            config: map[string]interface{}{
                "port": 9194,
            },
            wantError: false,
        },
        // Add more cases
    }

    for name, tc := range tests {
        t.Run(name, func(t *testing.T) {
            // Test implementation
        })
    }
}

func TestStartServices_PortConflict(t *testing.T) {
    // Test port already in use scenario
}

func TestStartServices_PluginManagerFailure(t *testing.T) {
    // Test plugin manager start failure
}

func TestStartServices_DatabaseNotInstalled(t *testing.T) {
    // Test behavior when DB not installed
}
```

### 2. stop_services_test.go

```go
package db_local

import "testing"

func TestStopServices_GracefulShutdown(t *testing.T) {
    // Test normal graceful shutdown
}

func TestStopServices_WithConnectedClients(t *testing.T) {
    // Test shutdown with active connections
}

func TestStopServices_ForceShutdown(t *testing.T) {
    // Test force stop
}

func TestStopServices_AlreadyStopped(t *testing.T) {
    // Test stopping already stopped service
}
```

### 3. install_test.go (enhance existing)

```go
package db_local

import "testing"

// Enhance the existing install_test.go

func TestInstallDB_FreshInstall(t *testing.T) {
    // Test fresh installation
}

func TestInstallDB_VersionMismatch(t *testing.T) {
    // Test upgrade scenario
}

func TestInstallDB_CorruptedInstallation(t *testing.T) {
    // Test corrupted DB directory
}
```

## Critical Test Cases

### Service Start
1. ✅ **Happy path** - Clean start with default config
2. ✅ **Custom port** - Start on non-default port
3. ✅ **Port conflict** - Port already in use
4. ✅ **Database not installed** - Should install first
5. ✅ **Plugin manager fails** - Handle PM failure gracefully
6. ✅ **Already running** - Detect already running service
7. ✅ **SSL cert generation** - Test cert creation
8. ✅ **Password management** - Test password setting

### Service Stop
1. ✅ **Graceful shutdown** - Normal stop
2. ✅ **Force shutdown** - Forced stop
3. ✅ **Active connections** - Stop with clients connected
4. ✅ **Already stopped** - Idempotent stop
5. ✅ **Plugin manager cleanup** - Ensure PM stopped
6. ✅ **Port release** - Verify port released

### Installation
1. ✅ **Fresh install** - No existing installation
2. ✅ **Reinstall** - Over existing installation
3. ✅ **Version mismatch** - Upgrade scenario
4. ✅ **Disk space** - Insufficient space handling
5. ✅ **Permissions** - Permission denied scenarios

## Mocking Strategy

### What to Mock
- File system operations (use temp directories)
- PostgreSQL process (mock exec.Command)
- Plugin manager client
- Network operations (port binding)
- Time (for timeouts)

### What to Keep Real
- Configuration parsing
- State management logic
- Error handling
- Validation logic

## Success Criteria

1. **Coverage Target**
   - 70%+ coverage on start_services.go
   - 70%+ coverage on stop_services.go
   - 60%+ coverage on install.go

2. **Test Quality**
   - [ ] All critical paths tested
   - [ ] Both success and error cases
   - [ ] Tests run fast (<100ms each)
   - [ ] Tests are independent
   - [ ] No flaky tests

3. **All Tests Pass**
   ```bash
   go test -v ./pkg/db/db_local/
   ```

4. **Existing Tests Still Pass**
   ```bash
   go test ./...
   cd tests/acceptance && ./run.sh service.bats
   ```

## Testing Your Work

```bash
# Run your new tests
go test -v ./pkg/db/db_local/ -run TestStart
go test -v ./pkg/db/db_local/ -run TestStop
go test -v ./pkg/db/db_local/ -run TestInstall

# Check coverage
go test -cover ./pkg/db/db_local/

# Run all Go tests
go test ./...

# Verify BATS tests still pass
cd tests/acceptance
./run.sh service.bats
```

## Dependencies

- **Requires:** Task 1 (test infrastructure) must be complete
- **Parallel with:** Tasks 3-7

## Estimated Time
4-5 hours

## Notes

- This is P0 CRITICAL - service must work
- Be thorough with error cases
- Look at existing BATS tests in `tests/acceptance/test_files/service.bats` for scenarios
- Start services requires sudo on some systems - consider this in tests
- Plugin manager is a subprocess - may need process mocking

## Report Format

Update `.ai/milestones/wave-1-foundation/STATUS.md`:
```markdown
## Task 2: Service Lifecycle Tests

**Status:** ✅ Complete
**Coverage Achieved:** X%
**Tests Added:** Y tests

### Files Created
- pkg/db/db_local/start_services_test.go (X tests)
- pkg/db/db_local/stop_services_test.go (Y tests)
- pkg/db/db_local/install_test.go (enhanced)

### Test Results
- All new tests passing: ✅
- Existing tests passing: ✅
- Coverage: X%

### Issues Encountered
- [Any issues and how you resolved them]
```

## Command to Run

```bash
# Requires Task 1 complete first!
# In a new terminal:
claude

# Then:
# "Please complete task-2-service-tests.md from .ai/milestones/wave-1-foundation/tasks/"
```
