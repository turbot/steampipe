package interactive

import (
	"bytes"
	"errors"
	"io"
	"log"
	"strings"
	"testing"

	"github.com/alecthomas/chroma"
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

type failingFormatter struct{ err error }

// Format writes partial output, then fails
func (f failingFormatter) Format(w io.Writer, _ *chroma.Style, _ chroma.Iterator) error {
	_, _ = w.Write([]byte("partial"))
	return f.err
}

type failingLexer struct {
	chroma.Lexer
	err error
}

func (l failingLexer) Tokenise(*chroma.TokeniseOptions, string) (chroma.Iterator, error) {
	return nil, l.err
}

// captureLog redirects the standard logger for the test and returns what was written
func captureLog(t *testing.T) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	orig := log.Writer()
	log.SetOutput(&buf)
	t.Cleanup(func() { log.SetOutput(orig) })
	return &buf
}

// TestHighlighterFormatterFailure: a formatter error returns no partial output (so the
// prompt falls back to the typed text) and is logged with its reason
func TestHighlighterFormatterFailure(t *testing.T) {
	logs := captureLog(t)
	h := newHighlighter(lexers.Get("sql"), failingFormatter{err: errors.New("formatter boom")}, styles.Native)

	result, err := h.Highlight(prompt.Document{Text: "SELECT 1"})

	if err == nil || !strings.Contains(err.Error(), "formatter boom") {
		t.Fatalf("expected the formatter error, got %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected no output on failure, got %q", result)
	}
	if !strings.Contains(logs.String(), "[WARN]") || !strings.Contains(logs.String(), "formatter boom") {
		t.Errorf("expected the failure to be logged with its reason, got %q", logs.String())
	}
}

// TestHighlighterLexerFailure: a lexer error returns no output and is logged with its reason
func TestHighlighterLexerFailure(t *testing.T) {
	logs := captureLog(t)
	h := newHighlighter(failingLexer{err: errors.New("lexer boom")}, formatters.Get("terminal256"), styles.Native)

	result, err := h.Highlight(prompt.Document{Text: "SELECT 1"})

	if err == nil || !strings.Contains(err.Error(), "lexer boom") {
		t.Fatalf("expected the lexer error, got %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected no output on failure, got %q", result)
	}
	if !strings.Contains(logs.String(), "lexer boom") {
		t.Errorf("expected the failure to be logged with its reason, got %q", logs.String())
	}
}

// TestHighlighterSuccessLogsNothing: working highlighting is unchanged and silent
func TestHighlighterSuccessLogsNothing(t *testing.T) {
	logs := captureLog(t)
	h := newHighlighter(lexers.Get("sql"), formatters.Get("terminal256"), styles.Native)

	result, err := h.Highlight(prompt.Document{Text: "SELECT 1"})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(string(result), "SELECT") || !strings.Contains(string(result), "\x1b[") {
		t.Errorf("expected coloured output, got %q", result)
	}
	if logs.Len() != 0 {
		t.Errorf("expected no log output, got %q", logs.String())
	}
}
