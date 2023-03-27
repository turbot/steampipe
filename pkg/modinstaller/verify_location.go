package modinstaller

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/error_helpers"
)

// verifyModLocation checks whether you are running from the home directory and asks for
// confirmation to continue
func VerifyModLocation(ctx context.Context, workspacePath string) bool {
	cmd := viper.Get(constants.ConfigKeyActiveCommand).(*cobra.Command)
	home, _ := os.UserHomeDir()
	if workspacePath == home {
		error_helpers.ShowWarning(fmt.Sprintf("You're in the home directory. It's recommended to create a new directory and run %s from there.\nDo you want to continue? (y/n)", constants.Bold(fmt.Sprintf("steampipe mod %s", cmd.Name()))))
		var userConfirm rune
		_, err := fmt.Scanf("%c", &userConfirm)
		if err != nil {
			log.Fatal(err)
		}
		keepMod := userConfirm == 'Y' || userConfirm == 'y'
		if !keepMod {
			return false
		}
	}
	return true
}
