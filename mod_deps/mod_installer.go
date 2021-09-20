package mod_deps

import (
	"fmt"
	"os"
	"path/filepath"

	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/steampipe/steampipeconfig"
	"github.com/turbot/steampipe/steampipeconfig/parse"

	"github.com/go-git/go-git/v5/plumbing"

	"github.com/turbot/steampipe/workspace"

	git "github.com/go-git/go-git/v5"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
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
	Workspace *workspace.Workspace
	ModsDir   string
}

func NewModInstaller(workspace *workspace.Workspace) *ModInstaller {
	return &ModInstaller{
		Workspace: workspace,
		ModsDir:   filepath.Join(workspace.Path, ".steampipe/mods"),
	}
}

// InstallModDependencies installs all dependencies of the mod
func (i *ModInstaller) InstallModDependencies(mod *modconfig.Mod) error {
	if mod.Requires == nil {
		return nil
	}

	if err := mod.Requires.ValidateSteampipeVersion(mod.Name()); err != nil {
		return err
	}
	// reset our dependency map
	dependencyMap := make(map[string]*ResolvedModRef)

	var errors []error

	for _, dep := range mod.Requires.Mods {
		// gather dependencies for this mod
		// NOTE - this mutates dependency map
		// get a resolved mod ref for this mod version
		resolvedRef, err := i.GetModRefForVersion(dep)
		if err != nil {
			return fmt.Errorf("dependency %s %s cannot be satisfied: %s", mod.Name, dep.VersionString, err.Error())
		}

		if err := i.GatherDependencies(resolvedRef, dependencyMap); err != nil {
			errors = append(errors, err)
		}
	}
	if len(errors) > 0 {
		return utils.CombineErrorsWithPrefix(fmt.Sprintf("%d dependencies failed to install", len(errors)), errors...)
	}

	// so now i.DependencyMap contains all dependencies - install them
	return i.installDependencies(dependencyMap)
}

func (i *ModInstaller) GatherDependencies(modRef *ResolvedModRef, dependencyMap map[string]*ResolvedModRef) error {

	modRequires, err := i.GetModRequires(modRef)
	if err != nil {
		return err
	}
	if modRequires == nil {
		return nil
	}

	if err := modRequires.ValidateSteampipeVersion(modRef.Name); err != nil {
		return err
	}

	var errors []error
	// for each dependency see whether it is already satisfied
	for _, dep := range modRequires.Mods {

		// have we already identified a dependency for this mod see if it satisfies this requirement
		if modRef, ok := dependencyMap[dep.Name]; ok {
			if modRef.Version.GreaterThanOrEqual(dep.Version) {
				continue
			}
		}
		// so either this dependency is not in the dependency map
		// or the version in the map does not satisfy the requirement
		// see if we can add this version (this checks replacements and workspace lock)
		resolvedRef, err := i.GetModRefForVersion(dep)
		if err != nil {
			errors = append(errors, fmt.Errorf("dependency %s %s cannot be satisfied: %s", dep.Name, dep.VersionString, err.Error()))
			continue
		}
		dependencyMap[dep.Name] = resolvedRef

		// gather dependencies for this dep
		if err := i.GatherDependencies(resolvedRef, dependencyMap); err != nil {
			errors = append(errors, err)
		}

	}
	return utils.CombineErrors(errors...)
}

func (i *ModInstaller) GetModRequires(modRef *ResolvedModRef) (*modconfig.Requires, error) {

	if err := i.installDependency(modRef); err != nil {
		return nil, err
	}

	// build options used to load workspace
	// set flags to create pseudo resources and a default mod if needed
	opts := &parse.ParseModOptions{

		ListOptions: &filehelpers.ListOptions{
			// listFlag specifies whether to load files recursively
			Flags: filehelpers.FilesFlat,
			// ignore hidden files and folders
			Exclude: []string{"**/.*"},
		},
	}

	m, err := steampipeconfig.LoadMod(filepath.Join(i.ModsDir, modRef.Name), opts)
	if err != nil {
		return nil, err
	}
	return m.Requires, nil
}

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
	return &ResolvedModRef{
		Name:         modVersion.Name,
		GitReference: plumbing.NewBranchReferenceName(modVersion.VersionString),
		Version:      modVersion.Version,
		raw:          modVersion.VersionString,
	}, nil
}

func (i *ModInstaller) installDependencies(dependencyMap map[string]*ResolvedModRef) error {

	var errors []error
	for _, dep := range dependencyMap {
		if err := i.installDependency(dep); err != nil {
			errors = append(errors, err)
		}
	}
	return utils.CombineErrors(errors...)
}

func (i *ModInstaller) installDependency(dependency *ResolvedModRef) error {
	if dependency.FilePath != "" {
		// if there is a file path, verify it exists
		if _, err := os.Stat(dependency.FilePath); os.IsNotExist(err) {
			return fmt.Errorf("dependency %s file path %s does not exist", dependency.Name, dependency.FilePath)
		}
		return nil
	}

	return i.installDependencyFromGit(dependency)
}

func (i *ModInstaller) installDependencyFromGit(dependency *ResolvedModRef) error {
	// ensure mod directory exists - create if necessary
	if err := os.MkdirAll(i.ModsDir, os.ModePerm); err != nil {
		return err
	}
	// TODO if the repo is clones, just switch to the approriate branch/tag
	// if it fails, try pulling

	// ig

	// get the mod from git
	_, err := git.PlainClone(i.ModsDir, false, &git.CloneOptions{
		URL:           dependency.Name,
		Progress:      os.Stdout,
		ReferenceName: dependency.GitReference,
	})

	return err
}
