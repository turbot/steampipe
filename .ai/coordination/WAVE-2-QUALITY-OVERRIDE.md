# URGENT: Wave 2 Quality Override

**Date:** 2025-11-09
**Priority:** CRITICAL - Read Before Continuing Wave 2

---

## ⚠️ Coverage Targets Are WRONG

Wave 2 task files have coverage targets (70%, 60%, etc.). **IGNORE THESE.**

### Why?
Wave 1.5 proved that coverage % is the wrong metric:
- Deleted 30% of tests → Found 4 bugs
- Coverage went DOWN, quality went UP
- 70% coverage with bad tests < 30% coverage with good tests

---

## ✅ Use These Success Criteria Instead

### Primary Goals (MUST ACHIEVE)
1. **Find Bugs** - At least 1 bug per domain (more = better)
2. **Zero Low-Value Tests** - No tests for:
   - Empty functions
   - Trivial getters
   - Constants/hardcoded strings
   - Fake test doubles that always succeed
   - "Doesn't panic" checks
3. **All Tests High-Value** - Every test must:
   - Test complex logic, error paths, or edge cases
   - Use real behavior (not mocked to always succeed)
   - Would fail if a bug were introduced

### Secondary Goals (Nice to Have)
4. **Test Value Score** - (HIGH tests × 3 + MEDIUM tests × 1) / Total > 2.0
5. **Coverage** - Track but don't chase (30-40% is fine if high quality)
6. **Test Speed** - Unit tests <100ms, integration <1s

---

## 🎯 What "Success" Looks Like

### Good Wave 2 Outcome
```
Tests added: 150
Bugs found: 5
High-value ratio: 100%
Coverage: 35%
Test value score: 2.8
```
✅ **This is success!** Found bugs, all tests valuable.

### Bad Wave 2 Outcome
```
Tests added: 300
Bugs found: 0
High-value ratio: 60%
Coverage: 70%
Test value score: 1.8
```
❌ **This is failure!** Hit coverage target but found no bugs, added low-value tests.

---

## 📋 Updated Success Checklist

Before marking your task complete, verify:

### Must Have (Mandatory)
- [ ] Found at least 1 real bug (or documented why code is bug-free)
- [ ] Zero tests for empty functions or trivial getters
- [ ] Zero tests that just check "doesn't panic"
- [ ] Zero tests of constants or hardcoded strings
- [ ] Zero fake test doubles that always succeed
- [ ] All tests test complex logic, error paths, or edge cases
- [ ] All tests would fail if bugs introduced

### Should Have (Strongly Recommended)
- [ ] Test value score > 2.0
- [ ] At least 50% of tests are HIGH value
- [ ] Tests for concurrent access (where applicable)
- [ ] Tests for error recovery and cleanup
- [ ] Tests for boundary conditions (nil, empty, max)
- [ ] Tests run with `-race` flag (no races found)

### Nice to Have (Optional)
- [ ] Coverage > 30% (but don't chase it!)
- [ ] Performance benchmarks for critical paths
- [ ] Documentation of complex test scenarios

---

## 🐛 How to Find Bugs

Based on Wave 1.5 success, bugs hide in:

### Memory Leaks
```go
// Test: Does collection grow unbounded?
for i := 0; i < 1000; i++ {
    collection.Add(item)
}
// BUG?: Is collection ever cleaned?
```

### Nil Panics
```go
// Test: What if fields are nil?
obj := &Struct{}  // nil fields
obj.Method()      // BUG?: Panic?
```

### Boundary Errors
```go
// Test: What if input is empty/nil/max?
Function("")      // BUG?: Panic?
Function(nil)     // BUG?: Panic?
Function(hugeInput) // BUG?: Hang?
```

### Race Conditions
```go
// Test: Concurrent access safe?
go test -race      // BUG?: Data race?
```

---

## 🚫 What NOT to Do

### Don't Write Tests Just for Coverage
```go
// BAD: Writing test for trivial getter just to hit coverage
func TestGetName(t *testing.T) {
    obj := &Obj{name: "test"}
    assert.Equal(t, "test", obj.GetName())
}
```
**Delete this immediately.** It adds no value.

### Don't Mock Everything to Always Succeed
```go
// BAD: Fake that never fails
type fakeExporter struct{}
func (f *fakeExporter) Export() error {
    return nil  // Always succeeds!
}
```
**This creates false confidence.** Use real behavior.

### Don't Test Language Features
```go
// BAD: Testing that Go's if statement works
func TestDisabled(t *testing.T) {
    if disabled {
        assert.Nil(t, feature())
    }
}
```
**Delete this.** Tests Go, not your code.

---

## ✅ What TO Do

### Test Complex Logic
```go
// GOOD: Tests actual business logic
func TestConnectionPoolExhaustion(t *testing.T) {
    pool := NewPool(maxConns: 5)
    // Acquire all connections
    for i := 0; i < 5; i++ {
        pool.Acquire()
    }
    // Try one more - should timeout or error
    ctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
    defer cancel()
    _, err := pool.Acquire()
    assert.Error(t, err)  // BUG?: Hangs? Wrong error?
}
```

### Test Error Paths
```go
// GOOD: Tests error handling
func TestExportDiskFull(t *testing.T) {
    exporter := NewExporter()
    mockFS := &mockFS{writeError: errors.New("disk full")}
    err := exporter.ExportWithFS(mockFS, data)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "disk full")
    // BUG?: Partial file left behind?
    // BUG?: File handle leaked?
}
```

### Test Edge Cases
```go
// GOOD: Tests boundary conditions
func TestValidation(t *testing.T) {
    tests := []struct{
        input string
        valid bool
    }{
        {"", false},           // Empty
        {nil, false},          // Nil
        {"x", true},           // Minimum
        {strings.Repeat("x", 10000), false}, // Maximum
    }
    // BUG?: Panics? Wrong validation?
}
```

---

## 🔄 If You Already Wrote Low-Value Tests

**Delete them immediately.** Better to have fewer high-quality tests than many low-quality tests.

### Quick Audit
Review each test and ask:
1. Does this test complex logic?
2. Would this fail if a bug were introduced?
3. Does this test more than a trivial getter?

If answer is "no" to all three: **DELETE IT.**

---

## 📊 Report These Metrics

When you complete your task, report:

### Primary Metrics
- **Bugs found:** [number] - List each bug
- **Tests added:** [number]
- **High-value tests:** [number] (%)
- **Medium-value tests:** [number] (%)
- **Low-value tests:** [number] ← Should be ZERO
- **Test value score:** [number] ← Should be > 2.0

### Secondary Metrics
- **Coverage achieved:** [%] - Track but don't chase
- **Tests deleted:** [number] - Deleting is good!
- **Execution time:** [seconds] - Should be <5s

### Quality Evidence
- **Race conditions:** Run with `-race`, report results
- **Error paths tested:** [% of functions with error returns]
- **Edge cases tested:** Nil, empty, boundary conditions

---

## 🎯 Remember

**From Wave 1.5:**
> "The goal is not to have lots of tests.
> The goal is to have tests that find bugs."

**Coverage % is a vanity metric.**
**Bugs found is the real metric.**

---

## Examples from Wave 1.5

We found 4 real bugs by:
1. Testing nil pointer handling → Found ResetPools panic
2. Testing boundary conditions → Found isValidDatabaseName panic
3. Testing resource cleanup → Found session map memory leak
4. Testing error handling → Found error prefix behavior

**Do more of this in Wave 2!**

---

## Questions?

**Q: But the task file says 70% coverage?**
A: Ignore it. It was written before Wave 1.5. Quality > coverage.

**Q: What if I can't find bugs?**
A: That's OK! Document that you tested thoroughly. But keep hunting - bugs are there.

**Q: What if my coverage is only 30%?**
A: That's fine! If all tests are high-value and you found bugs, that's success.

**Q: Should I write more tests to hit 70%?**
A: NO! Only write tests that add value. Quality > quantity.

---

**This overrides any coverage targets in task files.**

**Focus on finding bugs, not hitting coverage numbers.**
