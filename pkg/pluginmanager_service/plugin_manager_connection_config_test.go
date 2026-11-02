package pluginmanager_service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	sdkproto "github.com/turbot/steampipe-plugin-sdk/v6/grpc/proto"
)

// a plugin which is still starting has a placeholder runningPlugin with a nil client and reattach config.
// connection config changes arriving in that window must not panic, and must not be sent to the plugin.
// (RefreshConnections applies them once the plugin has started)

func newStartingPluginManager(t *testing.T, pluginInstance string) *PluginManager {
	t.Helper()
	pm := newTestPluginManager(t)
	pm.runningPluginMap[pluginInstance] = &runningPlugin{
		pluginInstance: pluginInstance,
		imageRef:       pluginInstance,
		initialized:    make(chan struct{}),
		failed:         make(chan struct{}),
	}
	return pm
}

func TestHandleAddedConnections_PluginStarting(t *testing.T) {
	const p = "hub.steampipe.io/plugins/turbot/aws@latest"
	pm := newStartingPluginManager(t, p)
	requestMap := map[string]*sdkproto.UpdateConnectionConfigsRequest{}
	added := map[string][]*sdkproto.ConnectionConfig{p: {newTestConnectionConfig(p, p, "aws_agg")}}

	require.NotPanics(t, func() { pm.handleAddedConnections(added, requestMap) })
	assert.Empty(t, requestMap, "a starting plugin should not be sent an update")
}

func TestHandleDeletedConnections_PluginStarting(t *testing.T) {
	const p = "hub.steampipe.io/plugins/turbot/aws@latest"
	pm := newStartingPluginManager(t, p)
	requestMap := map[string]*sdkproto.UpdateConnectionConfigsRequest{}
	deleted := map[string][]*sdkproto.ConnectionConfig{p: {newTestConnectionConfig(p, p, "aws_agg")}}

	require.NotPanics(t, func() { pm.handleDeletedConnections(deleted, requestMap) })
	assert.Empty(t, requestMap, "a starting plugin should not be sent an update")
}

func TestSendUpdateConnectionConfigs_PluginStarting(t *testing.T) {
	const p = "hub.steampipe.io/plugins/turbot/aws@latest"
	pm := newStartingPluginManager(t, p)
	requestMap := map[string]*sdkproto.UpdateConnectionConfigsRequest{
		p: {Added: []*sdkproto.ConnectionConfig{newTestConnectionConfig(p, p, "aws_agg")}},
	}

	var err error
	require.NotPanics(t, func() { err = pm.sendUpdateConnectionConfigs(requestMap) })
	assert.NoError(t, err)
}
