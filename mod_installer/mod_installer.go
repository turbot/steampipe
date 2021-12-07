package mod_installer

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/Masterminds/semver"
	git "github.com/go-git/go-git/v5"
	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/steampipe/constants"
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

Plugin Dependency
If the mod specifies plugin versions that are not installed, or have no connections, the user will be warned but the mod will be installed. The user should be warned at installation time, and also when starting Steampipe in the workspace.


Detecting conflicts
mod 1 require a@1.0
mod 2 require a@file:/foo

-> how do we detect if the file version satisfied constrainst of a - this is for dev purposes so always pass?

mod 1 require a@1.0
mod 2 require a@<branch>

-> how do we detect if the file version satisfied constraints of a - check branch?

*/

type ModInstaller struct {
	// all locally installed mod versions
	InstalledModVersions InstalledModMap
	// ALL the available versions for each mod that is required (we populate this in a lazy fashion)
	availableModVersions InstalledModMap
	// map of all installed dependencies, keyed by parent mod name
	workspaceLock modconfig.WorkspaceLock

	// list of mod installed by this installer
	recentlyInstalled []*ResolvedModRef

	modsPath      string
	workspacePath string

	// should we update dependencies to newer versions if they exist
	updateDependencies bool
	// are dependencies being added to the workspace with a get command
	// (if so, ignore locked version)
	GetMods []*modconfig.ModVersionConstraint
	// have specific mods been specified to update
	UpdateMods   map[string]bool
	workspaceMod *modconfig.Mod
}

func NewModInstaller(opts *InstallOpts) (*ModInstaller, error) {
	i := &ModInstaller{
		workspacePath:        opts.WorkspacePath,
		modsPath:             constants.WorkspaceModPath(opts.WorkspacePath),
		InstalledModVersions: make(InstalledModMap),
		availableModVersions: make(InstalledModMap),
		GetMods:              opts.GetMods,
		updateDependencies:   opts.ShouldUpdate,
	}

	// build list of ALL currently installed mods - by searching through the mods folder and finding mod.sp files
	installedMods, err := i.getInstalledMods()
	if err != nil {
		return nil, err
	}
	i.InstalledModVersions = installedMods

	// load sum file - ignore errors
	i.workspaceLock, _ = modconfig.LoadWorkspaceLock(i.workspacePath)
	// if we failed to load, create an empty map
	if i.workspaceLock == nil {
		log.Printf("[TRACE] no workspace.lock file loaded - creating a new one")
		i.workspaceLock = make(modconfig.WorkspaceLock)
	}

	// now load workspace mod, creating a default if needed
	mod, err := i.loadModfile(i.workspacePath, true)
	if err != nil {
		return nil, err
	}
	i.workspaceMod = mod

	if i.updateDependencies {
		// ensure we have a lock file, and that all mods requested to update are in it
		if err := i.verifyUpdates(opts.UpdateMods); err != nil {
			return nil, err
		}
	}

	return i, nil
}

// InstallWorkspaceDependencies installs all dependencies of the workspace mod
func (i *ModInstaller) InstallWorkspaceDependencies() error {
	// if additional dependencies were specified, add to the mod
	if len(i.GetMods) > 0 {
		i.workspaceMod.AddModDependencies(i.GetMods)
	}

	if i.workspaceMod.Requires == nil {
		return nil
	}

	mod := i.workspaceMod
	// first check our Steampipe version is sufficient
	if err := mod.Requires.ValidateSteampipeVersion(mod.Name()); err != nil {
		return err
	}

	//dependenciesToInstall := mod.Requires.Mods
	// if
	//
	if err := i.installModDependenciesRecursively(mod.Requires.Mods, mod); err != nil {
		return err
	}
	// write sum file
	if err := i.workspaceLock.Save(i.workspacePath); err != nil {
		return err
	}

	// if additional dependencies were added, save the mod file
	if len(i.GetMods) > 0 {
		mod.AddModDependencies(i.GetMods)
		if err := i.workspaceMod.Save(); err != nil {
			return err
		}
	}
	return nil
}

func (i *ModInstaller) installModDependenciesRecursively(mods []*modconfig.ModVersionConstraint, parent *modconfig.Mod) error {
	var errors []error

	for _, requiredModVersion := range mods {
		// get or create the installation data for this mod, adding in this mod version constraint
		availableVersions, err := i.getAvailableModVersions(requiredModVersion)
		if err != nil {
			errors = append(errors, err)
			continue
		}
		// if we are NOT updating, and have a locked version, update the required version to reflect this
		if i.respectLockFile() {
			lockedVersion, err := i.workspaceLock.GetLockedModVersion(requiredModVersion, parent)
			if err != nil {
				return err
			}
			if lockedVersion != nil {
				requiredModVersion = lockedVersion
			}
		}

		// check whether there is already a version which satisfies this mod version
		installedMod, err := i.getInstalledVersionForConstraint(requiredModVersion, availableVersions)
		if err != nil {
			return err
		}

		if installedMod != nil {
			// so we found an existing mod which will satisfy this requirement
			// update the workspace lock
			i.workspaceLock.Add(parent.Name(), requiredModVersion.Name, installedMod.Version)
			log.Printf("[TRACE] not installing %s with version constraint %s as version %s is already installed", requiredModVersion.Name, requiredModVersion.Constraint.Original, installedMod.Version)
		} else {
			// so we ARE installing

			// get a resolved mod ref that satisfies the version constraints
			resolvedRef, err := i.getModRefSatisfyingConstraints(requiredModVersion, availableVersions)
			if err != nil {
				errors = append(errors, fmt.Errorf("dependency %s %s cannot be satisfied: %s", requiredModVersion.Name, requiredModVersion.VersionString, err.Error()))
				continue
			}
			// install the mod
			installedMod, err = i.install(resolvedRef, parent)
			if err != nil {
				errors = append(errors, err)
			}
			if installedMod == nil {
				// this is unexpected but just ignore
				log.Printf("[TRACE] dependency %s does not define a mod definition - so there are no child dependencies to install", resolvedRef.Name)
				continue
			}
		}

		// to get here we have the dependency mod - either we installed it or it was already installed
		// recursively install its dependencies
		if err := i.installModDependenciesRecursively(installedMod.Requires.Mods, installedMod); err != nil {
			return err
		}

	}

	return utils.CombineErrorsWithPrefix(fmt.Sprintf("%d dependencies failed to install", len(errors)), errors...)
}

// if we are updating or 'getting' mods, do NOT respect the lock file
func (i *ModInstaller) respectLockFile() bool {
	return !i.updateDependencies && len(i.GetMods) == 0
}

// check whether there is a mod version installed that satisfies the version constraint (and update requirements)
func (i *ModInstaller) getInstalledVersionForConstraint(requiredModVersion *modconfig.ModVersionConstraint, availableVersions []*semver.Version) (*modconfig.Mod, error) {
	log.Printf("[TRACE] getInstalledVersionForConstraint required version %v", requiredModVersion)

	installedVersion := i.InstalledModVersions.GetVersionSatisfyingRequirement(requiredModVersion)
	if installedVersion == nil {
		log.Printf("[TRACE] no version of %s installed which satisfies version constrain %s", requiredModVersion.Name, requiredModVersion.Constraint.Original)
		return nil, nil
	}
	log.Printf("[TRACE] found installed version %s@%s", requiredModVersion.Name, installedVersion.Original())

	// so there IS a version installed which satisfies the constraint.
	// if we are updating, see if there is a newer verison
	if i.shouldUpdate(requiredModVersion.Name) {
		// so we should update if there is a newer version - check if there is
		newerModVersionFound, err := i.newerModVersionFound(requiredModVersion, installedVersion, availableVersions)
		if err != nil {
			return nil, err
		}
		if newerModVersionFound {
			// there is a newer version so we will NOT use the installed version - return nil
			return nil, nil
		}
	}

	// so we resolved an installed version which will satisfy
	// load the mod
	modPath := filepath.Join(i.modsPath, modVersionFullName(requiredModVersion.Name, installedVersion))
	installedMod, err := i.loadModfile(modPath, false)
	if err != nil {
		return nil, err
	}

	return installedMod, err
}

// determine whether there is a newer mod version avoilable which satisfies the dependency version constraint
func (i *ModInstaller) newerModVersionFound(requiredVersion *modconfig.ModVersionConstraint, currentVersion *semver.Version, availableVersions []*semver.Version) (bool, error) {
	latestVersion, err := i.getModRefSatisfyingConstraints(requiredVersion, availableVersions)
	if err != nil {
		return false, err
	}
	if latestVersion.Version.GreaterThan(currentVersion) {
		return true, nil
	}
	return false, nil
}

// getInstalledMods returns a map installed mods, and the versions installed for each
func (i *ModInstaller) getInstalledMods() (InstalledModMap, error) {
	// recursively search for all the mod.sp files under the .steampipe/mods folder, then build the mod name from the file path
	modFiles, err := filehelpers.ListFiles(i.modsPath, &filehelpers.ListOptions{
		Flags:   filehelpers.FilesRecursive,
		Include: []string{"**/mod.sp"},
	})
	if err != nil {
		return nil, err
	}

	// create result map - a list of version for each mod
	installedMods := make(InstalledModMap, len(modFiles))
	// collect errors
	var errors []error

	for _, modfilePath := range modFiles {
		modName, version, err := i.parseModPath(modfilePath)
		if err != nil {
			errors = append(errors, err)
			continue
		}
		// add this mod version to the map
		installedMods.Add(modName, version)
	}

	if len(errors) > 0 {
		return nil, utils.CombineErrors(errors...)
	}
	return installedMods, nil
}

// retrieve all available mod versions from Git
func (i *ModInstaller) getAvailableModVersions(modVersion *modconfig.ModVersionConstraint) ([]*semver.Version, error) {
	// have we already loaded the versions for this mod
	availableVersions, ok := i.availableModVersions[modVersion.Name]
	if !ok {
		var err error
		availableVersions, err = getTagVersionsFromGit(getGitUrl(modVersion.Name))
		if err != nil {
			return nil, err
		}

	}

	// update our cache
	i.availableModVersions[modVersion.Name] = availableVersions

	return availableVersions, nil
}

// install a mod
func (i *ModInstaller) install(dependency *ResolvedModRef, parent *modconfig.Mod) (_ *modconfig.Mod, err error) {
	defer func() {
		if err == nil {
			i.onModInstalled(dependency, parent)
		}
	}()
	var modPath string

	if dependency.FilePath != "" {
		// if there is a file path, verify it exists
		if _, err := os.Stat(dependency.FilePath); os.IsNotExist(err) {
			return nil, fmt.Errorf("dependency %s file path %s does not exist", dependency.Name, dependency.FilePath)
		}
		modPath = dependency.FilePath
	} else {
		modPath = filepath.Join(i.modsPath, dependency.FullName())

		// if the target path exists, this is a bug - we should never try to install over an existing directory
		if _, err := os.Stat(modPath); !os.IsNotExist(err) {
			return nil, fmt.Errorf("mod %s is already installed", dependency.FullName())
		}

		if err := i.installFromGit(dependency, modPath); err != nil {
			return nil, err
		}
	}

	// now load the installed mod and return it
	return i.loadModfile(modPath, false)

}

func (i *ModInstaller) installFromGit(dependency *ResolvedModRef, installPath string) error {
	// ensure mod directory exists - create if necessary
	if err := os.MkdirAll(i.modsPath, os.ModePerm); err != nil {
		return err
	}

	// NOTE: we need to check existing installed mods

	// get the mod from git
	gitUrl := getGitUrl(dependency.Name)
	_, err := git.PlainClone(installPath,
		false,
		&git.CloneOptions{
			URL: gitUrl,
			//Progress:      os.Stdout,
			ReferenceName: dependency.GitReference,
			Depth:         1,
			SingleBranch:  true,
		})

	return err
}

// update workspaceLock, InstalledModVersions and recentlyInstalled
func (i *ModInstaller) onModInstalled(dependency *ResolvedModRef, parent *modconfig.Mod) {
	// update workspace lock
	i.workspaceLock.Add(parent.Name(), dependency.Name, dependency.Version)
	// update the maps of all installed mods
	i.InstalledModVersions.Add(dependency.Name, dependency.Version)
	// update list of recently installed items
	i.recentlyInstalled = append(i.recentlyInstalled, dependency)
}

func (i *ModInstaller) loadModfile(modPath string, createDefault bool) (*modconfig.Mod, error) {
	if !parse.ModfileExists(modPath) {
		if createDefault {
			return modconfig.CreateDefaultMod(i.workspacePath), nil
		}
		return nil, nil
	}
	return parse.ParseModDefinition(modPath)
}

// get the most recent available mod version which satisfies the version constraint
func (i *ModInstaller) getModRefSatisfyingConstraints(modVersion *modconfig.ModVersionConstraint, availableVersions []*semver.Version) (*ResolvedModRef, error) {
	// find a git tag which satisfies the version constraint
	var version, err = i.getGitVersionSatisfyingConstraint(modVersion, availableVersions)
	if err != nil {
		return nil, err
	}
	if version == nil {
		return nil, fmt.Errorf("no version of %s found satisfying verison constraints: %s", modVersion.Name, modVersion.Constraint.Original)
	}

	return NewResolvedModRef(modVersion, version)
}

func (i *ModInstaller) getGitVersionSatisfyingConstraint(modVersion *modconfig.ModVersionConstraint, availableVersions []*semver.Version) (*semver.Version, error) {
	// search the reverse sorted versions, finding the highest version which satisfies ALL constraints
	for _, version := range availableVersions {
		if modVersion.Constraint.Check(version) {
			return version, nil
		}
	}
	return nil, nil
}

// extract the mod name and version from the modfile path
func (i *ModInstaller) parseModPath(modfilePath string) (modName string, modVersion *semver.Version, err error) {
	modLongName, err := filepath.Rel(i.modsPath, filepath.Dir(modfilePath))
	if err != nil {
		return
	}
	// we expect modLongName to be of form github.com/turbot/steampipe-mod-m2@v1.0
	// split to get the name and version
	parts := strings.Split(modLongName, "@")
	if len(parts) != 2 {
		err = fmt.Errorf("invalid mod path %s", modfilePath)
		return
	}
	modName = parts[0]
	modVersion, err = semver.NewVersion(parts[1])
	if err != nil {
		err = fmt.Errorf("mod path %s has invalid version", modfilePath)
		return
	}
	return
}

func (i *ModInstaller) InstallReport() string {
	if len(i.recentlyInstalled) == 0 {
		// TODO think about have we used an already installed mod for a get/update
		// mod declares no dependencies
		return "No dependencies installed"
	}
	strs := make([]string, len(i.recentlyInstalled))
	idx := 0
	for _, ref := range i.recentlyInstalled {
		strs[idx] = fmt.Sprintf("%s@%s", ref.Name, ref.Version.String())
		idx++
	}
	return fmt.Sprintf("\nInstalled %d %s:\n  - %s\n", len(i.recentlyInstalled), utils.Pluralize("dependency", len(i.recentlyInstalled)), strings.Join(strs, "\n  - "))
}
