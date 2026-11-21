package steampipeconfig

import (
	"os"
	"path/filepath"

	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/pipe-fittings/v2/app_specific"
	perror_helpers "github.com/turbot/pipe-fittings/v2/error_helpers"
	pparse "github.com/turbot/pipe-fittings/v2/parse"
)

// ValidConfigDir returns a folder holding only the .spc files in dir which parse, for loaders which
// stop at the first bad file and cannot be told to skip it (the workspace profile loader).
// A folder with no bad file is returned as it is. Otherwise the good files are copied into a scratch
// folder, and a warning is returned for each problem in the dropped files, each naming its file.
// Call cleanup when the loader is done with the folder. A file which cannot be read is not a parse
// error: the original folder is returned so the loader reports it as before.
func ValidConfigDir(dir string) (validDir string, warnings []string, cleanup func(), err error) {
	cleanup = func() {}
	configPaths, err := filehelpers.ListFiles(dir, &filehelpers.ListOptions{
		Flags:   filehelpers.FilesFlat,
		Include: filehelpers.InclusionsFromExtensions([]string{app_specific.ConfigExtension}),
	})
	if err != nil || len(configPaths) == 0 {
		return dir, nil, cleanup, nil
	}
	fileData, diags := pparse.LoadFileData(configPaths...)
	if diags.HasErrors() {
		return dir, nil, cleanup, nil
	}
	valid, dropped := DropFilesWithErrors(fileData)
	if !dropped.HasErrors() {
		return dir, nil, cleanup, nil
	}

	tmpDir, err := os.MkdirTemp("", "steampipe-config-")
	if err != nil {
		return "", nil, cleanup, err
	}
	cleanup = func() { _ = os.RemoveAll(tmpDir) }
	for path, data := range valid {
		if err := os.WriteFile(filepath.Join(tmpDir, filepath.Base(path)), data, 0600); err != nil {
			cleanup()
			return "", nil, func() {}, err
		}
	}
	return tmpDir, perror_helpers.HclDiagsToWarnings(downgradeToWarnings(dropped)), cleanup, nil
}
