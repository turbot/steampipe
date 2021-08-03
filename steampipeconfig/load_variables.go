package steampipeconfig

import (
	"fmt"
	"os"

	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/steampipeconfig/parse"
)

func LoadVariables(workspacePath string, opts *parse.ParseModOptions) (variables map[string]*modconfig.Variable, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = helpers.ToError(r)
		}
	}()

	// only parse variables
	opts.BlockTypes = []string{modconfig.BlockTypeVariable}

	// verify the mod folder exists
	if _, err := os.Stat(workspacePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("source folder %s does not exist", workspacePath)
	}

	// now parse the mod, passing the pseudo resources
	// load the raw data
	mod, err := parseMod(workspacePath, nil, opts)
	if err != nil {
		return nil, err
	}

	return mod.Variables, nil
}
