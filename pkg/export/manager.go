package export

import (
	"context"
	"fmt"
	"github.com/turbot/steampipe/pkg/error_helpers"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"golang.org/x/exp/maps"
	"path"
	"strings"
)

type Manager struct {
	registeredExporters  map[string]Exporter
	registeredExtensions map[string]Exporter
}

func NewManager() *Manager {
	return &Manager{
		registeredExporters:  make(map[string]Exporter),
		registeredExtensions: make(map[string]Exporter),
	}
}

func (r *Manager) Register(exporter Exporter) {
	r.registeredExporters[exporter.Name()] = exporter
	r.registeredExporters[exporter.FileExtension()] = exporter
}

func (r *Manager) resolveTargetsFromArgs(exportArgs []string, executionName string) ([]*Target, error) {

	var targets = make(map[string]*Target)
	var targetErrors []error

	for _, export := range exportArgs {
		export = strings.TrimSpace(export)
		if len(export) == 0 {
			// if this is an empty string, ignore
			continue
		}

		t, err := r.getExportTarget(export, executionName)
		if err != nil {
			targetErrors = append(targetErrors, err)
			continue
		}

		// add to map if not already there
		if _, ok := targets[t.filePath]; !ok {
			targets[t.filePath] = t
		}
	}

	// convert target map into array
	targetList := maps.Values(targets)
	return targetList, error_helpers.CombineErrors(targetErrors...)
}

func (r *Manager) getExportTarget(export, executionName string) (*Target, error) {
	if e, ok := r.registeredExporters[export]; ok {
		t := &Target{
			exporter: e,
			filePath: GenerateDefaultExportFileName(e, executionName),
		}
		return t, nil
	}

	if e, ok := r.registeredExtensions[path.Ext(export)]; ok {
		t := &Target{
			exporter: e,
			filePath: export,
		}
		return t, nil
	}
	return nil, fmt.Errorf("formatter satisfying '%s' not found", export)
}

func (r *Manager) DoExport(ctx context.Context, targetName string, source ExportSourceData, exports []string) error {

	if len(exports) == 0 {
		return nil
	}

	// get the short name for the target
	parsedResource, err := modconfig.ParseResourceName(targetName)
	if err != nil {
		return err
	}
	shortName := parsedResource.Name

	targets, err := r.resolveTargetsFromArgs(exports, shortName)
	if err != nil {
		return err
	}

	var errors []error
	for _, target := range targets {
		if err := target.Export(ctx, source); err != nil {
			errors = append(errors, err)
		}
	}
	return error_helpers.CombineErrors(errors...)
}
