# Blockers & Issues

**Last Updated:** 2025-11-08

## Active Blockers
None - project not yet started.

## Resolved Blockers
None yet.

## Potential Risks

### High Priority
1. **Mocking PostgreSQL** - May be complex to mock database operations
   - Mitigation: Use temp databases or in-memory SQLite for tests
   - Alternative: Focus on logic testing, not DB testing

2. **Plugin Process Mocking** - Plugins run as separate processes
   - Mitigation: Mock at the plugin manager level, not process level
   - Alternative: Use test plugins (chaos plugin already exists)

3. **gRPC Testing** - Plugin manager uses gRPC
   - Mitigation: Use grpc testing package
   - Alternative: Mock gRPC clients

### Medium Priority
4. **Test Execution Time** - May become slow with many tests
   - Mitigation: Keep unit tests fast, use parallel execution
   - Monitor: Track test execution time

5. **Flaky Tests** - Network/timing dependent tests may be flaky
   - Mitigation: Use deterministic mocks, control time
   - Detection: Run tests 10x to find flakes

6. **CI Resource Limits** - Coverage may slow down CI
   - Mitigation: Separate unit and integration test runs
   - Alternative: Run coverage on main branch only

### Low Priority
7. **Mock Maintenance** - Mocks may drift from real interfaces
   - Mitigation: Keep mocks simple, test against real when possible
   - Alternative: Use code generation for mocks

## How to Report Blockers

If you encounter a blocker while working on a task:

1. **Document it here** with:
   - Task ID
   - Description of blocker
   - What you tried
   - Potential solutions
   - Severity (High/Medium/Low)

2. **Update task STATUS.md** with blocker status

3. **Continue with non-blocked work** if possible

4. **Escalate if critical** - mark as ⛔ CRITICAL BLOCKER

## Blocker Template

```markdown
### [Task X] - Description
**Date:** YYYY-MM-DD
**Severity:** High/Medium/Low
**Status:** 🚫 Active / ✅ Resolved

**Problem:**
[Clear description of what's blocking progress]

**Tried:**
- Attempt 1
- Attempt 2

**Potential Solutions:**
1. Solution A (pros/cons)
2. Solution B (pros/cons)

**Decision:**
[What was decided]

**Resolution:**
[How it was resolved, if applicable]
```

## Questions & Answers

### Q: What if existing tests break?
**A:** STOP immediately. This is a critical issue.
1. Revert your changes
2. Document what broke in BLOCKERS.md
3. Investigate root cause
4. Fix the issue or change approach

### Q: What if coverage target not achievable?
**A:** Document why and adjust target.
1. Note in STATUS.md why target not reached
2. Document what would be needed
3. Adjust target for this wave
4. Plan additional coverage for next wave

### Q: What if tests are too slow?
**A:** Optimize or split tests.
1. Profile slow tests
2. Add mocks to remove I/O
3. Use parallel test execution
4. Consider integration test category

### Q: What if mocking is too complex?
**A:** Simplify or use real dependencies.
1. Simplify the mock interface
2. Use real dependencies in temp environment
3. Test at higher level of abstraction

## Current Status
✅ No blockers - project not yet started.
