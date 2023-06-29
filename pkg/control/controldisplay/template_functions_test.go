package controldisplay

import (
	"testing"
)

func BenchmarkToCsvCell(b *testing.B) {
	// the factory is called once per render execution
	toCsvCell := toCSVCellFnFactory("|")
	for i := 0; i < b.N; i++ {
		toCsvCell(i)
	}
}


func TestSafeFragmentId(t *testing.T) {
	cases := []struct {
		input, want string
	}{
		{"This is a test", "this-is-a-test"},
		{"Thîs ïs à Unicôde Strïng", "this-is-a-unicode-string"},
		{"With special & characters!", "with-special-and-characters"},
		{"", "id"}, // testing empty string, the function should return a default "id"
		{"!!!!", "id"}, // testing string with characters that will be removed
		{"'Single' & \"Double\" quotes", "single-and-double-quotes"},
		{"   Trim spaces   ", "trim-spaces"},
		{"-Trim-hyphens--", "trim-hyphens"},
		{"Mixed CASE", "mixed-case"},
		{"Spaces and_underscores", "spaces-and-underscores"},
		{"Trailing hyphen-", "trailing-hyphen"},
		{"_Leading and trailing underscore_", "leading-and-trailing-underscore"},
		{"01234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789", "0123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789"},
		{"Non-ASCII characters like éçà", "non-ascii-characters-like-eca"},
		{"Multiple spaces   in   a   row", "multiple-spaces-in-a-row"},
		{"Hyphens--in--a--row", "hyphens-in-a-row"},
		{"Underscores__in__a__row", "underscores-in-a-row"},
		{"-_-Leading and trailing hyphens and underscores -_-", "leading-and-trailing-hyphens-and-underscores"},
}

	for _, c := range cases {
		got := safeFragmentId(c.input)
		if got != c.want {
			t.Errorf("safeFragmentId(%q) == %q, want %q", c.input, got, c.want)
		}
	}
}