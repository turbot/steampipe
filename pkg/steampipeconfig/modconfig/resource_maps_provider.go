package modconfig

// GetDashboardInput looks for an input with a given parent dashboard
// this is required as GetResource does not support Inputs
func GetDashboardInput(provider ResourceMapsProvider, inputName, dashboardName string) (*DashboardInput, bool) {
	resourceMaps := provider.GetResourceMaps()

	dasboardInputs, ok := resourceMaps.DashboardInputs[dashboardName]
	if !ok {
		return nil, false
	}

	input, ok := dasboardInputs[inputName]

	return input, ok
}
