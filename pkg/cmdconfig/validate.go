package cmdconfig

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/viper"
	filehelpers "github.com/turbot/go-kit/files"
	pconstants "github.com/turbot/pipe-fittings/v2/constants"
	"github.com/turbot/pipe-fittings/v2/pipes"
	"github.com/turbot/pipe-fittings/v2/steampipeconfig"
	"github.com/turbot/steampipe/v2/pkg/error_helpers"
)

func ValidateSnapshotArgs(ctx context.Context) error {
	// only 1 of 'share' and 'snapshot' may be set
	share := viper.GetBool(pconstants.ArgShare)
	snapshot := viper.GetBool(pconstants.ArgSnapshot)
	if share && snapshot {
		return fmt.Errorf("only 1 of 'share' and 'snapshot' may be set")
	}

	// if neither share or snapshot are set, nothing more to do
	if !share && !snapshot {
		return nil
	}

	token := viper.GetString(pconstants.ArgPipesToken)

	// determine whether snapshot location is a cloud workspace or a file location
	// if a file location, check it exists
	if err := validateSnapshotLocation(ctx, token); err != nil {
		return err
	}

	// if workspace-database or snapshot-location are a cloud workspace handle, cloud token must be set
	requireCloudToken := steampipeconfig.IsPipesWorkspaceIdentifier(viper.GetString(pconstants.ArgWorkspaceDatabase)) ||
		steampipeconfig.IsPipesWorkspaceIdentifier(viper.GetString(pconstants.ArgSnapshotLocation))

	// verify cloud token and workspace has been set
	if requireCloudToken && token == "" {
		return error_helpers.MissingCloudTokenError
	}

	// should never happen as there is a default set
	if viper.GetString(pconstants.ArgPipesHost) == "" {
		return fmt.Errorf("to share snapshots, cloud host must be set")
	}

	return validateSnapshotTags()
}

func validateSnapshotLocation(ctx context.Context, cloudToken string) error {
	snapshotLocation := viper.GetString(pconstants.ArgSnapshotLocation)

	// if snapshot location is not set, set to the users default
	if snapshotLocation == "" {
		if cloudToken == "" {
			return error_helpers.MissingCloudTokenError
		}
		return setSnapshotLocationFromDefaultWorkspace(ctx, cloudToken)
	}

	// if it is NOT a workspace handle, assume it is a local file location:
	// tildefy it and ensure it exists
	if !steampipeconfig.IsPipesWorkspaceIdentifier(snapshotLocation) {
		var err error
		snapshotLocation, err = filehelpers.Tildefy(snapshotLocation)
		if err != nil {
			return err
		}

		// write back to viper
		viperMutex.Lock()
		viper.Set(pconstants.ArgSnapshotLocation, snapshotLocation)
		viperMutex.Unlock()

		if !filehelpers.DirectoryExists(snapshotLocation) {
			return fmt.Errorf("snapshot location %s does not exist", snapshotLocation)
		}
	}
	return nil
}

func setSnapshotLocationFromDefaultWorkspace(ctx context.Context, cloudToken string) error {
	workspaceHandle, err := pipes.GetUserWorkspaceHandle(ctx, cloudToken)
	if err != nil {
		return err
	}

	viperMutex.Lock()
	viper.Set(pconstants.ArgSnapshotLocation, workspaceHandle)
	viperMutex.Unlock()
	return nil
}

func validateSnapshotTags() error {
	tags := viper.GetStringSlice(pconstants.ArgSnapshotTag)
	for _, tagStr := range tags {
		if len(strings.Split(tagStr, "=")) != 2 {
			return fmt.Errorf("snapshot tags must be specified '--%s key=value'", pconstants.ArgSnapshotTag)
		}
	}
	return nil
}
