package serversettings

import (
	"context"
	"fmt"
	"log"
	"reflect"

	"github.com/jackc/pgx/v5"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/utils"
)

type valueWithType struct {
	val     any
	valType string
}

// SetupSql returns the set of SQL statements to fully replace any existing
// settings table with a new one and populates the values
func (s *ServerSettings) SetupTable(ctx context.Context, conn *pgx.Conn) (err error) {
	utils.LogTime("db_local.initializeServerSettingsTable start")
	defer utils.LogTime("db_local.initializeServerSettingsTable end")

	// start a transaction on this connection
	tx, err := conn.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		}
	}()

	// drop the old table (alternative is "if exists then truncate" which is more expensive)
	// this also allows us to modify the table structure without having to go through complex
	// migrations
	err = dropServerSettingsTable(ctx, tx)
	if err != nil {
		return err
	}
	err = createServerSettingsTable(ctx, tx)
	if err != nil {
		return err
	}
	err = setupGrantsOnServerSettingsTable(ctx, tx)
	if err != nil {
		return err
	}
	err = populateServerSettingsTable(ctx, tx, s)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func populateServerSettingsTable(ctx context.Context, tx pgx.Tx, settings *ServerSettings) error {
	settingsMap := settings.createMap(ctx)
	for name, value := range settingsMap {
		_, err := tx.Exec(
			ctx,
			fmt.Sprintf(
				// include the vartype for non-steampipe clients
				`INSERT INTO %s.%s (name,value,vartype) VALUES ($1,TO_JSONB($2::%s),$3)`,
				constants.InternalSchema,
				constants.ServerSettingsTable,
				value.valType,
			),
			name,
			value.val,
			value.valType,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func setupGrantsOnServerSettingsTable(ctx context.Context, tx pgx.Tx) error {
	_, err := tx.Exec(ctx, fmt.Sprintf(
		`GRANT SELECT ON TABLE %s.%s to %s;`,
		constants.InternalSchema,
		constants.ServerSettingsTable,
		constants.DatabaseUsersRole,
	))
	return err
}

func createServerSettingsTable(ctx context.Context, tx pgx.Tx) error {
	_, err := tx.Exec(ctx, fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.%s (
		name TEXT PRIMARY KEY,
		value JSONB NOT NULL,
		vartype TEXT NOT NULL);`,
		constants.InternalSchema, constants.ServerSettingsTable))

	return err
}

func dropServerSettingsTable(ctx context.Context, tx pgx.Tx) error {
	_, err := tx.Exec(ctx, fmt.Sprintf(
		`DROP TABLE IF EXISTS %s.%s;`,
		constants.InternalSchema,
		constants.ServerSettingsTable,
	))
	return err
}

// uses reflection to create a map of the settings struct that can be persisted
func (s *ServerSettings) createMap(ctx context.Context) map[string]valueWithType {
	mappedSettings := map[string]valueWithType{}

	// get the value of interface{}/ pointer point to
	val := reflect.Indirect(reflect.ValueOf(s))
	for i := 0; i < val.NumField(); i++ {
		field := val.Type().Field(i)
		tag := field.Tag.Get("setting_key")
		if tag == "" {
			continue
		}

		log.Println("[INFO] serversetting: persisting value of", field.Name, "into key", tag)
		if mappedValue := getValueWithType(val.Field(i)); mappedValue != nil {
			mappedSettings[tag] = *mappedValue
		}

	}
	return mappedSettings
}

func getValueWithType(field reflect.Value) *valueWithType {
	fieldValue := field.Interface()
	var valType string

	switch field.Kind() {
	case reflect.Bool:
		valType = "bool"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		valType = "integer"
	case reflect.String:
		valType = "text"
	case reflect.Struct:
		valType = "text"
		if timeString, converted := tryConvertFromTime(field, fieldValue); converted {
			fieldValue = timeString
		} else {
			// we don't know of any other struct types
			return nil
		}
	default:
		return nil
	}

	return &valueWithType{
		valType: valType,
		val:     fieldValue,
	}
}
