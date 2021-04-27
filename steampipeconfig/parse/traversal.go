package parse

import (
	"fmt"

	"github.com/turbot/go-kit/helpers"

	"github.com/hashicorp/hcl/v2"
)

// IsQualifiedTraversal :: a 'qualified traversal' is of form
// <mod>.<query|action|policy>.<name>.xxx.xxx
func IsQualifiedTraversal(traversal hcl.Traversal) bool {
	if len(traversal) < 3 {
		return false
	}
	s := traversal.SimpleSplit()
	if isReferenceable(s.Abs.RootName()) {
		return false
	}
	return isReferenceable(s.Rel[0].(hcl.TraverseAttr).Name)
}

func NameFromTraversal(traversal hcl.Traversal) (string, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	// we assume any traversal here is a fully qualified travers
	//if !IsQualifiedTraversal(traversal) {
	//	diags = append(diags, &hcl.Diagnostic{
	//		Severity: hcl.DiagError,
	//		Summary:  "Invalid traversal ",
	//		Detail:   fmt.Sprintf("NameFromTraversal failed for '%s'. Expected format: <mod>.<query|action|policy>.<name>"),
	//	})
	//}
	return TraversalAsString(traversal), diags
}

// TraversalAsString:: convert a traversal to a path string
// TODO feels wrong we have to write this
func TraversalAsString(traversal hcl.Traversal) string {
	s := traversal.SimpleSplit()
	name := s.Abs.RootName()
	for _, r := range s.Rel {
		name += fmt.Sprintf(".%s", r.(hcl.TraverseAttr).Name)
	}
	return name

}

//func (v unparsedInteractiveVariableValue) ParseVariableValue(mode configs.VariableParsingMode) (*terraform.InputValue, tfdiags.Diagnostics) {
//	var diags tfdiags.Diagnostics
//	val, valDiags := mode.Parse(v.Name, v.RawValue)
//	diags = diags.Append(valDiags)
//	if diags.HasErrors() {
//		return nil, diags
//	}
//	return &terraform.InputValue{
//		Value:      val,
//		SourceType: terraform.ValueFromInput,
//	}, diags
//}

func isReferenceable(name string) bool {
	// TODO USE block tpyes
	return helpers.StringSliceContains([]string{"mod", "control", "query", "control_group"}, name)
}
