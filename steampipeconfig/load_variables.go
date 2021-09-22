package steampipeconfig

import (
	"fmt"
	"os"
	"strings"

	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/steampipeconfig/parse"
)

// LoadVariables loads the workspace mod, only processing variables blocks
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

	// now parse the mod
	mod, err := loadAndParseModData(workspacePath, nil, opts)
	if err != nil {
		return nil, err
	}

	// TODO look into naming consistency
	// TACTICAL - as the tf derived code builds a map keyed by the short variable name, do the same
	res := make(map[string]*modconfig.Variable)
	for k, v := range mod.Variables {
		name := strings.Split(k, ".")[1]
		res[name] = v
	}
	return res, nil
}
