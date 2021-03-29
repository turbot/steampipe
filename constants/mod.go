package constants

import (
	"os"
	"path"

	"github.com/turbot/steampipe/utils"
)

// mod related constants
const (
	ModDataExtension = ".sp"
	WorkspaceModDir  = "mods"
	WorkspaceDataDir = ".steampipe"
)

func WorkspaceModPath(workspacePath string) string {
	loc := path.Join(workspacePath, WorkspaceDataDir, WorkspaceModDir)

	if _, err := os.Stat(loc); os.IsNotExist(err) {
		err = os.MkdirAll(loc, 0755)
		utils.FailOnErrorWithMessage(err, "could not create workspace mod directory")
	}
	return loc
}
