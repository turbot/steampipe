package connection

type pluginManager interface {
	OnConnectionConfigChanged(ConnectionConfigMap, LimiterMap)
	GetConnectionConfig() ConnectionConfigMap
	HandlePluginLimiterChanges(map[string]LimiterMap) error
}
