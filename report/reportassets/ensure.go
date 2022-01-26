package reportassets

import (
	_ "embed"
	"encoding/json"
	"log"
	"os"
	"path/filepath"

	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/filepaths"
	"github.com/turbot/steampipe/utils"
	"github.com/turbot/steampipe/version"
)

//go:embed assets.zip
var reportAssets []byte

const assetsZipFileName = "assets.zip"

func Ensure() error {
	// load report assets versions.json
	versionFile, err := LoadReportAssetVersionFile()
	if err != nil {
		return err
	}
	if versionFile.Version == version.SteampipeVersion.String() {
		return nil
	}

	reportAssetsPath := filepaths.ReportAssetsPath()

	zipPath := filepath.Join(os.TempDir(), assetsZipFileName)
	err = os.WriteFile(zipPath, reportAssets, 0744)
	defer os.RemoveAll(zipPath)

	if _, err := utils.Unzip(zipPath, reportAssetsPath); err != nil {
		return err
	}

	return nil
}

type ReportAssetsVersionFile struct {
	Version string `json:"version"`
}

func LoadReportAssetVersionFile() (*ReportAssetsVersionFile, error) {
	versionFilePath := filepaths.ReportAssetsVersionFilePath()
	if !helpers.FileExists(versionFilePath) {

		return &ReportAssetsVersionFile{}, nil
	}

	file, _ := os.ReadFile(versionFilePath)
	var versionFile ReportAssetsVersionFile
	if err := json.Unmarshal(file, &versionFile); err != nil {
		log.Println("[ERROR]", "Error while reading report assets version file", err)
		return nil, err
	}

	return &versionFile, nil

}
