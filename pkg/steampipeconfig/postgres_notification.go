package steampipeconfig

import (
	"github.com/turbot/pipe-fittings/v2/error_helpers"
)

const PostgresNotificationStructVersion = 20230306

type PostgresNotificationType int

const (
	PgNotificationSchemaUpdate PostgresNotificationType = iota + 1
	PgNotificationConnectionError
)

type PostgresNotification struct {
	StructVersion int
	Type          PostgresNotificationType
}

type ErrorsAndWarningsNotification struct {
	PostgresNotification
	Errors   []string
	Warnings []string
}

func NewSchemaUpdateNotification() *PostgresNotification {
	return &PostgresNotification{
		StructVersion: PostgresNotificationStructVersion,
		Type:          PgNotificationSchemaUpdate,
	}
}

func NewErrorsAndWarningsNotification(errorAndWarnings error_helpers.ErrorAndWarnings) *ErrorsAndWarningsNotification {
	res := &ErrorsAndWarningsNotification{
		PostgresNotification: PostgresNotification{
			StructVersion: PostgresNotificationStructVersion,
			Type:          PgNotificationConnectionError,
		},
	}

	if errorAndWarnings.Error != nil {
		res.Errors = []string{errorAndWarnings.Error.Error()}
	}
	res.Warnings = append(res.Warnings, errorAndWarnings.Warnings...)
	return res
}
