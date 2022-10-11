package export

import (
	"fmt"
	"github.com/turbot/steampipe/pkg/error_helpers"
	"golang.org/x/exp/maps"
	"path"
	"strings"
)

type Resolver struct {
	registeredExporters  map[string]Exporter
	registeredExtensions map[string]Exporter
}

func NewResolver() *Resolver {
	return &Resolver{
		registeredExporters:  make(map[string]Exporter),
		registeredExtensions: make(map[string]Exporter),
	}
}

func (r *Resolver) Register(exporter Exporter) {
	r.registeredExporters[exporter.Name()] = exporter
	r.registeredExporters[exporter.FileExtension()] = exporter
}

func (r *Resolver) ResolveTargetsFromArgs(exportArgs []string, executionName string) ([]*Target, error) {

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

func (r *Resolver) getExportTarget(export, executionName string) (*Target, error) {
	if e, ok := r.registeredExporters[export]; ok {
		t := &Target{
			exporter: e,
			filePath: generateDefaultExportFileName(e, executionName),
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
