package cmdconfig

import (
	"fmt"
	"github.com/spf13/viper"
	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/steampipe/pkg/cloud"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/steampipeconfig"
	"os"
	"strings"
)

func ValidateCloudArgs() error {
	defer displayConfig()

	// determine whether snapshot location is a cloud workspace or a file location
	// if a file location, check it exists
	if err := validateSnapshotLocation(); err != nil {
		return err
	}

	// only 1 of 'share' and 'snapshot' may be set
	share := viper.GetBool(constants.ArgShare)
	snapshot := viper.GetBool(constants.ArgSnapshot)
	if share && snapshot {
		return fmt.Errorf("only 1 of 'share' and 'snapshot' may be set")
	}

	// if neither share or snapshot are set, nothing more to do
	if !share && !snapshot {
		return nil
	}

	// verify cloud token and workspace has been set
	token := viper.GetString(constants.ArgCloudToken)
	if token == "" {
		return fmt.Errorf("to share snapshots, cloud token must be set")
	}

	// we should now have a value for workspace
	if !viper.IsSet(constants.ArgSnapshotLocation) {
		workspace, err := cloud.GetUserWorkspace(token)
		if err != nil {
			return err
		}
		viper.Set(constants.ArgSnapshotLocation, workspace)
	}

	// should never happen as there is a default set
	if viper.GetString(constants.ArgCloudHost) == "" {
		return fmt.Errorf("to share snapshots, cloud host must be set")
	}

	return validateSnapshotTags()
}

func displayConfig() {
	diagnostics := os.Getenv(constants.EnvDiagnostics)

	if strings.ToUpper(diagnostics) != "CONFIG" {
		return
	}

	var argNames = []string{
		constants.ArgInstallDir,
		constants.ArgModLocation,
		constants.ArgSnapshotLocation,
		constants.ArgWorkspaceProfile,
		constants.ArgWorkspaceDatabase,
		constants.ArgCloudHost,
		constants.ArgCloudToken,
	}
	maxLength := 0
	for _, a := range argNames {
		if l := len(a); l > maxLength {
			maxLength = l
		}
	}
	var b strings.Builder
	b.WriteString("\n================\nSteampipe Config\n================\n\n")
	fmtStr := `%-` + fmt.Sprintf("%d", maxLength) + `s: %v` + "\n"
	for _, a := range argNames {
		b.WriteString(fmt.Sprintf(fmtStr, a, viper.GetString(a)))
	}

	fmt.Println(b.String())
}

func validateSnapshotLocation() error {
	snapshotLocation := viper.GetString(constants.ArgSnapshotLocation)
	if !steampipeconfig.IsCloudWorkspaceIdentifier(snapshotLocation) {
		// if it is a file location tildefy it and ensure it exists
		var err error
		snapshotLocation, err = filehelpers.Tildefy(snapshotLocation)
		if err != nil {
			return err
		}

		// write back to viper
		viper.Set(constants.ArgSnapshotLocation, snapshotLocation)

		if !filehelpers.DirectoryExists(snapshotLocation) {
			return fmt.Errorf("snapshot location %s does not exist", snapshotLocation)
		}
	}
	return nil
}

func validateSnapshotTags() error {
	tags := viper.GetStringSlice(constants.ArgSnapshotTag)
	for _, tagStr := range tags {
		if len(strings.Split(tagStr, "=")) != 2 {
			return fmt.Errorf("snapshot tags must be specified '--tag key=value'")
		}
	}
	return nil
}
