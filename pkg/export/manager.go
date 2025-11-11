package export

import (
	"context"
	"fmt"
	"path"
	"strings"
	"sync"

	"github.com/turbot/pipe-fittings/v2/utils"
	"github.com/turbot/steampipe-plugin-sdk/v5/sperr"
	"github.com/turbot/steampipe/v2/pkg/error_helpers"
	"github.com/turbot/steampipe/v2/pkg/statushooks"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

type Manager struct {
	registeredExporters  map[string]Exporter
	registeredExtensions map[string]Exporter
	mu                   sync.RWMutex
}

func NewManager() *Manager {
	return &Manager{
		registeredExporters:  make(map[string]Exporter),
		registeredExtensions: make(map[string]Exporter),
	}
}

func (m *Manager) Register(exporter Exporter) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	name := exporter.Name()
	if _, ok := m.registeredExporters[name]; ok {
		return fmt.Errorf("failed to register exporter - duplicate name %s", name)
	}
	m.registeredExporters[exporter.Name()] = exporter

	// if the exporter has an alias, also register by alias
	if alias := exporter.Alias(); alias != "" {
		if _, ok := m.registeredExporters[alias]; ok {
			return fmt.Errorf("failed to register exporter - duplicate name %s", name)
		}
		m.registeredExporters[alias] = exporter
	}

	// now register extension
	ext := exporter.FileExtension()
	m.registerExporterByExtension(exporter, ext)
	// if the extension has multiple segments, try to register for the short version as well
	if shortExtension := path.Ext(ext); shortExtension != ext {
		m.registerExporterByExtension(exporter, shortExtension)
	}
	return nil
}

func (m *Manager) registerExporterByExtension(exporter Exporter, ext string) {
	// do we already have an exporter registered for this extension?
	if existing, ok := m.registeredExtensions[ext]; ok {

		// check if either the existing or new template is the default for extension
		existingIsDefaultForExt := isDefaultExporterForExtension(existing)
		newIsDefaultForExt := isDefaultExporterForExtension(exporter)

		// if  NEITHER are default for the extension, there is a clash which cannot be resolved -
		// we must remove the existing key
		if !newIsDefaultForExt && !existingIsDefaultForExt {
			delete(m.registeredExtensions, ext)
		}

		// if existing is default and new isn't, nothing to do
		if existingIsDefaultForExt {
			return
		}

		// to get here, new must be default exporter for extension
		// (it is impossible for both to be default as that implies duplicate exporter names)
		// fall through to...
	}

	// register the extension
	m.registeredExtensions[ext] = exporter
}

// an exporter is the 'default for extension' if the exporter name is the same as the extension name
// i.e. json exporter would be the default for the `.json` extension
func isDefaultExporterForExtension(existing Exporter) bool {
	return strings.TrimPrefix(existing.FileExtension(), ".") == existing.Name()
}

func (m *Manager) resolveTargetsFromArgs(exportArgs []string, executionName string) ([]*Target, error) {
	var targets = make(map[string]*Target)
	var targetErrors []error

	for _, export := range exportArgs {
		export = strings.TrimSpace(export)
		if len(export) == 0 {
			// if this is an empty string, ignore
			continue
		}

		t, err := m.getExportTarget(export, executionName)
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

func (m *Manager) getExportTarget(export, executionName string) (*Target, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if e, ok := m.registeredExporters[export]; ok {
		t := &Target{
			exporter: e,
			filePath: GenerateDefaultExportFileName(executionName, e.FileExtension()),
		}
		return t, nil
	}

	// now try by extension
	ext := path.Ext(export)
	if e, ok := m.registeredExtensions[ext]; ok {
		t := &Target{
			exporter:      e,
			filePath:      export,
			isNamedTarget: true,
		}
		return t, nil
	}

	return nil, fmt.Errorf("formatter satisfying '%s' not found", export)
}

func (m *Manager) DoExport(ctx context.Context, targetName string, source ExportSourceData, exports []string) ([]string, error) {
	var errors []error
	var msg string
	var expLocation []string

	if len(exports) == 0 {
		return nil, nil
	}

	targets, err := m.resolveTargetsFromArgs(exports, targetName)

	if err != nil {
		return nil, err
	}

	for idx, target := range targets {
		statushooks.SetStatus(ctx, fmt.Sprintf("Exporting %d of %d", idx+1, len(targets)))
		if msg, err = target.Export(ctx, source); err != nil {
			errors = append(errors, err)
		} else {
			expLocation = append(expLocation, msg)
		}
	}
	return expLocation, error_helpers.CombineErrors(errors...)
}

// HasNamedExport returns true if any of the export arguments has a filename (--export=file.json) instead of the format name (--export=json)
// panics if a target is not valid
func (m *Manager) HasNamedExport(exports []string) bool {
	for _, export := range exports {
		target, err := m.getExportTarget(export, "dummy_exec_name")
		error_helpers.FailOnError(err)
		if target.isNamedTarget {
			return true
		}
	}
	return false
}

func (m *Manager) ValidateExportFormat(exports []string) error {
	var invalidFormats []string
	var targets []*Target
	for _, export := range exports {
		target, err := m.getExportTarget(export, "dummy_exec_name")
		if err != nil {
			invalidFormats = append(invalidFormats, export)
		}
		targets = append(targets, target)
	}
	if invalidCount := len(invalidFormats); invalidCount > 0 {
		return fmt.Errorf("invalid export %s: '%s'", utils.Pluralize("format", invalidCount), strings.Join(invalidFormats, "','"))
	}
	// verify all are either named or unnamed but not both
	hasNamed := slices.ContainsFunc(targets, func(t *Target) bool { return t.isNamedTarget })
	hasUnnamed := slices.ContainsFunc(targets, func(t *Target) bool { return !t.isNamedTarget })

	if hasNamed && hasUnnamed {
		return sperr.New("combination of named and unnamed exports is not supported")
	}

	return nil
}
