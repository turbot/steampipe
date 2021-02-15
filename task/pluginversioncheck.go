package task

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/hashicorp/go-cleanhttp"
	"github.com/olekukonko/tablewriter"
	"github.com/turbot/steampipe/ociinstaller"
	"github.com/turbot/steampipe/ociinstaller/versionfile"
)

// VersionCheckReport ::
type VersionCheckReport struct {
	Plugin        *versionfile.InstalledVersion
	CheckResponse VersionCheckPayload
	CheckRequest  VersionCheckPayload
}

// VersionCheckPayload :: the payload that travels to-and-fro between steampipe and the server
type VersionCheckPayload struct {
	Org     string `json:"org"`
	Name    string `json:"name"`
	Stream  string `json:"stream"`
	Version string `json:"version"`
	Digest  string `json:"digest"`
}

// PluginVersionChecker :: wrapper struct over the plugin version check utilities
type PluginVersionChecker struct {
	installationID string
}

// check if there is a new version
func checkPluginVersions(id string) {
	// if SP_DISABLE_UPDATE_CHECK is set, do nothing
	if v, ok := os.LookupEnv(disableUpdatesCheckEnvVar); ok && strings.ToLower(v) == "true" {
		return
	}
	pvc := new(PluginVersionChecker)
	pvc.installationID = id
	pvc.CheckAndReportPluginUpdates()
}

// CheckAndReportPluginUpdates ::
func (pvc *PluginVersionChecker) CheckAndReportPluginUpdates() {
	versionFileData, err := versionfile.Load()
	if err != nil {
		log.Println("[TRACE]", "CheckAndReportPluginUpdates", "could not load versionfile")
		return
	}

	if versionFileData.Plugins == nil {
		versionFileData.Plugins = make(map[string](*versionfile.InstalledVersion))
	}

	pluginsToCheck := pvc.filterPluginsToCheck(versionFileData.Plugins)
	if len(pluginsToCheck) == 0 {
		// there's no plugin installed. no point continuing
		return
	}
	reports := pvc.getLatestVersionsForPlugins(pluginsToCheck)
	pluginsToUpdate := []VersionCheckReport{}

	for _, r := range reports {
		if r.CheckResponse.Digest != r.Plugin.ImageDigest {
			pluginsToUpdate = append(pluginsToUpdate, r)
			versionFileData.Plugins[r.Plugin.Name].LastCheckedDate = versionfile.FormatTime(time.Now())
		}
	}

	// now update the versionfile that these were checked
	// don't care if the write failed
	versionFileData.Save()

	if len(pluginsToUpdate) > 0 {
		pvc.showPluginUpdateNotification(pluginsToUpdate)
	}
}

func (pvc *PluginVersionChecker) filterPluginsToCheck(plugins map[string]*versionfile.InstalledVersion) map[string]*versionfile.InstalledVersion {
	pluginsToCheck := map[string]*versionfile.InstalledVersion{}
	for k, v := range plugins {
		if strings.HasPrefix(k, ociinstaller.DefaultImageRepoDisplayURL) {
			pluginsToCheck[k] = v
		}
	}
	return pluginsToCheck
}

func (pvc *PluginVersionChecker) showPluginUpdateNotification(reports []VersionCheckReport) {
	var updateCmdColor = color.New(color.FgHiYellow, color.Bold)
	var oldVersionColor = color.New(color.FgHiRed, color.Bold)
	var newVersionColor = color.New(color.FgHiGreen, color.Bold)

	var notificationLines = [][]string{
		{""},
		{"Updated versions of the following plugins are available:"},
		{""},
	}
	for _, report := range reports {
		thisName := fmt.Sprintf("%s/%s", report.CheckResponse.Org, report.CheckResponse.Name)
		line := ""
		if len(report.Plugin.Version) == 0 {
			line = fmt.Sprintf(
				"%-20s @ %-10s %24s",
				thisName,
				report.CheckResponse.Stream,
				newVersionColor.Sprintf(report.CheckResponse.Version),
			)
		} else {
			line = fmt.Sprintf(
				"%-20s @ %-10s %10s → %-10s",
				thisName,
				report.CheckResponse.Stream,
				oldVersionColor.Sprintf(report.Plugin.Version),
				newVersionColor.Sprintf(report.CheckResponse.Version),
			)
		}
		notificationLines = append(notificationLines, []string{line})
	}
	notificationLines = append(notificationLines, []string{""})
	notificationLines = append(notificationLines, []string{
		fmt.Sprintf("You can update by running\n %60s", updateCmdColor.Sprintf("steampipe plugin install --update-all")),
	})
	notificationLines = append(notificationLines, []string{""})

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

func (pvc *PluginVersionChecker) getLatestVersionsForPlugins(plugins map[string]*versionfile.InstalledVersion) map[string]VersionCheckReport {

	getMapKey := func(thisPayload VersionCheckPayload) string {
		return fmt.Sprintf("%s/%s/%s", thisPayload.Org, thisPayload.Name, thisPayload.Stream)
	}

	requestPayload := []VersionCheckPayload{}
	reports := map[string]VersionCheckReport{}

	for _, ref := range plugins {
		thisPayload := pvc.getPayloadFromInstalledData(ref)
		requestPayload = append(requestPayload, thisPayload)

		reports[getMapKey(thisPayload)] = VersionCheckReport{
			Plugin:       ref,
			CheckRequest: thisPayload,
		}
	}

	serverResponse := pvc.requestServerForLatest(requestPayload)
	if serverResponse == nil {
		// return a blank map
		return map[string]VersionCheckReport{}
	}

	for _, rD := range serverResponse {
		r := reports[getMapKey(rD)]
		r.CheckResponse = rD
		reports[getMapKey(rD)] = r
	}

	return reports
}

func (pvc *PluginVersionChecker) getPayloadFromInstalledData(plugin *versionfile.InstalledVersion) VersionCheckPayload {
	org, name, stream := splitNameIntoOrgNameAndStream(plugin.Name)
	payload := VersionCheckPayload{
		Org:     org,
		Name:    name,
		Stream:  stream,
		Version: plugin.Version,
		Digest:  plugin.ImageDigest,
	}
	return payload
}

func splitNameIntoOrgNameAndStream(name string) (string, string, string) {
	// plugin.Name looks like `hub.steampipe.io/plugins/turbot/aws@latest`
	split := strings.Split(name, "/")

	org := split[len(split)-2]
	pluginNameAndStream := strings.Split(split[len(split)-1], "@")

	return org, pluginNameAndStream[0], pluginNameAndStream[1]
}

func (pvc *PluginVersionChecker) getVersionCheckURL() url.URL {
	var u url.URL
	//https://hub-steampipe-io-git-development.turbot.vercel.app/api/plugin/version
	u.Scheme = "https"
	u.Host = "hub-steampipe-io-git-development.turbot.vercel.app"
	u.Path = "api/plugin/version"
	return u
}

func (pvc *PluginVersionChecker) requestServerForLatest(payload []VersionCheckPayload) []VersionCheckPayload {
	// Set a default timeout of 3 sec for the check request (in milliseconds)
	timeout := 3000
	byteContent, err := json.Marshal(payload)
	sendRequestTo := pvc.getVersionCheckURL()

	req, err := http.NewRequest("POST", sendRequestTo.String(), bytes.NewBuffer(byteContent))
	if err != nil {
		log.Println("[DEBUG] Could not construct request")
		return nil
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", constructUserAgent(pvc.installationID))

	client := cleanhttp.DefaultClient()

	// Use a short timeout since checking for new versions is not critical
	// enough to block on if the update server is broken/slow.
	client.Timeout = time.Duration(timeout) * time.Millisecond

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[DEBUG] Could not send request")
		return nil
	}

	if resp.StatusCode == 204 {
		log.Println("[DEBUG] Got 204")
		return nil
	}

	if resp.StatusCode != 200 {
		log.Printf("[DEBUG] Unknown response during version check: %d\n", resp.StatusCode)
		return nil
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("[DEBUG] Error reading body stream")
		return nil
	}
	defer resp.Body.Close()

	responseData := []VersionCheckPayload{}

	err = json.Unmarshal(bodyBytes, &responseData)
	if err != nil {
		fmt.Println("[DEBUG] Error in unmarshalling response", err)
		return nil
	}

	return responseData
}

// func spitJSON(msg string, d interface{}) {
// 	enc := json.NewEncoder(os.Stdout)
// 	enc.SetIndent(" ", " ")
// 	os.Stdout.WriteString(msg)
// 	enc.Encode(d)
// }
