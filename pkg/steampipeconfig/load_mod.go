package steampipeconfig

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/Masterminds/semver/v3"
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
func LoadMod(modPath string, parseCtx *parse.ModParseContext, opts ...LoadModOption) (mod *modconfig.Mod, errorsAndWarnings *modconfig.ErrorAndWarnings) {
	defer func() {
		if r := recover(); r != nil {
			errorsAndWarnings = modconfig.NewErrorsAndWarning(helpers.ToError(r))
		}
	}()

	mod, loadModResult := loadModDefinition(modPath, parseCtx)
	if loadModResult.Error != nil {
		return nil, loadModResult
	}

	// apply opts to mod
	for _, o := range opts {
		o(mod)
	}

	// set the current mod on the run context
	parseCtx.SetCurrentMod(mod)

	// load the mod dependencies
	if err := loadModDependencies(mod, parseCtx); err != nil {
		return nil, modconfig.NewErrorsAndWarning(err)
	}

	// populate the resource maps of the current mod using the dependency mods
	mod.ResourceMaps = parseCtx.GetResourceMaps()
	// now load the mod resource hcl (
	mod, errorsAndWarnings = loadModResources(mod, parseCtx)

	// add in any warnings from mod load
	errorsAndWarnings.AddWarning(loadModResult.Warnings...)
	return mod, errorsAndWarnings
}

func loadModDefinition(modPath string, parseCtx *parse.ModParseContext) (mod *modconfig.Mod, errorsAndWarnings *modconfig.ErrorAndWarnings) {
	errorsAndWarnings = &modconfig.ErrorAndWarnings{}
	// verify the mod folder exists
	_, err := os.Stat(modPath)
	if os.IsNotExist(err) {
		return nil, modconfig.NewErrorsAndWarning(fmt.Errorf("mod folder %s does not exist", modPath))
	}

	if parse.ModfileExists(modPath) {
		// load the mod definition to get the dependencies
		var res *parse.DecodeResult
		mod, res = parse.ParseModDefinition(modPath, parseCtx.EvalCtx)
		errorsAndWarnings = modconfig.DiagsToErrorsAndWarnings("mod load failed", res.Diags)
		if res.Diags.HasErrors() {
			return nil, errorsAndWarnings
		}
	} else {
		// so there is no mod file - should we create a default?
		if !parseCtx.ShouldCreateDefaultMod() {
			errorsAndWarnings.Error = fmt.Errorf("mod folder %s does not contain a mod resource definition", modPath)
			// ShouldCreateDefaultMod flag NOT set - fail
			return nil, errorsAndWarnings
		}
		// just create a default mod
		mod = modconfig.CreateDefaultMod(modPath)

	}
	return mod, errorsAndWarnings
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
			// have we already loaded a mod which satisfied this
			loadedMod, err := parseCtx.GetLoadedDependencyMod(requiredModVersion, mod)
			if err != nil {
				return err
			}
			if loadedMod != nil {
				continue
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
	dependencyDir, version, err := findInstalledDependency(modDependency, parentFolder)
	if err != nil {
		return err
	}

	// we need to modify the ListOptions to ensure we include hidden files - these are excluded by default
	prevExclusions := parseCtx.ListOptions.Exclude
	parseCtx.ListOptions.Exclude = nil
	defer func() { parseCtx.ListOptions.Exclude = prevExclusions }()

	childParseCtx := parse.NewChildModParseContext(parseCtx, dependencyDir)
	// NOTE: pass in the version and dependency path of the mod - these must be set before it loads its dependencies
	mod, errAndWarnings := LoadMod(dependencyDir, childParseCtx, WithDependencyConfig(modDependency.Name, version))
	if errAndWarnings.GetError() != nil {
		return errAndWarnings.GetError()
	}

	// update loaded dependency mods
	parseCtx.AddLoadedDependencyMod(mod)
	if parseCtx.ParentParseCtx != nil {
		parseCtx.ParentParseCtx.AddLoadedDependencyMod(mod)
		// add mod resources to parent parse context
		parseCtx.ParentParseCtx.AddModResources(mod)
	}

	return nil

}

func loadModResources(mod *modconfig.Mod, parseCtx *parse.ModParseContext) (*modconfig.Mod, *modconfig.ErrorAndWarnings) {
	// if flag is set, create pseudo resources by mapping files
	var pseudoResources []modconfig.MappableResource
	var err error
	if parseCtx.CreatePseudoResources() {
		// now execute any pseudo-resource creations based on file mappings
		pseudoResources, err = createPseudoResources(mod, parseCtx)
		if err != nil {
			return nil, modconfig.NewErrorsAndWarning(err)
		}
	}

	// get the source files
	sourcePaths, err := getSourcePaths(mod.ModPath, parseCtx.ListOptions)
	if err != nil {
		log.Printf("[WARN] LoadMod: failed to get mod file paths: %v\n", err)
		return nil, modconfig.NewErrorsAndWarning(err)
	}

	// load the raw file data
	fileData, diags := parse.LoadFileData(sourcePaths...)
	if diags.HasErrors() {
		return nil, modconfig.NewErrorsAndWarning(plugin.DiagsToError("Failed to load all mod files", diags))
	}

	// parse all hcl files (NOTE - this reads the CurrentMod out of ParseContext and adds to it)
	mod, errAndWarnings := parse.ParseMod(fileData, pseudoResources, parseCtx)

	return mod, errAndWarnings
}

// search the parent folder for a mod installation which satisfied the given mod dependency
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
func LoadModResourceNames(mod *modconfig.Mod, parseCtx *parse.ModParseContext) (resources *modconfig.WorkspaceResources, err error) {
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
	if _, err := os.Stat(mod.ModPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("mod folder %s does not exist", mod.ModPath)
	}

	// now execute any pseudo-resource creations based on file mappings
	pseudoResources, err := createPseudoResources(mod, parseCtx)
	if err != nil {
		return nil, err
	}

	// add pseudo resources to result
	for _, r := range pseudoResources {
		if strings.HasPrefix(r.Name(), "query.") || strings.HasPrefix(r.Name(), "local.query.") {
			resources.Query[r.Name()] = true
		}
	}

	sourcePaths, err := getSourcePaths(mod.ModPath, parseCtx.ListOptions)
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

// GetModFileExtensions returns list of all file extensions we care about
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
func createPseudoResources(mod *modconfig.Mod, parseCtx *parse.ModParseContext) ([]modconfig.MappableResource, error) {
	// create list options to find pseudo resources
	listOpts := &filehelpers.ListOptions{
		Flags:   parseCtx.ListOptions.Flags,
		Include: filehelpers.InclusionsFromExtensions(modconfig.RegisteredFileExtensions()),
		Exclude: parseCtx.ListOptions.Exclude,
	}
	// list all registered files
	sourcePaths, err := getSourcePaths(mod.ModPath, listOpts)
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
		resource, fileData, err := factory(mod.ModPath, path, parseCtx.CurrentMod)
		if err != nil {
			errors = append(errors, err)
			continue
		}
		if resource != nil {
			metadata, err := getPseudoResourceMetadata(mod, resource.Name(), path, fileData)
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

func getPseudoResourceMetadata(mod *modconfig.Mod, resourceName string, path string, fileData []byte) (*modconfig.ResourceMetadata, error) {
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
	m.SetMod(mod)

	return m, nil
}
