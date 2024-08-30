package parse

import (
	"fmt"
	"github.com/turbot/pipe-fittings/error_helpers"

	"io"
	"log"
	"os"
	"path/filepath"
	"sort"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/json"
	pconstants "github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe/pkg/filepaths"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"sigs.k8s.io/yaml"
)

// LoadFileData builds a map of filepath to file data
func LoadFileData(paths ...string) (map[string][]byte, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	var fileData = map[string][]byte{}

	for _, configPath := range paths {
		data, err := os.ReadFile(configPath)

		if err != nil {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagWarning,
				Summary:  fmt.Sprintf("failed to read config file %s", configPath),
				Detail:   err.Error()})
			continue
		}
		fileData[configPath] = data
	}
	return fileData, diags
}

// ParseHclFiles parses hcl file data and returns the hcl body object
func ParseHclFiles(fileData map[string][]byte) (hcl.Body, hcl.Diagnostics) {
	var parsedConfigFiles []*hcl.File
	var diags hcl.Diagnostics
	parser := hclparse.NewParser()

	// build ordered list of files so that we parse in a repeatable order
	filePaths := buildOrderedFileNameList(fileData)

	for _, filePath := range filePaths {
		var file *hcl.File
		var moreDiags hcl.Diagnostics
		ext := filepath.Ext(filePath)
		if ext == pconstants.JsonExtension {
			file, moreDiags = json.ParseFile(filePath)
		} else if pconstants.IsYamlExtension(ext) {
			file, moreDiags = parseYamlFile(filePath)
		} else {
			data := fileData[filePath]
			file, moreDiags = parser.ParseHCL(data, filePath)
		}

		if moreDiags.HasErrors() {
			diags = append(diags, moreDiags...)
			continue
		}
		parsedConfigFiles = append(parsedConfigFiles, file)
	}

	return hcl.MergeFiles(parsedConfigFiles), diags
}

func buildOrderedFileNameList(fileData map[string][]byte) []string {
	filePaths := make([]string, len(fileData))
	idx := 0
	for filePath := range fileData {
		filePaths[idx] = filePath
		idx++
	}
	sort.Strings(filePaths)
	return filePaths
}

// ModfileExists returns whether a mod file exists at the specified path and if so returns the filepath
func ModfileExists(modPath string) (string, bool) {
	for _, modFilePath := range filepaths.ModFilePaths(modPath) {
		if _, err := os.Stat(modFilePath); err == nil {
			return modFilePath, true
		}
	}
	return "", false
}

// parse a yaml file into a hcl.File object
func parseYamlFile(filename string) (*hcl.File, hcl.Diagnostics) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "Failed to open file",
				Detail:   fmt.Sprintf("The file %q could not be opened.", filename),
			},
		}
	}
	defer f.Close()

	src, err := io.ReadAll(f)
	if err != nil {
		return nil, hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "Failed to read file",
				Detail:   fmt.Sprintf("The file %q was opened, but an error occured while reading it.", filename),
			},
		}
	}
	jsonData, err := yaml.YAMLToJSON(src)
	if err != nil {
		return nil, hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "Failed to read convert YAML to JSON",
				Detail:   fmt.Sprintf("The file %q was opened, but an error occured while converting it to JSON.", filename),
			},
		}
	}
	return json.Parse(jsonData, filename)
}

func addPseudoResourcesToMod(pseudoResources []modconfig.MappableResource, hclResources map[string]bool, mod *modconfig.Mod) error_helpers.ErrorAndWarnings {
	res := error_helpers.EmptyErrorsAndWarning()
	for _, r := range pseudoResources {
		// is there a hcl resource with the same name as this pseudo resource - it takes precedence
		name := r.GetUnqualifiedName()
		if _, ok := hclResources[name]; ok {
			res.AddWarning(fmt.Sprintf("%s ignored as hcl resources of same name is already defined", r.GetDeclRange().Filename))
			log.Printf("[TRACE] %s ignored as hcl resources of same name is already defined", r.GetDeclRange().Filename)
			continue
		}
		// add pseudo resource to mod
		mod.AddResource(r.(modconfig.HclResource))
		// add to map of existing resources
		hclResources[name] = true
	}
	return res
}

// get names of all resources defined in hcl which may also be created as pseudo resources
// if we find a mod block, build a shell mod
func loadMappableResourceNames(content *hcl.BodyContent) (map[string]bool, error) {
	hclResources := make(map[string]bool)

	for _, block := range content.Blocks {
		// if this is a mod, build a shell mod struct (with just the name populated)
		switch block.Type {
		case modconfig.BlockTypeQuery:
			// for any mappable resource, store the resource name
			name := modconfig.BuildModResourceName(block.Type, block.Labels[0])
			hclResources[name] = true
		}
	}
	return hclResources, nil
}

// ParseModResourceNames parses all source hcl files for the mod path and associated resources,
// and returns the resource names
func ParseModResourceNames(fileData map[string][]byte) (*modconfig.WorkspaceResources, error) {
	var resources = modconfig.NewWorkspaceResources()
	body, diags := ParseHclFiles(fileData)
	if diags.HasErrors() {
		return nil, plugin.DiagsToError("Failed to load all mod source files", diags)
	}

	content, moreDiags := body.Content(WorkspaceBlockSchema)
	if moreDiags.HasErrors() {
		diags = append(diags, moreDiags...)
		return nil, plugin.DiagsToError("Failed to load mod", diags)
	}

	for _, block := range content.Blocks {
		// if this is a mod, build a shell mod struct (with just the name populated)
		switch block.Type {

		case modconfig.BlockTypeQuery:
			// for any mappable resource, store the resource name
			name := modconfig.BuildModResourceName(block.Type, block.Labels[0])
			resources.Query[name] = true
		case modconfig.BlockTypeControl:
			// for any mappable resource, store the resource name
			name := modconfig.BuildModResourceName(block.Type, block.Labels[0])
			resources.Control[name] = true
		case modconfig.BlockTypeBenchmark:
			// for any mappable resource, store the resource name
			name := modconfig.BuildModResourceName(block.Type, block.Labels[0])
			resources.Benchmark[name] = true
		}
	}
	return resources, nil
}
