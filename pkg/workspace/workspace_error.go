package workspace

import "errors"

var (
	ErrorNoModDefinition = errors.New("this command requires a mod definition file - could not find in the current directory tree")
)
