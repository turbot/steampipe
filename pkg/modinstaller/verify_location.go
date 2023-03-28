package modinstaller

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/error_helpers"
	"github.com/turbot/steampipe/pkg/utils"
)

// ValidateModLocation checks whether you are running from the home directory or if you have
// a lot of non .sql and .sp file in your current directory, and asks for confirmation to continue
func ValidateModLocation(ctx context.Context, workspacePath string) bool {
	cmd := viper.Get(constants.ConfigKeyActiveCommand).(*cobra.Command)
	home, _ := os.UserHomeDir()

	// check if running in home directory
	if workspacePath == home {
		error_helpers.ShowWarning(fmt.Sprintf("You're in the home directory. It's recommended to create a new directory and run %s from there.\nDo you want to continue? (y/n)", constants.Bold(fmt.Sprintf("steampipe mod %s", cmd.Name()))))
		return utils.UserConfirmation()
	} else {
		// else check if running in a directory containing lot of sql and sp files
		fileList, _ := filehelpers.ListFiles(workspacePath, &filehelpers.ListOptions{
			Flags:      filehelpers.FilesRecursive,
			Exclude:    filehelpers.InclusionsFromExtensions([]string{".sql", ".sp"}),
			MaxResults: 10,
		})
		if len(fileList) == 10 {
			error_helpers.ShowWarning(fmt.Sprintf("You're in a directory with a lot of files or subdirectories (>10 files that are not .sql or .sp). It's recommended to create a new directory and run %s from there.\nDo you want to continue? (y/n)", constants.Bold(fmt.Sprintf("steampipe mod %s", cmd.Name()))))
			return utils.UserConfirmation()
		}
	}
	return true
}
