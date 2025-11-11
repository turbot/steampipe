# AI Development Guide for Steampipe

This directory contains documentation, templates, and best practices for AI-assisted development on the Steampipe project.

## Quick Start

- **Generating Tests?** → See [docs/test-generation-guide.md](docs/test-generation-guide.md)
- **Found a Bug?** → See [docs/bug-workflow.md](docs/bug-workflow.md)
- **Creating PRs?** → See [docs/pr-commit-structure.md](docs/pr-commit-structure.md)
- **Working in Parallel?** → See [docs/git-worktree-guide.md](docs/git-worktree-guide.md)

## Directory Structure

```
.ai/
├── docs/           # Permanent documentation and guides
├── templates/      # Issue and PR templates
├── prompts/        # Reusable AI prompts for common tasks
├── examples/       # Reference implementations
└── wip/           # Temporary workspace (gitignored)
```

## Philosophy

Our approach to AI-assisted development prioritizes:

1. **Quality over Coverage** - Write valuable tests that catch real bugs
2. **Test-First Development** - Demonstrate bugs before fixing them
3. **Minimal Changes** - Keep PRs small and reviewable
4. **Parallel Execution** - Use git worktree to work on multiple tasks
5. **Documentation** - Capture learnings for future AI and human developers

## Workflow Overview

### 1. Test Generation Phase
- Generate high-quality tests focusing on complex logic, edge cases, and concurrency
- Look for bugs - keep testing until you find issues
- Value score: Aim for >2.0 (HIGH-value tests preferred)

### 2. Bug Discovery Phase
- When bugs are found, create detailed GitHub issues
- Document reproduction steps, impact, and suggested fixes
- Mark tests as skipped with issue references

### 3. Test Suite PR
- Combine all passing tests into a single PR
- Target: `develop` branch
- Include skip messages for bug-demonstrating tests

### 4. Bug Fix PRs
- One PR per bug
- Two commits:
  1. Unskip test (demonstrates bug - test fails)
  2. Fix bug (test passes)
- **Two-phase push**: Push commits separately to trigger CI runs
  - Phase 1: Push test commit → CI fails (proves bug exists)
  - Phase 2: Push fix commit → CI passes (proves fix works)
- Base on: `develop` branch

### 5. Review and Merge
- Bug fix PRs can merge independently
- Test suite PR provides comprehensive coverage
- All linked via GitHub issue numbers

## Resources

- [Full Test Generation Guide](docs/test-generation-guide.md)
- [Bug Workflow Guide](docs/bug-workflow.md)
- [Git Worktree Guide](docs/git-worktree-guide.md)
- [PR Structure Guide](docs/pr-commit-structure.md)

## For AI Agents

When working on this codebase:
1. Read the relevant guide from `docs/` before starting
2. Use templates from `templates/` for consistency
3. Use `wip/` directory for temporary files and coordination
4. Follow the 2-commit pattern for bug fixes
5. Use git worktree for parallel work

## Contributing

When you discover new best practices or patterns:
1. Document them in the appropriate guide
2. Update templates if needed
3. Add examples for clarity
4. Keep this README in sync
