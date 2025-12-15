# AI Development Guide for Steampipe

This directory contains documentation, templates, and conventions for AI-assisted development on the Steampipe project.

## Guides

- **[Bug Fix PRs](docs/bug-fix-prs.md)** - Two-commit pattern, branch naming, PR format for bug fixes
- **[GitHub Issues](docs/bug-workflow.md)** - Reporting bugs and issues
- **[Test Generation](docs/test-generation-guide.md)** - Writing effective tests
- **[Parallel Coordination](docs/parallel-coordination.md)** - Working with multiple agents in parallel

## Directory Structure

```
.ai/
├── docs/           # Permanent documentation and guides
├── templates/      # Issue and PR templates
└── wip/           # Temporary workspace (gitignored)
```

## Key Conventions

- **Base branch**: `develop` for all work
- **Bug fixes**: 2-commit pattern (demonstrate → fix)
- **Small PRs**: One logical change per PR
- **Issue linking**: PR title ends with `closes #XXXX`

## For AI Agents

- Reference the relevant guide in `docs/` for your task
- Use templates in `templates/` for PR descriptions
- Use `wip/<topic>/` for coordinated parallel work (gitignored)
- Follow project conventions for branches, commits, and PRs

**Parallel work pattern**: Create `.ai/wip/<topic>/` with task files, then agents can work independently. See [parallel-coordination.md](docs/parallel-coordination.md).
