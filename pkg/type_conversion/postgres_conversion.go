package type_conversion

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

// GoToPostgresString convert a go value into a postgres representation of the value
func GoToPostgresString(v any) (string, error) {
	// pass false to indicate we want a slice to be returned as a PostgresSlice
	return goToPostgresString(v, false)
}

func goToPostgresString(v any, sliceAsJson bool) (string, error) {
	var str string

	switch arg := v.(type) {
	case nil:
		str = "null"
	case int:
		str = strconv.FormatInt(int64(arg), 10)
	case int64:
		str = strconv.FormatInt(arg, 10)
	case float64:
		str = strconv.FormatFloat(arg, 'f', -1, 64)
	case bool:
		str = strconv.FormatBool(arg)
	case []byte:
		str = QuotePostgresBytes(arg)
	case string:
		str = QuotePostgresString(arg)
	case time.Time:
		str = arg.Truncate(time.Microsecond).Format("'2006-01-02 15:04:05.999999999Z07:00:00'")
	default:
		if !sliceAsJson {
			// is this an array
			val := reflect.ValueOf(v)

			if val.Kind() == reflect.Slice {
				return goSliceToPostgresString(val)

			}
		}

		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("'%s'::jsonb", string(jsonBytes)), nil
	}
	return str, nil
}

func goSliceToPostgresString(val reflect.Value) (string, error) {
	len := val.Len()
	if len == 0 {
		return "array[]", nil
	}

	var elemStrings = make([]string, len)

	var postgresArrayType string
	for i := 0; i < len; i++ {
		elem := val.Index(i).Interface()

		t := pgArraySuffixFromElem(elem)

		if postgresArrayType == "" {
			postgresArrayType = t
		} else {
			// all elements must be same type
			if t != postgresArrayType {
				return "", fmt.Errorf("goSliceToPostgresString failed: all elements of slice must bve the same type")
			}
		}

		// pass  true to indicate we want a slice to be returned as JSON
		str, err := goToPostgresString(elem, true)
		if err != nil {
			return "", err
		}
		elemStrings[i] = str
	}
	res := fmt.Sprintf("array[%s]::%s", strings.Join(elemStrings, ","), postgresArrayType)
	return res, nil
}

func pgArraySuffixFromElem(elem any) string {
	k := reflect.ValueOf(elem).Kind()
	log.Println(k)
	switch elem.(type) {
	case string:
		return "text[]"
	case bool:
		return "bool[]"
	case int, int8, int16, int32, int64, float32, float64:
		return "numeric[]"
	case time.Time:
		return "time[]"
	default:
		return "jsonb[]"
	}
}

// CtyToPostgresString convert a cty value into a postgres representation of the value
func CtyToPostgresString(v cty.Value) (valStr string, err error) {
	ty := v.Type()

	if ty.IsTupleType() || ty.IsListType() {
		return ctyListToPostgresString(v, ty)
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
			return
		} else {
			var targetf float64
			if err = gocty.FromCtyValue(v, &targetf); err == nil {
				valStr = fmt.Sprintf("%f", targetf)
			}
		}
	case cty.String:
		var target string
		if err := gocty.FromCtyValue(v, &target); err == nil {
			valStr = QuotePostgresString(target)
		}
	default:
		var json string
		// wrap as postgres string
		if json, err = CtyToJSON(v); err == nil {
			valStr = fmt.Sprintf("'%s'::jsonb", json)
		}
	}

	return valStr, err
}

// QuotePostgresString taken from github.com/jackc/pgx/v5@v4.17.2/internal/sanitize/sanitize.go
func QuotePostgresString(str string) string {
	return "'" + strings.ReplaceAll(str, "'", "''") + "'"
}

// QuotePostgresBytes taken from github.com/jackc/pgx/v5@v4.17.2/internal/sanitize/sanitize.go
func QuotePostgresBytes(buf []byte) string {
	return `'\x` + hex.EncodeToString(buf) + "'"
}

func ctyListToPostgresString(v cty.Value, ty cty.Type) (string, error) {
	var valStr string
	array, err := ctyTupleToArrayOfPgStrings(v)
	if err != nil {
		return "", err
	}

	suffix := ""
	if len(array) == 0 {
		t := ty.ElementType()
		// cast the empty array to the appropriate type
		switch t.FriendlyName() {
		case "string":
			suffix = "::text[]"
		case "bool":
			suffix = "::bool[]"
		case "number":
			suffix = "::numeric[]"
		}
	}
	valStr = fmt.Sprintf("array[%s]%s", strings.Join(array, ","), suffix)

	return valStr, nil
}
