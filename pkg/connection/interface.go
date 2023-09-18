package connection

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/turbot/steampipe/pkg/pluginmanager_service/grpc/shared"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
)

type pluginManager interface {
	shared.PluginManager
	OnConnectionConfigChanged(ConnectionConfigMap, map[string]*modconfig.Plugin)
	GetConnectionConfig() ConnectionConfigMap
	HandlePluginLimiterChanges(limiterMap PluginLimiterMap) error
	Pool() *pgxpool.Pool
	ShouldFetchRateLimiterDefs() bool
	LoadPluginRateLimiters(pluginConnectionMap map[string]string) (PluginLimiterMap, error)
}
