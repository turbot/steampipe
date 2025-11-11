# Parallel Agent Coordination

Simple patterns for coordinating multiple AI agents working in parallel.

## Basic Pattern

When working on multiple related tasks in parallel:

1. **Create a work directory** in `wip/`:
   ```bash
   mkdir -p .ai/wip/<topic-name>
   ```
   Example: `.ai/wip/bug-fixes-wave-1/` or `.ai/wip/test-snapshot-pkg/`

2. **Coordinator creates task files**:
   ```bash
   # In .ai/wip/<topic>/
   task-1-fix-bug-4767.md
   task-2-fix-bug-4768.md
   task-3-fix-bug-4769.md
   plan.md  # Overall coordination plan
   ```

3. **Parallel agents read and execute**:
   ```
   Agent 1: "See plan in .ai/wip/bug-fixes-wave-1/ and run task-1"
   Agent 2: "See plan in .ai/wip/bug-fixes-wave-1/ and run task-2"
   Agent 3: "See plan in .ai/wip/bug-fixes-wave-1/ and run task-3"
   ```

## Task File Format

Keep task files simple:

```markdown
# Task: Fix bug #4767

## Goal
Fix GetDbClient error handling bug

## Steps
1. Create worktree: /tmp/fix-4767
2. Branch: fix/4767-getdbclient
3. Unskip test in pkg/initialisation/init_data_test.go
4. Verify test fails
5. Implement fix
6. Verify test passes
7. Push (two-phase)
8. Create PR with title: "GetDbClient error handling (closes #4767)"

## Context
See issue #4767 for details
Test is already written and skipped
```

## Work Directory Structure

Example for a bug fixing session:

```
.ai/wip/bug-fixes-wave-1/
├── plan.md                    # Coordinator's overall plan
├── task-1-fix-4767.md        # Task for agent 1
├── task-2-fix-4768.md        # Task for agent 2
├── task-3-fix-4769.md        # Task for agent 3
└── status.md                  # Optional: track completion
```

Example for test generation:

```
.ai/wip/test-snapshot-pkg/
├── plan.md                    # What to test, approach
├── findings.md                # Bugs found during testing
└── test-checklist.md         # Coverage checklist
```

## Benefits

- **Isolated**: Each focus area has its own directory
- **Clean**: Old work directories can be deleted when done
- **Reusable**: Pattern works for any parallel work
- **Simple**: Just files and directories, no complex coordination

## Cleanup

When work is complete:

```bash
# Archive or delete the work directory
rm -rf .ai/wip/<topic-name>/
```

The `.ai/wip/` directory is gitignored, so these temporary files won't clutter the repo.

## Examples

**Parallel bug fixes:**
```
Coordinator: Creates .ai/wip/bug-fixes-wave-1/ with 10 task files
Agents 1-10: Each picks a task file and works independently
```

**Test generation with bug discovery:**
```
Coordinator: Creates .ai/wip/test-generation-phase-2/plan.md
Agent: Writes tests, documents bugs in findings.md
```

**Feature development:**
```
Coordinator: Creates .ai/wip/feature-auth/
            - task-1-backend.md
            - task-2-frontend.md
            - task-3-tests.md
Agents: Work in parallel on each component
```
