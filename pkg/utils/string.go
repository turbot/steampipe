package utils

import (
	"bytes"
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

// Resize resizes the string with the given length. It ellipses with '…' when the string's length exceeds
// the desired length or pads spaces to the right of the string when length is smaller than desired
func Resize(s string, length uint) string {
	n := int(length)
	if len(s) == n {
		return s
	}
	// Pads only when length of the string smaller than len needed
	s = PadRight(s, n, ' ')
	if len(s) > n {
		b := []byte(s)
		var buf bytes.Buffer
		for i := 0; i < n-1; i++ {
			buf.WriteByte(b[i])
		}
		buf.WriteString("…")
		s = buf.String()
	}
	return s
}

// PadRight returns a new string of a specified length in which the end of the current string is padded with spaces or with a specified Unicode character.
func PadRight(str string, length int, pad byte) string {
	if len(str) >= length {
		return str
	}
	buf := bytes.NewBufferString(str)
	for i := 0; i < length-len(str); i++ {
		buf.WriteByte(pad)
	}
	return buf.String()
}
