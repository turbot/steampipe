# Next Wave Planning

**Status:** Wave 1 not yet complete
**Planning Agent:** Not yet assigned

## Purpose
This file is maintained by a coordination agent that plans the next wave while the current wave is in progress. This allows seamless transition between waves.

## Wave 2: Core Functionality (PLANNING)

**Estimated Start:** After Wave 1 complete
**Coverage Target:** 35-40% (up from 15-20%)

### Preliminary Scope
Based on the testing strategy, Wave 2 will focus on:

1. **CLI Commands** (cmd/)
   - All command implementations
   - Argument parsing
   - Subcommand coordination

2. **Plugin Operations** (pkg/plugin/)
   - Plugin install/uninstall
   - Plugin list/update
   - Plugin configuration

3. **Interactive Console** (pkg/interactive/)
   - Interactive client
   - Autocomplete
   - Metaqueries
   - Command history

4. **Result Formatting** (pkg/query/queryresult/)
   - All output formats
   - Export functionality
   - Snapshot handling

5. **Error Handling** (pkg/error_helpers/)
   - Error utilities
   - Postgres error handling
   - Diagnostic messages

### Estimated Tasks
10-12 parallel agent tasks

### Dependencies
- Wave 1 must be complete
- Test infrastructure from Task 1
- Coverage reporting from Task 8

## Planning Process

When Wave 1 is 75% complete (Tasks 1-6 done), a coordination agent should:

1. Review Wave 1 results and lessons learned
2. Identify any gaps or issues
3. Create detailed Wave 2 task breakdown
4. Create Wave 2 task instruction files
5. Update this file with detailed plan
6. Prepare Wave 2 launch for when Wave 1 completes

## Coordination Agent Instructions

When you're assigned to plan Wave 2:

```
You are the Wave 2 planning agent. Your job is to prepare the next wave while Wave 1 is completing.

Review:
1. Wave 1 status and results
2. Testing strategy document
3. Change hotspots analysis
4. Coverage gaps after Wave 1

Create:
1. Wave 2 milestone directory structure
2. Wave 2 README.md
3. Wave 2 STATUS.md
4. Wave 2 task files (10-12 tasks)

Update:
1. This file with detailed Wave 2 plan
2. Timeline estimates
3. Resource requirements

Your work should be ready so that when Wave 1 completes, we can immediately launch Wave 2.
```

## Status
⏸️ Waiting for Wave 1 to reach 75% completion before planning begins.
