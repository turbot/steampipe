package steampipeconfig

import (
	"fmt"
	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/steampipe-plugin-sdk/v4/plugin"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/filepaths"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/pkg/steampipeconfig/parse"
	"os"
)

var GlobalWorkspaceProfile *modconfig.WorkspaceProfile

type WorkspaceProfileLoader struct {
	workspaceProfiles    map[string]*modconfig.WorkspaceProfile
	workspaceProfilePath string
}

func NewWorkspaceProfileLoader(workspaceProfilePath string) (*WorkspaceProfileLoader, error) {
	res := &WorkspaceProfileLoader{workspaceProfilePath: workspaceProfilePath}
	workspaceProfiles, err := res.load()
	if err != nil {
		return nil, err
	}
	res.workspaceProfiles = workspaceProfiles

	// now apply default values to all profiles
	if err := res.setDefaultValues(); err != nil {
		return nil, err
	}

	return res, nil
}

func (l *WorkspaceProfileLoader) load() (map[string]*modconfig.WorkspaceProfile, error) {
	// get all the config files in the directory
	configPaths, err := filehelpers.ListFiles(l.workspaceProfilePath, &filehelpers.ListOptions{
		Flags:   filehelpers.FilesFlat,
		Include: filehelpers.InclusionsFromExtensions([]string{constants.ConfigExtension}),
	})

	if err != nil {
		return nil, err
	}
	if len(configPaths) == 0 {
		return nil, nil
	}

	fileData, diags := parse.LoadFileData(configPaths...)
	if diags.HasErrors() {
		return nil, plugin.DiagsToError("Failed to load workspace profiles", diags)
	}

	body, diags := parse.ParseHclFiles(fileData)
	if diags.HasErrors() {
		return nil, plugin.DiagsToError("Failed to load workspace profiles", diags)
	}

	// do a partial decode
	content, moreDiags := body.Content(parse.WorkspaceProfileListBlockSchema)
	if moreDiags.HasErrors() {
		diags = append(diags, moreDiags...)
		return nil, plugin.DiagsToError("Failed to load workspace profiles", diags)
	}

	profileMap := map[string]*modconfig.WorkspaceProfile{}
	// build parse context
	parseContext := parse.NewParseContext(l.workspaceProfilePath)
	for _, block := range content.Blocks {

		workspaceProfile, res := parse.DecodeWorkspaceProfile(block, parseContext)
		if res.Success() {
			profileMap[workspaceProfile.Name] = workspaceProfile
		}
	}

	if diags.HasErrors() {
		return nil, plugin.DiagsToError("Failed to load config", diags)
	}

	// add in default if needed
	if _, ok := profileMap["default"]; !ok {
		profileMap["default"] = &modconfig.WorkspaceProfile{Name: "default"}
	}

	return profileMap, nil
}

func (l *WorkspaceProfileLoader) Get(name string) (*modconfig.WorkspaceProfile, error) {
	workspaceProfile, ok := l.workspaceProfiles[name]
	if !ok {
		return nil, fmt.Errorf("workspace profile %s does not exist", name)
	}

	return workspaceProfile, nil
}

// set default values on all profiles as necessary
func (l *WorkspaceProfileLoader) setDefaultValues() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	for _, p := range l.workspaceProfiles {
		if p.ModLocation == "" {
			p.ModLocation = cwd
		}
		if p.CloudHost == "" {
			p.CloudHost = constants.DefaultCloudHost
		}
		if p.InstallDir == "" {
			p.InstallDir = filepaths.DefaultInstallDir
		}
		if p.WorkspaceDatabase == "" {
			p.WorkspaceDatabase = constants.DefaultWorkspaceDatabase
		}
	}
	return nil
}
