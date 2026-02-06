package interactive

import (
	"strings"
	"testing"
)

// TestIsFirstWord tests the isFirstWord helper function
func TestIsFirstWord(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "single word",
			input:    "select",
			expected: true,
		},
		{
			name:     "two words",
			input:    "select *",
			expected: false,
		},
		{
			name:     "empty string",
			input:    "",
			expected: true,
		},
		{
			name:     "word with trailing space",
			input:    "select ",
			expected: false,
		},
		{
			name:     "multiple spaces",
			input:    "select  from",
			expected: false,
		},
		{
			name:     "unicode characters",
			input:    "選択",
			expected: true,
		},
		{
			name:     "emoji",
			input:    "🔥",
			expected: true,
		},
		{
			name:     "emoji with space",
			input:    "🔥 test",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isFirstWord(tt.input)
			if result != tt.expected {
				t.Errorf("isFirstWord(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

// TestLastWord tests the lastWord helper function
// Bug: #4787 - lastWord() panics on single word or empty string
func TestLastWord(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "two words",
			input:    "select *",
			expected: " *",
		},
		{
			name:     "multiple words",
			input:    "select * from users",
			expected: " users",
		},
		{
			name:     "trailing space",
			input:    "select * from ",
			expected: " ",
		},
		{
			name:     "unicode",
			input:    "select 你好",
			expected: " 你好",
		},
		{
			name:     "emoji",
			input:    "select 🔥",
			expected: " 🔥",
		},
		{
			name:     "single_word", // #4787
			input:    "select",
			expected: "select",
		},
		{
			name:     "empty_string", // #4787
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("lastWord(%q) panicked: %v", tt.input, r)
				}
			}()

			result := lastWord(tt.input)
			if result != tt.expected {
				t.Errorf("lastWord(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestLastIndexByteNot tests the lastIndexByteNot helper function
func TestLastIndexByteNot(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		char     byte
		expected int
	}{
		{
			name:     "no matching char",
			input:    "hello",
			char:     ' ',
			expected: 4,
		},
		{
			name:     "trailing spaces",
			input:    "hello   ",
			char:     ' ',
			expected: 4,
		},
		{
			name:     "all spaces",
			input:    "     ",
			char:     ' ',
			expected: -1,
		},
		{
			name:     "empty string",
			input:    "",
			char:     ' ',
			expected: -1,
		},
		{
			name:     "single char not matching",
			input:    "a",
			char:     ' ',
			expected: 0,
		},
		{
			name:     "single char matching",
			input:    " ",
			char:     ' ',
			expected: -1,
		},
		{
			name:     "mixed spaces",
			input:    "hello world  ",
			char:     ' ',
			expected: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := lastIndexByteNot(tt.input, tt.char)
			if result != tt.expected {
				t.Errorf("lastIndexByteNot(%q, %q) = %d, want %d", tt.input, tt.char, result, tt.expected)
			}
		})
	}
}

// TestGetPreviousWord tests the getPreviousWord helper function
func TestGetPreviousWord(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple case",
			input:    "select * from ",
			expected: "from",
		},
		{
			name:     "single word with trailing space",
			input:    "select ",
			expected: "select",
		},
		{
			name:     "single word",
			input:    "select",
			expected: "",
		},
		{
			name:     "multiple spaces",
			input:    "select  *  from  ",
			expected: "from",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "only spaces",
			input:    "   ",
			expected: "",
		},
		{
			name:     "unicode characters",
			input:    "select 你好 世界 ",
			expected: "世界",
		},
		{
			name:     "emoji",
			input:    "select 🔥 💥 ",
			expected: "💥",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getPreviousWord(tt.input)
			if result != tt.expected {
				t.Errorf("getPreviousWord(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestGetTable tests the getTable helper function
func TestGetTable(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple select",
			input:    "select * from users",
			expected: "users",
		},
		{
			name:     "qualified table",
			input:    "select * from public.users",
			expected: "public.users",
		},
		{
			name:     "no from clause",
			input:    "select 1",
			expected: "",
		},
		{
			name:     "from at end",
			input:    "select * from",
			expected: "",
		},
		{
			name:     "from with trailing text",
			input:    "select * from users where",
			expected: "users",
		},
		{
			name:     "double spaces",
			input:    "select  *  from  users",
			expected: "users",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "case sensitive - lowercase from",
			input:    "SELECT * from users",
			expected: "users",
		},
		{
			name:     "uppercase FROM",
			input:    "SELECT * FROM users",
			expected: "",
		},
		{
			name:     "unicode table name",
			input:    "select * from 用户表",
			expected: "用户表",
		},
		{
			name:     "emoji in table name",
			input:    "select * from users🔥",
			expected: "users🔥",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getTable(tt.input)
			if result != tt.expected {
				t.Errorf("getTable(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestIsEditingTable tests the isEditingTable helper function
func TestIsEditingTable(t *testing.T) {
	tests := []struct {
		name     string
		prevWord string
		expected bool
	}{
		{
			name:     "from keyword",
			prevWord: "from",
			expected: true,
		},
		{
			name:     "not from keyword",
			prevWord: "select",
			expected: false,
		},
		{
			name:     "empty string",
			prevWord: "",
			expected: false,
		},
		{
			name:     "FROM uppercase",
			prevWord: "FROM",
			expected: false,
		},
		{
			name:     "whitespace",
			prevWord: " from ",
			expected: false,
		},
		{
			name:     "table name after from",
			prevWord: "aws_s3_bucket",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isEditingTable(tt.prevWord)
			if result != tt.expected {
				t.Errorf("isEditingTable(%q) = %v, want %v", tt.prevWord, result, tt.expected)
			}
		})
	}
}

// TestGetQueryInfo tests the getQueryInfo function
// Bug: #4928 - autocomplete suggestions disappear when typing table name after 'from '
func TestGetQueryInfo(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedTable   string
		expectedEditing bool
	}{
		{
			name:            "editing table after from",
			input:           "select * from ",
			expectedTable:   "",
			expectedEditing: true,
		},
		{
			name:            "typing table name after from",
			input:           "select * from aws",
			expectedTable:   "aws",
			expectedEditing: true,
		},
		{
			name:            "typing partial table name",
			input:           "select * from aws_s3",
			expectedTable:   "aws_s3",
			expectedEditing: true,
		},
		{
			name:            "typing qualified table name",
			input:           "select * from aws.aws_s3_bucket",
			expectedTable:   "aws.aws_s3_bucket",
			expectedEditing: true,
		},
		{
			name:            "table specified with trailing space",
			input:           "select * from users ",
			expectedTable:   "users",
			expectedEditing: false,
		},
		{
			name:            "past table into where clause",
			input:           "select * from users where",
			expectedTable:   "users",
			expectedEditing: false,
		},
		{
			name:            "not at from clause",
			input:           "select * ",
			expectedTable:   "",
			expectedEditing: false,
		},
		{
			name:            "empty query",
			input:           "",
			expectedTable:   "",
			expectedEditing: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getQueryInfo(tt.input)
			if result.Table != tt.expectedTable {
				t.Errorf("getQueryInfo(%q).Table = %q, want %q", tt.input, result.Table, tt.expectedTable)
			}
			if result.EditingTable != tt.expectedEditing {
				t.Errorf("getQueryInfo(%q).EditingTable = %v, want %v", tt.input, result.EditingTable, tt.expectedEditing)
			}
		})
	}
}

// TestCleanBufferForWSL tests the WSL-specific buffer cleaning
func TestCleanBufferForWSL(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedOutput string
		expectedIgnore bool
	}{
		{
			name:           "normal text",
			input:          "hello",
			expectedOutput: "hello",
			expectedIgnore: false,
		},
		{
			name:           "empty string",
			input:          "",
			expectedOutput: "",
			expectedIgnore: false,
		},
		{
			name:           "escape sequence",
			input:          string([]byte{27, 65}), // ESC + 'A'
			expectedOutput: "",
			expectedIgnore: true,
		},
		{
			name:           "single escape",
			input:          string([]byte{27}),
			expectedOutput: string([]byte{27}),
			expectedIgnore: false,
		},
		{
			name:           "unicode",
			input:          "你好",
			expectedOutput: "你好",
			expectedIgnore: false,
		},
		{
			name:           "emoji",
			input:          "🔥",
			expectedOutput: "🔥",
			expectedIgnore: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, ignore := cleanBufferForWSL(tt.input)
			if output != tt.expectedOutput {
				t.Errorf("cleanBufferForWSL(%q) output = %q, want %q", tt.input, output, tt.expectedOutput)
			}
			if ignore != tt.expectedIgnore {
				t.Errorf("cleanBufferForWSL(%q) ignore = %v, want %v", tt.input, ignore, tt.expectedIgnore)
			}
		})
	}
}

// TestSanitiseTableName tests table name escaping (passing cases only)
func TestSanitiseTableName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple lowercase table",
			input:    "users",
			expected: "users",
		},
		{
			name:     "uppercase table",
			input:    "Users",
			expected: `"Users"`,
		},
		{
			name:     "table with space",
			input:    "user data",
			expected: `"user data"`,
		},
		{
			name:     "table with hyphen",
			input:    "user-data",
			expected: `"user-data"`,
		},
		{
			name:     "qualified table",
			input:    "schema.table",
			expected: "schema.table",
		},
		{
			name:     "qualified with uppercase",
			input:    "Schema.Table",
			expected: `"Schema"."Table"`,
		},
		{
			name:     "qualified with spaces",
			input:    "my schema.my table",
			expected: `"my schema"."my table"`,
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
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

// TestHelperFunctionsWithExtremeInput tests helper functions with extreme inputs
func TestHelperFunctionsWithExtremeInput(t *testing.T) {
	t.Run("very long string", func(t *testing.T) {
		longString := strings.Repeat("a ", 10000)

		// Test that these don't panic or hang
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Function panicked on long string: %v", r)
			}
		}()

		_ = isFirstWord(longString)
		_ = getTable(longString)
		_ = getPreviousWord(longString)
		_ = getQueryInfo(longString)
	})

	t.Run("null bytes", func(t *testing.T) {
		nullByteString := "select\x00from\x00users"

		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Function panicked on null bytes: %v", r)
			}
		}()

		_ = isFirstWord(nullByteString)
		_ = getTable(nullByteString)
		_ = getPreviousWord(nullByteString)
	})

	t.Run("control characters", func(t *testing.T) {
		controlString := "select\n\r\tfrom\n\rusers"

		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Function panicked on control chars: %v", r)
			}
		}()

		_ = isFirstWord(controlString)
		_ = getTable(controlString)
		_ = getPreviousWord(controlString)
	})

	t.Run("SQL injection attempts", func(t *testing.T) {
		injectionStrings := []string{
			"'; DROP TABLE users; --",
			"1' OR '1'='1",
			"1; DELETE FROM connections; --",
			"select * from users where id = 1' union select * from passwords --",
		}

		for _, injection := range injectionStrings {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Function panicked on injection string %q: %v", injection, r)
				}
			}()

			_ = isFirstWord(injection)
			_ = getTable(injection)
			_ = getPreviousWord(injection)
			_ = getQueryInfo(injection)
		}
	})
}
