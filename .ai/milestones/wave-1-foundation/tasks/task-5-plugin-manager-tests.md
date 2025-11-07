# Task 5: Plugin Manager Tests

## Context
Plugin manager coordinates all plugin operations. 24 commits to plugin_manager.go (856 lines!). PM failures prevent all queries.

## Files to Test
- `pkg/pluginmanager_service/plugin_manager.go` (856 lines, 24 changes) - CRITICAL
- `pkg/pluginmanager/lifecycle.go`
- `pkg/pluginmanager_service/plugin_manager_rate_limiters.go` (18 changes)
- `pkg/pluginmanager_service/plugin_manager_plugin_instance.go`

## Test Files to Create
1. `pkg/pluginmanager_service/plugin_manager_test.go`
2. `pkg/pluginmanager/lifecycle_test.go`
3. `pkg/pluginmanager_service/rate_limiter_test.go`

## Critical Test Cases
- Plugin Get requests (from FDW)
- Plugin process spawning
- Reattach config generation
- Connection config distribution
- Rate limiter coordination
- Plugin crash recovery
- Multiple plugin instances
- Plugin shutdown
- gRPC server lifecycle

## Coverage Target
50% of main workflows

## Dependencies
Requires Task 1, parallel with 2-4, 6-7

## Time Estimate
4-5 hours

## Command
```bash
claude
# "Please complete task-5-plugin-manager-tests.md from .ai/milestones/wave-1-foundation/tasks/"
```
