package db_local

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	"github.com/turbot/steampipe-plugin-sdk/v5/sperr"
	"github.com/turbot/steampipe/pkg/constants"
)

// SendPostgresNotification send a postgres notification that the schema has chganged
func SendPostgresNotification(ctx context.Context, conn *sql.Conn, notification any) error {
	notificationBytes, err := json.Marshal(notification)
	if err != nil {
		return sperr.WrapWithMessage(err, "error marshalling Postgres notification")
	}

	log.Printf("[TRACE] Send update notification")

	sql := fmt.Sprintf("select pg_notify('%s', $1)", constants_steampipe.PostgresNotificationChannel)
	_, err = conn.ExecContext(ctx, sql, notificationBytes)
	if err != nil {
		return sperr.WrapWithMessage(err, "error sending Postgres notification")
	}
	return nil
}
