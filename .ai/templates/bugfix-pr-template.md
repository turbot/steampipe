# Bug Fix PR Template

## PR Title

```
Brief description closes #<issue>
```

## PR Description

```markdown
## Summary
[1-2 sentences: what was wrong and how it's fixed]

## Changes

### Commit 1: Demonstrate Bug
- Unskipped test `TestName` in `pkg/path/file_test.go`
- Test **FAILS** with [error/panic/wrong result]

### Commit 2: Fix Bug
- Modified `pkg/path/file.go` to [change description]
- Test now **PASSES**

## Verification
CI history shows: ❌ (commit 1) → ✅ (commit 2)
```

## Branch and Commit Messages

**Branch:**
```
fix/<issue>-brief-description
```

**Commit 1:**
```
Unskip test demonstrating bug #<issue>: description
```

**Commit 2:**
```
Fix #<issue>: description of fix
```

## Checklist

- [ ] Exactly 2 commits in PR
- [ ] Test fails on commit 1
- [ ] Test passes on commit 2
- [ ] Pushed commits separately (two CI runs visible)
- [ ] PR title ends with "closes #XXXX"
- [ ] No unrelated changes
