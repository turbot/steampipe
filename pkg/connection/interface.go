package connection

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/turbot/pipe-fittings/v2/error_helpers"
	"github.com/turbot/pipe-fittings/v2/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe/v2/pkg/pluginmanager_service/grpc/shared"
)

type pluginManager interface {
	shared.PluginManager
	OnConnectionConfigChanged(context.Context, ConnectionConfigMap, map[string]*plugin.Plugin)
	GetConnectionConfig() ConnectionConfigMap
	HandlePluginLimiterChanges(PluginLimiterMap) error
	Pool() *pgxpool.Pool
	ShouldFetchRateLimiterDefs() bool
	LoadPluginRateLimiters(map[string]string) (PluginLimiterMap, error)
	SendPostgresSchemaNotification(context.Context) error
	SendPostgresErrorsAndWarningsNotification(context.Context, error_helpers.ErrorAndWarnings)
	UpdatePluginColumnsTable(context.Context, map[string]*proto.Schema, []string) error
}
