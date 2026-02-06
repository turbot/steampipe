package interactive

import (
	"context"
	"strings"
	"sync"
	"testing"

	"github.com/c-bata/go-prompt"
	pconstants "github.com/turbot/pipe-fittings/v2/constants"
	"github.com/turbot/steampipe/v2/pkg/cmdconfig"
)

// TestGetTableAndConnectionSuggestions_ReturnsEmptySliceNotNil tests that
// getTableAndConnectionSuggestions returns an empty slice instead of nil
// when no matching connection is found in the schema.
//
// This is important for proper API contract - functions that return slices
// should return empty slices rather than nil to avoid unexpected nil pointer
// issues in calling code.
//
// Bug: #4710
// PR: #4734
func TestGetTableAndConnectionSuggestions_ReturnsEmptySliceNotNil(t *testing.T) {
	tests := []struct {
		name     string
		word     string
		expected bool // true if we expect non-nil result
	}{
		{
			name:     "empty word should return non-nil",
			word:     "",
			expected: true,
		},
		{
			name:     "unqualified table should return non-nil",
			word:     "table",
			expected: true,
		},
		{
			name:     "non-existent connection should return non-nil",
			word:     "nonexistent.table",
			expected: true,
		},
		{
			name:     "qualified table with dot should return non-nil",
			word:     "aws.instances",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a minimal InteractiveClient with empty suggestions
			c := &InteractiveClient{
				suggestions: &autoCompleteSuggestions{
					schemas:          []prompt.Suggest{},
					unqualifiedTables: []prompt.Suggest{},
					tablesBySchema:    make(map[string][]prompt.Suggest),
				},
			}

			result := c.getTableAndConnectionSuggestions(tt.word)

			if tt.expected && result == nil {
				t.Errorf("getTableAndConnectionSuggestions(%q) returned nil, expected non-nil empty slice", tt.word)
			}

			// Additional check: even if not nil, should be empty in these test cases
			if result != nil && len(result) != 0 {
				t.Errorf("getTableAndConnectionSuggestions(%q) returned non-empty slice %v, expected empty slice", tt.word, result)
			}
		})
	}
}

// TestShouldExecute tests the shouldExecute logic for query execution
func TestShouldExecute(t *testing.T) {
	// Save and restore viper settings
	originalMultiline := cmdconfig.Viper().GetBool(pconstants.ArgMultiLine)
	defer func() {
		cmdconfig.Viper().Set(pconstants.ArgMultiLine, originalMultiline)
	}()

	tests := []struct {
		name         string
		query        string
		multiline    bool
		shouldExec   bool
		description  string
	}{
		{
			name:        "simple query without semicolon in non-multiline",
			query:       "SELECT * FROM users",
			multiline:   false,
			shouldExec:  true,
			description: "In non-multiline mode, execute without semicolon",
		},
		{
			name:        "simple query with semicolon in non-multiline",
			query:       "SELECT * FROM users;",
			multiline:   false,
			shouldExec:  true,
			description: "In non-multiline mode, execute with semicolon",
		},
		{
			name:        "simple query without semicolon in multiline",
			query:       "SELECT * FROM users",
			multiline:   true,
			shouldExec:  false,
			description: "In multiline mode, don't execute without semicolon",
		},
		{
			name:        "simple query with semicolon in multiline",
			query:       "SELECT * FROM users;",
			multiline:   true,
			shouldExec:  true,
			description: "In multiline mode, execute with semicolon",
		},
		{
			name:        "metaquery without semicolon in multiline",
			query:       ".help",
			multiline:   true,
			shouldExec:  true,
			description: "Metaqueries execute without semicolon even in multiline",
		},
		{
			name:        "metaquery with semicolon in multiline",
			query:       ".help;",
			multiline:   true,
			shouldExec:  true,
			description: "Metaqueries execute with semicolon in multiline",
		},
		{
			name:        "empty query",
			query:       "",
			multiline:   false,
			shouldExec:  true,
			description: "Empty query executes in non-multiline",
		},
		{
			name:        "empty query in multiline",
			query:       "",
			multiline:   true,
			shouldExec:  false,
			description: "Empty query doesn't execute in multiline",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &InteractiveClient{}
			cmdconfig.Viper().Set(pconstants.ArgMultiLine, tt.multiline)

			result := c.shouldExecute(tt.query)

			if result != tt.shouldExec {
				t.Errorf("shouldExecute(%q) in multiline=%v = %v, want %v\nReason: %s",
					tt.query, tt.multiline, result, tt.shouldExec, tt.description)
			}
		})
	}
}

// TestShouldExecuteEdgeCases tests edge cases for shouldExecute
func TestShouldExecuteEdgeCases(t *testing.T) {
	originalMultiline := cmdconfig.Viper().GetBool(pconstants.ArgMultiLine)
	defer func() {
		cmdconfig.Viper().Set(pconstants.ArgMultiLine, originalMultiline)
	}()

	c := &InteractiveClient{}
	cmdconfig.Viper().Set(pconstants.ArgMultiLine, true)

	tests := []struct {
		name  string
		query string
	}{
		{
			name:  "very long query with semicolon",
			query: strings.Repeat("SELECT * FROM users WHERE id = 1 AND ", 100) + "1=1;",
		},
		{
			name:  "unicode characters with semicolon",
			query: "SELECT '你好世界';",
		},
		{
			name:  "emoji with semicolon",
			query: "SELECT '🔥💥';",
		},
		{
			name:  "null bytes",
			query: "SELECT '\x00';",
		},
		{
			name:  "control characters",
			query: "SELECT '\n\r\t';",
		},
		{
			name:  "SQL injection with semicolon",
			query: "'; DROP TABLE users; --",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("shouldExecute(%q) panicked: %v", tt.query, r)
				}
			}()

			_ = c.shouldExecute(tt.query)
		})
	}
}

// TestBreakMultilinePrompt tests the breakMultilinePrompt function
func TestBreakMultilinePrompt(t *testing.T) {
	c := &InteractiveClient{
		interactiveBuffer: []string{"SELECT *", "FROM users", "WHERE"},
	}

	c.breakMultilinePrompt(nil)

	if len(c.interactiveBuffer) != 0 {
		t.Errorf("breakMultilinePrompt() didn't clear buffer, got %d items, want 0", len(c.interactiveBuffer))
	}
}

// TestBreakMultilinePromptEmpty tests breaking an already empty buffer
func TestBreakMultilinePromptEmpty(t *testing.T) {
	c := &InteractiveClient{
		interactiveBuffer: []string{},
	}

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("breakMultilinePrompt() panicked on empty buffer: %v", r)
		}
	}()

	c.breakMultilinePrompt(nil)

	if len(c.interactiveBuffer) != 0 {
		t.Errorf("breakMultilinePrompt() didn't maintain empty buffer, got %d items, want 0", len(c.interactiveBuffer))
	}
}

// TestBreakMultilinePromptNil tests breaking with nil buffer
func TestBreakMultilinePromptNil(t *testing.T) {
	c := &InteractiveClient{
		interactiveBuffer: nil,
	}

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("breakMultilinePrompt() panicked on nil buffer: %v", r)
		}
	}()

	c.breakMultilinePrompt(nil)

	if c.interactiveBuffer == nil {
		t.Error("breakMultilinePrompt() didn't initialize nil buffer")
	}

	if len(c.interactiveBuffer) != 0 {
		t.Errorf("breakMultilinePrompt() didn't create empty buffer, got %d items, want 0", len(c.interactiveBuffer))
	}
}

// TestIsInitialised tests the isInitialised method
func TestIsInitialised(t *testing.T) {
	tests := []struct {
		name                   string
		initialisationComplete bool
		expected               bool
	}{
		{
			name:                   "initialized",
			initialisationComplete: true,
			expected:               true,
		},
		{
			name:                   "not initialized",
			initialisationComplete: false,
			expected:               false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &InteractiveClient{}
			c.initialisationComplete.Store(tt.initialisationComplete)

			result := c.isInitialised()

			if result != tt.expected {
				t.Errorf("isInitialised() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestClientNil tests the client() method when initData is nil
func TestClientNil(t *testing.T) {
	c := &InteractiveClient{
		initData: nil,
	}

	client := c.client()

	if client != nil {
		t.Errorf("client() with nil initData should return nil, got %v", client)
	}
}

// TestAfterPromptCloseAction tests the AfterPromptCloseAction enum
func TestAfterPromptCloseAction(t *testing.T) {
	// Test that the enum values are distinct
	if AfterPromptCloseExit == AfterPromptCloseRestart {
		t.Error("AfterPromptCloseExit and AfterPromptCloseRestart should have different values")
	}

	// Test that they have the expected values
	if AfterPromptCloseExit != 0 {
		t.Errorf("AfterPromptCloseExit should be 0, got %d", AfterPromptCloseExit)
	}

	if AfterPromptCloseRestart != 1 {
		t.Errorf("AfterPromptCloseRestart should be 1, got %d", AfterPromptCloseRestart)
	}
}

// TestGetFirstWordSuggestionsEmptyWord tests getFirstWordSuggestions with empty input
func TestGetFirstWordSuggestionsEmptyWord(t *testing.T) {
	c := &InteractiveClient{
		suggestions: newAutocompleteSuggestions(),
	}

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("getFirstWordSuggestions panicked on empty input: %v", r)
		}
	}()

	suggestions := c.getFirstWordSuggestions("")

	// Should return suggestions (select, with, metaqueries)
	if len(suggestions) == 0 {
		t.Error("getFirstWordSuggestions(\"\") should return suggestions")
	}
}

// TestGetFirstWordSuggestionsQualifiedQuery tests qualified query suggestions
func TestGetFirstWordSuggestionsQualifiedQuery(t *testing.T) {
	c := &InteractiveClient{
		suggestions: newAutocompleteSuggestions(),
	}

	// Add mock data
	c.suggestions.queriesByMod = map[string][]prompt.Suggest{
		"mymod": {
			{Text: "mymod.query1", Description: "Query"},
		},
	}

	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "qualified with known mod",
			input: "mymod.",
		},
		{
			name:  "qualified with unknown mod",
			input: "unknownmod.",
		},
		{
			name:  "single word",
			input: "select",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("getFirstWordSuggestions(%q) panicked: %v", tt.input, r)
				}
			}()

			suggestions := c.getFirstWordSuggestions(tt.input)

			if suggestions == nil {
				t.Errorf("getFirstWordSuggestions(%q) returned nil", tt.input)
			}
		})
	}
}

// TestGetTableAndConnectionSuggestionsEdgeCases tests edge cases
func TestGetTableAndConnectionSuggestionsEdgeCases(t *testing.T) {
	c := &InteractiveClient{
		suggestions: newAutocompleteSuggestions(),
	}

	// Add mock data
	c.suggestions.schemas = []prompt.Suggest{
		{Text: "public", Description: "Schema"},
	}
	c.suggestions.unqualifiedTables = []prompt.Suggest{
		{Text: "users", Description: "Table"},
	}
	c.suggestions.tablesBySchema = map[string][]prompt.Suggest{
		"public": {
			{Text: "public.users", Description: "Table"},
		},
	}

	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "unqualified",
			input: "users",
		},
		{
			name:  "qualified with known schema",
			input: "public.users",
		},
		{
			name:  "empty string",
			input: "",
		},
		{
			name:  "just dot",
			input: ".",
		},
		{
			name:  "unicode",
			input: "用户.表",
		},
		{
			name:  "emoji",
			input: "schema🔥.table",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("getTableAndConnectionSuggestions(%q) panicked: %v", tt.input, r)
				}
			}()

			suggestions := c.getTableAndConnectionSuggestions(tt.input)

			if suggestions == nil {
				t.Errorf("getTableAndConnectionSuggestions(%q) returned nil", tt.input)
			}
		})
	}
}

// TestCancelActiveQueryIfAny tests the cancellation logic
func TestCancelActiveQueryIfAny(t *testing.T) {
	t.Run("no active query", func(t *testing.T) {
		c := &InteractiveClient{
			cancelActiveQuery: nil,
		}

		defer func() {
			if r := recover(); r != nil {
				t.Errorf("cancelActiveQueryIfAny() panicked with nil cancelFunc: %v", r)
			}
		}()

		c.cancelActiveQueryIfAny()

		if c.cancelActiveQuery != nil {
			t.Error("cancelActiveQueryIfAny() set cancelActiveQuery when it was nil")
		}
	})

	t.Run("with active query", func(t *testing.T) {
		cancelled := false
		cancelFunc := func() {
			cancelled = true
		}

		c := &InteractiveClient{
			cancelActiveQuery: cancelFunc,
		}

		c.cancelActiveQueryIfAny()

		if !cancelled {
			t.Error("cancelActiveQueryIfAny() didn't call the cancel function")
		}

		if c.cancelActiveQuery != nil {
			t.Error("cancelActiveQueryIfAny() didn't set cancelActiveQuery to nil")
		}
	})

	t.Run("multiple calls", func(t *testing.T) {
		callCount := 0
		cancelFunc := func() {
			callCount++
		}

		c := &InteractiveClient{
			cancelActiveQuery: cancelFunc,
		}

		// First call should cancel
		c.cancelActiveQueryIfAny()

		if callCount != 1 {
			t.Errorf("First cancelActiveQueryIfAny() call count = %d, want 1", callCount)
		}

		// Second call should be a no-op
		c.cancelActiveQueryIfAny()

		if callCount != 1 {
			t.Errorf("Second cancelActiveQueryIfAny() call count = %d, want 1 (should be idempotent)", callCount)
		}
	})
}

// TestInitialisationComplete_RaceCondition tests that concurrent access to
// the initialisationComplete flag does not cause data races.
//
// This test simulates the real-world scenario where:
// - One goroutine (init goroutine) writes to initialisationComplete
// - Other goroutines (query executor, notification handler) read from it
//
// Bug: #4803
func TestInitialisationComplete_RaceCondition(t *testing.T) {
	c := &InteractiveClient{}
	c.initialisationComplete.Store(false)

	var wg sync.WaitGroup

	// Simulate initialization goroutine writing to the flag
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			c.initialisationComplete.Store(true)
			c.initialisationComplete.Store(false)
		}
	}()

	// Simulate query executor reading the flag
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			_ = c.isInitialised()
		}
	}()

	// Simulate notification handler reading the flag
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			// Check the flag directly (as handleConnectionUpdateNotification does)
			if !c.initialisationComplete.Load() {
				continue
			}
		}
	}()

	wg.Wait()
}

// TestGetQueryInfo_FromDetection tests that getQueryInfo correctly detects
// when the user is editing a table name after typing "from ".
//
// This is important for autocomplete - when a user types "from " (with a space),
// the system should recognize they are about to enter a table name and enable
// table suggestions. It should also remain true while typing a table name so
// that autocomplete can filter suggestions as the user types.
//
// Bug: #4810, #4928
func TestGetQueryInfo_FromDetection(t *testing.T) {
	tests := []struct {
		name              string
		input             string
		expectedTable     string
		expectedEditTable bool
	}{
		{
			name:              "just_from_with_space",
			input:             "from ",
			expectedTable:     "",
			expectedEditTable: true,
		},
		{
			name:              "from_typing_table",
			input:             "from my_table",
			expectedTable:     "my_table",
			expectedEditTable: true, // Still editing - prevWord is "from"
		},
		{
			name:              "from_keyword_only",
			input:             "from",
			expectedTable:     "",
			expectedEditTable: false,
		},
		{
			name:              "from_table_done",
			input:             "from my_table ",
			expectedTable:     "my_table",
			expectedEditTable: false, // Done editing - prevWord is now "my_table"
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getQueryInfo(tt.input)

			if result.Table != tt.expectedTable {
				t.Errorf("getQueryInfo(%q).Table = %q, expected %q", tt.input, result.Table, tt.expectedTable)
			}

			if result.EditingTable != tt.expectedEditTable {
				t.Errorf("getQueryInfo(%q).EditingTable = %v, expected %v", tt.input, result.EditingTable, tt.expectedEditTable)
			}
		})
	}
}

// TestExecuteMetaquery_NotInitialised tests that executeMetaquery returns
// an error instead of panicking when the client is not initialized.
//
// Bug: #4789
func TestExecuteMetaquery_NotInitialised(t *testing.T) {
	// Create an InteractiveClient that is not initialized
	c := &InteractiveClient{}
	c.initialisationComplete.Store(false)

	ctx := context.Background()

	// Attempt to execute a metaquery before initialization
	// This should return an error, not panic
	err := c.executeMetaquery(ctx, ".inspect")

	// We expect an error
	if err == nil {
		t.Error("Expected error when executing metaquery before initialization, but got nil")
	}

	// The test passes if we get here without a panic
	t.Logf("Successfully received error instead of panic: %v", err)
}
