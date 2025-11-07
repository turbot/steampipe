# Steampipe Testing Project - .ai Directory

Welcome to the Steampipe testing project coordination directory. All project planning, coordination, and agent instructions live here.

## 📁 Directory Structure

```
.ai/
├── README.md                        ← You are here
├── GETTING-STARTED.md               ← Start here for quick start
├── 00-PROJECT-OVERVIEW.md           ← Project context and goals
├── 01-TESTING-STRATEGY.md           ← Detailed testing strategy
│
├── coordination/                     ← Project coordination
│   ├── CURRENT-WAVE.md              ← Active wave info
│   ├── NEXT-WAVE-PLAN.md            ← Future wave planning
│   └── BLOCKERS.md                  ← Issues tracking
│
├── milestones/                       ← Wave-based milestones
│   ├── wave-1-foundation/           ← Current wave
│   │   ├── README.md                ← Wave overview
│   │   ├── STATUS.md                ← Progress tracking
│   │   └── tasks/                   ← Agent task files
│   │       ├── task-1-test-infrastructure.md
│   │       ├── task-2-service-tests.md
│   │       ├── task-3-query-tests.md
│   │       ├── task-4-connection-tests.md
│   │       ├── task-5-plugin-manager-tests.md
│   │       ├── task-6-db-client-tests.md
│   │       ├── task-7-config-tests.md
│   │       └── task-8-coverage-ci.md
│   │
│   ├── wave-2-core/                 ← Future waves
│   ├── wave-3-integration/
│   └── wave-4-polish/
│
└── reference/                        ← Reference documentation
    └── testing-conventions.md       ← Testing standards
```

## 🚀 Quick Start

### First Time Here?
1. Read `GETTING-STARTED.md` (5 min quick start)
2. Read `00-PROJECT-OVERVIEW.md` (10 min context)
3. Read `01-TESTING-STRATEGY.md` (15 min strategy)
4. Check `coordination/CURRENT-WAVE.md` (current status)
5. Review wave-1 task files (plan understanding)

### Ready to Start?
```bash
# Verify tests pass
go test ./...
cd tests/acceptance && ./run.sh && cd ../..

# Create branch
git checkout -b testing-wave-1

# Launch Task 1
claude
# "Complete .ai/milestones/wave-1-foundation/tasks/task-1-test-infrastructure.md"
```

## 📊 Project Status

**Current Phase:** Ready to start Wave 1
**Coverage:** ~4% → Target: 15-20% (Wave 1)
**Approach:** Parallel agent coordination
**Waves:** 4 planned waves to 60%+ coverage

## 📖 Document Guide

### For Project Understanding
- **00-PROJECT-OVERVIEW.md** - Mission, principles, architecture, status
- **01-TESTING-STRATEGY.md** - Testing approach, priorities, infrastructure
- **GETTING-STARTED.md** - Quick start guide

### For Current Work
- **coordination/CURRENT-WAVE.md** - What's happening now
- **milestones/wave-1-foundation/README.md** - Current wave plan
- **milestones/wave-1-foundation/STATUS.md** - Progress tracking
- **milestones/wave-1-foundation/tasks/task-*.md** - Agent instructions

### For Coordination
- **coordination/BLOCKERS.md** - Issues and blockers
- **coordination/NEXT-WAVE-PLAN.md** - Future planning
- **reference/testing-conventions.md** - Coding standards

## 🎯 Wave 1 Overview

**Goal:** Test critical paths, achieve 15-20% coverage

**Structure:**
1. Task 1: Test Infrastructure (MUST DO FIRST)
2. Tasks 2-7: Core tests (PARALLEL)
3. Task 8: Coverage & CI (DO LAST)

**Execution:**
- Task 1: Sequential (creates foundation)
- Tasks 2-7: Parallel (6 agents simultaneously)
- Task 8: Sequential (integrates coverage)

**Duration:** ~10-15 hours with parallel agents

## 🔥 Critical Paths Being Tested

Wave 1 focuses on high-risk, high-change areas:
1. **Service Lifecycle** - 26 commits, can't break
2. **Query Execution** - 31 commits, core functionality
3. **Connection Management** - 20 commits, plugin coordination
4. **Plugin Manager** - 24 commits, critical infrastructure
5. **Database Client** - 19 commits, all queries go through this
6. **Configuration** - 33 commits, service start dependency

## 📈 Success Metrics

### Wave 1 Targets
- ✅ 15-20% code coverage
- ✅ 70% coverage on critical paths
- ✅ All existing tests still passing
- ✅ Test infrastructure created
- ✅ Coverage reporting enabled

### Overall Project Targets
- 🎯 60%+ code coverage
- 🎯 All critical paths tested
- 🎯 Performance benchmarks
- 🎯 Zero broken tests

## 🛠️ Tools & Commands

### Run Tests
```bash
go test ./...                         # All tests
go test -v ./pkg/example/            # Verbose
go test -cover ./...                 # With coverage
go test -coverprofile=coverage.out ./...  # Coverage file
go tool cover -html=coverage.out     # HTML report
```

### Check Status
```bash
cat .ai/coordination/CURRENT-WAVE.md
cat .ai/milestones/wave-1-foundation/STATUS.md
```

### Run BATS Tests
```bash
cd tests/acceptance && ./run.sh && cd ../..
```

## 💡 Key Principles

1. **DO NOT BREAK** - Existing functionality is sacred
2. **HIGH VALUE FIRST** - Critical paths before edge cases
3. **PARALLEL WORK** - Multiple agents simultaneously
4. **MILESTONE-BASED** - Complete waves, commit, repeat
5. **SIMPLE & CLEAR** - Easy to understand and maintain

## 🤝 How This Works

### Agent Coordination
1. You launch agents in separate terminals
2. Each agent gets a task file with instructions
3. Agents work independently on separate packages
4. Agents update STATUS.md when complete
5. Coordination happens through .ai files

### Parallel Execution
- Task 1 creates infrastructure (sequential)
- Tasks 2-7 work simultaneously (parallel)
- Task 8 integrates everything (sequential)
- Each wave builds on previous waves

### Communication
- **Task files** - Instructions for agents
- **STATUS.md** - Progress tracking
- **BLOCKERS.md** - Issue reporting
- **CURRENT-WAVE.md** - Active work
- **NEXT-WAVE-PLAN.md** - Future planning

## 🎓 Learning Resources

### Testing Patterns
- See `reference/testing-conventions.md`
- Look at existing test files for examples
- Check BATS tests for scenarios

### Steampipe Architecture
- See exploration reports in project overview
- Review critical paths section
- Check change hotspot analysis

## 🚦 Ready to Start?

Follow the quick start in `GETTING-STARTED.md` or jump right in:

```bash
# Read first (recommended)
cat .ai/GETTING-STARTED.md

# Or start immediately
git checkout -b testing-wave-1
claude
# "Complete .ai/milestones/wave-1-foundation/tasks/task-1-test-infrastructure.md"
```

## 📞 Need Help?

- Check `GETTING-STARTED.md` for guidance
- Review `BLOCKERS.md` for known issues
- Read task files carefully - they have everything
- Document new issues in `BLOCKERS.md`

## 🎉 Let's Build This!

You have everything you need:
- ✅ Comprehensive project plan
- ✅ Detailed task instructions
- ✅ Testing strategy
- ✅ Coordination system
- ✅ Success metrics

**Time to add those tests! 🚀**
