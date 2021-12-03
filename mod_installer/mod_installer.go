package mod_installer

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	git "github.com/go-git/go-git/v5"
	goVersion "github.com/hashicorp/go-version"
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

Plugin Dependency5
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
	ModsDir string
	// all installed mod versions
	AllInstalledMods InstalledModMap
	// dependencies installed during the current installation process
	NewInstalledMods InstalledModMap
	// should we update dependencies to newer versions if they exist
	ShouldUpdate bool
}

func NewModInstaller(workspacePath string) (*ModInstaller, error) {
	i := &ModInstaller{
		ModsDir:          constants.WorkspaceModPath(workspacePath),
		NewInstalledMods: make(InstalledModMap),
	}

	// build list of currently installed mods
	installedMods, err := i.getInstalledMods()
	if err != nil {
		return nil, err
	}
	i.AllInstalledMods = installedMods
	return i, nil
}

// InstallModDependencies installs all dependencies of the mod
func (i *ModInstaller) InstallModDependencies(mod *modconfig.Mod) error {
	if mod.Requires == nil {
		return nil
	}

	// first check our Steampipe version is sufficient
	if err := mod.Requires.ValidateSteampipeVersion(mod.Name()); err != nil {
		return err
	}

	return i.installModDependenciesRecursively(mod.Requires.Mods)
}

// getInstalledMods returns a map installed mods, and the versions installed for each
func (i *ModInstaller) getInstalledMods() (InstalledModMap, error) {
	// recursively search for all the mod.sp files under the .steampipe/mods folder, then build the mod name from the file path
	modFiles, err := filehelpers.ListFiles(i.ModsDir, &filehelpers.ListOptions{
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

func (i *ModInstaller) installModDependenciesRecursively(mods []*modconfig.ModVersion) error {
	var errors []error
	for _, modVersion := range mods {
		// check whether there is alreayd a version which satisfies this dependency
		// and if so - should we update it?
		shouldInstall, err := i.shouldInstallMod(modVersion)
		if err != nil {
			errors = append(errors, err)
		}
		if !shouldInstall {
			log.Printf("[TRACE] not installing %s with version constraint %s as it is already installed", modVersion.Name, modVersion.VersionConstraint.String())
			continue
		}

		// get a resolved mod ref that satisfies the version constraint for this mod version
		resolvedRef, err := i.getModRefForVersion(modVersion)
		if err != nil {
			errors = append(errors, fmt.Errorf("dependency %s %s cannot be satisfied: %s", modVersion.Name, modVersion.VersionString, err.Error()))
		}

		// install the mod
		err = i.install(resolvedRef)
		if err != nil {
			errors = append(errors, err)
		}
	}

	return utils.CombineErrorsWithPrefix(fmt.Sprintf("%d dependencies failed to install", len(errors)), errors...)
}

// get the most recent available mod version which satidsfies the version constraint
func (i *ModInstaller) getModRefForVersion(modVersion *modconfig.ModVersion) (*ResolvedModRef, error) {
	// TODO check whether the lock file contains this dependency and if so
	// does the locked version satisfy this version requirement-  return error if not
	// TODO check whether we are replacing this version

	// find a git tag which satisfies the version constraint
	version, err := i.getGitVersionSatisfyingConstraint(modVersion)
	if err != nil {
		return nil, err
	}
	if version == nil {
		return nil, fmt.Errorf("no version of %s found satisfying verison constraint %s", modVersion.Name, modVersion.VersionString)
	}

	return NewResolvedModRef(modVersion, version)
}

func (i *ModInstaller) getGitVersionSatisfyingConstraint(modVersion *modconfig.ModVersion) (*goVersion.Version, error) {
	modReporUrl := i.getGitUrl(modVersion.Name)
	// get a list of all tags, with corresponding versions - these come back reverse sorted
	sortedVersions, err := GetTagVersionsFromGit(modReporUrl)
	if err != nil {
		return nil, err
	}

	// search the reverse sorted versions, finding the highest version which satisfies the constraint
	for _, version := range sortedVersions {
		if modVersion.VersionConstraint.Check(version) {
			return version, nil
		}
	}
	return nil, nil
}

func (i *ModInstaller) install(dependency *ResolvedModRef) (err error) {
	defer func() {
		if err != nil {
			// update the maps of installed mods
			i.NewInstalledMods.Add(dependency.Name, dependency.Version)
			i.AllInstalledMods.Add(dependency.Name, dependency.Version)
		}
	}()
	var modPath string

	if dependency.FilePath != "" {
		// if there is a file path, verify it exists
		if _, err := os.Stat(dependency.FilePath); os.IsNotExist(err) {
			return fmt.Errorf("dependency %s file path %s does not exist", dependency.Name, dependency.FilePath)
		}
		modPath = dependency.FilePath
	} else {
		modPath = filepath.Join(i.ModsDir, dependency.FullName())

		// if the target path exists, this is a bug - we should never try to install over an existing directory
		if _, err := os.Stat(modPath); !os.IsNotExist(err) {
			return fmt.Errorf("mod %s is already installed", dependency.FullName())
		}

		if err := i.installDependencyFromGit(dependency, modPath); err != nil {
			return err
		}
	}

	// now load the installed mod and recusrively install _its_ dependencies
	if !parse.ModfileExists(modPath) {
		log.Printf("[TRACE] dependency %s does not define a mod definition - so there are no dependencies to install", dependency.Name)
		return nil
	}

	mod, err := parse.ParseModDefinition(modPath)
	if err != nil {
		return err
	}

	return i.installModDependenciesRecursively(mod.Requires.Mods)
}

func (i *ModInstaller) installDependencyFromGit(dependency *ResolvedModRef, installPath string) error {
	// ensure mod directory exists - create if necessary
	if err := os.MkdirAll(i.ModsDir, os.ModePerm); err != nil {
		return err
	}

	// NOTE: we need to check existing installed mods

	// get the mod from git
	gitUrl := i.getGitUrl(dependency.Name)
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

func (i *ModInstaller) getGitUrl(modName string) string {
	return fmt.Sprintf("https://%s", modName)
}

func (i *ModInstaller) InstallReport() string {
	if len(i.NewInstalledMods) == 0 {
		return "No dependencies installed"
	}
	strs := make([]string, len(i.NewInstalledMods))
	idx := 0
	for name, versions := range i.NewInstalledMods {
		for _, version := range versions {
			strs[idx] = fmt.Sprintf("%s@%s", name, version)
		}
	}
	return fmt.Sprintf("\nInstalled %d dependencies:\n  - %s\n", len(i.NewInstalledMods), strings.Join(strs, "\n  - "))

}

// extract the mod name and version from the modfile path
func (i *ModInstaller) parseModPath(modfilePath string) (modName string, modVersion *goVersion.Version, err error) {
	modLongName, err := filepath.Rel(i.ModsDir, filepath.Dir(modfilePath))
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
	modVersion, err = goVersion.NewVersion(parts[1])
	if err != nil {
		err = fmt.Errorf("mod path %s has invalid version", modfilePath)
		return
	}
	return
}

func (i *ModInstaller) shouldInstallMod(requiredVersion *modconfig.ModVersion) (bool, error) {
	// have we already got a version installed which satisfies this dependency?
	currentVersion := i.AllInstalledMods.GetVersionSatisfyingRequirement(requiredVersion)
	if currentVersion == nil {
		// no version installed - we SHOULD install it
		return true, nil
	}

	// so there is a version installed which satisfies the requirement

	// if the installer has updates disabled, we should NOT install
	if i.ShouldUpdate == false {
		return false, nil
	}

	// so we should update if there is a newer version - check if there is
	newerModVersionFound, err := i.newerModVersionFound(requiredVersion, currentVersion)
	if err != nil {
		return false, err
	}

	return newerModVersionFound, nil
}

// determine whether there is a newer mod version avoilable which satisfies the dependency version constraint
func (i *ModInstaller) newerModVersionFound(requiredVersion *modconfig.ModVersion, currentVersion *goVersion.Version) (bool, error) {
	latestVersion, err := i.getModRefForVersion(requiredVersion)
	if err != nil {
		return false, err
	}
	if latestVersion.Version.GreaterThan(currentVersion) {
		return true, nil
	}
	return false, nil
}
