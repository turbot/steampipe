package connection

type pluginManager interface {
	OnConnectionConfigChanged(ConnectionConfigMap, LimiterMap)
}
