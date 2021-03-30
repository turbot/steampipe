package modconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Machiel/slugify"
	"github.com/turbot/steampipe/constants"
)

// map of file extension to factory function to create
type factoryFunc func(path string) (MappableResource, error)

var ResourceTypeMap = map[string]factoryFunc{
	constants.ExtensionSql: func(path string) (MappableResource, error) { return QueryFromFile(path) },
}

func RegisteredFileExtensions() []string {
	var res []string
	for ext := range ResourceTypeMap {
		res = append(res, ext)
	}
	return res
}

// ResourceNameFromPath :: convert a filpath into a resource name:
// 1) convert into a relative path from the working folder
// 2) remove extension
// 3) sluggify, with '_' as the divider
func PseudoResourceNameFromPath(path string) (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory when converting sql files to query resources: %v", err)
	}
	relativePath, err := filepath.Rel(wd, path)
	if err != nil {
		return "", fmt.Errorf("failed to get relative path of sql file %s: %v", path, err)
	}
	// remove the extension
	relativePath = strings.TrimSuffix(relativePath, filepath.Ext(path))

	// now slugify this
	slugifier := slugify.New(slugify.Configuration{
		ReplaceCharacter: '_',
	})

	return slugifier.Slugify(relativePath), nil
}
