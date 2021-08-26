package parse

import (
	"fmt"
	"strings"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

func ctyObjectToPostgresMap(val cty.Value) (map[string]string, error) {
	res := make(map[string]string)
	it := val.ElementIterator()
	for it.Next() {
		k, v := it.Element()

		// decode key
		var key string
		gocty.FromCtyValue(k, &key)

		// decode the value into a postgres compatible
		valStr, err := ctyToPostgresString(v)
		if err != nil {
			err := fmt.Errorf("invalid value provided for param '%s': %v", key, err)
			return nil, err
		}

		res[key] = valStr
	}
	return res, nil
}

func ctyToPostgresString(v cty.Value) (valStr string, err error) {
	ty := v.Type()

	if ty.IsTupleType() {
		var array []string
		if array, err = ctyTupleToPostgresArray(v); err == nil {
			valStr = fmt.Sprintf("[%s]", strings.Join(array, ","))
		}
		return
	}

	switch ty {
	case cty.Bool:
		var target bool
		if err = gocty.FromCtyValue(v, &target); err == nil {
			valStr = fmt.Sprintf("%v", target)
		}
		return
	case cty.Number:
		var target int
		if err = gocty.FromCtyValue(v, &target); err == nil {
			valStr = fmt.Sprintf("%d", target)
			return
		} else {
			var targetf float64
			if err = gocty.FromCtyValue(v, &targetf); err == nil {
				valStr = fmt.Sprintf("%d", target)
			}
			return
		}

	case cty.String:
		var target string
		if err := gocty.FromCtyValue(v, &target); err == nil {
			valStr = fmt.Sprintf("'%s'", target)
		}
		return
	}

	return "", fmt.Errorf("unsupported type '%s'", ty.FriendlyName())
}

func ctyTupleToPostgresArray(val cty.Value) ([]string, error) {
	var res []string
	it := val.ElementIterator()
	for it.Next() {
		_, v := it.Element()
		// decode the value into a postgres compatible
		valStr, err := ctyToPostgresString(v)
		if err != nil {
			return nil, err
		}

		res = append(res, valStr)
	}
	return res, nil
}
