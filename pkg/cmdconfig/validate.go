package cmdconfig

import (
	"fmt"
	"github.com/spf13/viper"
	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/steampipe/pkg/cloud"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/steampipeconfig"
	"strings"
)

func ValidateCloudArgs() error {

	// determine whether snapshot location is a cloud workspace or a file location
	if snapshotLocation := viper.GetString(constants.ArgSnapshotLocation); snapshotLocation != "" {
		err := validateSnapshotLocation(snapshotLocation)
		if err != nil {
			return err
		}
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

func validateSnapshotLocation(snapshotLocation string) error {
	if steampipeconfig.IsCloudWorkspaceIdentifier(snapshotLocation) {
		// so snapshot location is a cloud workspace

		// if workspace-database has not been set, use snapshot location
		// NOTE: do this BEFORE populating workspace from share/snapshot args, if set
		if !viper.IsSet(constants.ArgWorkspaceDatabase) {
			viper.Set(constants.ArgWorkspaceDatabase, viper.GetString(constants.ArgSnapshotLocation))
		}
	} else {
		// if it is a file location tildefy it and ensure it exists
		var err error
		snapshotLocation, err = filehelpers.Tildefy(snapshotLocation)
		if err != nil {
			return err
		}
		if !filehelpers.DirectoryExists(snapshotLocation) {
			return fmt.Errorf("snapshot location %s does not exist", snapshotLocation#)
		}
		// write back to viper
		viper.Set(constants.ArgSnapshotLocation, snapshotLocation)
	}
	return nil
}

// determine whether SnapshotLocation is a local path or a cloud workspace
// if it is a cloud workspace it will have the form {identity_handle}/{workspace_handle}
// otherwise we assume it is a local path
func snapshotLocationIsFilePath() bool {
	if len(c.SnapshotLocation) == 0 {
		return false
	}
	parts := strings.Split(c.SnapshotLocation, "/")
	return len(parts) != 2
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
