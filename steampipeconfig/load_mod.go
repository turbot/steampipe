package steampipeconfig

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/turbot/steampipe/utils"

	goVersion "github.com/hashicorp/go-version"

	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe-plugin-sdk/plugin"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/steampipeconfig/parse"
)

// LoadModDefinition parses the mod.sp file only
// TODO do we need to think about variables?
func LoadModDefinition(modPath string) (mod *modconfig.Mod, err error) {
	return parseMod(modPath,
		nil,
		&parse.ParseModOptions{
			ListOptions: &filehelpers.ListOptions{
				Flags:   filehelpers.FilesFlat,
				Include: []string{"**/mod.sp"},
			}})
}

// LoadMod parses all hcl files in modPath and returns a single mod
// if CreatePseudoResources flag is set, construct hcl resources for files with specific extensions
// NOTE: it is an error if there is more than 1 mod defined, however zero mods is acceptable
// - a default mod will be created assuming there are any resource files
func LoadMod(modPath string, opts *parse.ParseModOptions) (mod *modconfig.Mod, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = helpers.ToError(r)
		}
	}()

	if opts == nil {
		opts = parse.NewParseModOptions()
	}
	// verify the mod folder exists
	if _, err := os.Stat(modPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("mod folder %s does not exist", modPath)
	}

	mod, err = LoadModDefinition(modPath)
	if err != nil {
		return nil, err
	}
	// if we have not set the root mod, this must be the root mod
	if opts.RootMod == nil {
		opts.RootMod = mod
	}
	// first load the mod dependencies
	if err := loadModDependencies(mod, opts); err != nil {
		return nil, err
	}

	// if flag is set, create pseudo resources by mapping files
	var pseudoResources []modconfig.MappableResource
	if opts.CreatePseudoResources() {
		// now execute any pseudo-resource creations based on file mappings
		pseudoResources, err = createPseudoResources(modPath, opts)
		if err != nil {
			return nil, err
		}
	}

	// now parse the mod, passing the pseudo resources
	// load the raw data
	mod, err = parseMod(modPath, pseudoResources, opts)
	if err != nil {
		return nil, err
	}

	// update loaded mods
	// TODO this is wrong - need modPath??
	opts.LoadedMods[mod.Name()] = mod

	return
}

func loadModDependencies(mod *modconfig.Mod, opts *parse.ParseModOptions) error {

	var errors []error
	if mod.Requires != nil {
		for _, dependencyMod := range mod.Requires.Mods {
			// have we already loaded a mod which satisfied this
			if loadedMod, ok := opts.LoadedMods[dependencyMod.Name]; ok {
				if loadedMod.Version.GreaterThanOrEqual(dependencyMod.VersionConstraint) {
					continue
				}
			}
			if err := loadModDependency(dependencyMod, opts); err != nil {
				errors = append(errors, err)
			}
		}
	}
	return utils.CombineErrors(errors...)
}

func loadModDependency(modDependency *modconfig.ModVersion, opts *parse.ParseModOptions) error {
	// dependency mods are installed to <mod path>/<mod nam>@version
	// for example workspace_folder/.steampipe/mods/github.com/turbot/steampipe-mod-aws-compliance@v1.0

	// we need to list all mod folder in the parent folder: workspace_folder/.steampipe/mods/github.com/turbot/
	// for each folder we parse the mod name and version and determine whether it meets the version constraint

	// we need to iterate through all mods in the parent folder and find one that sarifies requirements
	parentFolder := filepath.Dir(filepath.Join(opts.ModInstallationPath, modDependency.Name))
	// get th elast segment of mod name

	dependencyPath, version, err := findInstalledDependency(modDependency, parentFolder)
	if err != nil {
		return err
	}

	// we need to modify the ListOptions to ensure we include hidden files - these are excluded by default
	prevExclusions := opts.ListOptions.Exclude
	opts.ListOptions.Exclude = nil
	defer func() { opts.ListOptions.Exclude = prevExclusions }()

	mod, err := LoadMod(dependencyPath, opts)
	if err != nil {
		return err
	}

	// set the version on the mod
	mod.Version = version

	return err

}

func findInstalledDependency(modDependency *modconfig.ModVersion, parentFolder string) (string, *goVersion.Version, error) {
	shortDepName := filepath.Base(modDependency.Name)
	entries, err := ioutil.ReadDir(parentFolder)
	if err != nil {
		return "", nil, fmt.Errorf("mod dependency %s is not installed", modDependency.Name)
	}

	for _, entry := range entries {
		split := strings.Split(entry.Name(), "@")
		if len(split) != 2 {
			// invalid format - ignore
			continue
		}
		modName := split[0]
		versionString := strings.TrimPrefix(split[1], "v")
		if modName == shortDepName {
			v, err := goVersion.NewVersion(versionString)
			if err != nil {
				// invalid format - ignore
				continue
			}
			if v.GreaterThanOrEqual(modDependency.VersionConstraint) {
				return filepath.Join(parentFolder, entry.Name()), v, nil
			}
		}
	}

	return "", nil, fmt.Errorf("mod dependency %s is not installed", modDependency.Name)

}

func parseMod(modPath string, pseudoResources []modconfig.MappableResource, opts *parse.ParseModOptions) (*modconfig.Mod, error) {
	// if inclusions are not already set, build list of all filepaths we need to parse/load
	// NOTE: pseudo resource creation is already handler - we just need to look for .sp files
	if len(opts.ListOptions.Include) == 0 {
		opts.ListOptions.Include = filehelpers.InclusionsFromExtensions([]string{constants.ModDataExtension})
	}
	sourcePaths, err := getSourcePaths(modPath, opts)
	if err != nil {
		log.Printf("[WARN] LoadMod: failed to get mod file paths: %v\n", err)
		return nil, err
	}

	fileData, diags := parse.LoadFileData(sourcePaths...)
	if diags.HasErrors() {
		return nil, plugin.DiagsToError("Failed to load all mod files", diags)
	}

	// parse all hcl files.
	mod, err := parse.ParseMod(modPath, fileData, pseudoResources, opts)
	if err != nil {
		return nil, err
	}

	return mod, err
}

// LoadModResourceNames parses all hcl files in modPath and returns the names of all resources
func LoadModResourceNames(modPath string, opts *parse.ParseModOptions) (resources *modconfig.WorkspaceResources, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = helpers.ToError(r)
		}
	}()

	resources = modconfig.NewWorkspaceResources()
	if opts == nil {
		opts = &parse.ParseModOptions{}
	}
	// verify the mod folder exists
	if _, err := os.Stat(modPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("mod folder %s does not exist", modPath)
	}

	// now execute any pseudo-resource creations based on file mappings
	pseudoResources, err := createPseudoResources(modPath, opts)
	if err != nil {
		return nil, err
	}

	// add pseudo resources to result
	for _, r := range pseudoResources {
		if strings.HasPrefix(r.Name(), "query.") {
			resources.Query[r.Name()] = true
		}
	}

	// build list of all filepaths we need to parse/load
	// NOTE: pseudo resource creation is handled separately below
	opts.ListOptions.Include = filehelpers.InclusionsFromExtensions([]string{constants.ModDataExtension})
	sourcePaths, err := getSourcePaths(modPath, opts)
	if err != nil {
		log.Printf("[WARN] LoadMod: failed to get mod file paths: %v\n", err)
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
func getSourcePaths(modPath string, opts *parse.ParseModOptions) ([]string, error) {
	sourcePaths, err := filehelpers.ListFiles(modPath, opts.ListOptions)
	if err != nil {
		return nil, err
	}
	return sourcePaths, nil
}

// create pseudo-resources for any files whose extensions are registered
func createPseudoResources(modPath string, opts *parse.ParseModOptions) ([]modconfig.MappableResource, error) {
	// avoid mutating the passed in options
	prevListIncludeOptions := opts.ListOptions.Include
	defer func() {
		opts.ListOptions.Include = prevListIncludeOptions
	}()

	// list all registered files
	opts.ListOptions.Include = filehelpers.InclusionsFromExtensions(modconfig.RegisteredFileExtensions())
	sourcePaths, err := getSourcePaths(modPath, opts)
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
		resource, fileData, err := factory(modPath, path)
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
