package type_conversion

import (
	"fmt"
	"strings"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
	"github.com/zclconf/go-cty/cty/json"
)

// CtyToJSON converts a cty value to it;s JSON representation
func CtyToJSON(val cty.Value) (string, error) {

	if !val.IsWhollyKnown() {
		return "", fmt.Errorf("cannot serialize unknown values")
	}

	if val.IsNull() {
		return "{}", nil
	}

	buf, err := json.Marshal(val, val.Type())
	if err != nil {
		return "", err
	}

	return string(buf), nil

}

// CtyToString convert a cty value into a string representation of the value
func CtyToString(v cty.Value) (valStr string, err error) {
	if v.IsNull() || !v.IsWhollyKnown() {
		return "", nil
	}
	ty := v.Type()
	switch {
	case ty.IsTupleType(), ty.IsListType():
		{
			var array []string
			if array, err = ctyTupleToArrayOfPgStrings(v); err == nil {
				valStr = fmt.Sprintf("[%s]", strings.Join(array, ","))
			}
			return
		}
	}

	switch ty {
	case cty.Bool:
		var target bool
		if err = gocty.FromCtyValue(v, &target); err == nil {
			valStr = fmt.Sprintf("%v", target)
		}
	case cty.Number:
		var target int
		if err = gocty.FromCtyValue(v, &target); err == nil {
			valStr = fmt.Sprintf("%d", target)
		} else {
			var targetf float64
			if err = gocty.FromCtyValue(v, &targetf); err == nil {
				valStr = fmt.Sprintf("%d", target)
			}
		}
	case cty.String:
		var target string
		if err := gocty.FromCtyValue(v, &target); err == nil {
			valStr = target
		}
	default:
		var json string
		// wrap as postgres string
		if json, err = CtyToJSON(v); err == nil {
			valStr = json
		}

	}

	return valStr, err
}

func CtyToGo(v cty.Value) (val interface{}, err error) {
	if v.IsNull() {
		return nil, nil
	}
	ty := v.Type()
	switch {
	case ty.IsTupleType(), ty.IsListType():
		{
			var array []string
			if array, err = ctyTupleToArrayOfStrings(v); err == nil {
				val = array
			}
			return
		}
	}

	switch ty {
	case cty.Bool:
		var target bool
		if err = gocty.FromCtyValue(v, &target); err == nil {
			val = target
		}

	case cty.Number:
		var target int
		if err = gocty.FromCtyValue(v, &target); err == nil {
			val = target
		} else {
			var targetf float64
			if err = gocty.FromCtyValue(v, &targetf); err == nil {
				val = targetf
			}
		}
	case cty.String:
		var target string
		if err := gocty.FromCtyValue(v, &target); err == nil {
			val = target
		}

	default:
		var json string
		// wrap as postgres string
		if json, err = CtyToJSON(v); err == nil {
			val = json
		}
	}

	return
}

// CtyTypeToHclType converts a cty type to a hcl type
// accept multiple types and use the first non null and non dynamic one
func CtyTypeToHclType(types ...cty.Type) string {
	// find which if any of the types are non nil and not dynamic
	t := getKnownType(types)
	if t == cty.NilType {
		return ""
	}

	friendlyName := t.FriendlyName()

	// func to convert from ctyt aggregate syntax to hcl
	convertAggregate := func(prefix string) (string, bool) {
		if strings.HasPrefix(friendlyName, prefix) {
			return fmt.Sprintf("%s(%s)", strings.TrimSuffix(prefix, " of "), strings.TrimPrefix(friendlyName, prefix)), true
		}
		return "", false
	}

	if convertedName, isList := convertAggregate("list of "); isList {
		return convertedName
	}
	if convertedName, isMap := convertAggregate("map of "); isMap {
		return convertedName
	}
	if convertedName, isSet := convertAggregate("set of "); isSet {
		return convertedName
	}
	if friendlyName == "tuple" {
		elementTypes := t.TupleElementTypes()
		if len(elementTypes) == 0 {
			// we cannot determine the eleemnt type
			return "list"
		}
		// if there are element types, use the first one (assume homogeneous)
		underlyingType := elementTypes[0]
		return fmt.Sprintf("list(%s)", CtyTypeToHclType(underlyingType))
	}
	if friendlyName == "dynamic" {
		return ""
	}
	return friendlyName
}

// from a list oif cty typoes, return the first which is non nil and not dynamic
func getKnownType(types []cty.Type) cty.Type {
	for _, t := range types {
		if t != cty.NilType && !t.HasDynamicTypes() {
			return t
		}
	}
	return cty.NilType
}

func ctyTupleToArrayOfPgStrings(val cty.Value) ([]string, error) {
	var res []string
	it := val.ElementIterator()
	for it.Next() {
		_, v := it.Element()
		// decode the value into a postgres compatible
		valStr, err := CtyToPostgresString(v)
		if err != nil {
			return nil, err
		}

		res = append(res, valStr)
	}
	return res, nil
}

func ctyTupleToArrayOfStrings(val cty.Value) ([]string, error) {
	var res []string
	it := val.ElementIterator()
	for it.Next() {
		_, v := it.Element()

		var valStr string
		if err := gocty.FromCtyValue(v, &valStr); err != nil {
			return nil, err
		}

		res = append(res, valStr)
	}
	return res, nil
}
