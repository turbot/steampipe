# Task 7: Configuration Loading Tests

## Context
Configuration management has 33 commits each to load_config.go and steampipeconfig.go. Config errors prevent service start.

## Files to Test
- `pkg/steampipeconfig/load_config.go` (388 lines, 33 changes)
- `pkg/steampipeconfig/steampipeconfig.go` (362 lines, 33 changes)
- `pkg/steampipeconfig/connection_plugin.go` (16 changes)
- `pkg/cmdconfig/viper.go` (24 changes)

## Test Files to Create
1. Fix existing `pkg/steampipeconfig/load_config_test.go` (currently SKIPPED with TODO)
2. `pkg/steampipeconfig/steampipeconfig_test.go`
3. `pkg/steampipeconfig/connection_plugin_test.go`
4. `pkg/cmdconfig/viper_test.go`

## Critical Test Cases
- HCL config file parsing
- Connection config loading (.spc, .json, .yml)
- Plugin config resolution
- Workspace profiles
- Config validation
- Config file watching
- Invalid config handling
- Missing file handling
- Environment variable substitution

## Coverage Target
60% coverage

## Dependencies
Requires Task 1, parallel with 2-6

## Time Estimate
3-4 hours

## Special Notes
- Fix the existing skipped test first!
- Look for TODO comments in load_config_test.go

## Command
```bash
claude
# "Please complete task-7-config-tests.md from .ai/milestones/wave-1-foundation/tasks/"
```
