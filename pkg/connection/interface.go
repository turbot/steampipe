package connection

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe/pkg/error_helpers"
	"github.com/turbot/steampipe/pkg/pluginmanager_service/grpc/shared"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
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
	SendPostgresErrorsAndWarningsNotification(context.Context, error_helpers.ErrorAndWarnings)
	UpdatePluginColumnsTable(context.Context, map[string]*proto.Schema, []string) error
}
