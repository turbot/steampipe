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

func Load(ctx context.Context, conn *pgx.Conn) (*ServerSettings, error) {
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

		for i := 0; i < reflectedSettings.NumField(); i++ {
			fieldType := reflectedSettings.Type().Field(i)

			tag := fieldType.Tag.Get("setting_key")

			if tag == settingName {
				log.Println("[INFO] serversetting: loading value of", settingName, "into field", tag)

				// TODO: add panic handling
				reflectedSettings.Field(i).Set(reflect.ValueOf(settingValue).Convert(fieldType.Type))

				// we have a value
				// go to the next field
				break
			}
		}
	}
	return settings, nil
}
