package main

import (
	"encoding/json"
	"github.com/turbot/steampipe/pkg/version"
	"os"
)

type packageVersion struct {
	Version string `json:"version"`
}

func main() {
	spVersionString := version.SteampipeVersion.String()
	spVersion := packageVersion{Version: spVersionString}
	versionsFile, _ := json.MarshalIndent(spVersion, "", " ")
	err := os.WriteFile("build/versions.json", versionsFile, 0644)
	if err != nil {
		panic(err)
	}
}
