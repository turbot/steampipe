# Steampipe Testing Project

## Mission
Add comprehensive, high-value tests to Steampipe without breaking existing functionality.

## Project Status
**Current Phase:** Planning & Setup
**Started:** 2025-11-08
**Target Completion:** TBD based on milestone velocity

## Project Principles

### Core Commitments
1. **DO NOT BREAK** existing Steampipe functionality
2. **RUN EXISTING TESTS** before and after every change
3. **HIGH VALUE FIRST** - Focus on critical paths and change hotspots
4. **SIMPLE & CLEAR** - Tests should be easy to understand and maintain
5. **MILESTONE-BASED** - Each wave includes passing tests before commit

### Working Model
- **Parallel Agents:** Multiple agents work simultaneously on independent tasks
- **Manual Launch:** You launch agents in separate terminals for full control
- **Coordination via .ai/:** All agent communication happens through files
- **Wave-based:** Complete one wave, commit, then move to next

## Steampipe Overview

### What is Steampipe?
Zero-ETL tool that enables SQL queries against live APIs and services:
- SQL access to 140+ data sources via plugins
- Embedded PostgreSQL with custom Foreign Data Wrapper (FDW)
- CLI + service mode
- Plugin-based architecture (gRPC communication)

### Critical Paths (Cannot Break)
1. **Service Lifecycle** - start/stop/restart operations
2. **Query Execution** - SQL query processing and results
3. **Connection Management** - refresh, state, plugin coordination
4. **Plugin System** - install, lifecycle, manager coordination
5. **Database Client** - connection pooling, sessions, search paths

### Change Hotspots (Most Frequent Changes)
Based on 2-year git history analysis:
1. `pkg/query/queryexecute/execute.go` (31 changes)
2. `pkg/db/db_local/start_services.go` (26 changes)
3. `pkg/pluginmanager_service/plugin_manager.go` (24 changes)
4. `pkg/connection/refresh_connections_state.go` (20 changes)
5. `cmd/query.go` (46), `cmd/plugin.go` (31), `cmd/service.go` (22)
6. `pkg/steampipeconfig/` - Config management (33 changes)
7. `pkg/interactive/interactive_client.go` (22 changes)

## Current Test Status

### Strengths ✅
- **160+ BATS acceptance tests** covering all major features
- Mature test infrastructure (BATS framework)
- Good CI/CD integration (GitHub Actions)
- Multi-platform testing (Linux x86_64/ARM64, macOS)
- Real-world scenario coverage

### Critical Gaps ❌
- **Only ~4% unit test coverage** (9 test files for 229 source files)
- No code coverage tracking/reporting
- No performance benchmarks
- Minimal test isolation (integration-heavy)
- Skipped tests with TODOs

### Test Distribution
- **Unit Tests:** 9 files, ~9 test functions (MINIMAL)
- **BATS Tests:** 21 files, 160+ tests, 3,135+ lines (EXTENSIVE)
- **Focus:** End-to-end acceptance testing over unit testing

### Untested Critical Areas
Most packages have NO unit tests:
- `/cmd/*` - All command implementations
- `/pkg/plugin/` - Plugin operations
- `/pkg/pluginmanager/` - Plugin lifecycle
- `/pkg/query/` - Query execution engine
- `/pkg/connection/` - Connection management
- `/pkg/db/db_client/` - Database client
- Many more...

## Testing Strategy

### Approach
**"Critical Paths First, Unit Tests for Stability"**

1. **Phase 1: Safety Net** - Unit tests for critical paths
2. **Phase 2: Change Hotspots** - Test frequently modified code
3. **Phase 3: Integration Gaps** - Fill gaps in integration testing
4. **Phase 4: Coverage & Polish** - Achieve coverage goals

### Test Pyramid (Target)
```
        /\
       /  \      E2E Tests (Existing - Good!)
      /____\
     /      \    Integration Tests (Add targeted tests)
    /________\
   /          \  Unit Tests (MAJOR FOCUS - Build this!)
  /__________  \
```

### Testing Principles
1. **Table-Driven Tests** - Use Go's table-driven pattern
2. **Mock External Dependencies** - Database, plugins, file system
3. **Fast Tests** - Unit tests should run in milliseconds
4. **Clear Names** - Test names describe what they test
5. **Isolation** - Each test is independent
6. **Existing Tests Sacred** - NEVER break existing BATS tests

## Project Organization

### .ai Directory Structure
```
.ai/
├── 00-PROJECT-OVERVIEW.md          # This file - project overview
├── 01-TESTING-STRATEGY.md          # Detailed testing strategy
├── 02-ARCHITECTURE-ANALYSIS.md     # Steampipe architecture deep dive
├── 03-TEST-GAPS-ANALYSIS.md        # Detailed gap analysis
├──
├── milestones/
│   ├── wave-1-foundation/
│   │   ├── README.md               # Wave overview
│   │   ├── tasks/                  # Individual agent tasks
│   │   │   ├── task-1-service-tests.md
│   │   │   ├── task-2-query-tests.md
│   │   │   └── ...
│   │   └── STATUS.md               # Task completion tracking
│   │
│   ├── wave-2-core/
│   ├── wave-3-integration/
│   └── wave-4-polish/
│
├── coordination/
│   ├── CURRENT-WAVE.md             # Points to active wave
│   ├── NEXT-WAVE-PLAN.md           # Coordination agent planning
│   └── BLOCKERS.md                 # Track blockers/issues
│
└── reference/
    ├── test-helpers.md             # Shared test utilities
    ├── mock-patterns.md            # Mocking patterns
    └── conventions.md              # Testing conventions
```

### Agent Communication Pattern
1. **Task Files** - Each task has a markdown file with:
   - Goal
   - Context
   - Specific files to test
   - Success criteria
   - Command to run

2. **Status Files** - Track progress:
   - Task status (todo/in-progress/done)
   - Test coverage achieved
   - Blockers encountered

3. **Coordination** - Planning agent:
   - Monitors current wave progress
   - Plans next wave while current executes
   - Updates NEXT-WAVE-PLAN.md

## Milestone Overview

### Wave 1: Foundation (Safety Net)
**Goal:** Test critical paths that absolutely cannot break
**Coverage Target:** 15-20%
**Estimated Tasks:** 8-10 parallel agents

Focus areas:
1. Service lifecycle (start/stop)
2. Query execution core
3. Connection state management
4. Plugin manager core
5. Database client basics
6. Configuration loading

### Wave 2: Core Functionality
**Goal:** Test main user-facing features
**Coverage Target:** 35-40%
**Estimated Tasks:** 10-12 parallel agents

Focus areas:
1. All CLI commands
2. Plugin operations
3. Interactive console
4. Query result formatting
5. Error handling
6. Connection refresh logic

### Wave 3: Integration & Edge Cases
**Goal:** Fill integration gaps and test edge cases
**Coverage Target:** 50-55%
**Estimated Tasks:** 8-10 parallel agents

Focus areas:
1. Integration test framework
2. Error scenarios
3. Concurrent operations
4. Resource exhaustion
5. Performance benchmarks
6. Mock infrastructure

### Wave 4: Polish & Coverage
**Goal:** Achieve coverage goals and polish
**Coverage Target:** 60%+
**Estimated Tasks:** 6-8 parallel agents

Focus areas:
1. Remaining untested packages
2. Documentation
3. Test utilities
4. CI/CD improvements
5. Coverage reporting
6. Technical debt (skipped tests)

## Success Criteria

### Per Milestone
- ✅ All new tests pass
- ✅ All existing BATS tests still pass
- ✅ Code builds successfully
- ✅ No regressions in functionality
- ✅ Coverage target achieved
- ✅ Changes committed to git

### Overall Project
- 🎯 **60%+ unit test coverage**
- 🎯 **All critical paths have unit tests**
- 🎯 **Change hotspots have comprehensive tests**
- 🎯 **Code coverage tracking enabled**
- 🎯 **Performance benchmarks established**
- 🎯 **Zero broken existing tests**
- 🎯 **Clear testing conventions documented**

## Risk Mitigation

### Safety Measures
1. **Pre-flight Check:** Run existing tests before starting any wave
2. **Continuous Testing:** Run tests after every agent completes
3. **Git Discipline:** Commit only when all tests pass
4. **Rollback Plan:** Git revert if something breaks
5. **Parallel Safety:** Agents work on independent packages

### Monitoring
- Track test execution times
- Monitor for flaky tests
- Watch for CI/CD failures
- Review coverage trends

## Next Steps
See `CURRENT-WAVE.md` for active work and `milestones/wave-1-foundation/` for first wave details.
