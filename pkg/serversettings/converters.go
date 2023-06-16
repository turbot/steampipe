package serversettings

import (
	"reflect"
	"time"
)

// tryConvertFromTime tries to serialize the given value(v) into a RFC3339 formatted
// string if the given field is a time.Time
func tryConvertFromTime(field reflect.Value, v any) (_ string, converted bool) {
	structInterface := field.Interface()
	if _, ok := structInterface.(time.Time); ok {
		return v.(time.Time).Format(time.RFC3339), true
	}
	return "", false
}

// tryConvertToTime tries to deserialize the value(v) into a time.Time
// if the given field is a time.Time
func tryConvertToTime(field reflect.Value, v reflect.Value) (_ time.Time, converted bool) {
	structInterface := field.Interface()
	if _, ok := structInterface.(time.Time); ok {
		parsedTime, err := time.Parse(time.RFC3339, v.String())
		if err == nil {
			return parsedTime, true
		}
	}
	return time.Now(), false
}
