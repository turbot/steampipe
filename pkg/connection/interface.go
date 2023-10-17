package connection

import (
	"context"
	"database/sql"

	"github.com/turbot/steampipe/pkg/error_helpers"
	"github.com/turbot/steampipe/pkg/pluginmanager_service/grpc/shared"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
)

type pluginManager interface {
	shared.PluginManager
	OnConnectionConfigChanged(context.Context, ConnectionConfigMap, map[string]*modconfig.Plugin)
	GetConnectionConfig() ConnectionConfigMap
	HandlePluginLimiterChanges(PluginLimiterMap) error
	Pool() *sql.DB
	ShouldFetchRateLimiterDefs() bool
	LoadPluginRateLimiters(map[string]string) (PluginLimiterMap, error)
	SendPostgresSchemaNotification(context.Context) error
	SendPostgresErrorsAndWarningsNotification(context.Context, *error_helpers.ErrorAndWarnings)
}
