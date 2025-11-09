package interactive

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsFirstWord(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected bool
	}{
		"single word": {
			input:    "select",
			expected: true,
		},
		"two words": {
			input:    "select *",
			expected: false,
		},
		"multiple words": {
			input:    "select * from table",
			expected: false,
		},
		"empty string": {
			input:    "",
			expected: true,
		},
		"word with trailing space": {
			input:    "select ",
			expected: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := isFirstWord(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestLastWord(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected string
	}{
		"two words": {
			input:    "select *",
			expected: " *",
		},
		"multiple words": {
			input:    "select * from",
			expected: " from",
		},
		"with trailing space": {
			input:    "select * from ",
			expected: " ",
		},
		"table with schema": {
			input:    "select * from aws_ec2",
			expected: " aws_ec2",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := lastWord(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestGetTable(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected string
	}{
		"simple select from": {
			input:    "select * from users",
			expected: "users",
		},
		"qualified table name": {
			input:    "select * from aws_ec2.instances",
			expected: "aws_ec2.instances",
		},
		"from without table": {
			input:    "select * from",
			expected: "",
		},
		"no from clause": {
			input:    "select * where id = 1",
			expected: "",
		},
		"multiple spaces": {
			input:    "select  *  from  users",
			expected: "users",
		},
		"from in middle": {
			input:    "select * from users where name = 'test'",
			expected: "users",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := getTable(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestGetPreviousWord(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected string
	}{
		"simple two words": {
			input:    "select *",
			expected: "", // getPreviousWord returns the word BEFORE the last word
		},
		"three words": {
			input:    "select * from",
			expected: "*",
		},
		"four words": {
			input:    "select * from users",
			expected: "from",
		},
		"single word": {
			input:    "select",
			expected: "",
		},
		"with trailing spaces": {
			input:    "select *   ",
			expected: "*", // Trailing spaces are trimmed, so previous word is "*"
		},
		"empty string": {
			input:    "",
			expected: "",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := getPreviousWord(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsEditingTable(t *testing.T) {
	tests := map[string]struct {
		prevWord string
		expected bool
	}{
		"from keyword": {
			prevWord: "from",
			expected: true,
		},
		"FROM uppercase": {
			prevWord: "FROM",
			expected: false, // The function checks lowercase only
		},
		"select keyword": {
			prevWord: "select",
			expected: false,
		},
		"where keyword": {
			prevWord: "where",
			expected: false,
		},
		"empty string": {
			prevWord: "",
			expected: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := isEditingTable(tc.prevWord)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestGetQueryInfo(t *testing.T) {
	tests := map[string]struct {
		input                string
		expectedTable        string
		expectedEditingTable bool
	}{
		"editing table after from": {
			input:                "select * from ",
			expectedTable:        "",
			expectedEditingTable: true,
		},
		"with table name": {
			input:                "select * from users",
			expectedTable:        "users",
			expectedEditingTable: true, // Previous word is "from", so we're still editing table
		},
		"qualified table name": {
			input:                "select * from aws_ec2.instances",
			expectedTable:        "aws_ec2.instances",
			expectedEditingTable: true, // Previous word is "from", so we're still editing table
		},
		"no from clause": {
			input:                "select *",
			expectedTable:        "",
			expectedEditingTable: false,
		},
		"from in where clause": {
			input:                "select * from users where name = 'from'",
			expectedTable:        "users",
			expectedEditingTable: false, // Previous word is not "from"
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := getQueryInfo(tc.input)
			assert.Equal(t, tc.expectedTable, result.Table)
			assert.Equal(t, tc.expectedEditingTable, result.EditingTable)
		})
	}
}

func TestLastIndexByteNot(t *testing.T) {
	tests := map[string]struct {
		input    string
		char     byte
		expected int
	}{
		"no trailing spaces": {
			input:    "hello",
			char:     ' ',
			expected: 4,
		},
		"with trailing spaces": {
			input:    "hello   ",
			char:     ' ',
			expected: 4,
		},
		"all spaces": {
			input:    "   ",
			char:     ' ',
			expected: -1,
		},
		"empty string": {
			input:    "",
			char:     ' ',
			expected: -1,
		},
		"single non-space char": {
			input:    "a",
			char:     ' ',
			expected: 0,
		},
		"single space": {
			input:    " ",
			char:     ' ',
			expected: -1,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := lastIndexByteNot(tc.input, tc.char)
			assert.Equal(t, tc.expected, result)
		})
	}
}
