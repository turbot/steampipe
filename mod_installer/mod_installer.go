package mod_installer

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/Masterminds/semver"
	git "github.com/go-git/go-git/v5"
	"github.com/spf13/viper"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/steampipeconfig/parse"
	"github.com/turbot/steampipe/steampipeconfig/version_map"
	"github.com/turbot/steampipe/utils"
)

type ModInstaller struct {
	workspaceMod  *modconfig.Mod
	modsPath      string
	workspacePath string

	installData *InstallData

	// should we update dependencies to newer versions if they exist
	updating bool
	// are dependencies being added to the workspace
	mods   version_map.VersionConstraintMap
	dryRun bool
}

func NewModInstaller(opts *InstallOpts) (*ModInstaller, error) {
	i := &ModInstaller{
		workspacePath: opts.WorkspacePath,
		updating:      opts.Updating,
		mods:          opts.ModArgs,
		dryRun:        opts.DryRun,
	}
	if err := i.setModsPath(); err != nil {
		return nil, err
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
	return i, nil
}

func (i *ModInstaller) setModsPath() error {
	// if this is a dry run, install mods to temp dir which will be deleted
	if i.dryRun {
		dir, err := os.MkdirTemp(os.TempDir(), "sp_dr_*")
		if err != nil {
			return err
		}
		i.modsPath = dir
	} else {
		// fall back to setting real mod path
		i.modsPath = constants.WorkspaceModPath(i.workspacePath)
	}
	return nil
}

func (i *ModInstaller) UninstallWorkspaceDependencies() error {
	workspaceMod := i.workspaceMod

	// remove required dependencies from the mod file
	if len(i.mods) == 0 {
		workspaceMod.RemoveAllModDependencies()

	} else {
		workspaceMod.RemoveModDependencies(i.mods)
	}

	// uninstall by calling Install
	if err := i.installMods(workspaceMod.Require.Mods, workspaceMod); err != nil {
		return err
	}

	if workspaceMod.Require.Empty() {
		workspaceMod.Require = nil
	}

	// if this is a dry run, return now
	if i.dryRun {

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
		if _, err := i.Prune(); err != nil {
			return err
		}
	}
	return nil
}

// InstallWorkspaceDependencies installs all dependencies of the workspace mod
func (i *ModInstaller) InstallWorkspaceDependencies() error {
	workspaceMod := i.workspaceMod
	defer func() {
		if i.dryRun {
			os.RemoveAll(i.modsPath)
		}
	}()

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
	if i.dryRun {
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
		if _, err := i.Prune(); err != nil {
			return err
		}
	}

	return nil
}

func (i *ModInstaller) GetModList() string {
	return i.installData.Lock.GetModList(i.workspaceMod.GetModDependencyPath())
}

func (i *ModInstaller) installMods(mods []*modconfig.ModVersionConstraint, parent *modconfig.Mod) error {
	var errors []error
	for _, requiredModVersion := range mods {
		update := i.updating
		modToUse, err := i.getCurrentlyInstalledVersionToUse(requiredModVersion, parent, update)
		if err != nil {
			errors = append(errors, err)
			continue
		}

		// if the mod is not installed or needs updating, pass shouldUpdate=true into installModDependencesRecursively
		// this ensures that we update any dependencies which have updates available
		shouldUpdate := modToUse == nil
		if err := i.installModDependencesRecursively(requiredModVersion, modToUse, parent, shouldUpdate); err != nil {
			errors = append(errors, err)
		}
	}

	// update the lock to be the new lock, and record any uninstalled mods
	i.installData.onInstallComplete()

	return i.buildInstallError(errors)
}

func (i *ModInstaller) buildInstallError(errors []error) error {
	if len(errors) == 0 {
		return nil
	}
	verb := "install"
	if i.updating {
		verb = "update"
	}
	prefix := fmt.Sprintf("%d %s failed to %s", len(errors), utils.Pluralize("dependency", len(errors)), verb)
	err := utils.CombineErrorsWithPrefix(prefix, errors...)
	return err
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
		childDependencyMod, err := i.getCurrentlyInstalledVersionToUse(dep, parent, shouldUpdate)
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

func (i *ModInstaller) getCurrentlyInstalledVersionToUse(requiredModVersion *modconfig.ModVersionConstraint, parent *modconfig.Mod, forceUpdate bool) (*modconfig.Mod, error) {
	// do we have an installed version of this mod matching the required mod constraint
	installedVersion, err := i.installData.Lock.GetLockedModVersion(requiredModVersion, parent)
	if err != nil {
		return nil, err
	}
	if installedVersion == nil {
		// if we are updating, the a version of th emod MUST be installed
		if i.updating {
			return nil, fmt.Errorf("%s is not installed", requiredModVersion.Name)
		}
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
	modFile, err := i.loadModfile(destPath, false)
	if err != nil {
		return nil, err
	}
	if modFile == nil {
		return nil, fmt.Errorf("'%s' has no mod definition file", dependency.FullName())
	}
	return modFile, nil

}

func (i *ModInstaller) installFromGit(dependency *ResolvedModRef, installPath string) error {
	// ensure mod directory exists - create if necessary
	if err := os.MkdirAll(i.modsPath, os.ModePerm); err != nil {
		return err
	}

	// get the mod from git
	gitUrl := getGitUrl(dependency.Name)
	_, err := git.PlainClone(installPath,
		false,
		&git.CloneOptions{
			URL:           gitUrl,
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
	mod, err := parse.ParseModDefinition(modPath)
	if err != nil {
		return nil, err
	}
	// if this is NOT the workspace mod, set ModDependencyPath - determine relative path from mod root
	if modPath != i.workspacePath {
		mod.ModDependencyPath, err = filepath.Rel(i.modsPath, modPath)
		if err != nil {
			return nil, err
		}
	}
	return mod, nil
}
