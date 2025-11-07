# Task 6: Database Client Tests

## Context
Database client is the interface to PostgreSQL. All queries go through here. 19 commits to db_client_execute.go.

## Files to Test
- `pkg/db/db_client/db_client.go` (316 lines)
- `pkg/db/db_client/db_client_connect.go`
- `pkg/db/db_client/db_client_execute.go` (19 changes)
- `pkg/db/db_client/db_client_search_path.go`
- `pkg/db/db_client/db_client_session.go`

## Test Files to Create
1. `pkg/db/db_client/db_client_test.go`
2. `pkg/db/db_client/db_client_connect_test.go`
3. `pkg/db/db_client/db_client_execute_test.go`
4. `pkg/db/db_client/db_client_search_path_test.go`

## Critical Test Cases
- Connection pool management (user + management pools)
- Connection acquisition/release
- Search path setup and modification
- Query execution with retry
- Session management
- Connection timeout
- Pool exhaustion
- Connection errors
- Server settings loading

## Coverage Target
60% of critical paths

## Dependencies
Requires Task 1, parallel with 2-5, 7

## Time Estimate
4 hours

## Command
```bash
claude
# "Please complete task-6-db-client-tests.md from .ai/milestones/wave-1-foundation/tasks/"
```
