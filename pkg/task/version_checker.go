package task

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/url"
	"time"

	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/turbot/pipe-fittings/v2/app_specific"
	"github.com/turbot/pipe-fittings/v2/utils"
)

type CLIVersionCheckResponse struct {
	NewVersion   string    `json:"latest_version,omitempty"` // `json:"current_version"`
	DownloadURL  string    `json:"download_url,omitempty"`   // `json:"download_url"`
	ChangelogURL string    `json:"html,omitempty"`           // `json:"changelog_url"`
	Alerts       []*string `json:"alerts,omitempty"`
}

// VersionChecker :: the version checker struct composition container.
// This MUST not be instantiated manually. Use `CreateVersionChecker` instead
type versionChecker struct {
	checkResult *CLIVersionCheckResponse // a channel to store the HTTP response
	signature   string                   // flags whether update check should be done
}

// get the latest available version of the CLI
func fetchAvailableCLIVersion(ctx context.Context, installationId string) (*CLIVersionCheckResponse, error) {
	v := new(versionChecker)
	v.signature = installationId
	err := v.doCheckRequest(ctx)
	if err != nil {
		return nil, err
	}
	return v.checkResult, nil
}

// contact the Turbot Artifacts Server and retrieve the latest released version
func (c *versionChecker) doCheckRequest(ctx context.Context) error {
	payload := utils.BuildRequestPayload(c.signature, map[string]interface{}{})
	sendRequestTo := c.versionCheckURL()
	timeout := 5 * time.Second

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	resp, err := utils.SendRequest(ctx, c.signature, "POST", sendRequestTo, payload)
	if err != nil {
		return err
	}
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	bodyString := string(bodyBytes)
	defer resp.Body.Close()

	if resp.StatusCode == 204 {
		return nil
	}

	if resp.StatusCode != 200 {
		log.Printf("[TRACE] Unknown response during version check: %d\n", resp.StatusCode)
		return http.NewErr(resp)
	}

	c.checkResult = c.decodeResult(bodyString)
	return nil
}

func (c *versionChecker) decodeResult(body string) *CLIVersionCheckResponse {
	var result CLIVersionCheckResponse

	if err := json.Unmarshal([]byte(body), &result); err != nil {
		return nil
	}
	return &result
}

func (c *versionChecker) versionCheckURL() url.URL {
	var u url.URL
	//https://hub.steampipe.io/api/cli/version/latest
	u.Scheme = "https"
	u.Host = app_specific.VersionCheckHost
	u.Path = app_specific.VersionCheckPath
	return u
}
