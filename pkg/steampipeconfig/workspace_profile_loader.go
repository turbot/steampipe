package steampipeconfig

import (
	"fmt"
	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/steampipe-plugin-sdk/v4/plugin"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/pkg/steampipeconfig/parse"
	"log"
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

	return res, nil
}

func (l *WorkspaceProfileLoader) load() (map[string]*modconfig.WorkspaceProfile, error) {
	// create profile map to populate
	// create a default profile, which will be overwritten if one is defined
	profileMap := map[string]*modconfig.WorkspaceProfile{"default": {Name: "default"}}

	// get all the config files in the directory
	configPaths, err := filehelpers.ListFiles(l.workspaceProfilePath, &filehelpers.ListOptions{
		Flags:   filehelpers.FilesFlat,
		Include: filehelpers.InclusionsFromExtensions([]string{constants.ConfigExtension}),
	})
	if err != nil {
		return nil, err
	}
	if len(configPaths) == 0 {
		// be sure to return the default
		return profileMap, nil
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
	content, diags := body.Content(parse.WorkspaceProfileListBlockSchema)
	if diags.HasErrors() {

		return nil, plugin.DiagsToError("Failed to load workspace profiles", diags)
	}

	// build parse context
	parseContext := parse.NewParseContext(l.workspaceProfilePath)
	for _, block := range content.Blocks {

		workspaceProfile, res := parse.DecodeWorkspaceProfile(block, parseContext)
		if res.Success() {
			// success - add to map
			profileMap[workspaceProfile.Name] = workspaceProfile
		}
		// TODO handle failure and dependencies
	}

	// add in default if needed
	if _, ok := profileMap["default"]; !ok {
		profileMap["default"] = &modconfig.WorkspaceProfile{Name: "default"}
	}

	//
	return profileMap, nil
}

func (l *WorkspaceProfileLoader) Get(name string) (*modconfig.WorkspaceProfile, error) {
	if workspaceProfile, ok := l.workspaceProfiles[name]; ok {
		return workspaceProfile, nil
	}

	if implicitWorkspace := l.getImplicitWorkspace(name); implicitWorkspace != nil {
		return implicitWorkspace, nil
	}

	return nil, fmt.Errorf("workspace profile %s does not exist", name)
}

/*
Named workspaces follow normal standards for hcl identities, thus they cannot contain the slash (/) character.

If you pass a value to --workspace or STEAMPIPE_WORKSPACE in the form of {identity_handle}/{workspace_handle},
it will be interpreted as an implicit workspace.

Implicit workspaces, as the name suggests, do not need to be specified in the workspaces.spc file.

Instead they will be assumed to refer to a Steampipe Cloud workspace,
which will be used as both the database and snapshot location.

Essentially, --workspace acme/dev is equivalent to:

	workspace "acme/dev" {
	  workspace_database = "acme/dev"
	  snapshot_location  = "acme/dev"
	}
*/
func (l *WorkspaceProfileLoader) getImplicitWorkspace(name string) *modconfig.WorkspaceProfile {
	if IsCloudWorkspaceIdentifier(name) {
		log.Printf("[TRACE] getImplicitWorkspace - %s is implicit workspace: SnapshotLocation=%s, WorkspaceDatabase=%s", name, name, name)
		return &modconfig.WorkspaceProfile{
			SnapshotLocation:  name,
			WorkspaceDatabase: name,
		}
	}
	return nil
}
