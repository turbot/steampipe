package steampipeconfig

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/Masterminds/semver"
	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/error_helpers"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/pkg/steampipeconfig/parse"
)

// LoadMod parses all hcl files in modPath and returns a single mod
// if CreatePseudoResources flag is set, construct hcl resources for files with specific extensions
// NOTE: it is an error if there is more than 1 mod defined, however zero mods is acceptable
// - a default mod will be created assuming there are any resource files
func LoadMod(modPath string, parseCtx *parse.ModParseContext) (mod *modconfig.Mod, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = helpers.ToError(r)
		}
	}()

	mod, err = loadModDefinition(modPath, parseCtx)
	if err != nil {
		return nil, err
	}
	// load the mod dependencies
	if err := loadModDependencies(mod, parseCtx); err != nil {
		return nil, err
	}
	// now we have loaded dependencies, set the current mod on the run context
	parseCtx.CurrentMod = mod
	// populate the resource maps of the current mod using the dependency mods
	mod.ResourceMaps = parseCtx.GetResourceMaps()
	// now load the mod resource hcl
	return loadModResources(modPath, parseCtx, mod)
}

func loadModDefinition(modPath string, parseCtx *parse.ModParseContext) (*modconfig.Mod, error) {
	var mod *modconfig.Mod
	// verify the mod folder exists
	_, err := os.Stat(modPath)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("mod folder %s does not exist", modPath)
	}

	if parse.ModfileExists(modPath) {
		// load the mod definition to get the dependencies
		mod, err = parse.ParseModDefinition(modPath)
		if err != nil {
			return nil, err
		}
		// now we have loaded the mod, if this is a dependency mod, add in any variables we have loaded
		if parseCtx.ParentParseCtx != nil {
			parseCtx.Variables = parseCtx.ParentParseCtx.DependencyVariables[mod.ShortName]
			parseCtx.SetVariablesForDependencyMod(mod, parseCtx.ParentParseCtx.DependencyVariables)
		}

	} else {
		// so there is no mod file - should we create a default?
		if !parseCtx.ShouldCreateDefaultMod() {
			// ShouldCreateDefaultMod flag NOT set - fail
			return nil, fmt.Errorf("mod folder %s does not contain a mod resource definition", modPath)
		}
		// just create a default mod
		mod = modconfig.CreateDefaultMod(modPath)

	}
	return mod, nil
}

func loadModDependencies(mod *modconfig.Mod, parseCtx *parse.ModParseContext) error {
	var errors []error

	if mod.Require != nil {
		// now ensure there is a lock file - if we have any mod dependnecies there MUST be a lock file -
		// otherwise 'steampipe install' must be run
		if err := parseCtx.EnsureWorkspaceLock(mod); err != nil {
			return err
		}

		for _, requiredModVersion := range mod.Require.Mods {
			// if we have a locked version, update the required version to reflect this
			lockedVersion, err := parseCtx.WorkspaceLock.GetLockedModVersionConstraint(requiredModVersion, mod)
			if err != nil {
				errors = append(errors, err)
				continue
			}
			if lockedVersion != nil {
				requiredModVersion = lockedVersion
			}

			// have we already loaded a mod which satisfied this
			if loadedMod, ok := parseCtx.LoadedDependencyMods[requiredModVersion.Name]; ok {
				if requiredModVersion.Constraint.Check(loadedMod.Version) {
					continue
				}
			}
			if err := loadModDependency(requiredModVersion, parseCtx); err != nil {
				errors = append(errors, err)
			}
		}
	}

	return error_helpers.CombineErrors(errors...)
}

func loadModDependency(modDependency *modconfig.ModVersionConstraint, parseCtx *parse.ModParseContext) error {
	// dependency mods are installed to <mod path>/<mod nam>@version
	// for example workspace_folder/.steampipe/mods/github.com/turbot/steampipe-mod-aws-compliance@v1.0

	// we need to list all mod folder in the parent folder: workspace_folder/.steampipe/mods/github.com/turbot/
	// for each folder we parse the mod name and version and determine whether it meets the version constraint

	// we need to iterate through all mods in the parent folder and find one that satisfies requirements
	parentFolder := filepath.Dir(filepath.Join(parseCtx.WorkspaceLock.ModInstallationPath, modDependency.Name))

	// search the parent folder for a mod installation which satisfied the given mod dependency
	dependencyPath, version, err := findInstalledDependency(modDependency, parentFolder)
	if err != nil {
		return err
	}

	// we need to modify the ListOptions to ensure we include hidden files - these are excluded by default
	prevExclusions := parseCtx.ListOptions.Exclude
	parseCtx.ListOptions.Exclude = nil
	defer func() { parseCtx.ListOptions.Exclude = prevExclusions }()

	// create a child run context
	childRunCtx := parse.NewModParseContext(
		parseCtx.WorkspaceLock,
		dependencyPath,
		parse.CreatePseudoResources,
		&filehelpers.ListOptions{
			// listFlag specifies whether to load files recursively
			Flags: filehelpers.FilesRecursive,
			// only load .sp files
			Include: filehelpers.InclusionsFromExtensions([]string{constants.ModDataExtension}),
		})
	childRunCtx.BlockTypes = parseCtx.BlockTypes
	childRunCtx.ParentParseCtx = parseCtx

	mod, err := LoadMod(dependencyPath, childRunCtx)
	if err != nil {
		return err
	}

	// set the version and dependency path of the mod
	mod.Version = version
	mod.ModDependencyPath = modDependency.Name

	// update loaded dependency mods
	parseCtx.LoadedDependencyMods[modDependency.Name] = mod
	if parseCtx.ParentParseCtx != nil {
		parseCtx.ParentParseCtx.LoadedDependencyMods[modDependency.Name] = mod
	}

	return err

}

func loadModResources(modPath string, parseCtx *parse.ModParseContext, mod *modconfig.Mod) (*modconfig.Mod, error) {
	// if flag is set, create pseudo resources by mapping files
	var pseudoResources []modconfig.MappableResource
	var err error
	if parseCtx.CreatePseudoResources() {
		// now execute any pseudo-resource creations based on file mappings
		pseudoResources, err = createPseudoResources(modPath, parseCtx)
		if err != nil {
			return nil, err
		}
	}

	// get the source files
	sourcePaths, err := getSourcePaths(modPath, parseCtx.ListOptions)
	if err != nil {
		log.Printf("[WARN] LoadMod: failed to get mod file paths: %v\n", err)
		return nil, err
	}

	// load the raw file data
	fileData, diags := parse.LoadFileData(sourcePaths...)
	if diags.HasErrors() {
		return nil, plugin.DiagsToError("Failed to load all mod files", diags)
	}

	// parse all hcl files.
	mod, err = parse.ParseMod(modPath, fileData, pseudoResources, parseCtx)
	if err != nil {
		return nil, err
	}

	// now add fully populated mod to the parent run context
	if parseCtx.ParentParseCtx != nil {
		parseCtx.ParentParseCtx.CurrentMod = mod
		parseCtx.ParentParseCtx.AddMod(mod)
	}

	return mod, err
}

// search the parent folder for a mod installatio which satisfied the given mod dependency
func findInstalledDependency(modDependency *modconfig.ModVersionConstraint, parentFolder string) (string, *semver.Version, error) {
	shortDepName := filepath.Base(modDependency.Name)
	entries, err := os.ReadDir(parentFolder)
	if err != nil {
		return "", nil, fmt.Errorf("mod satisfying '%s' is not installed", modDependency)
	}

	// results vars
	var dependencyPath string
	var dependencyVersion *semver.Version

	for _, entry := range entries {
		split := strings.Split(entry.Name(), "@")
		if len(split) != 2 {
			// invalid format - ignore
			continue
		}
		modName := split[0]
		versionString := strings.TrimPrefix(split[1], "v")
		if modName == shortDepName {
			v, err := semver.NewVersion(versionString)
			if err != nil {
				// invalid format - ignore
				continue
			}
			if modDependency.Constraint.Check(v) {
				// if there is more than 1 mod which satisfied the dependency, fail (for now)
				if dependencyVersion != nil {
					return "", nil, fmt.Errorf("more than one mod found which satisfies dependency %s@%s", modDependency.Name, modDependency.VersionString)
				}
				dependencyPath = filepath.Join(parentFolder, entry.Name())
				dependencyVersion = v
			}
		}
	}

	// did we find a result?
	if dependencyVersion != nil {
		return dependencyPath, dependencyVersion, nil
	}

	return "", nil, fmt.Errorf("mod satisfying '%s' is not installed", modDependency)
}

// LoadModResourceNames parses all hcl files in modPath and returns the names of all resources
func LoadModResourceNames(modPath string, parseCtx *parse.ModParseContext) (resources *modconfig.WorkspaceResources, err error) {
	// TODO support dependencies
	defer func() {
		if r := recover(); r != nil {
			err = helpers.ToError(r)
		}
	}()

	resources = modconfig.NewWorkspaceResources()
	if parseCtx == nil {
		parseCtx = &parse.ModParseContext{}
	}
	// verify the mod folder exists
	if _, err := os.Stat(modPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("mod folder %s does not exist", modPath)
	}

	// now execute any pseudo-resource creations based on file mappings
	pseudoResources, err := createPseudoResources(modPath, parseCtx)
	if err != nil {
		return nil, err
	}

	// add pseudo resources to result
	for _, r := range pseudoResources {
		if strings.HasPrefix(r.Name(), "query.") || strings.HasPrefix(r.Name(), "local.query.") {
			resources.Query[r.Name()] = true
		}
	}

	sourcePaths, err := getSourcePaths(modPath, parseCtx.ListOptions)
	if err != nil {
		log.Printf("[WARN] LoadModResourceNames: failed to get mod file paths: %v\n", err)
		return nil, err
	}

	fileData, diags := parse.LoadFileData(sourcePaths...)
	if diags.HasErrors() {
		return nil, plugin.DiagsToError("Failed to load all mod files", diags)
	}

	parsedResourceNames, err := parse.ParseModResourceNames(fileData)
	if err != nil {
		return nil, err
	}
	return resources.Merge(parsedResourceNames), nil
}

// GetModFileExtensions :: return list of all file extensions we care about
// this will be the mod data extension, plus any registered extensions registered in fileToResourceMap
func GetModFileExtensions() []string {
	return append(modconfig.RegisteredFileExtensions(), constants.ModDataExtension, constants.VariablesExtension)
}

// build list of all filepaths we need to parse/load the mod
// this will include hcl files (with .sp extension)
// as well as any other files with extensions that have been registered for pseudo resource creation
// (see steampipeconfig/modconfig/resource_type_map.go)
func getSourcePaths(modPath string, listOpts *filehelpers.ListOptions) ([]string, error) {
	sourcePaths, err := filehelpers.ListFiles(modPath, listOpts)
	if err != nil {
		return nil, err
	}
	return sourcePaths, nil
}

// create pseudo-resources for any files whose extensions are registered
func createPseudoResources(modPath string, parseCtx *parse.ModParseContext) ([]modconfig.MappableResource, error) {
	// create list options to find pseudo resources
	listOpts := &filehelpers.ListOptions{
		Flags:   parseCtx.ListOptions.Flags,
		Include: filehelpers.InclusionsFromExtensions(modconfig.RegisteredFileExtensions()),
		Exclude: parseCtx.ListOptions.Exclude,
	}
	// list all registered files
	sourcePaths, err := getSourcePaths(modPath, listOpts)
	if err != nil {
		return nil, err
	}

	var errors []error
	var res []modconfig.MappableResource

	// for every source path:
	// - if it is NOT a registered type, skip
	// [- if an existing resource has already referred directly to this file, skip] *not yet*
	for _, path := range sourcePaths {
		factory, ok := modconfig.ResourceTypeMap[filepath.Ext(path)]
		if !ok {
			continue
		}
		resource, fileData, err := factory(modPath, path, parseCtx.CurrentMod)
		if err != nil {
			errors = append(errors, err)
			continue
		}
		if resource != nil {
			metadata, err := getPseudoResourceMetadata(resource.Name(), path, fileData)
			if err != nil {
				return nil, err
			}
			resource.SetMetadata(metadata)
			res = append(res, resource)
		}
	}

	// show errors as trace logging
	if len(errors) > 0 {
		for _, err := range errors {
			log.Printf("[TRACE] failed to convert local file into resource: %v", err)
		}
	}

	return res, nil
}

func getPseudoResourceMetadata(resourceName string, path string, fileData []byte) (*modconfig.ResourceMetadata, error) {
	sourceDefinition := string(fileData)
	split := strings.Split(sourceDefinition, "\n")
	lineCount := len(split)

	// convert the name into a short name
	parsedName, err := modconfig.ParseResourceName(resourceName)
	if err != nil {
		return nil, err
	}

	m := &modconfig.ResourceMetadata{
		ResourceName:     parsedName.Name,
		FileName:         path,
		StartLineNumber:  1,
		EndLineNumber:    lineCount,
		IsAutoGenerated:  true,
		SourceDefinition: sourceDefinition,
	}

	return m, nil
}
