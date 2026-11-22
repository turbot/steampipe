package interactive

import (
	"bytes"
	"log"

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

// Highlight returns the document text formatted with colour control characters.
//
// If lexing or formatting fails, the failure is logged and returned with no
// output. go-prompt treats an error as "no highlighting" and renders the
// plain text the user typed, so a failure never garbles the prompt or stops
// the user typing or running a query.
func (h *Highlighter) Highlight(d prompt.Document) ([]byte, error) {
	buffer := bytes.NewBuffer([]byte{})
	tokens, err := h.lexer.Tokenise(nil, d.Text)
	if err != nil {
		log.Printf("[WARN] syntax highlighting failed, showing plain input: could not tokenise: %s", err)
		return nil, err
	}
	if err := h.formatter.Format(buffer, h.style, tokens); err != nil {
		// discard any partial output so the prompt falls back to the typed text
		log.Printf("[WARN] syntax highlighting failed, showing plain input: could not format: %s", err)
		return nil, err
	}
	return buffer.Bytes(), nil
}
