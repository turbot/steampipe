package cmdconfig

// ShouldShowSpinner :: utility to control whether to show spinner
func ShouldShowSpinner() bool {
	if globalViperInstance.GetBool("query-cmd") {
		if globalViperInstance.GetBool("interactive") {
			return true
		}
		return false
	}
	return true
}
