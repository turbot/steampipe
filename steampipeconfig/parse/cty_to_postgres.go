package parse

import (
	"fmt"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
	"github.com/zclconf/go-cty/cty/json"
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
			return
		} else {
			var targetf float64
			if err = gocty.FromCtyValue(v, &targetf); err == nil {
				valStr = fmt.Sprintf("%d", target)
			}
		}

	case cty.String:
		var target string
		if err := gocty.FromCtyValue(v, &target); err == nil {
			valStr = fmt.Sprintf("'%s'", target)
		}
	default:
		var json string
		// wrap as postgres string
		if json, err = ctyToJSON(v); err == nil {
			valStr = fmt.Sprintf("'%s'", json)
		}

	}

	return valStr, err
}

func ctyToJSON(val cty.Value) (string, error) {

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
