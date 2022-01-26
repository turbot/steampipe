package main

import (
	"encoding/json"
	"github.com/turbot/steampipe/report/reportassets"
	"github.com/turbot/steampipe/version"
	"io/ioutil"
)

func main() {
	spVersionString := version.SteampipeVersion.String()
	spVersion := reportassets.ReportAssetsVersionFile{Version: spVersionString}
	versionsFile, _ := json.MarshalIndent(spVersion, "", " ")
	err := ioutil.WriteFile("build/versions.json", versionsFile, 0644)
	if err != nil {
		panic(err)
	}
}
