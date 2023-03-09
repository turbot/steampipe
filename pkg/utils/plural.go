package utils

import (
	"github.com/gertd/go-pluralize"
)

// Pluralize :: pluralizes a word (if applicable) based on provided count
func Pluralize(base string, count int) string {
	pluralizer := pluralize.NewClient()
	pluralizer.AddIrregularRule("it", "them")
	pluralizer.AddIrregularRule("a", "")
	pluralizer.AddIrregularRule("is", "are")
	pluralizer.AddIrregularRule("it is", "they are")
	pluralizer.AddIrregularRule("this plugin", "these plugins")
	return pluralizer.Pluralize(base, count, false)
}
