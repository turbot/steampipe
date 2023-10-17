package db_common

import (
	"database/sql"
	"reflect"
)

func CollectOneToStructByName[T any](rows *sql.Rows) (*T, error) {
	collection, err := CollectToStructByName[T](rows)
	if err != nil {
		return nil, err
	}
	return &collection[0], nil
}

func CollectToStructByName[T any](rows *sql.Rows) ([]T, error) {
	dest := []T{}
	destType := reflect.TypeOf(dest).Elem()
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		destElem := reflect.New(destType).Elem()

		scanArgs := make([]interface{}, len(columns))
		for i := range columns {
			scanArgs[i] = new(interface{})
		}

		if err := rows.Scan(scanArgs...); err != nil {
			return nil, err
		}

		values := make(map[string]interface{})
		for i, colName := range columns {
			values[colName] = *scanArgs[i].(*interface{})
		}

		for i := 0; i < destType.NumField(); i++ {
			field := destElem.Field(i)
			colName := destType.Field(i).Tag.Get("db")

			if value, ok := values[colName]; ok {
				valueReflect := reflect.ValueOf(value)
				if valueReflect.IsValid() && valueReflect.Type().AssignableTo(field.Type()) {
					field.Set(valueReflect)
				}
			}
		}

		dest = append(dest, destElem.Interface().(T))
	}

	return dest, nil
}
