package plugin

import "github.com/hashicorp/hcl/v2"

type ConnectionConfigRange struct {
	ConnectionName string
	DeclRange      *hcl.Range
}
