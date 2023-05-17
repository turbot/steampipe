package steampipeconfig

import (
	"fmt"
)

type ValidationFailure struct {
	Plugin             string
	ConnectionName     string
	Message            string
	ShouldDropIfExists bool
}

func (v ValidationFailure) String() string {
	return fmt.Sprintf(
		"Connection: %s\nPlugin:     %s\nError:      %s",
		v.ConnectionName,
		v.Plugin,
		v.Message,
	)
}
