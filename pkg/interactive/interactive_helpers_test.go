package interactive

import "testing"

// TestLastWord tests the lastWord function for various inputs
// Bug: #4787 - lastWord() panics on single word or empty string
func TestLastWord(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "multiple words",
			input:    "select * from",
			expected: " from",
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
		{
			name:     "trailing space",
			input:    "select from ",
			expected: " ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := lastWord(tt.input)
			if result != tt.expected {
				t.Errorf("lastWord(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}
