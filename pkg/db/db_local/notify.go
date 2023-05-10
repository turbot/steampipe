package db_local

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/sperr"
	"log"
)

// SendPostgresNotification send a postgres notification that the schema has chganged
func SendPostgresNotification(_ context.Context, conn *pgx.Conn, notification any) error {
	notificationBytes, err := json.Marshal(notification)
	if err != nil {
		return sperr.WrapWithMessage(err, "error marshalling Postgres notification")
	}

	log.Printf("[TRACE] Send update notification")

	sql := fmt.Sprintf("select pg_notify('%s', $1)", constants.PostgresNotificationChannel)
	_, err = conn.Exec(context.Background(), sql, notificationBytes)
	if err != nil {
		return sperr.WrapWithMessage(err, "error sending Postgres notification")
	}
	return nil
}
