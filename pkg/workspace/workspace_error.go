package workspace

import (
	"fmt"

	"github.com/turbot/steampipe/pkg/constants"
)

var (
	ErrorNoModDefinition = fmt.Errorf("This command requires a mod definition file(mod.sp) - could not find in the current directory tree.\nYou can either %s or install a mod using %s and run this command from the cloned/installed mod directory.\nPlease refer to: https://steampipe.io/docs/mods/overview", constants.Bold("clone a mod repository"), constants.Bold("steampipe mod install"))
)
