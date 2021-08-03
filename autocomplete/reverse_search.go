package autocomplete

import (
	"fmt"
	"unicode/utf8"
)

type ReverseSearch struct {
	Search       string
	currentIndex int
	Inputs       []string
	Matches      []string
}

func NewReverseSearch(searchString string, inputStrings []string) ReverseSearch {
	r := ReverseSearch{
		Search:       searchString,
		currentIndex: 0,
		Inputs:       inputStrings,
	}
	r.updateMatches()
	return r
}

func (r *ReverseSearch) updateMatches()    {}
func (r *ReverseSearch) Current() string   { return "" }
func (r *ReverseSearch) GoToPrevious()     {}
func (r *ReverseSearch) GoToNext()         {}
func (r *ReverseSearch) GetPrompt() string { return "" }

func (r *ReverseSearch) PushCharacters(chars string) {
	r.SetSearch(fmt.Sprintf("%s%s", r.Search, chars))
}
func (r *ReverseSearch) PopLastCharacter() {
	if len(r.Search) > 0 {
		_, size := utf8.DecodeLastRuneInString(r.Search)
		r.SetSearch(r.Search[:len(r.Search)-size])
	}
}
func (r *ReverseSearch) SetSearch(search string) { r.Search = search; r.updateMatches() }
