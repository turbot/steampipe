package modinstaller

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/Masterminds/semver"
	git "github.com/go-git/go-git/v5"
	"github.com/otiai10/copy"
	"github.com/spf13/viper"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/error_helpers"
	"github.com/turbot/steampipe/pkg/filepaths"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/pkg/steampipeconfig/parse"
	"github.com/turbot/steampipe/pkg/steampipeconfig/versionmap"
	"github.com/turbot/steampipe/pkg/utils"
	"github.com/turbot/steampipe/sperr"
)

type ModInstaller struct {
	installData *InstallData

	workspaceMod *modconfig.Mod
	mods         versionmap.VersionConstraintMap

	// the final resting place of all dependency mods
	modsPath string
	// a shadow directory for installing mods
	// this is necessary to make mod installation transactional
	shadowDirPath string

	workspacePath string

	// what command is being run
	command string
	// are dependencies being added to the workspace
	dryRun bool
}

func NewModInstaller(opts *InstallOpts) (*ModInstaller, error) {
	i := &ModInstaller{
		workspacePath: opts.WorkspacePath,
		command:       opts.Command,
		dryRun:        opts.DryRun,
	}
	if err := i.setModsPath(); err != nil {
		return nil, err
	}

	// load workspace mod, creating a default if needed
	workspaceMod, err := i.loadModfile(i.workspacePath, true)
	if err != nil {
		return nil, err
	}
	i.workspaceMod = workspaceMod

	// load lock file
	workspaceLock, err := versionmap.LoadWorkspaceLock(i.workspacePath)
	if err != nil {
		return nil, err
	}

	// create install data
	i.installData = NewInstallData(workspaceLock, workspaceMod)

	// parse args to get the required mod versions
	requiredMods, err := i.GetRequiredModVersionsFromArgs(opts.ModArgs)
	if err != nil {
		return nil, err
	}
	i.mods = requiredMods

	return i, nil
}

func (i *ModInstaller) removeOldShadowDirectories() error {
	removeErrors := []error{}
	// get the parent of the 'mods' directory - all shadow directories are siblings of this
	parent := filepath.Base(i.modsPath)
	entries, err := os.ReadDir(parent)
	if err != nil {
		return err
	}
	for _, dir := range entries {
		if dir.IsDir() && filepaths.IsModInstallShadowPath(dir.Name()) {
			err := os.RemoveAll(filepath.Join(parent, dir.Name()))
			if err != nil {
				removeErrors = append(removeErrors, err)
			}
		}
	}
	return error_helpers.CombineErrors(removeErrors...)
}

func (i *ModInstaller) setModsPath() error {
	i.modsPath = filepaths.WorkspaceModPath(i.workspacePath)
	if err := i.removeOldShadowDirectories(); err != nil {
		log.Println("[INFO] could not remove old mod installation shadow directory", err)
	}
	i.shadowDirPath = filepaths.WorkspaceModShadowPath(i.workspacePath)
	return nil
}

func (i *ModInstaller) UninstallWorkspaceDependencies() error {
	workspaceMod := i.workspaceMod

	// remove required dependencies from the mod file
	if len(i.mods) == 0 {
		workspaceMod.RemoveAllModDependencies()

	} else {
		// verify all the mods specifed in the args exist in the modfile
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
func (i *ModInstaller) InstallWorkspaceDependencies() (err error) {
	workspaceMod := i.workspaceMod
	defer func() {
		// tidy unused mods
		// (put in defer so it still gets called in case of errors)
		if viper.GetBool(constants.ArgPrune) {
			// be sure not to overwrite an existing return error
			_, pruneErr := i.Prune()
			if pruneErr != nil && err == nil {
				err = pruneErr
			}
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

	if !workspaceMod.HasDependentMods() {
		// there are no dependencies - delete the cache
		i.installData.Lock.Delete()
	}
	return nil
}

func (i *ModInstaller) GetModList() string {
	return i.installData.Lock.GetModList(i.workspaceMod.GetInstallCacheKey())
}

// commitShadow recursively copies over the contents of the shadow directory
// to the mods directory, replacing conflicts as it goes
// (uses `os.Create(dest)` under the hood - which truncates the target)
func (i *ModInstaller) commitShadow() (err error) {
	if _, err := os.Stat(i.shadowDirPath); os.IsNotExist(err) {
		// nothing to do here
		// there's no shadow directory to commit
		// this is not an error and may happen when install does not make any changes
		return nil
	}
	entries, err := os.ReadDir(i.shadowDirPath)
	if err != nil {
		return sperr.WrapWithRootMessage(err, "could not read shadow directory")
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		source := filepath.Join(i.shadowDirPath, entry.Name())
		destination := filepath.Join(i.modsPath, entry.Name())
		log.Println("[TRACE] copying", source, destination)
		if err := copy.Copy(source, destination); err != nil {
			return sperr.WrapWithRootMessage(err, "could not commit shadow directory '%s'", entry.Name())
		}
	}
	return nil
}

func (i *ModInstaller) installMods(mods []*modconfig.ModVersionConstraint, parent *modconfig.Mod) (err error) {
	defer func() {
		if err == nil {
			// everything went well
			// copy whatever we installed to the mods directory
			err = i.commitShadow()
		}

		// force remove the shadow directory
		os.RemoveAll(i.shadowDirPath)
	}()

	var errors []error
	for _, requiredModVersion := range mods {
		modToUse, err := i.getCurrentlyInstalledVersionToUse(requiredModVersion, parent, i.updating())
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
	if i.updating() {
		verb = "update"
	}
	prefix := fmt.Sprintf("%d %s failed to %s", len(errors), utils.Pluralize("dependency", len(errors)), verb)
	err := error_helpers.CombineErrorsWithPrefix(prefix, errors...)
	return err
}

func (i *ModInstaller) installModDependencesRecursively(requiredModVersion *modconfig.ModVersionConstraint, dependencyMod *modconfig.Mod, parent *modconfig.Mod, shouldUpdate bool) error {
	// get available versions for this mod
	includePrerelease := requiredModVersion.Constraint.IsPrerelease()
	availableVersions, err := i.installData.getAvailableModVersions(requiredModVersion.Name, includePrerelease)

	if err != nil {
		return err
	}

	if dependencyMod == nil {
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
		if err = dependencyMod.ValidateSteampipeVersion(); err != nil {
			return err
		}
	} else {
		// update the install data
		i.installData.addExisting(requiredModVersion.Name, dependencyMod, requiredModVersion.Constraint, parent)
		log.Printf("[TRACE] not installing %s with version constraint %s as version %s is already installed", requiredModVersion.Name, requiredModVersion.Constraint.Original, dependencyMod.Version)
	}
	// to get here we have the dependency mod - either we installed it or it was already installed
	// recursively install its dependencies
	var errors []error

	for _, childDependency := range dependencyMod.Require.Mods {
		childDependencyMod, err := i.getCurrentlyInstalledVersionToUse(childDependency, dependencyMod, shouldUpdate)
		if err != nil {
			errors = append(errors, err)
			continue
		}
		if err := i.installModDependencesRecursively(childDependency, childDependencyMod, dependencyMod, shouldUpdate); err != nil {
			errors = append(errors, err)
			continue
		}
	}

	return error_helpers.CombineErrorsWithPrefix(fmt.Sprintf("%d child %s failed to install", len(errors), utils.Pluralize("dependency", len(errors))), errors...)
}

func (i *ModInstaller) getCurrentlyInstalledVersionToUse(requiredModVersion *modconfig.ModVersionConstraint, parent *modconfig.Mod, forceUpdate bool) (*modconfig.Mod, error) {
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

// loadDependencyMod tries to load the mod definition from the shadow directory
// and falls back to the 'mods' directory of the root mod
func (i *ModInstaller) loadDependencyMod(modVersion *versionmap.ResolvedVersionConstraint) (*modconfig.Mod, error) {
	// construct the dependency path - this is the relative path of the dependency we are installing
	dependencyPath := modVersion.DependencyPath()

	// first try loading from the shadow dir
	modDefinition, err := i.loadDependencyModFromRoot(i.shadowDirPath, dependencyPath)
	if err != nil {
		return nil, err
	}

	// failed to load from shadow dir, try mods dir
	if modDefinition == nil {
		modDefinition, err = i.loadDependencyModFromRoot(i.modsPath, dependencyPath)
		if err != nil {
			return nil, err
		}
	}

	// if we still failed, give up
	if modDefinition == nil {
		return nil, fmt.Errorf("could not find dependency mod '%s'", dependencyPath)
	}

	// set the DependencyName, DependencyPath and Version properties on the mod
	if err := i.setModDependencyConfig(modDefinition, dependencyPath); err != nil {
		return nil, err
	}

	return modDefinition, nil
}

func (i *ModInstaller) loadDependencyModFromRoot(modInstallRoot string, dependencyPath string) (*modconfig.Mod, error) {
	log.Printf("[TRACE] loadDependencyModFromRoot: trying to load %s from root %s", dependencyPath, modInstallRoot)

	modPath := path.Join(modInstallRoot, dependencyPath)
	modDefinition, err := i.loadModfile(modPath, false)
	if err != nil {
		return nil, sperr.WrapWithMessage(err, "failed to load mod definition for %s from %s", dependencyPath, modInstallRoot)
	}
	return modDefinition, nil
}

// determine if we should update this mod, and if so whether there is an update available
func (i *ModInstaller) canUpdateMod(installedVersion *versionmap.ResolvedVersionConstraint, requiredModVersion *modconfig.ModVersionConstraint, forceUpdate bool) (bool, error) {
	// so should we update?
	// if forceUpdate is set or if the required version constraint is different to the locked version constraint, update
	// TODO check * vs latest - maybe need a custom equals?
	if forceUpdate || installedVersion.Constraint != requiredModVersion.Constraint.Original {
		// get available versions for this mod
		includePrerelease := requiredModVersion.Constraint.IsPrerelease()
		availableVersions, err := i.installData.getAvailableModVersions(requiredModVersion.Name, includePrerelease)
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
		return nil, fmt.Errorf("no version of %s found satisfying version constraint: %s", modVersion.Name, modVersion.Constraint.Original)
	}

	return NewResolvedModRef(modVersion, version)
}

// install a mod
func (i *ModInstaller) install(dependency *ResolvedModRef, parent *modconfig.Mod) (_ *modconfig.Mod, err error) {
	var modDef *modconfig.Mod
	// get the temp location to install the mod to
	dependencyPath := dependency.DependencyPath()
	destPath := i.getDependencyShadowPath(dependencyPath)

	defer func() {
		if err == nil {
			i.installData.onModInstalled(dependency, modDef, parent)
		}
	}()
	// if the target path exists, use the exiting file
	// if it does not exist (the usual case), install it
	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		log.Println("[TRACE] installing", dependencyPath, "in", destPath)
		if err := i.installFromGit(dependency, destPath); err != nil {
			return nil, err
		}
	}

	// now load the installed mod and return it
	modDef, err = i.loadModfile(destPath, false)
	if err != nil {
		return nil, err
	}
	if modDef == nil {
		return nil, fmt.Errorf("'%s' has no mod definition file", dependencyPath)
	}

	if !i.dryRun {
		// now the mod is installed in its final location, set mod dependency path
		if err := i.setModDependencyConfig(modDef, dependencyPath); err != nil {
			return nil, err
		}
	}

	return modDef, nil
}

func (i *ModInstaller) installFromGit(dependency *ResolvedModRef, installPath string) error {
	// get the mod from git
	gitUrl := getGitUrl(dependency.Name)
	log.Println("[TRACE] >>> cloning", gitUrl, dependency.GitReference)
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

// build the path of the temp location to copy this depednency to
func (i *ModInstaller) getDependencyDestPath(dependencyFullName string) string {
	return filepath.Join(i.modsPath, dependencyFullName)
}

// build the path of the temp location to copy this depednency to
func (i *ModInstaller) getDependencyShadowPath(dependencyFullName string) string {
	return filepath.Join(i.shadowDirPath, dependencyFullName)
}

// set the mod dependency path
func (i *ModInstaller) setModDependencyConfig(mod *modconfig.Mod, dependencyPath string) error {
	return mod.SetDependencyConfig(dependencyPath)
}

func (i *ModInstaller) loadModfile(modPath string, createDefault bool) (*modconfig.Mod, error) {
	if !parse.ModfileExists(modPath) {
		if createDefault {
			mod := modconfig.CreateDefaultMod(i.workspacePath)
			return mod, nil
		}
		return nil, nil
	}

	mod, err := parse.ParseModDefinition(modPath)
	if err != nil {
		return nil, err
	}

	return mod, nil
}

func (i *ModInstaller) updating() bool {
	return i.command == "update"
}

func (i *ModInstaller) uninstalling() bool {
	return i.command == "uninstall"
}
