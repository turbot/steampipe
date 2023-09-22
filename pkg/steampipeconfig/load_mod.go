package steampipeconfig

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/error_helpers"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/pkg/steampipeconfig/parse"
	"github.com/turbot/steampipe/pkg/steampipeconfig/versionmap"
)

// LoadMod parses all hcl files in modPath and returns a single mod
// if CreatePseudoResources flag is set, construct hcl resources for files with specific extensions
// NOTE: it is an error if there is more than 1 mod defined, however zero mods is acceptable
// - a default mod will be created assuming there are any resource files
func LoadMod(ctx context.Context, modPath string, parseCtx *parse.ModParseContext) (mod *modconfig.Mod, errorsAndWarnings *error_helpers.ErrorAndWarnings) {
	defer func() {
		if r := recover(); r != nil {
			errorsAndWarnings = error_helpers.NewErrorsAndWarning(helpers.ToError(r))
		}
	}()

	mod, loadModResult := loadModDefinition(ctx, modPath, parseCtx)
	if loadModResult.Error != nil {
		return nil, loadModResult
	}

	// if this is a dependency mod, initialise the dependency config
	if parseCtx.DependencyConfig != nil {
		parseCtx.DependencyConfig.SetModProperties(mod)
	}

	// set the current mod on the run context
	if err := parseCtx.SetCurrentMod(mod); err != nil {
		return nil, error_helpers.NewErrorsAndWarning(err)
	}

	// load the mod dependencies
	if err := loadModDependencies(ctx, mod, parseCtx); err != nil {
		return nil, error_helpers.NewErrorsAndWarning(err)
	}

	// populate the resource maps of the current mod using the dependency mods
	mod.ResourceMaps = parseCtx.GetResourceMaps()
	// now load the mod resource hcl (
	mod, errorsAndWarnings = loadModResources(ctx, mod, parseCtx)

	// add in any warnings from mod load
	errorsAndWarnings.AddWarning(loadModResult.Warnings...)
	return mod, errorsAndWarnings
}

func loadModDefinition(ctx context.Context, modPath string, parseCtx *parse.ModParseContext) (mod *modconfig.Mod, errorsAndWarnings *error_helpers.ErrorAndWarnings) {
	errorsAndWarnings = &error_helpers.ErrorAndWarnings{}
	// verify the mod folder exists
	_, err := os.Stat(modPath)
	if os.IsNotExist(err) {
		return nil, error_helpers.NewErrorsAndWarning(fmt.Errorf("mod folder %s does not exist", modPath))
	}

	if parse.ModfileExists(modPath) {
		// load the mod definition to get the dependencies
		var res *parse.DecodeResult
		mod, res = parse.ParseModDefinition(modPath, parseCtx.EvalCtx)
		errorsAndWarnings = error_helpers.DiagsToErrorsAndWarnings("mod load failed", res.Diags)
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

func loadModDependencies(ctx context.Context, parent *modconfig.Mod, parseCtx *parse.ModParseContext) error {
	var errors []error
	if parent.Require != nil {
		// now ensure there is a lock file - if we have any mod dependnecies there MUST be a lock file -
		// otherwise 'steampipe install' must be run
		if err := parseCtx.EnsureWorkspaceLock(parent); err != nil {
			return err
		}

		for _, requiredModVersion := range parent.Require.Mods {
			// get the locked version ofd this dependency
			lockedVersion, err := parseCtx.WorkspaceLock.GetLockedModVersion(requiredModVersion, parent)
			if err != nil {
				return err
			}
			if lockedVersion == nil {
				return fmt.Errorf("not all dependencies are installed - run 'steampipe mod install'")
			}
			if err := loadModDependency(ctx, lockedVersion, parseCtx); err != nil {
				errors = append(errors, err)
			}
		}
	}

	return error_helpers.CombineErrors(errors...)
}

func loadModDependency(ctx context.Context, modDependency *versionmap.ResolvedVersionConstraint, parseCtx *parse.ModParseContext) error {
	// dependency mods are installed to <mod path>/<mod nam>@version
	// for example workspace_folder/.steampipe/mods/github.com/turbot/steampipe-mod-aws-compliance@v1.0

	// we need to list all mod folder in the parent folder: workspace_folder/.steampipe/mods/github.com/turbot/
	// for each folder we parse the mod name and version and determine whether it meets the version constraint

	// search the parent folder for a mod installation which satisfied the given mod dependency
	dependencyDir, err := parseCtx.WorkspaceLock.FindInstalledDependency(modDependency)
	if err != nil {
		return err
	}

	// we need to modify the ListOptions to ensure we include hidden files - these are excluded by default
	prevExclusions := parseCtx.ListOptions.Exclude
	parseCtx.ListOptions.Exclude = nil
	defer func() { parseCtx.ListOptions.Exclude = prevExclusions }()

	childParseCtx := parse.NewChildModParseContext(parseCtx, modDependency, dependencyDir)
	// NOTE: pass in the version and dependency path of the mod - these must be set before it loads its dependencies
	dependencyMod, errAndWarnings := LoadMod(ctx, dependencyDir, childParseCtx)
	if errAndWarnings.GetError() != nil {
		return errAndWarnings.GetError()
	}

	// update loaded dependency mods
	parseCtx.AddLoadedDependencyMod(dependencyMod)
	// TODO IS THIS NEEDED????
	if parseCtx.ParentParseCtx != nil {
		// add mod resources to parent parse context
		parseCtx.ParentParseCtx.AddModResources(dependencyMod)
	}
	return nil

}

func loadModResources(ctx context.Context, mod *modconfig.Mod, parseCtx *parse.ModParseContext) (*modconfig.Mod, *error_helpers.ErrorAndWarnings) {
	// if flag is set, create pseudo resources by mapping files
	var pseudoResources []modconfig.MappableResource
	var err error
	if parseCtx.CreatePseudoResources() {
		// now execute any pseudo-resource creations based on file mappings
		pseudoResources, err = createPseudoResources(ctx, mod, parseCtx)
		if err != nil {
			return nil, error_helpers.NewErrorsAndWarning(err)
		}
	}

	// get the source files
	sourcePaths, err := getSourcePaths(ctx, mod.ModPath, parseCtx.ListOptions)
	if err != nil {
		log.Printf("[WARN] LoadMod: failed to get mod file paths: %v\n", err)
		return nil, error_helpers.NewErrorsAndWarning(err)
	}

	// load the raw file data
	fileData, diags := parse.LoadFileData(sourcePaths...)
	if diags.HasErrors() {
		return nil, error_helpers.NewErrorsAndWarning(plugin.DiagsToError("Failed to load all mod files", diags))
	}

	// parse all hcl files (NOTE - this reads the CurrentMod out of ParseContext and adds to it)
	mod, errAndWarnings := parse.ParseMod(ctx, fileData, pseudoResources, parseCtx)

	return mod, errAndWarnings
}

// LoadModResourceNames parses all hcl files in modPath and returns the names of all resources
func LoadModResourceNames(ctx context.Context, mod *modconfig.Mod, parseCtx *parse.ModParseContext) (resources *modconfig.WorkspaceResources, err error) {
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
	pseudoResources, err := createPseudoResources(ctx, mod, parseCtx)
	if err != nil {
		return nil, err
	}

	// add pseudo resources to result
	for _, r := range pseudoResources {
		if strings.HasPrefix(r.Name(), "query.") || strings.HasPrefix(r.Name(), "local.query.") {
			resources.Query[r.Name()] = true
		}
	}

	sourcePaths, err := getSourcePaths(ctx, mod.ModPath, parseCtx.ListOptions)
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
func getSourcePaths(ctx context.Context, modPath string, listOpts *filehelpers.ListOptions) ([]string, error) {
	sourcePaths, err := filehelpers.ListFilesWithContext(ctx, modPath, listOpts)
	if err != nil {
		return nil, err
	}
	return sourcePaths, nil
}

// create pseudo-resources for any files whose extensions are registered
func createPseudoResources(ctx context.Context, mod *modconfig.Mod, parseCtx *parse.ModParseContext) ([]modconfig.MappableResource, error) {
	// create list options to find pseudo resources
	listOpts := &filehelpers.ListOptions{
		Flags:   parseCtx.ListOptions.Flags,
		Include: filehelpers.InclusionsFromExtensions(modconfig.RegisteredFileExtensions()),
		Exclude: parseCtx.ListOptions.Exclude,
	}
	// list all registered files
	sourcePaths, err := getSourcePaths(ctx, mod.ModPath, listOpts)
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
