# Bug Report: Export Package

**Date Found:** 2025-11-09
**Found By:** Bug hunting tests in Task 9 (Wave 2 Quality Focus)

---

## Bugs Found: 3 Total

1. **BUG #1:** Race Condition in Manager - **CRITICAL** 🔴
2. **BUG #2:** Nil Exporter Panic in Target - **HIGH** 🟠
3. **BUG #3:** Partial File Left on Error - **MEDIUM** 🟡

---

# BUG #1: Race Condition in Manager

**Severity:** **CRITICAL** 🔴
**Status:** Confirmed by race detector

## Bug Summary

**Manager has no mutex/lock protection, causing data races and crashes when Register() is called concurrently.**

---

## Bug Details

### Location
- **File:** `pkg/export/manager.go`
- **Functions Affected:**
  - `Manager.Register()` (lines 29-52)
  - `Manager.registerExporterByExtension()` (lines 54-80)
  - `Manager.getExportTarget()` (line 117+)
  - `Manager.ValidateExportFormat()` (line 178+)
  - `Manager.HasNamedExport()` (line 167+)

### Root Cause
The `Manager` struct has two maps that are accessed without synchronization:
```go
type Manager struct {
    registeredExporters  map[string]Exporter  // NO MUTEX!
    registeredExtensions map[string]Exporter  // NO MUTEX!
}
```

Any concurrent access to these maps (read+write or write+write) will cause:
1. **Data races** (detected by `-race` flag)
2. **Panics** with "fatal error: concurrent map read and map write"
3. **Application crashes**

### Race Conditions Detected

**Race #1: registeredExporters map**
- **manager.go:31** - Read: `if _, ok := m.registeredExporters[name]; ok {`
- **manager.go:34** - Write: `m.registeredExporters[exporter.Name()] = exporter`
- Multiple goroutines reading/writing simultaneously

**Race #2: registeredExtensions map**
- **manager.go:56** - Read: `if existing, ok := m.registeredExtensions[ext]; ok {`
- **manager.go:65** - Delete: `delete(m.registeredExtensions, ext)`
- **manager.go:79** - Write: `m.registeredExtensions[ext] = exporter`
- Multiple goroutines accessing during registration

### How to Reproduce

```go
m := NewManager()

// Register exporters concurrently (simulates plugin initialization)
var wg sync.WaitGroup
for i := 0; i < 100; i++ {
    wg.Add(1)
    go func(n int) {
        defer wg.Done()
        exporter := &testExporter{name: fmt.Sprintf("exp%d", n), extension: ".test"}
        m.Register(exporter)  // RACE CONDITION HERE
    }(i)
}
wg.Wait()
```

**Run with:** `go test -race`

**Result:** Multiple data race warnings + fatal error

---

## Impact Assessment

### Severity: CRITICAL

**When Bug Occurs:**
- During application startup if exporters are registered concurrently
- During plugin initialization if done in parallel
- Any scenario where Register() called from multiple goroutines

**Consequences:**
1. **Application crashes** - "fatal error: concurrent map read and map write"
2. **Data corruption** - Partial registrations, missing exporters
3. **Undefined behavior** - Race conditions can cause unpredictable results

**Likelihood:**
- **High** if exporters are registered during concurrent initialization
- **Low** if Register() only called serially during single-threaded startup

### Real-World Scenarios

1. **Plugin Manager** initializes multiple exporters in parallel
2. **Hot Reload** re-registers exporters while exports are running
3. **Test Suites** that register exporters concurrently
4. **Service Restart** where cleanup and initialization overlap

---

## Evidence

### Race Detector Output

```
==================
WARNING: DATA RACE
Read at 0x00c000627080 by goroutine 29:
  github.com/turbot/steampipe/v2/pkg/export.(*Manager).Register()
      /Users/nathan/src/steampipe/pkg/export/manager.go:31 +0x6c

Previous write at 0x00c000627080 by goroutine 30:
  github.com/turbot/steampipe/v2/pkg/export.(*Manager).Register()
      /Users/nathan/src/steampipe/pkg/export/manager.go:34 +0xb8
==================

fatal error: concurrent map read and map write
```

### Test That Found Bug
- **Test:** `TestManagerConcurrentRegistration_RaceCondition`
- **File:** `pkg/export/bug_hunting_test.go`
- **Command:** `go test -race`

---

## Recommended Fix

### Solution: Add Mutex Protection

```go
type Manager struct {
    mu                   sync.RWMutex         // Add this
    registeredExporters  map[string]Exporter
    registeredExtensions map[string]Exporter
}

func (m *Manager) Register(exporter Exporter) error {
    m.mu.Lock()         // Lock for write
    defer m.mu.Unlock()

    // ... existing code ...
}

func (m *Manager) getExportTarget(export, executionName string) (*Target, error) {
    m.mu.RLock()        // Lock for read
    defer m.mu.RUnlock()

    // ... existing code ...
}

// Similar changes for:
// - ValidateExportFormat (RLock)
// - HasNamedExport (RLock)
// - registerExporterByExtension (called from Register, already locked)
```

### Alternative: Use sync.Map
```go
type Manager struct {
    registeredExporters  sync.Map  // map[string]Exporter
    registeredExtensions sync.Map  // map[string]Exporter
}
```

**Recommendation:** Use `sync.RWMutex` for simplicity and compatibility with existing code.

---

## Verification Plan

After fix is applied:

1. ✅ Run `go test -race` - Should show NO races
2. ✅ Run concurrent registration test - Should pass
3. ✅ Run concurrent read/write test - Should pass
4. ✅ Verify no performance regression
5. ✅ Update tests to verify thread-safety

---

## Related Issues

### Other Functions That Need Review
- `DoExport()` - Reads from maps, needs RLock
- `resolveTargetsFromArgs()` - Reads from maps, needs RLock

### Questions for Code Owner
1. Is Register() only called during initialization (single-threaded)?
2. Are there plans for hot-reload or dynamic registration?
3. Should we prevent registration after first use?

---

## Historical Context

This bug likely went unnoticed because:
1. Exporters are typically registered once during startup (single-threaded)
2. No concurrent registration in normal workflows
3. Tests don't typically run with `-race` flag
4. No stress testing or concurrent scenarios

Wave 1.5 testing strategy (quality over coverage) successfully found this bug by:
- Writing tests specifically designed to find race conditions
- Running with `-race` flag
- Testing concurrent access patterns
- Not assuming "it works in practice"

---

## Summary

- ✅ **Bug confirmed:** Data race causing crashes
- ✅ **Severity:** Critical - can crash application
- ✅ **Fix:** Add mutex protection (simple, well-understood)
- ✅ **Test:** Bug-hunting test will verify fix

**This bug validates the Wave 1.5 approach: Quality > Coverage**

---

# BUG #2: Nil Exporter Panic in Target

**Severity:** **HIGH** 🟠
**Status:** Confirmed by panic test

## Bug Summary

**Target.Export() panics with nil pointer dereference if Target.exporter is nil.**

---

## Bug Details

### Location
- **File:** `pkg/export/target.go`
- **Function:** `Target.Export()` (line 15-23)

### Root Cause
The `Target.Export()` method doesn't check if `t.exporter` is nil before calling methods on it:

```go
func (t *Target) Export(ctx context.Context, input ExportSourceData) (string, error) {
    err := t.exporter.Export(ctx, input, t.filePath)  // PANIC if t.exporter is nil!
    if err != nil {
        return "", err
    } else {
        pwd, _ := os.Getwd()
        return fmt.Sprintf("File exported to %s/%s", pwd, t.filePath), nil
    }
}
```

### How to Reproduce

```go
target := &Target{
    exporter: nil,  // This should not be allowed
    filePath: "output.json",
}

_, err := target.Export(ctx, data)  // PANICS: invalid memory address or nil pointer dereference
```

---

## Impact Assessment

### Severity: HIGH

**When Bug Occurs:**
- If `getExportTarget()` has a bug and creates Target with nil exporter
- If Target is constructed manually without validation
- During error conditions where exporter lookup fails but Target is still created

**Consequences:**
1. **Application panic** - Crashes the entire process
2. **No graceful error handling** - Can't recover
3. **Poor error message** - Just "nil pointer dereference"

**Likelihood:**
- **Low** in normal operation (Target created by Manager has proper exporter)
- **High** if there are bugs in Manager's target resolution logic

---

## Evidence

### Test Output

```
bug_hunting_test.go:130: Found potential bug: panic on nil exporter: runtime error: invalid memory address or nil pointer dereference
```

### Test That Found Bug
- **Test:** `TestTargetExport_NilExporter`
- **File:** `pkg/export/bug_hunting_test.go`

---

## Recommended Fix

### Solution: Add Nil Check

```go
func (t *Target) Export(ctx context.Context, input ExportSourceData) (string, error) {
    // Add nil check
    if t.exporter == nil {
        return "", fmt.Errorf("target has nil exporter - invalid target")
    }

    err := t.exporter.Export(ctx, input, t.filePath)
    if err != nil {
        return "", err
    }

    pwd, _ := os.Getwd()
    return fmt.Sprintf("File exported to %s/%s", pwd, t.filePath), nil
}
```

### Alternative: Validate in Constructor
Make Target construction private and validate:

```go
func newTarget(exporter Exporter, filePath string, isNamed bool) (*Target, error) {
    if exporter == nil {
        return nil, fmt.Errorf("exporter cannot be nil")
    }
    return &Target{
        exporter:      exporter,
        filePath:      filePath,
        isNamedTarget: isNamed,
    }, nil
}
```

---

# BUG #3: Partial File Left on Error

**Severity:** **MEDIUM** 🟡
**Status:** Confirmed by file system test

## Bug Summary

**Write() helper leaves partial file on disk if io.Copy() fails, potentially confusing users or wasting disk space.**

---

## Bug Details

### Location
- **File:** `pkg/export/helpers.go`
- **Function:** `Write()` (lines 16-26)

### Root Cause
When `io.Copy()` fails after writing some data, the file remains on disk with partial content:

```go
func Write(filePath string, exportData io.Reader) error {
    destination, err := os.Create(filePath)
    if err != nil {
        return err
    }
    defer destination.Close()

    _, err = io.Copy(destination, exportData)
    return err  // File exists with partial data if this fails!
}
```

### How to Reproduce

```go
// Create a reader that fails after 10 bytes
failingReader := &failAfterNReader{
    data:      []byte("This is test data that will fail"),
    failAfter: 10,
}

err := Write("/tmp/test.txt", failingReader)
// err != nil, BUT file exists with "This is te"
```

---

## Impact Assessment

### Severity: MEDIUM

**When Bug Occurs:**
- Disk full during write
- Network failure for network-backed readers
- Source data stream errors
- Out of memory during large writes

**Consequences:**
1. **Partial files left on disk** - User sees file but it's incomplete
2. **Disk space wasted** - Especially for large exports
3. **Confusion** - User might think export succeeded
4. **No cleanup** - Files accumulate over time

**Likelihood:**
- **Medium** - IO errors are common (disk full, permissions, etc.)

---

## Evidence

### Test Output

```
bug_hunting_test.go:187: Partial file exists with 10 bytes: "This is te"
bug_hunting_test.go:188: Potential issue: partial file not cleaned up on error
```

### Test That Found Bug
- **Test:** `TestWrite_PartialWriteFailure`
- **File:** `pkg/export/bug_hunting_test.go`

---

## Recommended Fix

### Solution: Remove File on Error

```go
func Write(filePath string, exportData io.Reader) error {
    destination, err := os.Create(filePath)
    if err != nil {
        return err
    }
    defer destination.Close()

    _, err = io.Copy(destination, exportData)
    if err != nil {
        // Clean up partial file on error
        os.Remove(filePath)
        return err
    }

    return nil
}
```

### Alternative: Write to Temp File First

```go
func Write(filePath string, exportData io.Reader) error {
    // Write to temp file first
    tempPath := filePath + ".tmp"
    destination, err := os.Create(tempPath)
    if err != nil {
        return err
    }

    _, err = io.Copy(destination, exportData)
    destination.Close()

    if err != nil {
        os.Remove(tempPath)
        return err
    }

    // Atomic rename
    return os.Rename(tempPath, filePath)
}
```

**Recommendation:** Use temp file + atomic rename for better reliability.

---

## Summary of All Bugs

| # | Bug | Severity | Impact | Fix Complexity |
|---|-----|----------|--------|----------------|
| 1 | Race Condition in Manager | CRITICAL 🔴 | App crash | Low (add mutex) |
| 2 | Nil Exporter Panic | HIGH 🟠 | App panic | Low (add nil check) |
| 3 | Partial File on Error | MEDIUM 🟡 | User confusion | Low (cleanup on error) |

**All bugs found through aggressive bug-hunting tests.**
**All bugs have straightforward fixes.**
**Wave 1.5 quality-focused testing strategy validated: Finding bugs > hitting coverage %.**
