package modconfig

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

// Local is a struct representing a Local resource
type Local struct {
	HclResourceBase
	ResourceWithMetadataBase

	ShortName string
	FullName  string `cty:"name"`

	Value     cty.Value
	DeclRange hcl.Range
	Mod       *Mod `cty:"mod"`
}

func NewLocal(name string, val cty.Value, declRange hcl.Range, mod *Mod) *Local {
	l := &Local{
		ShortName: name,
		FullName:  fmt.Sprintf("local.%s", name),
		Value:     val,
		DeclRange: declRange,
	}
	l.SetMod(mod)
	return l
}

// Name implements HclResource, ResourceWithMetadata
func (l *Local) Name() string {
	return l.FullName
}

// OnDecoded implements HclResource
func (l *Local) OnDecoded(*hcl.Block) hcl.Diagnostics { return nil }

// AddReference implements HclResource
func (l *Local) AddReference(*ResourceReference) {}

// SetMod implements HclResource
func (l *Local) SetMod(mod *Mod) {
	l.Mod = mod
	// add mod name to full name
	l.FullName = fmt.Sprintf("%s.%s", mod.ShortName, l.FullName)
}

// GetMod implements HclResource
func (l *Local) GetMod() *Mod {
	return l.Mod
}

// CtyValue implements HclResource
func (l *Local) CtyValue() (cty.Value, error) {
	return l.Value, nil
}

// GetDeclRange implements HclResource
func (l *Local) GetDeclRange() *hcl.Range {
	return &l.DeclRange
}
