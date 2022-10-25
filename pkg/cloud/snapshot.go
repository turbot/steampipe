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

	cloudWorkspace := viper.GetString(constants.ArgSnapshotLocation)
	baseUrl := getBaseApiUrl()
	parts := strings.Split(cloudWorkspace, "/")
	if len(parts) != 2 {
		return "", fmt.Errorf("failed to resolve username and workspace handle from workspace %s", cloudWorkspace)
	}
	user := parts[0]
	worskpaceHandle := parts[1]

	urlPath, err := url.JoinPath(baseUrl,
		fmt.Sprintf("api/v0/user/%s/workspace/%s/snapshot", user, worskpaceHandle))
	if err != nil {
		return "", err
	}
	// get the cloud token (we have already verifuied this was provided)
	token := viper.GetString(constants.ArgCloudToken)
	// create a 'bearer' string by appending the access token
	var bearer = "Bearer " + token

	client := &http.Client{}

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

	result, err := postToAPI(urlPath, bearer, string(bodyStr), client)
	if err != nil {
		return "", err
	}

	snapshotId := result["id"].(string)
	snapshotUrl := fmt.Sprintf("https://%s/user/%s/workspace/%s/snapshot/%s",
		viper.GetString(constants.ArgCloudHost),
		user,
		worskpaceHandle,
		snapshotId)

	return snapshotUrl, nil
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
