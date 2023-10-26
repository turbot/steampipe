package connection

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/steampipe/pkg/pluginmanager_service/grpc/shared"
)

type pluginManager interface {
	shared.PluginManager
	OnConnectionConfigChanged(context.Context, ConnectionConfigMap, map[string]*modconfig.Plugin)
	GetConnectionConfig() ConnectionConfigMap
	HandlePluginLimiterChanges(PluginLimiterMap) error

	Pool() *pgxpool.Pool
	ShouldFetchRateLimiterDefs() bool
	LoadPluginRateLimiters(map[string]string) (PluginLimiterMap, error)
	SendPostgresSchemaNotification(context.Context) error
	SendPostgresErrorsAndWarningsNotification(context.Context, *error_helpers.ErrorAndWarnings)
}
