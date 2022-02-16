package dashboardassets

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe-plugin-sdk/v3/logging"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/filepaths"
	"github.com/turbot/steampipe/ociinstaller"
	"github.com/turbot/steampipe/statushooks"
)

func Ensure(ctx context.Context) error {
	logging.LogTime("reportassets.Ensure start")
	defer logging.LogTime("reportassets.Ensure end")

	// load report assets versions.json
	versionFile, err := loadReportAssetVersionFile()
	if err != nil {
		return err
	}

	if versionFile.Version == constants.DashboardAssetsVersion {
		return nil
	}

	statushooks.SetStatus(ctx, "Installing reporting server...")
	defer statushooks.Done(ctx)
	reportAssetsPath := filepaths.EnsureDashboardAssetsDir()
	return ociinstaller.InstallAssets(ctx, reportAssetsPath)
}

type ReportAssetsVersionFile struct {
	Version string `json:"version"`
}

func loadReportAssetVersionFile() (*ReportAssetsVersionFile, error) {
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
