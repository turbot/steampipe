package task

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"runtime"

	"github.com/fatih/color"
	SemVer "github.com/hashicorp/go-version"
	"github.com/olekukonko/tablewriter"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/utils"
	"github.com/turbot/steampipe/version"
)

const legacyDisableUpdatesCheckEnvVar = "SP_DISABLE_UPDATE_CHECK"
const updatesCheckEnvVar = "STEAMPIPE_UPDATE_CHECK"

// the current version of the Steampipe CLI application
var currentVersion = version.Version

type versionCheckResponse struct {
	NewVersion   string    `json:"latest_version,omitempty"` // `json:"current_version"`
	DownloadURL  string    `json:"download_url,omitempty"`   // `json:"download_url"`
	ChangelogURL string    `json:"html,omitempty"`           // `json:"changelog_url"`
	Alerts       []*string `json:"alerts,omitempty"`
}

type versionCheckRequest struct {
	Version    string `json:"version,omitempty"`
	OsPlatform string `json:"os_platform,omitempty"`
	OsArch     string `json:"arch,omitempty"`
	Signature  string `json:"signature"`
}

// VersionChecker :: the version checker struct composition container.
// This MUST not be instantiated manually. Use `CreateVersionChecker` instead
type versionChecker struct {
	checkResult *versionCheckResponse // a channel to store the HTTP response
	disabled    bool
	signature   string // flags whether update check should be done
}

// check if there is a new version
func checkSteampipeVersion(id string) {
	if !shouldDoUpdateCheck() {
		return
	}

	v := new(versionChecker)
	v.signature = id
	v.GetVersionResp()
	v.Notify()
}

// RunCheck :: Communicates with the Turbot Artifacts Server retrieves
// the latest released version
func (c *versionChecker) GetVersionResp() {
	c.doCheckRequest()
}

// Notify :: Notifies the user if a new version is available
func (c *versionChecker) Notify() {
	info := c.checkResult
	if info == nil {
		return
	}

	if info.NewVersion == "" {
		return
	}

	newVersion, err := SemVer.NewVersion(info.NewVersion)
	if err != nil {
		return
	}
	currentVersion, err := SemVer.NewVersion(currentVersion)

	if err != nil {
		fmt.Println(fmt.Errorf("there's something wrong with the Current Version"))
		fmt.Println(err)
	}

	if newVersion.GreaterThan(currentVersion) {
		displayUpdateNotification(info, currentVersion, newVersion)
	}
}

func displayUpdateNotification(info *versionCheckResponse, currentVersion *SemVer.Version, newVersion *SemVer.Version) {

	var downloadURLColor = color.New(color.FgYellow)

	var notificationLines = [][]string{
		[]string{""},
		[]string{fmt.Sprintf("A new version of Steampipe is available! %s → %s", constants.Bold(currentVersion), constants.Bold(newVersion))},
		[]string{fmt.Sprintf("You can update by downloading from %s", downloadURLColor.Sprint("https://steampipe.io/downloads"))},
		[]string{""},
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{})                // no headers please
	table.SetAlignment(tablewriter.ALIGN_LEFT) // we align to the left
	table.SetAutoWrapText(false)               // let's not wrap the text
	table.SetBorder(true)                      // there needs to be a border to give the dialog feel
	table.AppendBulk(notificationLines)        // Add Bulk Data

	fmt.Println()
	table.Render()
	fmt.Println()
}

func versionDiff(oldVersion *SemVer.Version, newVersion *SemVer.Version) string {
	// find out the difference between the two
	nSegments := newVersion.Segments()
	cSegments := oldVersion.Segments()
	var diff = ""

	if len(nSegments) > 0 && len(cSegments) > 0 && nSegments[0] != cSegments[0] {
		diff = "major"
	} else if len(nSegments) > 1 && len(cSegments) > 1 && nSegments[1] != cSegments[1] {
		diff = "minor"
	} else if len(nSegments) > 2 && len(cSegments) > 2 && nSegments[2] != cSegments[2] {
		diff = "patch"
	}

	if diff == "" && newVersion.Prerelease() == "" && oldVersion.Prerelease() != "" {
		diff = "stable"
	}

	if newVersion.Prerelease() != "" {
		diff = "pre-" + diff
	}

	return diff
}

func (c *versionChecker) doCheckRequest() {
	payload := utils.BuildRequestPayload(c.signature, map[string]interface{}{})
	sendRequestTo := c.versionCheckURL()

	resp, err := utils.SendRequest(c.signature, "POST", sendRequestTo, payload)
	if err != nil {
		return
	}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	bodyString := string(bodyBytes)
	defer resp.Body.Close()

	if resp.StatusCode == 204 {
		return
	}

	if resp.StatusCode != 200 {
		log.Printf("[DEBUG] Unknown response during version check: %d\n", resp.StatusCode)
		return
	}

	c.checkResult = c.decodeResult(bodyString)
}

func (c *versionChecker) buildJSONPayload() *bytes.Buffer {
	id := c.signature
	body := &versionCheckRequest{
		Version:    currentVersion,
		OsPlatform: runtime.GOOS,
		OsArch:     runtime.GOARCH,
		Signature:  id,
	}
	jsonStr, _ := json.Marshal(body)
	return bytes.NewBuffer(jsonStr)
}

func (c *versionChecker) decodeResult(body string) *versionCheckResponse {
	var result versionCheckResponse

	if err := json.Unmarshal([]byte(body), &result); err != nil {
		return nil
	}
	return &result
}

func (c *versionChecker) versionCheckURL() url.URL {
	var u url.URL
	//https://hub.steampipe.io/api/cli/version/latest
	u.Scheme = "https"
	u.Host = "hub.steampipe.io"
	u.Path = "api/cli/version/latest"
	return u
}
