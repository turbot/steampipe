# Wave 1: Foundation (Safety Net)

## Overview
**Goal:** Test critical paths that absolutely cannot break
**Coverage Target:** 15-20% overall, 70%+ on critical paths
**Timeline:** First milestone - establishes testing foundation

## Success Criteria
- ✅ All critical path entry points have unit tests
- ✅ Service lifecycle thoroughly tested
- ✅ Query execution core paths tested
- ✅ Connection state management tested
- ✅ Plugin manager core tested
- ✅ All existing BATS tests still pass
- ✅ Test helpers and mocking infrastructure created
- ✅ Coverage reporting enabled

## Focus Areas

### 1. Service Lifecycle (CRITICAL - P0)
**Risk:** Service failures prevent all Steampipe operations
**Change Frequency:** 26 commits to start_services.go

Files to test:
- `pkg/db/db_local/start_services.go` (676 lines)
- `pkg/db/db_local/stop_services.go` (363 lines)
- `pkg/db/db_local/install.go` (542 lines)
- `pkg/pluginmanager/lifecycle.go`

Critical paths:
- Service start with various configurations
- Graceful shutdown
- Force shutdown
- Database installation
- Port binding conflicts
- Plugin manager startup/shutdown
- Error recovery

### 2. Query Execution (CRITICAL - P0)
**Risk:** Query failures are core functionality failures
**Change Frequency:** 31 commits to execute.go, 46 to query.go

Files to test:
- `pkg/query/queryexecute/execute.go`
- `pkg/query/init_data.go`
- `pkg/db/db_client/db_client_execute.go` (19 changes)
- `cmd/query.go` (46 changes)

Critical paths:
- Query parsing
- Batch execution
- Interactive execution
- Result streaming
- Error handling
- Timeout handling
- Transaction management

### 3. Connection Management (CRITICAL - P0)
**Risk:** Connection failures break plugin functionality
**Change Frequency:** 20 commits to refresh_connections_state.go

Files to test:
- `pkg/connection/refresh_connections_state.go` (915 lines!)
- `pkg/steampipeconfig/connection_updates.go` (544 lines)
- `pkg/steampipeconfig/connection_state.go`

Critical paths:
- Connection state transitions (pending → ready → error)
- Config change detection
- Schema refresh/clone
- Connection state synchronization
- Rate limiter management

### 4. Plugin Manager (CRITICAL - P0)
**Risk:** Plugin manager failures prevent all queries
**Change Frequency:** 24 commits to plugin_manager.go

Files to test:
- `pkg/pluginmanager_service/plugin_manager.go` (856 lines!)
- `pkg/pluginmanager/lifecycle.go`
- `pkg/pluginmanager_service/plugin_manager_rate_limiters.go` (18 changes)

Critical paths:
- Plugin Get requests (from FDW)
- Plugin process spawning
- Reattach config generation
- Connection config distribution
- Rate limiter coordination
- Plugin crash handling

### 5. Database Client (CRITICAL - P0)
**Risk:** Client failures break all database operations
**Change Frequency:** 19 commits to db_client_execute.go

Files to test:
- `pkg/db/db_client/db_client.go` (316 lines)
- `pkg/db/db_client/db_client_connect.go`
- `pkg/db/db_client/db_client_execute.go`
- `pkg/db/db_client/db_client_search_path.go`

Critical paths:
- Connection pool management
- User pool vs management pool
- Search path setup
- Query execution with retry
- Session management
- Connection acquisition/release

### 6. Configuration Loading (HIGH - P1)
**Risk:** Config failures prevent service start
**Change Frequency:** 33 commits to load_config.go, steampipeconfig.go

Files to test:
- `pkg/steampipeconfig/load_config.go` (388 lines, 33 changes)
- `pkg/steampipeconfig/steampipeconfig.go` (362 lines, 33 changes)
- `pkg/steampipeconfig/connection_plugin.go` (16 changes)

Critical paths:
- HCL config parsing
- Connection config loading
- Plugin config resolution
- Workspace profiles
- Config validation
- Config file watching

## Test Infrastructure to Build

### 1. Mock Components
Create in `pkg/test/mocks/`:

```go
// MockDBClient - Mock database client
type MockDBClient struct {
    ExecuteFunc func(ctx context.Context, sql string) (*sql.Rows, error)
    ConnectFunc func() error
    CloseFunc func() error
}

// MockPluginManager - Mock plugin manager
type MockPluginManager struct {
    GetPluginFunc func(connectionName string) (*PluginInstance, error)
    StartFunc func() error
    StopFunc func() error
}

// MockConnection - Mock connection
type MockConnection struct {
    NameField string
    StateField string
}
```

### 2. Test Helpers
Create in `pkg/test/helpers/`:

```go
// Database helpers
func CreateTestDatabase(t *testing.T) (*sql.DB, func())
func CreateTempDir(t *testing.T) string

// Config helpers
func NewTestConfig() *SteampipeConfig
func LoadTestConfig(filename string) (*SteampipeConfig, error)

// Assertion helpers
func AssertNoError(t *testing.T, err error)
func AssertError(t *testing.T, err error, expectedMsg string)
func AssertEqual(t *testing.T, expected, actual interface{})
```

### 3. Test Fixtures
Create in testdata directories:

- Sample connection configs (.spc)
- Sample plugin configs
- Expected query results (golden files)
- Error message fixtures

### 4. Coverage Infrastructure
Add to CI/CD:

- Coverage report generation
- Coverage upload to Codecov
- Coverage badge
- Per-package coverage tracking

## Task Breakdown

### Task 1: Test Infrastructure Setup
**Agent Focus:** Build testing foundation
**Files to Create:**
- `pkg/test/mocks/db_client.go`
- `pkg/test/mocks/plugin_manager.go`
- `pkg/test/helpers/database.go`
- `pkg/test/helpers/config.go`
- `pkg/test/helpers/assertions.go`

**Success Criteria:**
- Mocks compile and are usable
- Helpers work in example tests
- Documentation for using test infrastructure

### Task 2: Service Lifecycle Tests
**Agent Focus:** pkg/db/db_local/
**Files to Create:**
- `pkg/db/db_local/start_services_test.go`
- `pkg/db/db_local/stop_services_test.go`
- `pkg/db/db_local/install_test.go` (enhance existing)

**Coverage Target:** 70% of critical functions

### Task 3: Query Execution Tests
**Agent Focus:** pkg/query/queryexecute/
**Files to Create:**
- `pkg/query/queryexecute/execute_test.go`
- `pkg/query/init_data_test.go`

**Coverage Target:** 60% of core paths

### Task 4: Connection Management Tests
**Agent Focus:** pkg/connection/
**Files to Create:**
- `pkg/connection/refresh_connections_state_test.go`
- `pkg/steampipeconfig/connection_updates_test.go`
- `pkg/steampipeconfig/connection_state_test.go`

**Coverage Target:** 60% of state transitions

### Task 5: Plugin Manager Tests
**Agent Focus:** pkg/pluginmanager_service/
**Files to Create:**
- `pkg/pluginmanager_service/plugin_manager_test.go`
- `pkg/pluginmanager/lifecycle_test.go`

**Coverage Target:** 50% of main workflows

### Task 6: Database Client Tests
**Agent Focus:** pkg/db/db_client/
**Files to Create:**
- `pkg/db/db_client/db_client_test.go`
- `pkg/db/db_client/db_client_connect_test.go`
- `pkg/db/db_client/db_client_execute_test.go`

**Coverage Target:** 60% of critical paths

### Task 7: Configuration Loading Tests
**Agent Focus:** pkg/steampipeconfig/
**Files to Create:**
- `pkg/steampipeconfig/load_config_test.go` (fix existing skipped test)
- `pkg/steampipeconfig/steampipeconfig_test.go`

**Coverage Target:** 60% coverage

### Task 8: Coverage & CI Integration
**Agent Focus:** CI/CD pipeline
**Files to Modify:**
- `.github/workflows/11-test-acceptance.yaml`
- Add coverage reporting
- Add coverage gates

**Success Criteria:**
- Coverage reports in CI
- Coverage badge in README
- Failing build if coverage drops

## Parallel Execution Plan

### Phase 1: Foundation (Sequential)
1. **Task 1: Test Infrastructure** (MUST complete first)
   - Creates mocks and helpers needed by all other tasks
   - Estimated time: 2-3 hours

### Phase 2: Core Testing (Parallel - 6 agents)
After Task 1 completes, launch in parallel:
2. **Task 2: Service Lifecycle** ⟷ parallel
3. **Task 3: Query Execution** ⟷ parallel
4. **Task 4: Connection Management** ⟷ parallel
5. **Task 5: Plugin Manager** ⟷ parallel
6. **Task 6: Database Client** ⟷ parallel
7. **Task 7: Configuration Loading** ⟷ parallel

Each agent works independently on separate packages.

### Phase 3: Integration (Sequential)
8. **Task 8: Coverage & CI** (After all tests pass)
   - Integrates coverage reporting
   - Validates all tests work in CI

## Agent Instruction Files

Each task has a detailed instruction file in `tasks/` directory:
- `task-1-test-infrastructure.md` - Setup mocks and helpers
- `task-2-service-tests.md` - Service lifecycle tests
- `task-3-query-tests.md` - Query execution tests
- `task-4-connection-tests.md` - Connection management tests
- `task-5-plugin-manager-tests.md` - Plugin manager tests
- `task-6-db-client-tests.md` - Database client tests
- `task-7-config-tests.md` - Configuration loading tests
- `task-8-coverage-ci.md` - Coverage and CI integration

## Pre-flight Checklist

Before starting Wave 1:
- [ ] Run existing tests: `go test ./...`
- [ ] Run existing BATS tests: `cd tests/acceptance && ./run.sh`
- [ ] Verify all pass
- [ ] Create git branch: `git checkout -b testing-wave-1`
- [ ] Review task files
- [ ] Launch Task 1 agent first

## Post-wave Checklist

After completing Wave 1:
- [ ] All new tests pass
- [ ] All existing tests still pass
- [ ] Coverage ≥ 15%
- [ ] CI/CD updated with coverage
- [ ] Documentation updated
- [ ] Commit changes: `git commit -m "Wave 1: Foundation tests"`
- [ ] Create PR or merge to develop

## Monitoring

### Track These Metrics
- Coverage percentage (target: 15-20%)
- Test execution time (should stay fast)
- Number of tests added
- Number of critical paths covered
- CI/CD success rate

### Blockers to Watch
- Difficulty mocking external dependencies
- Slow test execution
- Flaky tests
- CI/CD failures
- Existing test breakage

## Next Steps
1. Review task instruction files in `tasks/` directory
2. Complete pre-flight checklist
3. Launch Task 1 agent to build test infrastructure
4. Wait for Task 1 completion
5. Launch Tasks 2-7 in parallel
6. Launch Task 8 after all tests pass
7. Complete post-wave checklist
8. Move to Wave 2!
