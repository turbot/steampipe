package modconfig

import (
	"path/filepath"
	"strings"

	"github.com/Machiel/slugify"
	"github.com/turbot/steampipe/constants"
)

// map of file extension to factory function to create
type factoryFunc func(modPath, filePath string) (MappableResource, []byte, error)

var ResourceTypeMap = map[string]factoryFunc{
	constants.SqlExtension: func(modPath, filePath string) (MappableResource, []byte, error) {
		return QueryFromFile(modPath, filePath)
	},
}

func RegisteredFileExtensions() []string {
	var res []string
	for ext := range ResourceTypeMap {
		res = append(res, ext)
	}
	return res
}

// PseudoResourceNameFromPath converts  a filepath into a resource name
//
// It operates as follows:
// 	1) get filename
// 	2) remove extension
// 	3) sluggify, with '_' as the divider
func PseudoResourceNameFromPath(modPath, filePath string) (string, error) {
	filename := filepath.Base(filePath)
	// remove the extension
	filename = strings.TrimSuffix(filename, filepath.Ext(filePath))

	// now slugify this
	slugifier := slugify.New(slugify.Configuration{
		ReplaceCharacter: '_',
	})

	return slugifier.Slugify(filename), nil
}
