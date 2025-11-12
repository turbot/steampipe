package interactive

import (
	"strings"
	"testing"

	"github.com/alecthomas/chroma/formatters"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
	"github.com/c-bata/go-prompt"
)

// TestNewHighlighter tests highlighter creation
func TestNewHighlighter(t *testing.T) {
	lexer := lexers.Get("sql")
	formatter := formatters.Get("terminal256")
	style := styles.Native

	h := newHighlighter(lexer, formatter, style)

	if h == nil {
		t.Fatal("newHighlighter returned nil")
	}

	if h.lexer == nil {
		t.Error("highlighter lexer is nil")
	}

	if h.formatter == nil {
		t.Error("highlighter formatter is nil")
	}

	if h.style == nil {
		t.Error("highlighter style is nil")
	}
}

// TestHighlighterHighlight tests the Highlight function
func TestHighlighterHighlight(t *testing.T) {
	h := newHighlighter(
		lexers.Get("sql"),
		formatters.Get("terminal256"),
		styles.Native,
	)

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "simple select",
			input:   "SELECT * FROM users",
			wantErr: false,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: false,
		},
		{
			name:    "multiline query",
			input:   "SELECT *\nFROM users\nWHERE id = 1",
			wantErr: false,
		},
		{
			name:    "unicode characters",
			input:   "SELECT '你好世界'",
			wantErr: false,
		},
		{
			name:    "emoji",
			input:   "SELECT '🔥💥✨'",
			wantErr: false,
		},
		{
			name:    "null bytes",
			input:   "SELECT '\x00'",
			wantErr: false,
		},
		{
			name:    "control characters",
			input:   "SELECT '\n\r\t'",
			wantErr: false,
		},
		{
			name:    "very long query",
			input:   "SELECT " + strings.Repeat("a, ", 1000) + "* FROM users",
			wantErr: false,
		},
		{
			name:    "SQL injection attempt",
			input:   "'; DROP TABLE users; --",
			wantErr: false,
		},
		{
			name:    "malformed SQL",
			input:   "SELECT FROM WHERE",
			wantErr: false,
		},
		{
			name:    "special characters",
			input:   "SELECT '\\', '/', '\"', '`'",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := prompt.Document{
				Text: tt.input,
			}

			result, err := h.Highlight(doc)

			if (err != nil) != tt.wantErr {
				t.Errorf("Highlight() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && result == nil {
				t.Error("Highlight() returned nil result without error")
			}

			// Verify result is not empty for non-empty input
			if !tt.wantErr && tt.input != "" && len(result) == 0 {
				t.Error("Highlight() returned empty result for non-empty input")
			}
		})
	}
}

// TestGetHighlighter tests the getHighlighter function
func TestGetHighlighter(t *testing.T) {
	tests := []struct {
		name  string
		theme string
	}{
		{
			name:  "default theme",
			theme: "",
		},
		{
			name:  "dark theme",
			theme: "dark",
		},
		{
			name:  "light theme",
			theme: "light",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := getHighlighter(tt.theme)

			if h == nil {
				t.Fatal("getHighlighter returned nil")
			}

			if h.lexer == nil {
				t.Error("highlighter lexer is nil")
			}

			if h.formatter == nil {
				t.Error("highlighter formatter is nil")
			}
		})
	}
}

// TestHighlighterConcurrency tests concurrent highlighting
func TestHighlighterConcurrency(t *testing.T) {
	h := newHighlighter(
		lexers.Get("sql"),
		formatters.Get("terminal256"),
		styles.Native,
	)

	queries := []string{
		"SELECT * FROM users",
		"SELECT id FROM posts",
		"SELECT name FROM companies",
	}

	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func(idx int) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Concurrent Highlight panicked: %v", r)
				}
				done <- true
			}()

			doc := prompt.Document{
				Text: queries[idx%len(queries)],
			}

			_, err := h.Highlight(doc)
			if err != nil {
				t.Errorf("Concurrent Highlight error: %v", err)
			}
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

// TestHighlighterMemoryLeak tests for memory leaks with repeated highlighting
func TestHighlighterMemoryLeak(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory leak test in short mode")
	}

	h := newHighlighter(
		lexers.Get("sql"),
		formatters.Get("terminal256"),
		styles.Native,
	)

	// Highlight the same query many times to check for memory leaks
	doc := prompt.Document{
		Text: "SELECT * FROM users WHERE id = 1",
	}

	for i := 0; i < 10000; i++ {
		_, err := h.Highlight(doc)
		if err != nil {
			t.Fatalf("Highlight failed at iteration %d: %v", i, err)
		}
	}

	// If we get here without OOM, the test passes
}
