package modconfig

import "github.com/turbot/go-kit/helpers"

// we cannot use the Base directly as decoded as the runtime dependencies and resource metadata
// are not stored in the evaluation context
// instead, resolve the base from the run context (passed as a ResourceMapsProvider)
func resolveBase(base HclResource, resourceMapProvider ResourceMapsProvider) (HclResource, bool) {
	if helpers.IsNil(base) {
		return nil, false
	}
	parsedName, err := ParseResourceName(base.Name())
	if err != nil {
		return nil, false
	}
	return GetResource(resourceMapProvider, parsedName)
}
