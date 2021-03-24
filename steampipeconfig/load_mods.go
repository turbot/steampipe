package steampipeconfig

import (
	"fmt"
	"log"

	"github.com/turbot/steampipe/steampipeconfig/modconfig"

	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe-plugin-sdk/plugin"
)

// parse all mod files in modPath and return an array of mods
func loadMod(modPath string) (mod *modconfig.Mod, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = helpers.ToError(r)
		}
	}()
	// get all the config files in the directory
	configPaths, err := getFilePaths(modPath, modExtension)
	if err != nil {
		log.Printf("[WARN] loadConfig: failed to get mod file paths: %v\n", err)
		return nil, err
	}
	if len(configPaths) == 0 {
		return nil, nil
	}
	fileData, diags := loadFileData(configPaths)
	if diags.HasErrors() {
		log.Printf("[WARN] loadConfig: failed to load all mod files: %v\n", err)
		return nil, plugin.DiagsToError("Failed to load all mod files", diags)
	}

	body, diags := parseHclFiles(fileData)
	if diags.HasErrors() {
		return nil, plugin.DiagsToError("Failed to load all config files", diags)
	}

	content, moreDiags := body.Content(modFileSchema)
	if moreDiags.HasErrors() {
		diags = append(diags, moreDiags...)
		return nil, plugin.DiagsToError("Failed to load config", diags)
	}

	for _, block := range content.Blocks {
		switch block.Type {
		case "variable":
			// TODO
		case "mod":
			// if there is more than one mod, fa\il
			if mod != nil {
				return nil, fmt.Errorf("more than 1 mod defintion found in %s", modPath)
			}

			mod, moreDiags = parseMod(block)
			if moreDiags.HasErrors() {
				diags = append(diags, moreDiags...)
			}
		}
	}

	if diags.HasErrors() {
		err = plugin.DiagsToError("Failed to load config", diags)
	}
	return
}
