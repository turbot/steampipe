package steampipeconfig

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/pkg/steampipeconfig/parse"
	"github.com/turbot/steampipe/pkg/utils"
	"github.com/turbot/steampipe/sperr"
)

var GlobalWorkspaceProfile *modconfig.WorkspaceProfile
var defaultWorkspaceSampleFileName = "workspaces.spc.sample"

type WorkspaceProfileLoader struct {
	workspaceProfiles    map[string]*modconfig.WorkspaceProfile
	workspaceProfilePath string
	DefaultProfile       *modconfig.WorkspaceProfile
	ConfiguredProfile    *modconfig.WorkspaceProfile
}

func ensureDefaultWorkspaceFile(configFolder string) error {
	// always write the workspaces.spc.sample file
	err := os.MkdirAll(configFolder, 0755)
	if err != nil {
		return err
	}
	defaultWorkspaceSampleFile := filepath.Join(configFolder, defaultWorkspaceSampleFileName)
	err = os.WriteFile(defaultWorkspaceSampleFile, []byte(constants.DefaultWorkspaceContent), 0755)
	if err != nil {
		return err
	}
	return nil
}

func NewWorkspaceProfileLoader(workspaceProfilePath string) (*WorkspaceProfileLoader, error) {
	// write the workspaces.spc.sample file
	if err := ensureDefaultWorkspaceFile(workspaceProfilePath); err != nil {
		return nil,
			sperr.WrapWithMessage(
				err,
				"could not create sample workspace",
			)
	}
	loader := &WorkspaceProfileLoader{workspaceProfilePath: workspaceProfilePath}
	workspaceProfiles, err := loader.load()
	if err != nil {
		return nil, err
	}
	loader.workspaceProfiles = workspaceProfiles

	defaultProfile, err := loader.get("default")
	if err != nil {
		// there must always be a default - this should have been added by parse.LoadWorkspaceProfiles
		return nil, err
	}
	loader.DefaultProfile = defaultProfile

	if viper.IsSet(constants.ArgWorkspaceProfile) {
		configuredProfile, err := loader.get(viper.GetString(constants.ArgWorkspaceProfile))
		if err != nil {
			// could not find configured profile
			return nil, err
		}
		loader.ConfiguredProfile = configuredProfile
	}

	return loader, nil
}

func (l *WorkspaceProfileLoader) GetActiveWorkspaceProfile() *modconfig.WorkspaceProfile {
	if l.ConfiguredProfile != nil {
		return l.ConfiguredProfile
	}
	return l.DefaultProfile
}

func (l *WorkspaceProfileLoader) get(name string) (*modconfig.WorkspaceProfile, error) {
	if workspaceProfile, ok := l.workspaceProfiles[name]; ok {
		return workspaceProfile, nil
	}

	if implicitWorkspace := l.getImplicitWorkspace(name); implicitWorkspace != nil {
		return implicitWorkspace, nil
	}

	return nil, fmt.Errorf("workspace profile %s does not exist", name)
}

func (l *WorkspaceProfileLoader) load() (map[string]*modconfig.WorkspaceProfile, error) {
	// get all the config files in the directory
	return parse.LoadWorkspaceProfiles(l.workspaceProfilePath)
}

/*
Named workspaces follow normal standards for hcl identities, thus they cannot contain the slash (/) character.

If you pass a value to --workspace or STEAMPIPE_WORKSPACE in the form of {identity_handle}/{workspace_handle},
it will be interpreted as an implicit workspace.

Implicit workspaces, as the name suggests, do not need to be specified in the workspaces.spc file.

Instead they will be assumed to refer to a Turbot Pipes workspace,
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
			SnapshotLocation:  utils.ToStringPointer(name),
			WorkspaceDatabase: utils.ToStringPointer(name),
		}
	}
	return nil
}
