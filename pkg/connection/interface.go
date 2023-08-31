package connection

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

type pluginManager interface {
	OnConnectionConfigChanged(ConnectionConfigMap, LimiterMap)
	GetConnectionConfig() ConnectionConfigMap
	HandlePluginLimiterChanges(map[string]LimiterMap) error
	Pool() *pgxpool.Pool
	ShouldFetchRateLimiterDefs() bool
	GetPluginExemplarConnections() map[string]string
}
