package modconfig

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Machiel/slugify"
	"github.com/turbot/steampipe/constants"
)

// map of file extension to factory function to create
type factoryFunc func(modPath, filePath string) (MappableResource, error)

var ResourceTypeMap = map[string]factoryFunc{
	constants.SqlExtension: func(modPath, filePath string) (MappableResource, error) { return QueryFromFile(modPath, filePath) },
}

func RegisteredFileExtensions() []string {
	var res []string
	for ext := range ResourceTypeMap {
		res = append(res, ext)
	}
	return res
}

// ResourceNameFromPath :: convert a filepath into a resource name:
// 1) convert into a relative path from the working folder
// 2) remove extension
// 3) sluggify, with '_' as the divider
func PseudoResourceNameFromPath(modPath, filePath string) (string, error) {
	relativePath, err := filepath.Rel(modPath, filePath)
	if err != nil {
		return "", fmt.Errorf("failed to get relative path of sql file '%s' with base folder '%s': %v", filePath, modPath, err)
	}
	// remove the extension
	relativePath = strings.TrimSuffix(relativePath, filepath.Ext(filePath))

	// now slugify this
	slugifier := slugify.New(slugify.Configuration{
		ReplaceCharacter: '_',
	})

	return slugifier.Slugify(relativePath), nil
}
