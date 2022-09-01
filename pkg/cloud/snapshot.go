package cloud

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/spf13/viper"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/dashboard/dashboardtypes"
	"io"
	"net/http"
	"strings"
)

func UploadSnapshot(snapshot *dashboardtypes.SteampipeSnapshot, share bool) (string, error) {

	cloudWorkspace := viper.GetString(constants.ArgWorkspace)
	parts := strings.Split(cloudWorkspace, "/")
	if len(parts) != 2 {
		return "", fmt.Errorf("failed to resolve username and workspace handle from workspace %s", cloudWorkspace)
	}
	user := parts[0]
	worskpaceHandle := parts[1]

	url := fmt.Sprintf("https://%s/api/v0/user/%s/workspace/%s/snapshot",
		viper.GetString(constants.ArgCloudHost),
		user,
		worskpaceHandle)

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

	body := struct {
		Data       *dashboardtypes.SteampipeSnapshot `json:"data"`
		Tags       map[string]interface{}            `json:"tags"`
		Visibility string                            `json:"visibility"`
	}{
		Data:       snapshot,
		Tags:       map[string]interface{}{"generated_by": "cli"},
		Visibility: visibility,
	}

	bodyStr, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(bodyStr))
	if err != nil {
		return "", err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", bearer)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 206 {
		return "", fmt.Errorf("%s", resp.Status)
	}

	var result map[string]interface{}
	err = json.Unmarshal(bodyBytes, &result)
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
