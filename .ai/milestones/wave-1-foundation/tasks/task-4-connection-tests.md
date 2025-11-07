# Task 4: Connection Management Tests

## Context
Connection management is critical - 20 commits to refresh_connections_state.go (915 lines!). Connection failures break plugin functionality.

## Files to Test
- `pkg/connection/refresh_connections_state.go` (915 lines, 20 changes) - CRITICAL
- `pkg/steampipeconfig/connection_updates.go` (544 lines)
- `pkg/steampipeconfig/connection_state.go`
- `pkg/steampipeconfig/connection_state_map.go` (14 changes)

## Test Files to Create
1. `pkg/connection/refresh_connections_state_test.go`
2. `pkg/steampipeconfig/connection_updates_test.go`
3. `pkg/steampipeconfig/connection_state_test.go`

## Critical Test Cases
- State transitions: pending → ready → error
- Config change detection
- Schema refresh and cloning
- Connection state synchronization
- Rate limiter management
- Concurrent state updates
- Connection addition/removal
- Plugin connection coordination

## Coverage Target
60% on critical paths

## Dependencies
Requires Task 1 complete, parallel with Tasks 2, 3, 5-7

## Time Estimate
4-5 hours

## Command
```bash
claude
# "Please complete task-4-connection-tests.md from .ai/milestones/wave-1-foundation/tasks/"
```
