package pluginmanager_service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	sdkproto "github.com/turbot/steampipe-plugin-sdk/v6/grpc/proto"
	"github.com/turbot/steampipe/v2/pkg/error_helpers"
	pb "github.com/turbot/steampipe/v2/pkg/pluginmanager_service/grpc/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// supportedOperationsErrorPlugin is a plugin whose GetSupportedOperations call fails with err
type supportedOperationsErrorPlugin struct {
	*fakePlugin
	err error
}

func (p *supportedOperationsErrorPlugin) GetSupportedOperations(*sdkproto.GetSupportedOperationsRequest) (*sdkproto.GetSupportedOperationsResponse, error) {
	return nil, p.err
}

func initializeSupportedOperationsErrorPlugin(t *testing.T, err error) error {
	t.Helper()
	setDatabaseOptions(t, nil)
	pm := newCacheTestPluginManager(t, connectionCounts{"aws": 1})
	connectionConfigs := pm.pluginConnectionConfigMap["aws"]
	client := serveFakePlugin(t, connectionConfigs[0].Plugin, &supportedOperationsErrorPlugin{fakePlugin: &fakePlugin{}, err: err})
	_, initErr := pm.initializePlugin(connectionConfigs, client, &pb.GetRequest{})
	return initErr
}

// A plugin that does not answer in time is not too old: the start fails with the real cause, not the advice to update
// the plugin. The plugin SDK ends the call with DeadlineExceeded when a plugin does not answer within its limit.
func TestInitializePlugin_SupportedOperationsErrorIsReported(t *testing.T) {
	for name, err := range map[string]error{
		"no answer in time": status.Error(codes.DeadlineExceeded, "context deadline exceeded"),
		"plugin error":      status.Error(codes.Internal, "the plugin failed"),
	} {
		t.Run(name, func(t *testing.T) {
			initErr := initializeSupportedOperationsErrorPlugin(t, err)
			require.Error(t, initErr)
			assert.NotContains(t, initErr.Error(), error_helpers.PluginSdkCompatibilityError)
			assert.Contains(t, initErr.Error(), status.Convert(err).Message())
		})
	}
}

// A plugin built with an SDK older than the GetSupportedOperations call answers Unimplemented: that plugin is too old.
func TestInitializePlugin_SupportedOperationsUnimplementedIsTooOld(t *testing.T) {
	initErr := initializeSupportedOperationsErrorPlugin(t, status.Error(codes.Unimplemented, "unknown method GetSupportedOperations"))
	require.Error(t, initErr)
	assert.Equal(t, error_helpers.PluginSdkCompatibilityError, initErr.Error())
}
