package cloud

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/spf13/viper"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/dashboard/dashboardtypes"
	"github.com/turbot/steampipe/pkg/export"
	"github.com/turbot/steampipe/pkg/steampipeconfig"
	"net/http"
	"net/url"
	"path"
	"strings"
)

func PublishSnapshot(snapshot *dashboardtypes.SteampipeSnapshot, share bool) (string, error) {
	snapshotLocation := viper.GetString(constants.ArgSnapshotLocation)
	// snapshotLocation must be set (validation should ensure this)
	if snapshotLocation == "" {
		return "", fmt.Errorf("to share a snapshot, snapshot-locationmust be set")
	}

	// if snapshot location is a workspace handle, upload it
	if steampipeconfig.IsCloudWorkspaceIdentifier(snapshotLocation) {
		url, err := uploadSnapshot(snapshot, share)
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
	return fmt.Sprintf("\nSnapshot copied to %s\n", filePath), nil
}

func exportSnapshot(snapshot *dashboardtypes.SteampipeSnapshot) (string, error) {
	exporter := &export.SnapshotExporter{}

	fileName := export.GenerateDefaultExportFileName(exporter, snapshot.Layout.Name)
	dirName := viper.GetString(constants.ArgSnapshotLocation)
	filePath := path.Join(dirName, fileName)

	err := exporter.Export(context.Background(), snapshot, filePath)
	if err != nil {
		return "", err
	}
	return filePath, nil
}

func uploadSnapshot(snapshot *dashboardtypes.SteampipeSnapshot, share bool) (string, error) {
	baseUrl := getBaseApiUrl()
	client := &http.Client{}
	bearer := getBearerToken(viper.GetString(constants.ArgCloudToken))
	cloudWorkspace := viper.GetString(constants.ArgSnapshotLocation)

	parts := strings.Split(cloudWorkspace, "/")
	if len(parts) != 2 {
		return "", fmt.Errorf("failed to resolve username and workspace handle from workspace %s", cloudWorkspace)
	}
	identity := parts[0]
	worskpaceHandle := parts[1]

	// no determine whether this is a user or org workspace
	workspaceType, err := getWorkspaceType(identity, worskpaceHandle, baseUrl, bearer, client)
	if err != nil {
		return "", err
	}

	urlPath, err := url.JoinPath(baseUrl,
		fmt.Sprintf("api/v0/%s/%s/workspace/%s/snapshot", workspaceType, identity, worskpaceHandle))
	if err != nil {
		return "", err
	}

	// set the visibility
	visibility := "workspace"
	if share {
		visibility = "anyone_with_link"
	}

	// populate map of tags tags been set?
	tags := getTags()

	body := struct {
		Data       *dashboardtypes.SteampipeSnapshot `json:"data"`
		Tags       map[string]any                    `json:"tags"`
		Visibility string                            `json:"visibility"`
	}{
		Data:       snapshot,
		Tags:       tags,
		Visibility: visibility,
	}

	bodyStr, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	var resp = map[string]any{}
	err = postToAPI(urlPath, bearer, string(bodyStr), client, &resp)
	if err != nil {
		return "", err
	}

	snapshotId := resp["id"].(string)
	snapshotUrl := fmt.Sprintf("https://%s/%s/%s/workspace/%s/snapshot/%s",
		viper.GetString(constants.ArgCloudHost),
		workspaceType,
		identity,
		worskpaceHandle,
		snapshotId)

	return snapshotUrl, nil
}

func getWorkspaceType(identityHandle, workspaceHandle, baseUrl, bearer string, client *http.Client) (string, error) {
	workspaces, err := getWorkspaces(baseUrl, bearer, client)
	if err != nil {
		return "", err
	}
	for _, w := range workspaces {
		workspace := w.(map[string]any)
		if workspace["handle"].(string) == workspaceHandle {
			identity := workspace["identity"].(map[string]any)
			if identity["handle"].(string) == identityHandle {
				workspaceType := identity["type"].(string)
				return workspaceType, nil
			}
		}
	}
	return "", fmt.Errorf("workspace %s not found", workspaceHandle)
}

func getTags() map[string]any {
	tags := viper.GetStringSlice(constants.ArgSnapshotTag)
	res := map[string]any{}
	if len(tags) == 0 {
		// if no tags were specified, add the default
		res["generated_by"] = "cli"
		return res
	}

	for _, tagStr := range tags {
		parts := strings.Split(tagStr, "=")
		if len(parts) != 2 {
			continue
		}
		res[parts[0]] = parts[1]
	}
	return res
}
