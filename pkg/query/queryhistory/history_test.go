package queryhistory

import (
	"fmt"
	"testing"

	"github.com/turbot/steampipe/v2/pkg/constants"
)

// TestQueryHistory_BoundedSize tests that query history doesn't grow unbounded.
// This test demonstrates bug #4811 where history could grow without limit in memory
// during a session, even though Push() limits new additions.
//
// Bug: #4811
func TestQueryHistory_BoundedSize(t *testing.T) {
	// t.Skip("Test demonstrates bug #4811: query history grows unbounded in memory during session")

	// Simulate a scenario where history is pre-populated (e.g., from a corrupted file or direct manipulation)
	// This represents the in-memory history during a long-running session
	oversizedHistory := make([]string, constants.HistorySize+100)
	for i := 0; i < len(oversizedHistory); i++ {
		oversizedHistory[i] = fmt.Sprintf("SELECT %d;", i)
	}

	history := &QueryHistory{history: oversizedHistory}

	// Even with pre-existing oversized history, operations should enforce the limit
	// Get() should never return more than HistorySize entries
	retrieved := history.Get()
	if len(retrieved) > constants.HistorySize {
		t.Errorf("Get() returned %d entries, exceeds limit %d", len(retrieved), constants.HistorySize)
	}

	// After any operation, the internal history should be bounded
	history.Push("SELECT new;")
	if len(history.history) > constants.HistorySize {
		t.Errorf("After Push(), history size %d exceeds limit %d", len(history.history), constants.HistorySize)
	}
}
