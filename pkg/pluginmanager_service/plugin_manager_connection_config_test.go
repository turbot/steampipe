package pluginmanager_service

import (
	"context"
	"testing"

	goplugin "github.com/hashicorp/go-plugin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/turbot/pipe-fittings/v2/ociinstaller"
	sdkproto "github.com/turbot/steampipe-plugin-sdk/v6/grpc/proto"
	"github.com/turbot/steampipe/v2/pkg/connection"
	pb "github.com/turbot/steampipe/v2/pkg/pluginmanager_service/grpc/proto"
)

// a plugin which is still starting has a placeholder runningPlugin with a nil client and reattach config.
// connection config and rate limiter changes arriving in that window must not panic, and are not sent
// to the plugin while it is starting (startPlugin sends the changes made since it started, once it has started)

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

func TestSetRateLimitersForPlugin_PluginStarting(t *testing.T) {
	// key the starting plugin the same way setRateLimitersForPlugin looks it up
	p := ociinstaller.NewImageRef("aws").DisplayImageRef()
	pm := newStartingPluginManager(t, p)

	var err error
	require.NotPanics(t, func() { err = pm.setRateLimitersForPlugin("aws") })
	assert.NoError(t, err)
	assert.True(t, pm.runningPluginMap[p].limitersChangedWhileStarting, "the limiter change must be sent once the plugin has started")
}

// a plugin started with startedWith whose connection config changes while it is starting must be sent the
// difference once it has started
func TestChangesMadeWhileStarting_AreSentOnceStarted(t *testing.T) {
	const p = "hub.steampipe.io/plugins/turbot/aws@latest"
	pm := newStartingPluginManager(t, p)

	keep := newTestConnectionConfig(p, p, "aws_keep")
	changed := newTestConnectionConfig(p, p, "aws_changed")
	deleted := newTestConnectionConfig(p, p, "aws_deleted")
	startedWith := []*sdkproto.ConnectionConfig{keep, changed, deleted}
	pm.connectionConfigMap = connection.ConnectionConfigMap{"aws_keep": keep, "aws_changed": changed, "aws_deleted": deleted}
	pm.populatePluginConnectionConfigs()

	changedAfter := newTestConnectionConfig(p, p, "aws_changed")
	changedAfter.Config = "region = \"eu-west-1\""
	added := newTestConnectionConfig(p, p, "aws_added")
	newMap := connection.ConnectionConfigMap{"aws_keep": keep, "aws_changed": changedAfter, "aws_added": added}

	// the change arrives while the plugin is starting
	require.NoError(t, pm.handleConnectionConfigChanges(context.Background(), newMap))

	// the plugin finishes starting
	running := pm.runningPluginMap[p]
	running.client = &goplugin.Client{}
	running.reattach = &pb.ReattachConfig{Connections: []string{"aws_keep", "aws_changed", "aws_deleted"}}

	requestMap := pm.requestsForChangesMadeWhileStarting(p, startedWith)

	require.Contains(t, requestMap, p)
	req := requestMap[p]
	require.Len(t, req.Added, 1)
	assert.Equal(t, "aws_added", req.Added[0].Connection)
	require.Len(t, req.Deleted, 1)
	assert.Equal(t, "aws_deleted", req.Deleted[0].Connection)
	require.Len(t, req.Changed, 1)
	assert.Equal(t, "aws_changed", req.Changed[0].Connection)
	assert.ElementsMatch(t, []string{"aws_keep", "aws_changed", "aws_added"}, running.reattach.Connections)
}

func TestChangesMadeWhileStarting_NoChange(t *testing.T) {
	const p = "hub.steampipe.io/plugins/turbot/aws@latest"
	pm := newStartingPluginManager(t, p)
	keep := newTestConnectionConfig(p, p, "aws_keep")
	pm.connectionConfigMap = connection.ConnectionConfigMap{"aws_keep": keep}
	pm.populatePluginConnectionConfigs()

	requestMap := pm.requestsForChangesMadeWhileStarting(p, []*sdkproto.ConnectionConfig{keep})
	assert.Empty(t, requestMap)
}
