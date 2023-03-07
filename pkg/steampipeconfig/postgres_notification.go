package steampipeconfig

const PostgresNotificationStructVersion = 20230306

type PostgresNotificationType int

const (
	PgNotificationSchemaUpdate PostgresNotificationType = iota + 1
)

type PostgresNotification struct {
	StructVersion int
	Type          PostgresNotificationType
}

type SchemaUpdateNotification struct {
	StructVersion int
	Type          PostgresNotificationType
	Update        []string
	Delete        []string
}

func NewSchemaUpdateNotification(update, delete []string) *SchemaUpdateNotification {
	return &SchemaUpdateNotification{
		StructVersion: PostgresNotificationStructVersion,
		Type:          PgNotificationSchemaUpdate,
		Update:        update,
		Delete:        delete,
	}
}
