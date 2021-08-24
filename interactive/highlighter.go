package interactive

import (
	"bytes"
	"strings"

	"github.com/alecthomas/chroma"
	"github.com/c-bata/go-prompt"
)

type Highlighter struct {
	lexer     chroma.Lexer
	formatter chroma.Formatter
	style     *chroma.Style
}

func newHighlighter(lexer chroma.Lexer, formatter chroma.Formatter, style *chroma.Style) Highlighter {
	return Highlighter{
		lexer:     lexer,
		formatter: formatter,
		style:     style,
	}
}

func (h *Highlighter) Highlight(d prompt.Document) ([]byte, error) {
	leftIterator, err := h.lexer.Tokenise(nil, strings.TrimSuffix(d.TextBeforeCursor(), d.GetWordBeforeCursor()))
	if err != nil {
		return nil, err
	}
	rightIterator, err := h.lexer.Tokenise(nil, strings.TrimPrefix(d.TextAfterCursor(), d.GetWordAfterCursor()))
	if err != nil {
		return nil, err
	}

	buffer := bytes.NewBuffer([]byte{})

	// format till the second last word before the cursor
	h.formatter.Format(buffer, h.style, leftIterator)

	// leave the last word before the cursor unformatted
	buffer.WriteString(d.GetWordBeforeCursor())
	// and the next word as well
	buffer.WriteString(d.GetWordAfterCursor())

	// format the rest
	h.formatter.Format(buffer, h.style, rightIterator)

	return buffer.Bytes(), nil
}
