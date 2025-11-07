# Getting Started with Steampipe Testing Project

## Welcome! 🎯

You're about to add comprehensive tests to Steampipe using a parallel agent coordination system. This guide will get you started.

## Quick Overview

**What we're doing:** Adding high-value unit tests to Steampipe
**How:** Using parallel Claude Code agents, each working on independent tasks
**Structure:** Organized into "waves" (milestones), each with multiple parallel tasks
**Current Wave:** Wave 1 - Foundation (critical paths)

## Prerequisites

### 1. Verify Environment
```bash
cd /Users/nathan/src/steampipe

# Check existing tests pass
go test ./...
# Should see: 6 tests PASS, 2 tests SKIPPED

cd tests/acceptance
./run.sh
# All BATS tests should pass

cd ../..
```

### 2. Understand the Structure
```
.ai/
├── 00-PROJECT-OVERVIEW.md          # ← START HERE - Read first
├── 01-TESTING-STRATEGY.md          # Testing approach
├── GETTING-STARTED.md              # ← This file
│
├── coordination/
│   ├── CURRENT-WAVE.md             # Current wave info
│   ├── NEXT-WAVE-PLAN.md           # Future planning
│   └── BLOCKERS.md                 # Issues tracking
│
├── milestones/
│   └── wave-1-foundation/
│       ├── README.md               # Wave 1 overview
│       ├── STATUS.md               # Progress tracking
│       └── tasks/
│           ├── task-1-test-infrastructure.md
│           ├── task-2-service-tests.md
│           ├── task-3-query-tests.md
│           ├── task-4-connection-tests.md
│           ├── task-5-plugin-manager-tests.md
│           ├── task-6-db-client-tests.md
│           ├── task-7-config-tests.md
│           └── task-8-coverage-ci.md
│
└── reference/
    └── testing-conventions.md      # Coding standards
```

## Wave 1: Foundation - Step by Step

### Step 1: Read the Docs (10 minutes)
```bash
# Essential reading (in order):
cat .ai/00-PROJECT-OVERVIEW.md          # Project context
cat .ai/01-TESTING-STRATEGY.md          # Testing approach
cat .ai/milestones/wave-1-foundation/README.md  # Wave 1 plan
```

### Step 2: Pre-flight Check (5 minutes)
```bash
# Verify tests pass
go test ./...
cd tests/acceptance && ./run.sh && cd ../..

# Create git branch
git checkout -b testing-wave-1

# You're ready!
```

### Step 3: Launch Task 1 (2-3 hours)
**IMPORTANT:** Task 1 MUST complete before others!

```bash
# Terminal 1:
cd /Users/nathan/src/steampipe
claude

# Tell Claude:
# "Please complete the task in .ai/milestones/wave-1-foundation/tasks/task-1-test-infrastructure.md"
```

Wait for Task 1 to finish. It creates the test infrastructure (mocks, helpers) that all other tasks need.

### Step 4: Launch Tasks 2-7 in Parallel (4-5 hours)
Once Task 1 completes, launch 6 agents in parallel:

**Terminal 2: Service Tests**
```bash
claude
# "Complete .ai/milestones/wave-1-foundation/tasks/task-2-service-tests.md"
```

**Terminal 3: Query Tests**
```bash
claude
# "Complete .ai/milestones/wave-1-foundation/tasks/task-3-query-tests.md"
```

**Terminal 4: Connection Tests**
```bash
claude
# "Complete .ai/milestones/wave-1-foundation/tasks/task-4-connection-tests.md"
```

**Terminal 5: Plugin Manager Tests**
```bash
claude
# "Complete .ai/milestones/wave-1-foundation/tasks/task-5-plugin-manager-tests.md"
```

**Terminal 6: DB Client Tests**
```bash
claude
# "Complete .ai/milestones/wave-1-foundation/tasks/task-6-db-client-tests.md"
```

**Terminal 7: Config Tests**
```bash
claude
# "Complete .ai/milestones/wave-1-foundation/tasks/task-7-config-tests.md"
```

Let all 6 agents run in parallel. They work on independent packages so no conflicts.

### Step 5: Launch Task 8 (2 hours)
After Tasks 2-7 complete:

**Terminal 8: Coverage & CI**
```bash
claude
# "Complete .ai/milestones/wave-1-foundation/tasks/task-8-coverage-ci.md"
```

### Step 6: Verify & Commit
```bash
# Run all tests
go test ./...

# Run BATS tests
cd tests/acceptance && ./run.sh && cd ../..

# Check coverage
go test -cover ./...
# Should be 15-20%

# Commit!
git add .
git commit -m "Wave 1: Foundation tests - 15-20% coverage

- Added test infrastructure (mocks, helpers)
- Service lifecycle tests (70% coverage)
- Query execution tests (60% coverage)
- Connection management tests (60% coverage)
- Plugin manager tests (50% coverage)
- Database client tests (60% coverage)
- Configuration tests (60% coverage)
- Coverage reporting in CI/CD

All existing tests still passing.
"
```

## Monitoring Progress

### Check Overall Status
```bash
cat .ai/coordination/CURRENT-WAVE.md
cat .ai/milestones/wave-1-foundation/STATUS.md
```

### Check Individual Task
```bash
# Each agent should update STATUS.md as they work
cat .ai/milestones/wave-1-foundation/STATUS.md
```

### Run Tests Anytime
```bash
# Unit tests
go test ./...

# With coverage
go test -cover ./...

# Specific package
go test ./pkg/db/db_local/

# BATS tests
cd tests/acceptance && ./run.sh && cd ../..
```

## Tips for Success

### 1. Sequential vs Parallel
- **MUST BE SEQUENTIAL:** Task 1 before others, Task 8 after others
- **CAN BE PARALLEL:** Tasks 2-7 (they don't conflict)

### 2. Managing Multiple Agents
- Use separate terminal windows/tabs
- Name terminals (e.g., "Task 2: Service")
- Check STATUS.md to track progress
- Agents are independent - can work at different speeds

### 3. When Things Go Wrong
- **Existing tests break:** STOP. Revert. Document in BLOCKERS.md
- **Coverage too low:** Document why, adjust target
- **Agent stuck:** Check BLOCKERS.md for known issues
- **Conflicts:** Shouldn't happen - tasks work on different files

### 4. Communication
All coordination happens through files:
- Update STATUS.md when task complete
- Document blockers in BLOCKERS.md
- Check CURRENT-WAVE.md for guidance
- Each agent is independent

## After Wave 1

When Wave 1 is complete:
1. Review what worked / what didn't
2. Update approach if needed
3. Move to Wave 2 (or let planning agent prepare it)
4. Rinse and repeat!

## Quick Reference

### Key Commands
```bash
# Run tests
go test ./...
go test -v ./pkg/example/
go test -cover ./...

# BATS tests
cd tests/acceptance && ./run.sh

# Coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Build
go build ./...
```

### Key Files
- Task instructions: `.ai/milestones/wave-1-foundation/tasks/task-*.md`
- Progress tracking: `.ai/milestones/wave-1-foundation/STATUS.md`
- Current wave info: `.ai/coordination/CURRENT-WAVE.md`
- Testing conventions: `.ai/reference/testing-conventions.md`

### Help!
- Read task file carefully - has all instructions
- Check existing tests for patterns
- See `testing-conventions.md` for coding standards
- Document blockers in BLOCKERS.md
- Review BATS tests for test scenarios

## Ready?

1. ✅ Read PROJECT-OVERVIEW.md
2. ✅ Tests pass
3. ✅ Branch created
4. ✅ Understand the flow

**Launch Task 1 and let's go! 🚀**

```bash
claude
# "Complete .ai/milestones/wave-1-foundation/tasks/task-1-test-infrastructure.md"
```
