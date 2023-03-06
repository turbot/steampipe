package steampipeconfig

import (
	"golang.org/x/exp/maps"
)

const ConnectionUpdateNotificationStructVersion = 20230306
const PostgresNotificationStructVersion = 20230306

type PostgresNotificationType int

const (
	PgNotificationSchemaUpdate PostgresNotificationType = iota + 1
)

type PostgresNotification struct {
	StructVersion int
	Type          PostgresNotificationType
	Payload       any
}

type ConnectionUpdateNotification struct {
	StructVersion int
	Update        []string
	Delete        []string
}

func NewConnectionUpdateNotification(updates *ConnectionUpdates) *ConnectionUpdateNotification {
	return &ConnectionUpdateNotification{
		StructVersion: ConnectionUpdateNotificationStructVersion,
		Update:        maps.Keys(updates.Update),
		Delete:        maps.Keys(updates.Delete),
	}
}
