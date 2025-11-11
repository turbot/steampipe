package interactive

import (
	"testing"
)

// TestSanitiseTableName tests that sanitiseTableName properly escapes table names
// that require PostgreSQL identifier escaping.
//
// PostgreSQL requires escaping for:
// - Names with spaces
// - Names with hyphens
// - Names with uppercase letters (to preserve case)
// - Names with unicode/emoji characters
//
// Bug #4801: sanitiseTableName doesn't escape unicode/emoji
func TestSanitiseTableName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple lowercase",
			input:    "users",
			expected: "users",
		},
		{
			name:     "uppercase requires escaping",
			input:    "Users",
			expected: `"Users"`,
		},
		{
			name:     "space requires escaping",
			input:    "user data",
			expected: `"user data"`,
		},
		{
			name:     "hyphen requires escaping",
			input:    "user-data",
			expected: `"user-data"`,
		},
		{
			name:     "qualified table with schema",
			input:    "public.Users",
			expected: `public."Users"`,
		},
		{
			name:     "unicode table name",
			input:    "用户",
			expected: `"用户"`,
		},
		{
			name:     "emoji in table name",
			input:    "table_😀_data",
			expected: `"table_😀_data"`,
		},
		{
			name:     "qualified with unicode",
			input:    "schema.用户表",
			expected: `schema."用户表"`,
		},
		{
			name:     "mixed unicode and ascii",
			input:    "données_utilisateur",
			expected: `"données_utilisateur"`,
		},
		{
			name:     "cyrillic characters",
			input:    "таблица",
			expected: `"таблица"`,
		},
		{
			name:     "arabic characters",
			input:    "جدول",
			expected: `"جدول"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitiseTableName(tt.input)
			if result != tt.expected {
				t.Errorf("sanitiseTableName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
