package serversettings

import (
	"context"
	"fmt"
	"log"
	"reflect"

	"github.com/jackc/pgx/v5"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/sperr"
)

func Load(ctx context.Context, conn *pgx.Conn) (_ *ServerSettings, e error) {
	defer func() {
		// this function uses reflection to extract and convert values
		// we need to be able to recover from panics while using reflection
		if r := recover(); r != nil {
			e = sperr.ToError(r, sperr.WithMessage("error loading server settings"))
		}
	}()

	rows, err := conn.Query(ctx, fmt.Sprintf("SELECT name,value FROM %s.%s", constants.InternalSchema, constants.ServerSettingsTable))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	settings := new(ServerSettings)
	reflectedSettings := reflect.Indirect(reflect.ValueOf(settings))
	for rows.Next() {
		var settingName string
		var settingValue any
		if err := rows.Scan(&settingName, &settingValue); err != nil {
			return nil, sperr.WrapWithMessage(err, "error reading row from %s.%s", constants.InternalSchema, constants.ServerSettingsTable)
		}
		value := reflect.ValueOf(settingValue)

		for i := 0; i < reflectedSettings.NumField(); i++ {
			tag := reflectedSettings.Type().Field(i).Tag.Get("setting_key")

			if tag != settingName {
				continue
			}

			log.Println("[INFO] serversetting: loading value of", settingName, "into field", tag)
			field := reflectedSettings.Field(i)
			value = tryConvert(field, value, tag)

			reflectedSettings.Field(i).Set(value)
		}
	}
	return settings, nil
}

func tryConvert(field reflect.Value, value reflect.Value, tag string) reflect.Value {
	kind := field.Kind()
	// check if the target field is a struct
	if kind == reflect.Struct {
		if parsedTime, ok := tryConvertToTime(field, value); ok {
			value = reflect.ValueOf(parsedTime)
		} else {
			// we don't know of any other struct types
			log.Printf("[INFO] serversetting : unknown struct type for '%s' (%v)", tag, value)
		}
	} else {
		// for primitive types, we need to convert to the data type of the target field
		// this is mostly to handle integer fields, since the postgres uses
		// float64 for number fields, which need to be converted
		value = value.Convert(field.Type())
	}
	return value
}
