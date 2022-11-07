package cloud

import (
	"context"
	"fmt"
	"github.com/spf13/viper"
	steampipecloud "github.com/turbot/steampipe-cloud-sdk-go"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/dashboard/dashboardtypes"
	"github.com/turbot/steampipe/pkg/export"
	"github.com/turbot/steampipe/pkg/steampipeconfig"
	"log"
	"path"
	"strings"
)

func PublishSnapshot(ctx context.Context, snapshot *dashboardtypes.SteampipeSnapshot, share bool) (string, error) {
	snapshotLocation := viper.GetString(constants.ArgSnapshotLocation)
	// snapshotLocation must be set (validation should ensure this)
	if snapshotLocation == "" {
		return "", fmt.Errorf("to share a snapshot, snapshot-locationmust be set")
	}

	// if snapshot location is a workspace handle, upload it
	if steampipeconfig.IsCloudWorkspaceIdentifier(snapshotLocation) {
		url, err := uploadSnapshot(ctx, snapshot, share)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("\nSnapshot uploaded to %s\n", url), nil
	}

	// otherwise assume snapshot location is a file path
	filePath, err := exportSnapshot(snapshot)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("\nSnapshot saved to %s\n", filePath), nil
}

func exportSnapshot(snapshot *dashboardtypes.SteampipeSnapshot) (string, error) {
	exporter := &export.SnapshotExporter{}

	fileName := export.GenerateDefaultExportFileName(snapshot.FileNameRoot, exporter.FileExtension())
	dirName := viper.GetString(constants.ArgSnapshotLocation)
	filePath := path.Join(dirName, fileName)

	err := exporter.Export(context.Background(), snapshot, filePath)
	if err != nil {
		return "", err
	}
	return filePath, nil
}

func uploadSnapshot(ctx context.Context, snapshot *dashboardtypes.SteampipeSnapshot, share bool) (string, error) {
	client := newSteampipeCloudClient(viper.GetString(constants.ArgCloudToken))

	cloudWorkspace := viper.GetString(constants.ArgSnapshotLocation)
	parts := strings.Split(cloudWorkspace, "/")
	if len(parts) != 2 {
		return "", fmt.Errorf("failed to resolve username and workspace handle from workspace %s", cloudWorkspace)
	}
	identityHandle := parts[0]
	workspaceHandle := parts[1]

	// no determine whether this is a user or org workspace
	// get the identity
	identity, _, err := client.Identities.Get(ctx, identityHandle).Execute()
	if err != nil {
		return "", err
	}

	workspaceType := identity.Type

	// set the visibility
	visibility := "workspace"
	if share {
		visibility = "anyone_with_link"
	}

	// resolve the snapshot title
	title := resolveSnapshotTitle(snapshot)
	log.Printf("[TRACE] Uploading snapshot with title %s", title)
	// populate map of tags tags been set?
	tags := getTags()

	cloudSnapshot, err := snapshot.AsCloudSnapshot()
	if err != nil {
		return "", err
	}

	// strip verbose/sensitive fields
	dashboardtypes.StripSnapshot(cloudSnapshot)

	req := steampipecloud.CreateWorkspaceSnapshotRequest{Data: *cloudSnapshot, Tags: tags, Visibility: &visibility}
	req.SetTitle(title)

	var uploadedSnapshot steampipecloud.WorkspaceSnapshot
	if identity.Type == "user" {
		uploadedSnapshot, _, err = client.UserWorkspaceSnapshots.Create(ctx, identityHandle, workspaceHandle).Request(req).Execute()
	} else {
		uploadedSnapshot, _, err = client.OrgWorkspaceSnapshots.Create(ctx, identityHandle, workspaceHandle).Request(req).Execute()
	}
	if err != nil {
		return "", err
	}

	snapshotId := uploadedSnapshot.Id
	snapshotUrl := fmt.Sprintf("https://%s/%s/%s/workspace/%s/snapshot/%s",
		viper.GetString(constants.ArgCloudHost),
		workspaceType,
		identityHandle,
		workspaceHandle,
		snapshotId)

	return snapshotUrl, nil
}

func resolveSnapshotTitle(snapshot *dashboardtypes.SteampipeSnapshot) string {
	if titleArg := viper.GetString(constants.ArgSnapshotTitle); titleArg != "" {
		return titleArg
	}
	// is there a title property set on the snapshot
	if snapshotTitle := snapshot.Title; snapshotTitle != "" {
		return snapshotTitle
	}
	// fall back to the fully qualified name of the root resource (which is also the FileNameRoot)
	return snapshot.FileNameRoot
}

func getTags() map[string]any {
	tags := viper.GetStringSlice(constants.ArgSnapshotTag)
	res := map[string]any{}

	for _, tagStr := range tags {
		parts := strings.Split(tagStr, "=")
		if len(parts) != 2 {
			continue
		}
		res[parts[0]] = parts[1]
	}
	return res
}
