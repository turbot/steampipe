package mod_installer

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/viper"

	"github.com/turbot/steampipe/steampipeconfig/version_map"

	"github.com/Masterminds/semver"
	git "github.com/go-git/go-git/v5"
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
	workspaceMod  *modconfig.Mod
	modsPath      string
	workspacePath string

	installData *InstallData

	// should we update dependencies to newer versions if they exist
	shouldUpdate bool
	// are dependencies being added to the workspace
	mods version_map.VersionConstraintMap
}

func NewModInstaller(opts *InstallOpts) (*ModInstaller, error) {
	i := &ModInstaller{
		workspacePath: opts.WorkspacePath,
		modsPath:      constants.WorkspaceModPath(opts.WorkspacePath),
		shouldUpdate:  opts.Updating,
		mods:          opts.ModArgs,
	}

	// load workspace mod, creating a default if needed
	mod, err := i.loadModfile(i.workspacePath, true)
	if err != nil {
		return nil, err
	}
	i.workspaceMod = mod

	// load lock file
	workspaceLock, err := version_map.LoadWorkspaceLock(i.workspacePath)
	if err != nil {
		return nil, err
	}

	// create install data
	i.installData = NewInstallData(workspaceLock)

	// TODO think about if we need verifyCanUpdate
	//// if we are updating ensure we have a non empty lock file, and that all mods requested to update are in it
	//if err := i.setModArgs(opts); err != nil {
	//	return nil, err
	//}
	return i, nil
}

func (i *ModInstaller) UninstallWorkspaceDependencies() error {
	workspaceMod := i.workspaceMod

	// if no mods specified, just delete the lock file and tidy
	if len(i.mods) == 0 {
		workspaceMod.RemoveAllModDependencies()
	} else {
		workspaceMod.RemoveModDependencies(i.mods)
	}
	if err := i.installMods(workspaceMod.Require.Mods, workspaceMod); err != nil {
		return err
	}

	if workspaceMod.Require.Empty() {
		workspaceMod.Require = nil
	}

	// if this is a dry run, return now
	if viper.GetBool(constants.ArgDryRun) {
		log.Printf("[TRACE] UninstallWorkspaceDependencies - dry-run=true, returning before saving mod file and cache\n")
		return nil
	}

	// write the lock file
	if err := i.installData.Lock.Save(); err != nil {
		return err
	}

	//  now safe to save the mod file
	if err := i.workspaceMod.Save(); err != nil {
		return err
	}

	// tidy unused mods
	if viper.GetBool(constants.ArgPrune) {
		if _, err := i.Tidy(); err != nil {
			return err
		}
	}

	return nil
}

// InstallWorkspaceDependencies installs all dependencies of the workspace mod
func (i *ModInstaller) InstallWorkspaceDependencies() error {
	workspaceMod := i.workspaceMod

	// first check our Steampipe version is sufficient
	if err := workspaceMod.Require.ValidateSteampipeVersion(workspaceMod.Name()); err != nil {
		return err
	}

	// if mod args have been provided, add them to the the workspace mod requires
	// (this will replace any existing dependencies of same name)
	if len(i.mods) > 0 {
		workspaceMod.AddModDependencies(i.mods)
	}

	// if there are no dependencies, we have nothing to do
	if !workspaceMod.HasDependentMods() {
		// there are no dependencies - delete the cache
		i.installData.Lock.Delete()
		return nil
	}

	if err := i.installMods(workspaceMod.Require.Mods, workspaceMod); err != nil {
		return err
	}

	// if this is a dry run, return now
	if viper.GetBool(constants.ArgDryRun) {
		log.Printf("[TRACE] InstallWorkspaceDependencies - dry-run=true, returning before saving mod file and cache\n")
		return nil
	}

	// write the lock file
	if err := i.installData.Lock.Save(); err != nil {
		return err
	}

	//  now safe to save the mod file
	if len(i.mods) > 0 {
		if err := i.workspaceMod.Save(); err != nil {
			return err
		}
	}

	// tidy unused mods
	if viper.GetBool(constants.ArgPrune) {
		if _, err := i.Tidy(); err != nil {
			return err
		}
	}

	return nil
}

func (i *ModInstaller) installMods(mods []*modconfig.ModVersionConstraint, parent *modconfig.Mod) error {
	var errors []error
	for _, requiredModVersion := range mods {
		update := i.shouldUpdate
		modToUse, err := i.getCurrentlyInstalledVersion(requiredModVersion, parent, update)
		if err != nil {
			errors = append(errors, err)
			continue
		}
		// set the force updfate flag - this is passed down the depdnecy tree to ensure we update downstream dependecies
		shouldUpdate := modToUse == nil
		if err := i.installModDependencesRecursively(requiredModVersion, modToUse, parent, shouldUpdate); err != nil {
			errors = append(errors, err)
		}
	}

	i.installData.Lock = i.installData.NewLock
	return utils.CombineErrorsWithPrefix(fmt.Sprintf("%d %s failed to install", len(errors), utils.Pluralize("dependency", len(errors))), errors...)
}

func (i *ModInstaller) installModDependencesRecursively(requiredModVersion *modconfig.ModVersionConstraint, dependencyMod *modconfig.Mod, parent *modconfig.Mod, shouldUpdate bool) error {
	// get available versions for this mod
	availableVersions, err := i.installData.getAvailableModVersions(requiredModVersion.Name)
	if err != nil {
		return err
	}

	if dependencyMod == nil {
		// so we ARE installing

		// get a resolved mod ref that satisfies the version constraints
		resolvedRef, err := i.getModRefSatisfyingConstraints(requiredModVersion, availableVersions)
		if err != nil {
			return err
		}

		// install the mod
		dependencyMod, err = i.install(resolvedRef, parent)
		if err != nil {
			return err
		}
		if dependencyMod == nil {
			// this is unexpected but just ignore
			log.Printf("[TRACE] dependency %s does not define a mod definition - so there are no child dependencies to install", resolvedRef.Name)
			return nil
		}
	} else {
		// so we found an existing mod which will satisfy this requirement

		// update the install data
		i.installData.addExisting(requiredModVersion.Name, dependencyMod.Version, requiredModVersion.Constraint, parent)
		log.Printf("[TRACE] not installing %s with version constraint %s as version %s is already installed", requiredModVersion.Name, requiredModVersion.Constraint.Original, dependencyMod.Version)
	}

	// to get here we have the dependency mod - either we installed it or it was already installed
	// recursively install its dependencies
	var errors []error
	// now update the parent to dependency mod and install its child dependencies
	parent = dependencyMod
	for _, dep := range dependencyMod.Require.Mods {
		childDependencyMod, err := i.getCurrentlyInstalledVersion(dep, parent, shouldUpdate)
		if err != nil {
			errors = append(errors, err)
			continue
		}
		if err := i.installModDependencesRecursively(dep, childDependencyMod, parent, shouldUpdate); err != nil {
			errors = append(errors, err)
			continue
		}
	}

	return utils.CombineErrorsWithPrefix(fmt.Sprintf("%d child %s failed to install", len(errors), utils.Pluralize("dependency", len(errors))), errors...)
}

func (i *ModInstaller) getCurrentlyInstalledVersion(requiredModVersion *modconfig.ModVersionConstraint, parent *modconfig.Mod, forceUpdate bool) (*modconfig.Mod, error) {
	// do we have an installed version of this mod matching the required mod constraint
	installedVersion, err := i.installData.Lock.GetLockedModVersion(requiredModVersion, parent)
	if err != nil {
		return nil, err
	}
	if installedVersion == nil {
		return nil, nil
	}

	// can we update this
	canUpdate, err := i.canUpdateMod(installedVersion, requiredModVersion, forceUpdate)
	if err != nil {
		return nil, err

	}
	if canUpdate {
		// return nil mod to indicate we should update
		return nil, nil
	}

	// load the existing mod and return
	return i.loadDependencyMod(installedVersion)
}

// determine if we should update this mod, and if so whether there is an update available
func (i *ModInstaller) canUpdateMod(installedVersion *version_map.ResolvedVersionConstraint, requiredModVersion *modconfig.ModVersionConstraint, forceUpdate bool) (bool, error) {
	// so should we update?
	// if forceUpdate is set or if the required version constraint is different to the locked version constraint, update
	// TODO check * vs latest - maybe need a custom equals?
	if forceUpdate || installedVersion.Constraint != requiredModVersion.Constraint.Original {
		// get available versions for this mod
		availableVersions, err := i.installData.getAvailableModVersions(requiredModVersion.Name)
		if err != nil {
			return false, err
		}

		return i.updateAvailable(requiredModVersion, installedVersion.Version, availableVersions)
	}
	return false, nil

}

// determine whether there is a newer mod version avoilable which satisfies the dependency version constraint
func (i *ModInstaller) updateAvailable(requiredVersion *modconfig.ModVersionConstraint, currentVersion *semver.Version, availableVersions []*semver.Version) (bool, error) {
	latestVersion, err := i.getModRefSatisfyingConstraints(requiredVersion, availableVersions)
	if err != nil {
		return false, err
	}
	if latestVersion.Version.GreaterThan(currentVersion) {
		return true, nil
	}
	return false, nil
}

// get the most recent available mod version which satisfies the version constraint
func (i *ModInstaller) getModRefSatisfyingConstraints(modVersion *modconfig.ModVersionConstraint, availableVersions []*semver.Version) (*ResolvedModRef, error) {
	// find a version which satisfies the version constraint
	var version = getVersionSatisfyingConstraint(modVersion.Constraint, availableVersions)
	if version == nil {
		return nil, fmt.Errorf("no version of %s found satisfying verison constraint: %s", modVersion.Name, modVersion.Constraint.Original)
	}

	return NewResolvedModRef(modVersion, version)
}

// install a mod
func (i *ModInstaller) install(dependency *ResolvedModRef, parent *modconfig.Mod) (_ *modconfig.Mod, err error) {
	defer func() {
		if err == nil {
			i.installData.onModInstalled(dependency, parent)
		}
	}()

	if viper.GetBool(constants.ArgDryRun) {
		log.Printf("[TRACE] install %s - dry-run=truem, returning before install\n", dependency.FullName())
		return nil, nil
	}
	fullName := dependency.FullName()
	destPath := i.getDependencyDestPath(fullName)

	// if the target path exists, use the exiting file
	// if it does not exist (the usual case), install it
	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		if err := i.installFromGit(dependency, destPath); err != nil {
			return nil, err
		}
	}

	// now load the installed mod and return it
	return i.loadModfile(destPath, false)

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

func (i *ModInstaller) getDependencyDestPath(dependencyFullName string) string {
	return filepath.Join(i.modsPath, dependencyFullName)
}

func (i *ModInstaller) loadDependencyMod(modVersion *version_map.ResolvedVersionConstraint) (*modconfig.Mod, error) {
	modPath := i.getDependencyDestPath(modconfig.ModVersionFullName(modVersion.Name, modVersion.Version))
	return i.loadModfile(modPath, false)

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

//func (i *ModInstaller) setModArgs(opts *InstallOpts) error {
//	// list of mods to add
//	var addMods version_map.VersionConstraintMap
//	if i.updating {
//		if err := i.verifyCanUpdate(); err != nil {
//			return err
//		}
//	}
//	// add all mod args to the installer
//	// if any mods have mods versions which are NOT already in the lock file
//		// (i/.e/ if same mod name and constraint exist, skip)
//		for name, version := range opts.ModArgs {
//			if i.installData.Lock.ContainsModConstraint(name, version.Constraint) {
//				addMods[name] = version
//			}
//		}
//	} else {
//		// we are not updating, so any set the AddMods - exclude anything already in the mod requires
//		addMods = opts.ModArgs
//	}
//
//	// if we have any add mods, add them
//	i.setAddMods(addMods)
//
//	return nil
//}

//
//// check whether there is a mod version installed that satisfies the version constraint (and update requirements)
//func (i *ModInstaller) getInstalledVersionForConstraint(requiredModVersion *modconfig.ModVersionConstraint, availableVersions []*semver.Version, parent *modconfig.Mod) (*modconfig.Mod, error) {
//	// does this required version exist in in the lock file
//	log.Printf("[TRACE] getInstalledVersionForConstraint required version %v", requiredModVersion)
//	installedVersion, err := i.installData.Lock.EnsureLockedModVersion(requiredModVersion, parent)
//	if installedVersion == nil {
//		log.Printf("[TRACE] no version of %s installed for parent %s which satisfies version constrain %s", requiredModVersion.Name, parent.Name(), requiredModVersion.Constraint.Original)
//		return nil, nil
//	}
//
//	log.Printf("[TRACE] found installed version %s@%s", requiredModVersion.Name, installedVersion.Version)
//
//	// so there IS a version installed which satisfies the constraint.
//	// if we are updating, see if there is a newer verison
//	if i.shouldUpdate(requiredModVersion.Name) {
//		// so we should update if there is a newer version - check if there is
//		newerModVersionFound, err := i.updateAvailable(requiredModVersion, installedVersion.Version, availableVersions)
//		if err != nil {
//			return nil, err
//		}
//		if newerModVersionFound {
//			// there is a newer version so we will NOT use the installed version - return nil
//			return nil, nil
//		}
//	}
//
//	// so we resolved an installed version which will satisfy
//	// load the mod
//	modPath := filepath.Join(i.modsPath, modconfig.ModVersionFullName(requiredModVersion.Name, installedVersion.Version))
//	installedMod, err := i.loadModfile(modPath, false)
//	if err != nil {
//		return nil, err
//	}
//
//	return installedMod, nil
//}

//// check whether there is a mod version installed that satisfies the version constraint (and update requirements)
//func (i *ModInstaller) getInstalledVersionForConstraintOLD(requiredModVersion *modconfig.ModVersionConstraint, availableVersions []*semver.Version, parent *modconfig.Mod) (*modconfig.Mod, error) {
//	// does this required version exist in in the lock file
//	log.Printf("[TRACE] getInstalledVersionForConstraint required version %v", requiredModVersion)
//	installedVersion, err := i.installData.Lock.EnsureLockedModVersion(requiredModVersion, parent)
//	if installedVersion == nil {
//		log.Printf("[TRACE] no version of %s installed for parent %s which satisfies version constrain %s", requiredModVersion.Name, parent.Name(), requiredModVersion.Constraint.Original)
//		return nil, nil
//	}
//
//	log.Printf("[TRACE] found installed version %s@%s", requiredModVersion.Name, installedVersion.Version)
//
//	// so there IS a version installed which satisfies the constraint.
//	// if we are updating, see if there is a newer verison
//	if i.shouldUpdate(requiredModVersion.Name) {
//		// so we should update if there is a newer version - check if there is
//		newerModVersionFound, err := i.updateAvailable(requiredModVersion, installedVersion.Version, availableVersions)
//		if err != nil {
//			return nil, err
//		}
//		if newerModVersionFound {
//			// there is a newer version so we will NOT use the installed version - return nil
//			return nil, nil
//		}
//	}
//
//	// so we resolved an installed version which will satisfy
//	// load the mod
//	modPath := filepath.Join(i.modsPath, modconfig.ModVersionFullName(requiredModVersion.Name, installedVersion.Version))
//	installedMod, err := i.loadModfile(modPath, false)
//	if err != nil {
//		return nil, err
//	}
//
//	return installedMod, nil
//}

//func (i *ModInstaller) setAddMods(addMods version_map.VersionConstraintMap) {
//	for name, contraint := range addMods {
//		// does the workspace mod already have a dependency on this mod - if so DO NOT add to AddMods
//		if !i.workspaceMod.DependsOnMod(contraint) {
//			i.mods[name] = contraint
//		}
//	}
//}
