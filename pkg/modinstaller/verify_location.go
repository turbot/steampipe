package modinstaller

import (
	"context"
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/utils"
)

// ValidateModLocation checks whether you are running from the home directory or if you have
// a lot of non .sql and .sp file in your current directory, and asks for confirmation to continue
func ValidateModLocation(ctx context.Context, workspacePath string) bool {
	const MaxResults = 10
	cmd := viper.Get(constants.ConfigKeyActiveCommand).(*cobra.Command)
	home, _ := os.UserHomeDir()

	// check if running in home directory
	if workspacePath == home {
		return utils.UserConfirmation(fmt.Sprintf("%s: Creating a mod file in the home directory is not recommended.\nBest practice is to create a new directory and run %s from there.\nDo you want to continue? (y/n)", color.YellowString("Warning"), constants.Bold(fmt.Sprintf("steampipe mod %s", cmd.Name()))))
	}
	// else check if running in a directory containing lot of sql and sp files
	fileList, _ := filehelpers.ListFiles(workspacePath, &filehelpers.ListOptions{
		Flags:      filehelpers.FilesRecursive,
		Exclude:    filehelpers.InclusionsFromExtensions([]string{".sql", ".sp"}),
		MaxResults: MaxResults,
	})
	if len(fileList) == MaxResults {
		return utils.UserConfirmation(fmt.Sprintf("%s: Creating a mod file in a directory with a lot of files or subdirectories is not recommended.\nBest practice is to create a new directory and run %s from there.\nDo you want to continue? (y/n)", color.YellowString("Warning"), constants.Bold(fmt.Sprintf("steampipe mod %s", cmd.Name()))))
	}

	return true
}
