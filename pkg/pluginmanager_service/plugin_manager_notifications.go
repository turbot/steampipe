package pluginmanager_service

import (
	"context"
	"github.com/turbot/steampipe/pkg/db/db_local"
	"github.com/turbot/steampipe/pkg/error_helpers"
	"github.com/turbot/steampipe/pkg/steampipeconfig"
	"log"
)

func (m *PluginManager) SendPostgresSchemaNotification(ctx context.Context) error {
	log.Println("[DEBUG] refreshConnectionState.sendPostgreSchemaNotification start")
	defer log.Println("[DEBUG] refreshConnectionState.sendPostgreSchemaNotification end")

	return m.sendPostgresNotification(ctx, steampipeconfig.NewSchemaUpdateNotification())

}

func (m *PluginManager) SendPostgresErrorsAndWarningsNotification(ctx context.Context, errorAndWarnings error_helpers.ErrorAndWarnings) {
	if err := m.sendPostgresNotification(ctx, steampipeconfig.NewErrorsAndWarningsNotification(errorAndWarnings)); err != nil {

		log.Printf("[WARN] failed to send error notification, error")
	}

}
func (m *PluginManager) sendPostgresNotification(ctx context.Context, notification any) error {
	conn, err := m.pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	return db_local.SendPostgresNotification(ctx, conn.Conn(), notification)
}
