package interactive

import (
	"bytes"

	"github.com/alecthomas/chroma"
	"github.com/c-bata/go-prompt"
)

type Highlighter struct {
	lexer     chroma.Lexer
	formatter chroma.Formatter
	style     *chroma.Style
}

func newHighlighter(lexer chroma.Lexer, formatter chroma.Formatter, style *chroma.Style) *Highlighter {
	h := new(Highlighter)
	h.formatter = formatter
	h.lexer = lexer
	h.style = style
	return h
}

func (h *Highlighter) Highlight(d prompt.Document) ([]byte, error) {
	buffer := bytes.NewBuffer([]byte{})
	tokens, err := h.lexer.Tokenise(nil, d.Text)
	if err != nil {
		return nil, err
	}
	h.formatter.Format(buffer, h.style, tokens)
	return buffer.Bytes(), nil
}
