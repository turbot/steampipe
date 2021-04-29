package modconfig

import (
	"log"
	"reflect"

	"github.com/turbot/go-kit/helpers"
)

// HclProperties
// all properties parsed from hcl have a json tag
// return map of property name to the pointer to the destination property
// this is used to populate a control during decoding
func HclProperties(item interface{}) map[string]interface{} {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[WARN] HclProperties failed with panic: %v", r)
		}
	}()
	r := make(map[string]interface{})
	t := reflect.TypeOf(helpers.DereferencePointer(item))
	val := reflect.ValueOf(item)
	if val.Type().Kind() == reflect.Ptr {
		val = val.Elem()
	}
	for i := 0; i < val.NumField(); i++ {
		structField := t.Field(i)
		attribute, ok := structField.Tag.Lookup("json")
		if ok && attribute != "-" {
			r[attribute] = val.Field(i).Addr().Interface()
		}
	}
	return r
}
