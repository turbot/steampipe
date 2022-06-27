package utils

import (
	"encoding/csv"
	"strings"
)

// SplitByRune uses the CSV decoder to parse out the tokens - even if they are quoted and/or escaped
func SplitByRune(str string, r rune) []string {
	csvDecoder := csv.NewReader(strings.NewReader(str))
	csvDecoder.Comma = r
	csvDecoder.LazyQuotes = true
	csvDecoder.TrimLeadingSpace = true
	split, _ := csvDecoder.Read()
	return split
}

// SplitByWhitespace splits by the ' ' rune
func SplitByWhitespace(str string) []string {
	return SplitByRune(str, ' ')
}
