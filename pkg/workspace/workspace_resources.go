package workspace

import (
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"log"
)

func (w *Workspace) GetQueryProvider(queryName string) (modconfig.QueryProvider, bool) {
	parsedName, err := modconfig.ParseResourceName(queryName)
	if err != nil {
		return nil, false
	}
	// try to find the resource
	if resource, ok := w.GetResource(parsedName); ok {
		// found a resource - is itr a query provider
		if qp := resource.(modconfig.QueryProvider); ok {
			return qp, true
		}
		log.Printf("[TRACE] GetQueryProviderImpl found a resource for '%s' but it is not a query provider", queryName)
	}

	return nil, false
}

// GetResourceMaps implements ResourceMapsProvider
func (w *Workspace) GetResourceMaps() *modconfig.ResourceMaps {
	w.loadLock.Lock()
	defer w.loadLock.Unlock()

	// if this a source snapshot workspace, create a ResourceMaps containing ONLY source snapshot paths
	if len(w.SourceSnapshots) != 0 {
		return modconfig.NewSourceSnapshotModResources(w.SourceSnapshots)
	}
	return w.Mod.ResourceMaps
}

func (w *Workspace) GetResource(parsedName *modconfig.ParsedResourceName) (resource modconfig.HclResource, found bool) {
	return w.GetResourceMaps().GetResource(parsedName)
}
