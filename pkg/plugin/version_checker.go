package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/turbot/steampipe/pkg/ociinstaller"
	"github.com/turbot/steampipe/pkg/ociinstaller/versionfile"
	"github.com/turbot/steampipe/pkg/utils"
)

// VersionCheckReport ::
type VersionCheckReport struct {
	Plugin        *versionfile.InstalledVersion
	CheckResponse versionCheckResponsePayload
	CheckRequest  versionCheckRequestPayload
}

func (vr *VersionCheckReport) ShortName() string {
	return fmt.Sprintf("%s/%s", vr.CheckResponse.Org, vr.CheckResponse.Name)
}

// VersionChecker :: wrapper struct over the plugin version check utilities
type VersionChecker struct {
	pluginsToCheck []*versionfile.InstalledVersion
	signature      string
}

// GetUpdateReport looks up and reports the updated version of selective turbot plugins which are listed in versions.json
func GetUpdateReport(ctx context.Context, installationID string, check []*versionfile.InstalledVersion) map[string]VersionCheckReport {
	versionChecker := new(VersionChecker)
	versionChecker.signature = installationID

	for _, c := range check {
		if strings.HasPrefix(c.Name, ociinstaller.DefaultImageRepoDisplayURL) {
			versionChecker.pluginsToCheck = append(versionChecker.pluginsToCheck, c)
		}
	}

	return versionChecker.reportPluginUpdates(ctx)
}

// GetAllUpdateReport looks up and reports the updated version of all turbot plugins which are listed in versions.json
func GetAllUpdateReport(ctx context.Context, installationID string) map[string]VersionCheckReport {
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

	return versionChecker.reportPluginUpdates(ctx)
}

func (v *VersionChecker) reportPluginUpdates(ctx context.Context) map[string]VersionCheckReport {
	versionFileData, err := versionfile.LoadPluginVersionFile()
	if err != nil {
		log.Println("[TRACE]", "CheckAndReportPluginUpdates", "could not load versionfile")
		return nil
	}

	if len(v.pluginsToCheck) == 0 {
		// there's no plugin installed. no point continuing
		return nil
	}
	reports := v.getLatestVersionsForPlugins(ctx, v.pluginsToCheck)

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

func (v *VersionChecker) getLatestVersionsForPlugins(ctx context.Context, plugins []*versionfile.InstalledVersion) map[string]VersionCheckReport {

	getMapKey := func(thisPayload versionCheckRequestPayload) string {
		return fmt.Sprintf("%s/%s/%s", thisPayload.Org, thisPayload.Name, thisPayload.Stream)
	}

	var requestPayload []versionCheckRequestPayload
	reports := map[string]VersionCheckReport{}

	for _, ref := range plugins {
		thisPayload := v.getPayloadFromInstalledData(ref)
		requestPayload = append(requestPayload, thisPayload)

		reports[getMapKey(thisPayload)] = VersionCheckReport{
			Plugin:        ref,
			CheckRequest:  thisPayload,
			CheckResponse: versionCheckResponsePayload{},
		}
	}

	serverResponse, err := v.requestServerForLatest(ctx, requestPayload)
	if err != nil {
		log.Printf("[TRACE] PluginVersionChecker getLatestVersionsForPlugins returned error: %s", err.Error())
		// return a blank map
		return map[string]VersionCheckReport{}
	}

	log.Println("[TRACE] serverResponse:", serverResponse)

	for _, pluginResponseData := range serverResponse {
		r := reports[pluginResponseData.getMapKey()]
		r.CheckResponse = pluginResponseData
		reports[pluginResponseData.getMapKey()] = r
	}

	return reports
}

func (v *VersionChecker) getPayloadFromInstalledData(plugin *versionfile.InstalledVersion) versionCheckRequestPayload {
	ref := ociinstaller.NewSteampipeImageRef(plugin.Name)
	org, name, stream := ref.GetOrgNameAndStream()
	payload := versionCheckRequestPayload{
		Org:     org,
		Name:    name,
		Stream:  stream,
		Version: plugin.Version,
		Digest:  plugin.ImageDigest,
	}
	// if Digest field is missing, populate with dummy field
	// - this will force and update an in doing so fix the versions.json
	// https://github.com/turbot/steampipe/issues/2030
	if payload.Digest == "" {
		payload.Digest = "no digest"
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

func (v *VersionChecker) requestServerForLatest(ctx context.Context, payload []versionCheckRequestPayload) ([]versionCheckResponsePayload, error) {
	// Set a default timeout of 3 sec for the check request (in milliseconds)
	sendRequestTo := v.getVersionCheckURL()
	requestBody := utils.BuildRequestPayload(v.signature, map[string]interface{}{
		"plugins": payload,
	})

	resp, err := utils.SendRequest(ctx, v.signature, "POST", sendRequestTo, requestBody)
	if err != nil {
		log.Printf("[TRACE] Could not send request")
		return nil, err
	}

	if resp.StatusCode != 200 {
		log.Printf("[TRACE] Unknown response during version check: %d\n", resp.StatusCode)
		return nil, fmt.Errorf("requestServerForLatest failed - SendRequest returned %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("[TRACE] Error reading body stream")
		return nil, err
	}
	defer resp.Body.Close()

	var responseData []versionCheckResponsePayload

	err = json.Unmarshal(bodyBytes, &responseData)
	if err != nil {
		log.Println("[TRACE] Error in unmarshalling plugin update response", err)
		return nil, err
	}

	return responseData, nil
}
