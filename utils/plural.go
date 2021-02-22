package utils

import (
	"github.com/gertd/go-pluralize"
)

// Pluralize :: pluralizes a word (if applicable) based on provided count
func Pluralize(base string, count int) string {
	pluralizer := pluralize.NewClient()
	pluralizer.AddIrregularRule("it", "they")
	return pluralizer.Pluralize(base, count, false)
}
