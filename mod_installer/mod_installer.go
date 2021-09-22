package mod_installer

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/turbot/steampipe/constants"

	git "github.com/go-git/go-git/v5"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/steampipeconfig/parse"
	"github.com/turbot/steampipe/utils"
)

/*
mog get

A user may install a mod with steampipe mod get modname[@version]

version may be:

- Not specified: steampipe mod get aws-cis
	The latest version (highest version tag) will be installed.
	A dependency is added to the requires block specifying the version that was downloaded
- A major version: steampipe mod get aws-cis@3
	The latest release (highest version tag) of the specified major version will be installed.
	A dependency is added to the requires block specifying the version that was downloaded
- A monotonic version tag: steampipe mod get aws-cis@v2.21
	The specified version is downloaded and added as requires dependency.
- A branch name: steampipe mod get aws-cis@staging
	The current version of the specified branch is downloaded.
	The branch dependency is added to the requires list. Note that a branch is considered a distinct "major" release, it is not cached in the registry, and has no minor version.
	Branch versions do not auto-update - you have to run steampipe mod update to get a newer version.
	Branch versioning is meant to simplify development and testing - published mods should ONLY include version tag dependencies, NOT branch dependencies.
- A local file path: steampipe mod get "file:~/my_path/aws-core"
	The mod from the local filesystem is added to the namespace, but nothing is downloaded.
	The local dependency is added to the requires list. Note that a local mod is considered a distinct "major" release, it is not cached in the registry, and has no minor version.
	Local versioning is meant to simplify development and testing - published mods should ONLY include version tag dependencies, NOT local dependencies.


Steampipe Version Dependency
If the installed version of Steampipe does not meet the dependency criteria, the user will be warned and the mod will NOT be installed.

Plugin Dependency5
If the mod specifies plugin versions that are not installed, or have no connections, the user will be warned but the mod will be installed. The user should be warned at installation time, and also when starting Steampipe in the workspace.


Detecting conflicts
mod 1 require a@1.0
mod 2 require a@file:/foo

-> how do we detect if the file version satisfied constrainst of a - this is for dev purposes so always pass?

mod 1 require a@1.0
mod 2 require a@<branch>

-> how do we detect if the file version satisfied constraints of a - check branch?

Checking whethr version is satisfied


MOD INSTA
*/

type ModInstaller struct {
	ModsDir string
}

func NewModInstaller(workspacePath string) *ModInstaller {
	return &ModInstaller{
		ModsDir: constants.WorkspaceModPath(workspacePath),
	}
}

// InstallModDependencies installs all dependencies of the mod
func (i *ModInstaller) InstallModDependencies(mod *modconfig.Mod) error {
	dependencyMap := make(map[string]*ResolvedModRef)
	return i.installModDependenciesRecursively(mod, dependencyMap)
}

func (i *ModInstaller) installModDependenciesRecursively(mod *modconfig.Mod, dependencyMap map[string]*ResolvedModRef) error {
	if mod.Requires == nil {
		return nil
	}

	// first check our Steampipe version is sufficient
	if err := mod.Requires.ValidateSteampipeVersion(mod.Name()); err != nil {
		return err
	}

	var errors []error
	for _, modVersion := range mod.Requires.Mods {
		// get a resolved mod ref for this mod version
		resolvedRef, err := i.GetModRefForVersion(modVersion)
		if err != nil {
			return fmt.Errorf("dependency %s %s cannot be satisfied: %s", mod.Name, modVersion.VersionString, err.Error())
		}

		// install this mod
		// NOTE - this mutates dependency map
		if err := i.installDependency(resolvedRef, dependencyMap); err != nil {
			errors = append(errors, err)
		}
	}

	return utils.CombineErrorsWithPrefix(fmt.Sprintf("%d dependencies failed to install", len(errors)), errors...)
}

//func (i *ModInstaller) GatherDependencies(modRef *ResolvedModRef, dependencyMap map[string]*ResolvedModRef) error {
//
//	modRequires, err := i.GetModRequires(modRef)
//	if err != nil {
//		return err
//	}
//	if modRequires == nil {
//		return nil
//	}
//
//	if err := modRequires.ValidateSteampipeVersion(modRef.Name); err != nil {
//		return err
//	}
//
//	var errors []error
//	// for each dependency see whether it is already satisfied
//	for _, dep := range modRequires.Mods {
//
//		// have we already identified a dependency for this mod see if it satisfies this requirement
//		if modRef, ok := dependencyMap[dep.Name]; ok {
//			if modRef.Version.GreaterThanOrEqual(dep.VersionConstraint) {
//				continue
//			}
//		}
//		// so either this dependency is not in the dependency map
//		// or the version in the map does not satisfy the requirement
//		// see if we can add this version (this checks replacements and workspace lock)
//		resolvedRef, err := i.GetModRefForVersion(dep)
//		if err != nil {
//			errors = append(errors, fmt.Errorf("dependency %s %s cannot be satisfied: %s", dep.Name, dep.VersionString, err.Error()))
//			continue
//		}
//		dependencyMap[dep.Name] = resolvedRef
//
//		// gather dependencies for this dep
//		if err := i.GatherDependencies(resolvedRef, dependencyMap); err != nil {
//			errors = append(errors, err)
//		}
//
//	}
//	return utils.CombineErrors(errors...)
//}

//func (i *ModInstaller) GetModRequires(modRef *ResolvedModRef) (*modconfig.Requires, error) {
//	modPath, err := i.installDependency(modRef, nil)
//	if err != nil {
//		return nil, err
//	}
//
//	m, err := parse.ParseModDefinition(modPath)
//	if err != nil {
//		return nil, err
//	}
//	return m.Requires, nil
//}

func (i *ModInstaller) GetModRefForVersion(modVersion *modconfig.ModVersion) (*ResolvedModRef, error) {

	// TODO check whether the lock file contains this dependency and if so
	//  does the locked version satisy this version requirement
	// return error if not

	// TODO check whether we are replacing this version
	// if so does the locked version satisy this version requirement
	// return error if not

	// so we need to resolve this mod version
	// TODO  for now assume github
	// get the most recent minor version fo rthis major version from github
	return i.getLatestCompatibleVersionFromGithub(modVersion)
}

func (i *ModInstaller) getLatestCompatibleVersionFromGithub(modVersion *modconfig.ModVersion) (*ResolvedModRef, error) {
	// TODO for now assume the mod is specified with a full version
	return NewResolvedModRef(modVersion)
}

func (i *ModInstaller) installDependency(dependency *ResolvedModRef, dependencyMap map[string]*ResolvedModRef) error {
	// have we already installed a mod which satisfies this dependency
	if modRef, ok := dependencyMap[dependency.Name]; ok {
		if modRef.Version.GreaterThanOrEqual(dependency.Version) {
			return nil
		}
	}

	// add this dependency into the map (if we fail to install,m the whole installation process will terminate,
	// so no need to check for errors
	dependencyMap[dependency.Name] = dependency

	var modPath string
	if dependency.FilePath != "" {
		// if there is a file path, verify it exists
		if _, err := os.Stat(dependency.FilePath); os.IsNotExist(err) {
			return fmt.Errorf("dependency %s file path %s does not exist", dependency.Name, dependency.FilePath)
		}
		modPath = dependency.FilePath
	} else {
		modPath = filepath.Join(i.ModsDir, dependency.FullName())
		if err := i.installDependencyFromGit(dependency, modPath); err != nil {
			return err
		}
	}
	// no load the installed mod and install _its_ dependencies
	mod, err := parse.ParseModDefinition(modPath)
	if err != nil {
		return err
	}
	return i.installModDependenciesRecursively(mod, dependencyMap)

}

func (i *ModInstaller) installDependencyFromGit(dependency *ResolvedModRef, installPath string) error {
	// ensure mod directory exists - create if necessary
	if err := os.MkdirAll(i.ModsDir, os.ModePerm); err != nil {
		return err
	}
	// TODO if the repo is cloned, just switch to the approriate branch/tag
	// TODO HACK FOR NOW

	// if it fails, try pulling

	// get the mod from git

	gitUrl := fmt.Sprintf("https://%s", dependency.Name)
	_, err := git.PlainClone(installPath,
		false,
		&git.CloneOptions{
			URL:           gitUrl,
			Progress:      os.Stdout,
			ReferenceName: dependency.GitReference,
			Depth:         1,
			SingleBranch:  true,
		})

	return err
}
