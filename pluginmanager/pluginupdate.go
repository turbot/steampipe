package pluginmanager

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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

// versionCheckPayload :: the payload that travels to-and-fro between steampipe and the server
type versionCheckPayload struct {
	Org     string `json:"org"`
	Name    string `json:"name"`
	Stream  string `json:"stream"`
	Version string `json:"version"`
	Digest  string `json:"digest"`
}

// PluginVersionChecker :: wrapper struct over the plugin version check utilities
type PluginVersionChecker struct {
	pluginsToCheck []*versionfile.InstalledVersion
	signature      string
}

// GetPluginUpdateReport :: looks up and reports the updated version of selective turbot plugins which are listed in versions.json
func GetPluginUpdateReport(installationID string, check []*versionfile.InstalledVersion) map[string]VersionCheckReport {
	pvc := new(PluginVersionChecker)
	pvc.signature = installationID

	for _, c := range check {
		if strings.HasPrefix(c.Name, ociinstaller.DefaultImageRepoDisplayURL) {
			pvc.pluginsToCheck = append(pvc.pluginsToCheck, c)
		}
	}

	return pvc.reportPluginUpdates()
}

// GetAllPluginUpdateReport :: looks up and reports the updated version of all turbot plugins which are listed in versions.json
func GetAllPluginUpdateReport(installationID string) map[string]VersionCheckReport {
	pvc := new(PluginVersionChecker)
	pvc.signature = installationID
	pvc.pluginsToCheck = []*versionfile.InstalledVersion{}

	versionFileData, err := versionfile.Load()
	if err != nil {
		log.Println("[TRACE]", "CheckAndReportPluginUpdates", "could not load versionfile")
		return nil
	}

	if versionFileData.Plugins == nil {
		versionFileData.Plugins = make(map[string](*versionfile.InstalledVersion))
	}

	for _, p := range versionFileData.Plugins {
		if strings.HasPrefix(p.Name, ociinstaller.DefaultImageRepoDisplayURL) {
			pvc.pluginsToCheck = append(pvc.pluginsToCheck, p)
		}
	}

	return pvc.reportPluginUpdates()
}

func (pvc *PluginVersionChecker) reportPluginUpdates() map[string]VersionCheckReport {
	versionFileData, err := versionfile.Load()
	if err != nil {
		log.Println("[TRACE]", "CheckAndReportPluginUpdates", "could not load versionfile")
		return nil
	}

	if versionFileData.Plugins == nil {
		versionFileData.Plugins = make(map[string](*versionfile.InstalledVersion))
	}

	if len(pvc.pluginsToCheck) == 0 {
		// there's no plugin installed. no point continuing
		return nil
	}
	reports := pvc.getLatestVersionsForPlugins(pvc.pluginsToCheck)

	// update the version file
	for _, plugin := range pvc.pluginsToCheck {
		versionFileData.Plugins[plugin.Name].LastCheckedDate = versionfile.FormatTime(time.Now())
	}
	versionFileData.Save()

	return reports
}

func (pvc *PluginVersionChecker) getLatestVersionsForPlugins(plugins []*versionfile.InstalledVersion) map[string]VersionCheckReport {

	getMapKey := func(thisPayload versionCheckPayload) string {
		return fmt.Sprintf("%s/%s/%s", thisPayload.Org, thisPayload.Name, thisPayload.Stream)
	}

	requestPayload := []versionCheckPayload{}
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

func (pvc *PluginVersionChecker) getPayloadFromInstalledData(plugin *versionfile.InstalledVersion) versionCheckPayload {
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

func (pvc *PluginVersionChecker) getVersionCheckURL() url.URL {
	var u url.URL
	//https://hub-steampipe-io-git-development.turbot.vercel.app/api/plugin/version
	u.Scheme = "https"
	u.Host = "hub.steampipe.io"
	u.Path = "api/plugin/version"
	return u
}

func (pvc *PluginVersionChecker) requestServerForLatest(payload []versionCheckPayload) []versionCheckPayload {
	// Set a default timeout of 3 sec for the check request (in milliseconds)
	sendRequestTo := pvc.getVersionCheckURL()
	requestBody := utils.BuildRequestPayload(pvc.signature, map[string]interface{}{
		"plugins": payload,
	})

	resp, err := utils.SendRequest(pvc.signature, "POST", sendRequestTo, requestBody)
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

	responseData := []versionCheckPayload{}

	err = json.Unmarshal(bodyBytes, &responseData)
	if err != nil {
		fmt.Println("[DEBUG] Error in unmarshalling response", err)
		return nil
	}

	return responseData
}
