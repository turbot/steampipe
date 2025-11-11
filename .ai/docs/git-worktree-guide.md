# Git Worktree Guide for Parallel Development

## Why Use Git Worktrees?

Git worktrees allow you to work on multiple branches simultaneously without switching contexts. This is perfect for:
- Fixing multiple bugs in parallel
- Running tests in one worktree while developing in another
- Keeping your main worktree clean while experimenting

## Basic Concepts

A **worktree** is a separate checkout of your repository. You can have:
- Main worktree: `/Users/nathan/src/steampipe`
- Additional worktrees: `/Users/nathan/src/steampipe-4767`, `/Users/nathan/src/steampipe-4768`, etc.

Each worktree:
- Has its own working directory
- Shares the same `.git` repository
- Can be on a different branch
- Can be worked on independently

## Creating Worktrees

### For a New Bug Fix

```bash
# Create worktree for bug fix
git worktree add /Users/nathan/src/steampipe-4767 -b fix/4767-description develop

# Breakdown:
# - /Users/nathan/src/steampipe-4767  # Path to new worktree
# - -b fix/4767-description           # Create new branch
# - develop                            # Base branch
```

### For an Existing Branch

```bash
# Create worktree for existing branch
git worktree add /Users/nathan/src/steampipe-4767 fix/4767-description

# Or use a temporary location
git worktree add /tmp/fix-4767 fix/4767-description
```

### Naming Convention

Use consistent naming:
```
/Users/nathan/src/steampipe-<issue-number>
/tmp/steampipe-<issue-number>
/tmp/fix-<issue-number>
```

## Working with Multiple Worktrees

### Parallel Bug Fixing Example

```bash
# Main worktree: Continue test generation
cd /Users/nathan/src/steampipe

# Create worktrees for 3 bugs
git worktree add /tmp/fix-4767 -b fix/4767-getdbclient develop
git worktree add /tmp/fix-4768 -b fix/4768-goroutine-leak develop
git worktree add /tmp/fix-4769 -b fix/4769-nil-exporter develop

# Work on each in parallel (different terminals or tabs)
cd /tmp/fix-4767 && code .
cd /tmp/fix-4768 && code .
cd /tmp/fix-4769 && code .
```

### Using with AI Agents

Launch parallel agents with worktrees:

```bash
# In Claude Code or similar tool
Task: Fix bug 4767
Working directory: /tmp/fix-4767

Task: Fix bug 4768
Working directory: /tmp/fix-4768

# Both run in parallel without conflicts
```

## Common Operations

### List All Worktrees

```bash
git worktree list
```

Output:
```
/Users/nathan/src/steampipe              365f928e [feature/wave-3-quality-tests]
/tmp/fix-4767                             0b3e08c4 [fix/4767-getdbclient]
/tmp/fix-4768                             5b509863 [fix/4768-goroutine-leak]
```

### Remove a Worktree

```bash
# After you're done with it
git worktree remove /tmp/fix-4767

# If it has uncommitted changes, force remove
git worktree remove -f /tmp/fix-4767
```

### Move a Worktree

```bash
# Move the directory
mv /tmp/fix-4767 /Users/nathan/src/steampipe-4767

# Update git's tracking
git worktree repair
```

### Prune Deleted Worktrees

```bash
# If you manually deleted worktree directories
git worktree prune
```

## Workflow Patterns

### Pattern 1: Test and Fix in Parallel

```bash
# Main worktree: Generate tests
cd /Users/nathan/src/steampipe
# Writing tests, finding bugs...

# Side worktree: Fix a bug while testing continues
git worktree add /tmp/fix-bug -b fix/bug-description develop
cd /tmp/fix-bug
# Implement fix...
```

### Pattern 2: Multiple Bug Fixes

```bash
# Create worktrees for all bugs found
for issue in 4767 4768 4769 4770; do
    git worktree add /tmp/fix-$issue -b fix/$issue-description develop
done

# Fix each independently (can use parallel agents)
# Each can be pushed and PR'd independently
```

### Pattern 3: Review and Develop

```bash
# Main worktree: Current development
cd /Users/nathan/src/steampipe

# Review worktree: Check out PR for review
git worktree add /tmp/review-pr-4765 pr-branch-name

# Can test the PR without disrupting development
cd /tmp/review-pr-4765
go test ./...
```

## Rebasing in Worktrees

### Rebase onto Latest Develop

```bash
# In the worktree
cd /tmp/fix-4767

# Fetch latest
git fetch origin develop

# Rebase
git rebase origin/develop

# Force push (if already pushed)
git push --force-with-lease
```

### Handling Conflicts

```bash
# If rebase conflicts occur
git status  # See conflicted files
# Resolve conflicts
git add resolved-file.go
git rebase --continue

# Or abort if needed
git rebase --abort
```

## Best Practices

### DO:
- ✅ Use descriptive worktree paths
- ✅ Clean up worktrees when done
- ✅ Use worktrees for parallel bug fixes
- ✅ Keep worktrees short-lived
- ✅ Use consistent naming conventions

### DON'T:
- ❌ Commit directly to develop in any worktree
- ❌ Have the same branch checked out in multiple worktrees
- ❌ Keep worktrees around forever
- ❌ Mix unrelated changes in one worktree

## Integration with Claude Code

### Task Tool with Worktrees

```javascript
// Launch parallel agents with worktrees
Task({
  subagent_type: "general-purpose",
  description: "Fix bug 4767",
  prompt: `
    Working directory: /tmp/fix-4767

    1. Unskip test for bug 4767
    2. Verify test fails
    3. Implement fix
    4. Verify test passes
    5. Push to origin
  `
})

// Launch another in parallel
Task({
  subagent_type: "general-purpose",
  description: "Fix bug 4768",
  prompt: `
    Working directory: /tmp/fix-4768
    [same pattern]
  `
})
```

## Common Scenarios

### Scenario 1: Found 10 Bugs, Fix Them All

```bash
# Create test suite PR first
git checkout -b feature/comprehensive-tests develop
# Generate all tests, mark bugs as skipped
git commit -m "Add comprehensive tests"
git push

# Create worktrees for each bug
for i in {4767..4776}; do
    git worktree add /tmp/fix-$i -b fix/$i-description develop
done

# Use Task tool to fix all in parallel
# Or work on them sequentially in different terminals
```

### Scenario 2: Need to Test Against Different Branches

```bash
# Test against develop
git worktree add /tmp/test-develop develop
cd /tmp/test-develop && go test ./...

# Test against feature branch
git worktree add /tmp/test-feature feature/branch
cd /tmp/test-feature && go test ./...

# Compare results
```

### Scenario 3: Emergency Hotfix While Developing

```bash
# Current work in progress in main worktree
cd /Users/nathan/src/steampipe
# On feature branch, uncommitted changes

# Create worktree for hotfix
git worktree add /tmp/hotfix -b hotfix/critical-bug main
cd /tmp/hotfix

# Fix and push
# Main worktree is undisturbed
```

## Cleanup

### Regular Cleanup

```bash
# List all worktrees
git worktree list

# Remove finished worktrees
git worktree remove /tmp/fix-4767
git worktree remove /tmp/fix-4768

# Prune deleted worktrees
git worktree prune
```

### Bulk Cleanup

```bash
# Remove all worktrees in /tmp
for wt in /tmp/fix-*; do
    git worktree remove -f "$wt"
done

# Or use find
find /tmp -maxdepth 1 -name "fix-*" -type d -exec git worktree remove -f {} \;
```

## Troubleshooting

### Issue: "Already checked out"
```
fatal: 'fix/bug-description' is already checked out at '/tmp/fix-4767'
```

**Solution**: Remove the worktree first or use a different branch name

### Issue: "Prunable" Worktree
```
/tmp/fix-4767  prunable
```

**Solution**: The directory was deleted manually
```bash
git worktree prune
```

### Issue: Can't Remove Worktree
```
fatal: validation failed, cannot remove working tree
```

**Solution**: Use force flag
```bash
git worktree remove -f /tmp/fix-4767
```

## Advanced Usage

### Sharing Git Config

All worktrees share `.git/config`, so:
- Remotes are shared
- Global settings are shared
- Hooks are shared

### Worktree-Specific Config

Use worktree-specific config file:
```bash
cd /tmp/fix-4767
git config --worktree user.email "test@example.com"
```

### Sparse Checkouts

For large repos, use sparse checkout in worktrees:
```bash
git worktree add --no-checkout /tmp/minimal develop
cd /tmp/minimal
git sparse-checkout set pkg/specific/path
git checkout develop
```

## Summary

Worktrees enable:
- ⚡ Parallel development on multiple bugs
- 🔀 Context switching without branch switching
- 🧪 Testing without disrupting development
- 🤖 AI agents working concurrently
- 🚀 Faster bug fix iterations

Use them liberally for bug fixing workflows!

## Next Steps

- [Bug Workflow](bug-workflow.md) - How to fix bugs
- [PR Structure](pr-commit-structure.md) - 2-commit pattern
- [Test Generation](test-generation-guide.md) - Finding bugs to fix
