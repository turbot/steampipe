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

type SteampipeNotification struct {
	StructVersion int
	Type          PostgresNotificationType
}

func NewSchemaUpdateNotification(notificationType PostgresNotificationType) *SteampipeNotification {
	return &SteampipeNotification{
		StructVersion: PostgresNotificationStructVersion,
		Type:          notificationType,
	}
}
