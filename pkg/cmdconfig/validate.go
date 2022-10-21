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

	token := viper.GetString(constants.ArgCloudToken)

	// determine whether snapshot location is a cloud workspace or a file location
	// if a file location, check it exists
	if err := validateSnapshotLocation(token); err != nil {
		return err
	}

	// if workspace-database or snapshot-location are a cloud workspace handle, cloud token must be set
	requireCloudToken := steampipeconfig.IsCloudWorkspaceIdentifier(viper.GetString(constants.ArgWorkspaceDatabase)) ||
		steampipeconfig.IsCloudWorkspaceIdentifier(viper.GetString(constants.ArgSnapshotLocation))

	// verify cloud token and workspace has been set
	if requireCloudToken && token == "" {
		return constants.MissingCloudTokenError
	}

	// should never happen as there is a default set
	if viper.GetString(constants.ArgCloudHost) == "" {
		return fmt.Errorf("to share snapshots, cloud host must be set")
	}

	return validateSnapshotTags()
}

func validateSnapshotLocation(cloudToken string) error {
	snapshotLocation := viper.GetString(constants.ArgSnapshotLocation)

	// if snapshot location is not set, set to the users default
	if snapshotLocation == "" {
		if cloudToken == "" {
			return constants.MissingCloudTokenError
		}
		return setSnapshotLocationFromDefaultWorkspace(cloudToken)
	}

	// if it is NOT a workspace handle, assume it is a local file location:
	// tildefy it and ensure it exists
	if !steampipeconfig.IsCloudWorkspaceIdentifier(snapshotLocation) {
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

func setSnapshotLocationFromDefaultWorkspace(cloudToken string) error {
	workspaces, userHandle, err := cloud.GetUserWorkspaceHandles(cloudToken)
	if err != nil {
		return err
	}
	if len(workspaces) == 0 {
		return fmt.Errorf("snapshot-location is not specified and no workspaces exist for user %s", userHandle)
	}
	if len(workspaces) > 1 {
		return fmt.Errorf("more than one workspace found for user %s", userHandle)
	}

	viper.Set(constants.ArgSnapshotLocation, workspaces[0])
	return nil
}

func validateSnapshotTags() error {
	tags := viper.GetStringSlice(constants.ArgSnapshotTag)
	for _, tagStr := range tags {
		if len(strings.Split(tagStr, "=")) != 2 {
			return fmt.Errorf("snapshot tags must be specified '--%s key=value'", constants.ArgSnapshotTag)
		}
	}
	return nil
}
