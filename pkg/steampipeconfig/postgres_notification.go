package steampipeconfig

import (
	"github.com/turbot/steampipe/pkg/error_helpers"
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

type ConnectionErrorNotification struct {
	PostgresNotification
	Errors []string
	// TODO separate Warning
}

func NewSchemaUpdateNotification() *PostgresNotification {
	return &PostgresNotification{
		StructVersion: PostgresNotificationStructVersion,
		Type:          PgNotificationSchemaUpdate,
	}
}

func NewConnectionErrorNotification(errorAndWarnings error_helpers.ErrorAndWarnings) *ConnectionErrorNotification {
	res := &ConnectionErrorNotification{
		PostgresNotification: PostgresNotification{
			StructVersion: PostgresNotificationStructVersion,
			Type:          PgNotificationConnectionError,
		},
	}
	// TODO colour - add Error:/Warning prefix?
	if errorAndWarnings.Error != nil {
		res.Errors = []string{errorAndWarnings.Error.Error()}
	}
	res.Errors = append(res.Errors, errorAndWarnings.Warnings...)
	return res
}
