package plugin

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/turbot/steampipe/ociinstaller"
	"github.com/turbot/steampipe/ociinstaller/versionfile"
	"github.com/turbot/steampipe/utils"
)

// VersionCheckReport ::
type VersionCheckReport struct {
	Plugin        *versionfile.InstalledVersion
	CheckResponse versionCheckPayload
	CheckRequest  versionCheckPayload
}

func (vr *VersionCheckReport) ShortName() string {
	return fmt.Sprintf("%s/%s", vr.CheckResponse.Org, vr.CheckResponse.Name)
}

// the payload that travels to-and-fro between steampipe and the server
type versionCheckPayload struct {
	Org     string `json:"org"`
	Name    string `json:"name"`
	Stream  string `json:"stream"`
	Version string `json:"version"`
	Digest  string `json:"digest"`
}

// VersionChecker :: wrapper struct over the plugin version check utilities
type VersionChecker struct {
	pluginsToCheck []*versionfile.InstalledVersion
	signature      string
}

// GetUpdateReport looks up and reports the updated version of selective turbot plugins which are listed in versions.json
func GetUpdateReport(installationID string, check []*versionfile.InstalledVersion) map[string]VersionCheckReport {
	versionChecker := new(VersionChecker)
	versionChecker.signature = installationID

	for _, c := range check {
		if strings.HasPrefix(c.Name, ociinstaller.DefaultImageRepoDisplayURL) {
			versionChecker.pluginsToCheck = append(versionChecker.pluginsToCheck, c)
		}
	}

	return versionChecker.reportPluginUpdates()
}

// GetAllUpdateReport looks up and reports the updated version of all turbot plugins which are listed in versions.json
func GetAllUpdateReport(installationID string) map[string]VersionCheckReport {
	versionChecker := new(VersionChecker)
	versionChecker.signature = installationID
	versionChecker.pluginsToCheck = []*versionfile.InstalledVersion{}

	versionFileData, err := versionfile.LoadPluginVersionFile()
	if err != nil {
		log.Println("[TRACE]", "CheckAndReportPluginUpdates", "could not load versionfile")
		return nil
	}

	for _, p := range versionFileData.Plugins {
		if strings.HasPrefix(p.Name, ociinstaller.DefaultImageRepoDisplayURL) {
			versionChecker.pluginsToCheck = append(versionChecker.pluginsToCheck, p)
		}
	}

	return versionChecker.reportPluginUpdates()
}

func (v *VersionChecker) reportPluginUpdates() map[string]VersionCheckReport {
	versionFileData, err := versionfile.LoadPluginVersionFile()
	if err != nil {
		log.Println("[TRACE]", "CheckAndReportPluginUpdates", "could not load versionfile")
		return nil
	}

	if len(v.pluginsToCheck) == 0 {
		// there's no plugin installed. no point continuing
		return nil
	}
	reports := v.getLatestVersionsForPlugins(v.pluginsToCheck)

	// remove elements from `reports` which have empty strings in CheckResponse
	// this happens if we have sent a plugin to the API which doesn't exist
	// in the registry
	for key, value := range reports {
		if value.CheckResponse.Name == "" {
			// delete this key
			delete(reports, key)
		}
	}

	// update the version file
	for _, plugin := range v.pluginsToCheck {
		versionFileData.Plugins[plugin.Name].LastCheckedDate = versionfile.FormatTime(time.Now())
	}

	if err = versionFileData.Save(); err != nil {
		log.Println("[TRACE]", "CheckAndReportPluginUpdates", "could not save versionfile")
		return nil
	}

	return reports
}

func (v *VersionChecker) getLatestVersionsForPlugins(plugins []*versionfile.InstalledVersion) map[string]VersionCheckReport {

	getMapKey := func(thisPayload versionCheckPayload) string {
		return fmt.Sprintf("%s/%s/%s", thisPayload.Org, thisPayload.Name, thisPayload.Stream)
	}

	requestPayload := []versionCheckPayload{}
	reports := map[string]VersionCheckReport{}

	for _, ref := range plugins {
		thisPayload := v.getPayloadFromInstalledData(ref)
		requestPayload = append(requestPayload, thisPayload)

		reports[getMapKey(thisPayload)] = VersionCheckReport{
			Plugin:        ref,
			CheckRequest:  thisPayload,
			CheckResponse: versionCheckPayload{},
		}
	}

	serverResponse := v.requestServerForLatest(requestPayload)
	if serverResponse == nil {
		log.Println("[TRACE]", "PluginVersionChecker", "getLatestVersionsForPlugins", "response nil")
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

func (v *VersionChecker) getPayloadFromInstalledData(plugin *versionfile.InstalledVersion) versionCheckPayload {
	ref := ociinstaller.NewSteampipeImageRef(plugin.Name)
	org, name, stream := ref.GetOrgNameAndStream()
	payload := versionCheckPayload{
		Org:     org,
		Name:    name,
		Stream:  stream,
		Version: plugin.Version,
		Digest:  plugin.ImageDigest,
	}
	return payload
}

func (v *VersionChecker) getVersionCheckURL() url.URL {
	var u url.URL
	u.Scheme = "https"
	u.Host = "hub.steampipe.io"
	u.Path = "api/plugin/version"
	return u
}

func (v *VersionChecker) requestServerForLatest(payload []versionCheckPayload) []versionCheckPayload {
	// Set a default timeout of 3 sec for the check request (in milliseconds)
	sendRequestTo := v.getVersionCheckURL()
	requestBody := utils.BuildRequestPayload(v.signature, map[string]interface{}{
		"plugins": payload,
	})

	resp, err := utils.SendRequest(v.signature, "POST", sendRequestTo, requestBody)
	if err != nil {
		log.Printf("[TRACE] Could not send request")
		return nil
	}

	if resp.StatusCode == 204 {
		log.Println("[TRACE] Got 204")
		return nil
	}

	if resp.StatusCode != 200 {
		log.Printf("[TRACE] Unknown response during version check: %d\n", resp.StatusCode)
		return nil
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("[TRACE] Error reading body stream")
		return nil
	}
	defer resp.Body.Close()

	var responseData []versionCheckPayload

	err = json.Unmarshal(bodyBytes, &responseData)
	if err != nil {
		log.Println("[TRACE] Error in unmarshalling plugin update response", err)
		return nil
	}

	return responseData
}
