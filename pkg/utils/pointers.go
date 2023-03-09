package utils

// ToStringPointer converts a string into its pointer
func ToStringPointer(s string) *string {
	return &s
}

// ToIntegerPointer converts an integer into its pointer
func ToIntegerPointer(i int) *int {
	return &i
}
