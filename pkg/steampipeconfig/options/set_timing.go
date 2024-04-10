package options

import "github.com/hashicorp/hcl/v2"

type CanSetTiming interface {
	SetTiming(flag string, r hcl.Range) hcl.Diagnostics
}
