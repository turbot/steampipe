# Current Wave: Wave 1 - Foundation

**Status:** Ready to Start
**Started:** Not yet
**Expected Completion:** TBD

## Quick Links
- [Wave 1 Overview](../milestones/wave-1-foundation/README.md)
- [Wave 1 Status](../milestones/wave-1-foundation/STATUS.md)
- [Task Files](../milestones/wave-1-foundation/tasks/)

## Current Focus
Wave 1 is establishing the testing foundation with focus on critical paths.

## Goal
Test critical paths that absolutely cannot break, achieve 15-20% coverage.

## Active Tasks
See [STATUS.md](../milestones/wave-1-foundation/STATUS.md) for current task status.

## Quick Start Guide

### Step 1: Pre-flight Check
```bash
# Verify existing tests pass
cd /Users/nathan/src/steampipe
go test ./...
cd tests/acceptance && ./run.sh

# Create branch
git checkout -b testing-wave-1
```

### Step 2: Launch Task 1 (REQUIRED FIRST)
```bash
# In Terminal 1:
cd /Users/nathan/src/steampipe
claude

# Then tell Claude:
# "Please complete the task described in .ai/milestones/wave-1-foundation/tasks/task-1-test-infrastructure.md"
```

### Step 3: Wait for Task 1 Completion
Task 1 creates the test infrastructure needed by all other tasks. MUST finish before launching others.

### Step 4: Launch Tasks 2-7 in Parallel
Once Task 1 is complete, open 6 terminals and launch:

**Terminal 2:**
```bash
claude
# "Please complete .ai/milestones/wave-1-foundation/tasks/task-2-service-tests.md"
```

**Terminal 3:**
```bash
claude
# "Please complete .ai/milestones/wave-1-foundation/tasks/task-3-query-tests.md"
```

**Terminal 4:**
```bash
claude
# "Please complete .ai/milestones/wave-1-foundation/tasks/task-4-connection-tests.md"
```

**Terminal 5:**
```bash
claude
# "Please complete .ai/milestones/wave-1-foundation/tasks/task-5-plugin-manager-tests.md"
```

**Terminal 6:**
```bash
claude
# "Please complete .ai/milestones/wave-1-foundation/tasks/task-6-db-client-tests.md"
```

**Terminal 7:**
```bash
claude
# "Please complete .ai/milestones/wave-1-foundation/tasks/task-7-config-tests.md"
```

### Step 5: Launch Task 8 (After Tasks 2-7)
Once all tests are passing:

**Terminal 8:**
```bash
claude
# "Please complete .ai/milestones/wave-1-foundation/tasks/task-8-coverage-ci.md"
```

### Step 6: Wave Complete!
```bash
# Verify everything
go test ./...
cd tests/acceptance && ./run.sh

# Check coverage
go test -cover ./...

# Commit
git add .
git commit -m "Wave 1: Foundation tests - coverage 15-20%"
```

## Monitoring Progress

Check status anytime:
```bash
cat .ai/milestones/wave-1-foundation/STATUS.md
```

## Need Help?

- See [Project Overview](../00-PROJECT-OVERVIEW.md) for overall context
- See [Testing Strategy](../01-TESTING-STRATEGY.md) for testing approach
- See individual task files for detailed instructions
- Check STATUS.md for current progress
