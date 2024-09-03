package modconfig

import (
	"github.com/zclconf/go-cty/cty"
)

type CtyValueProvider interface {
	CtyValue() (cty.Value, error)
}
