package inputvars

//
//import (
//	"fmt"
//
//	"github.com/hashicorp/hcl/v2"
//	"github.com/hashicorp/terraform/tfdiags"
//	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
//	"github.com/zclconf/go-cty/cty"
//	"github.com/zclconf/go-cty/cty/convert"
//)
//
//// InputValue represents a value for a variable in the configuration, provided
//// as part of the definition of an operation.
//type InputValue struct {
//	Value      cty.Value
//	SourceType ValueSourceType
//
//	// SourceRange provides source location information for values whose
//	// SourceType is either ValueFromConfig or ValueFromFile. It is not
//	// populated for other source types, and so should not be used.
//	SourceRange tfdiags.SourceRange
//}
//
//// ValueSourceType describes what broad category of source location provided
//// a particular value.
//type ValueSourceType rune
//
//const (
//	// ValueFromUnknown is the zero value of ValueSourceType and is not valid.
//	ValueFromUnknown ValueSourceType = 0
//
//	// ValueFromConfig indicates that a value came from a .tf or .tf.json file,
//	// e.g. the default value defined for a variable.
//	ValueFromConfig ValueSourceType = 'C'
//
//	// ValueFromAutoFile indicates that a value came from a "values file", like
//	// a .tfvars file, that was implicitly loaded by naming convention.
//	ValueFromAutoFile ValueSourceType = 'F'
//
//	// ValueFromNamedFile indicates that a value came from a named "values file",
//	// like a .tfvars file, that was passed explicitly on the command line (e.g.
//	// -var-file=foo.tfvars).
//	ValueFromNamedFile ValueSourceType = 'N'
//
//	// ValueFromCLIArg indicates that the value was provided directly in
//	// a CLI argument. The name of this argument is not recorded and so it must
//	// be inferred from context.
//	ValueFromCLIArg ValueSourceType = 'A'
//
//	// ValueFromEnvVar indicates that the value was provided via an environment
//	// variable. The name of the variable is not recorded and so it must be
//	// inferred from context.
//	ValueFromEnvVar ValueSourceType = 'E'
//
//	// ValueFromInput indicates that the value was provided at an interactive
//	// input prompt.
//	ValueFromInput ValueSourceType = 'I'
//
//	// ValueFromModFile indicates that the value was provided in the 'Require' section of a mod file
//	ValueFromModFile ValueSourceType = 'M'
//)
//
//func (v *InputValue) GoString() string {
//	if (v.SourceRange != tfdiags.SourceRange{}) {
//		return fmt.Sprintf("&InputValue{Value: %#v, SourceType: %#v, SourceRange: %#v}", v.Value, v.SourceType, v.SourceRange)
//	} else {
//		return fmt.Sprintf("&InputValue{Value: %#v, SourceType: %#v}", v.Value, v.SourceType)
//	}
//}
//
//func (v *InputValue) SourceTypeString() string {
//	switch v.SourceType {
//	case ValueFromConfig:
//		return "config"
//	case ValueFromAutoFile:
//		return "auto file"
//	case ValueFromNamedFile:
//		return "name file"
//	case ValueFromCLIArg:
//		return "CLI arg"
//	case ValueFromEnvVar:
//		return "env var"
//	case ValueFromInput:
//		return "user input"
//	default:
//		return "unknown"
//	}
//}
//
////go:generate go run golang.org/x/tools/cmd/stringer -type ValueSourceType
//
//// InputValues is a map of InputValue instances.
//type InputValues map[string]*InputValue
//
//// Override merges the given value maps with the receiver, overriding any
//// conflicting keys so that the latest definition wins.
//func (vv InputValues) Override(others ...InputValues) InputValues {
//	ret := make(InputValues)
//	for k, v := range vv {
//		ret[k] = v
//	}
//	for _, other := range others {
//		for k, v := range other {
//			ret[k] = v
//		}
//	}
//	return ret
//}
//
//// JustValues returns a map that just includes the values, discarding the
//// source information.
//func (vv InputValues) JustValues() map[string]cty.Value {
//	ret := make(map[string]cty.Value, len(vv))
//	for k, v := range vv {
//		ret[k] = v.Value
//	}
//	return ret
//}
//
//// SameValues returns true if the given InputValues has the same values as
//// the receiver, disregarding the source types and source ranges.
////
//// Values are compared using the cty "RawEquals" method, which means that
//// unknown values can be considered equal to one another if they are of the
//// same type.
//func (vv InputValues) SameValues(other InputValues) bool {
//	if len(vv) != len(other) {
//		return false
//	}
//
//	for k, v := range vv {
//		ov, exists := other[k]
//		if !exists {
//			return false
//		}
//		if !v.Value.RawEquals(ov.Value) {
//			return false
//		}
//	}
//
//	return true
//}
//
//// HasValues returns true if the reciever has the same values as in the given
//// map, disregarding the source types and source ranges.
////
//// Values are compared using the cty "RawEquals" method, which means that
//// unknown values can be considered equal to one another if they are of the
//// same type.
//func (vv InputValues) HasValues(vals map[string]cty.Value) bool {
//	if len(vv) != len(vals) {
//		return false
//	}
//
//	for k, v := range vv {
//		oVal, exists := vals[k]
//		if !exists {
//			return false
//		}
//		if !v.Value.RawEquals(oVal) {
//			return false
//		}
//	}
//
//	return true
//}
//
//// Identical returns true if the given InputValues has the same values,
//// source types, and source ranges as the receiver.
////
//// Values are compared using the cty "RawEquals" method, which means that
//// unknown values can be considered equal to one another if they are of the
//// same type.
////
//// This method is primarily for testing. For most practical purposes, it's
//// better to use SameValues or HasValues.
//func (vv InputValues) Identical(other InputValues) bool {
//	if len(vv) != len(other) {
//		return false
//	}
//
//	for k, v := range vv {
//		ov, exists := other[k]
//		if !exists {
//			return false
//		}
//		if !v.Value.RawEquals(ov.Value) {
//			return false
//		}
//		if v.SourceType != ov.SourceType {
//			return false
//		}
//		if v.SourceRange != ov.SourceRange {
//			return false
//		}
//	}
//
//	return true
//}
//
//func (vv InputValues) DefaultTo(other InputValues) {
//	for k, otherVal := range other {
//		if val, ok := vv[k]; !ok || !val.Value.IsKnown() {
//			vv[k] = otherVal
//		}
//	}
//}
//
//
//// SetVariableValues determines whether the given variable is a public variable and if so sets its value
//func (vv InputValues) SetVariableValues(m *modconfig.ModVariableMap) {
//	for name, inputValue := range vv {
//		variable, ok := m.PublicVariables[name]
//		// if this variable does not exist in public variables, skip
//		if !ok {
//			// we should have already caught this
//			continue
//		}
//		variable.SetInputValue(
//			inputValue.Value,
//			inputValue.SourceTypeString(),
//			inputValue.SourceRange)
//	}
//}
