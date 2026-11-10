package pluginmanager_service

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/hashicorp/go-hclog"
	goplugin "github.com/hashicorp/go-plugin"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	sdkproto "github.com/turbot/steampipe-plugin-sdk/v6/grpc/proto"
	sdkshared "github.com/turbot/steampipe-plugin-sdk/v6/grpc/shared"
	sdkplugin "github.com/turbot/steampipe-plugin-sdk/v6/plugin"
	"github.com/turbot/steampipe/v2/pkg/connection"
	"github.com/turbot/steampipe/v2/pkg/options"
	pb "github.com/turbot/steampipe/v2/pkg/pluginmanager_service/grpc/proto"
)

// The query cache size limit is one total for all plugins. Each plugin gets a share of the
// total, in proportion to its number of connections (aggregators are not counted).
//
// These tests start fake plugins in this process and initialize them with the real
// PluginManager.initializePlugin, so they check the cache size each plugin receives.

// fakePlugin is a plugin server that records the requests that set its cache size.
// It implements only the calls that the plugin manager makes when it initializes a plugin.
type fakePlugin struct {
	sdkshared.WrapperPluginServer
	supportsSetCacheOptions bool

	mut                  sync.Mutex
	allConnectionConfigs *sdkproto.SetAllConnectionConfigsRequest
	cacheOptions         *sdkproto.SetCacheOptionsRequest
}

func (p *fakePlugin) GetSupportedOperations(*sdkproto.GetSupportedOperationsRequest) (*sdkproto.GetSupportedOperationsResponse, error) {
	return &sdkproto.GetSupportedOperationsResponse{
		MultipleConnections: true,
		SetCacheOptions:     p.supportsSetCacheOptions,
	}, nil
}

func (p *fakePlugin) SetAllConnectionConfigs(req *sdkproto.SetAllConnectionConfigsRequest) (*sdkproto.SetConnectionConfigResponse, error) {
	p.mut.Lock()
	defer p.mut.Unlock()
	p.allConnectionConfigs = req
	return &sdkproto.SetConnectionConfigResponse{}, nil
}

func (p *fakePlugin) SetCacheOptions(req *sdkproto.SetCacheOptionsRequest) (*sdkproto.SetCacheOptionsResponse, error) {
	p.mut.Lock()
	defer p.mut.Unlock()
	p.cacheOptions = req
	return &sdkproto.SetCacheOptionsResponse{}, nil
}

func (p *fakePlugin) GetSchema(*sdkproto.GetSchemaRequest) (*sdkproto.GetSchemaResponse, error) {
	return &sdkproto.GetSchemaResponse{Schema: &sdkproto.Schema{Mode: sdkplugin.SchemaModeStatic}}, nil
}

// received returns the requests the plugin received
func (p *fakePlugin) received() (*sdkproto.SetAllConnectionConfigsRequest, *sdkproto.SetCacheOptionsRequest) {
	p.mut.Lock()
	defer p.mut.Unlock()
	return p.allConnectionConfigs, p.cacheOptions
}

// cacheSizeMb returns the cache size the plugin received: from SetCacheOptions if the plugin
// supports it, otherwise from SetAllConnectionConfigs
func (p *fakePlugin) cacheSizeMb(t *testing.T) int64 {
	t.Helper()
	allConnectionConfigs, cacheOptions := p.received()
	require.NotNil(t, allConnectionConfigs, "plugin did not receive SetAllConnectionConfigs")
	if !p.supportsSetCacheOptions {
		require.Nil(t, cacheOptions, "plugin without SetCacheOptions support received SetCacheOptions")
		return allConnectionConfigs.MaxCacheSizeMb
	}
	// the plugin manager sends -1 in SetAllConnectionConfigs, so that the plugin waits for SetCacheOptions
	require.Equal(t, int64(-1), allConnectionConfigs.MaxCacheSizeMb)
	require.NotNil(t, cacheOptions, "plugin did not receive SetCacheOptions")
	return cacheOptions.MaxSizeMb
}

// serveFakePlugin serves impl in this process and returns a started client for it,
// the way startPluginProcess returns a started client for a plugin process
func serveFakePlugin(t *testing.T, pluginName string, impl *fakePlugin) *goplugin.Client {
	t.Helper()

	ctx, cancel := context.WithCancel(context.Background())
	reattachCh := make(chan *goplugin.ReattachConfig, 1)
	closeCh := make(chan struct{})
	go goplugin.Serve(&goplugin.ServeConfig{
		HandshakeConfig: sdkshared.Handshake,
		Plugins:         map[string]goplugin.Plugin{pluginName: &sdkshared.WrapperPlugin{Impl: impl}},
		GRPCServer:      goplugin.DefaultGRPCServer,
		Logger:          hclog.NewNullLogger(),
		Test:            &goplugin.ServeTestConfig{Context: ctx, ReattachConfigCh: reattachCh, CloseCh: closeCh},
	})

	var reattach *goplugin.ReattachConfig
	select {
	case reattach = <-reattachCh:
	case <-time.After(10 * time.Second):
		cancel()
		t.Fatal("fake plugin did not start")
	}

	client := goplugin.NewClient(&goplugin.ClientConfig{
		HandshakeConfig:  sdkshared.Handshake,
		Plugins:          map[string]goplugin.Plugin{pluginName: &sdkshared.WrapperPlugin{}},
		Reattach:         reattach,
		AllowedProtocols: []goplugin.Protocol{goplugin.ProtocolGRPC},
		Logger:           hclog.NewNullLogger(),
	})
	t.Cleanup(func() {
		client.Kill()
		cancel()
		<-closeCh
	})
	_, err := client.Start()
	require.NoError(t, err)
	return client
}

// startFakePlugin starts a fake plugin for a plugin instance and initializes it with all the
// connections the plugin manager has for that instance, as startPlugin does
func startFakePlugin(t *testing.T, pm *PluginManager, pluginInstance string, supportsSetCacheOptions bool) *fakePlugin {
	t.Helper()
	connectionConfigs := pm.pluginConnectionConfigMap[pluginInstance]
	require.NotEmpty(t, connectionConfigs, "no connections for plugin instance %s", pluginInstance)

	impl := &fakePlugin{supportsSetCacheOptions: supportsSetCacheOptions}
	client := serveFakePlugin(t, connectionConfigs[0].Plugin, impl)
	_, err := pm.initializePlugin(connectionConfigs, client, &pb.GetRequest{})
	require.NoError(t, err)
	return impl
}

// setDatabaseOptions sets viper from options "database" { cache = true, cache_max_ttl = 300, cache_max_size_mb = ... }
// in the same way as the config loader. A nil cacheMaxSizeMb leaves cache_max_size_mb unset.
func setDatabaseOptions(t *testing.T, cacheMaxSizeMb *int) {
	t.Helper()
	viper.Reset()
	t.Cleanup(viper.Reset)

	cacheEnabled := true
	cacheMaxTtl := 300
	databaseOptions := &options.Database{Cache: &cacheEnabled, CacheMaxTtl: &cacheMaxTtl, CacheMaxSizeMb: cacheMaxSizeMb}
	for k, v := range databaseOptions.ConfigMap() {
		viper.SetDefault(k, v)
	}
}

// connectionCounts is the number of (non aggregator) connections for each plugin instance
type connectionCounts map[string]int

// buildConnectionConfigMap builds a connection config map with the given number of connections
// for each plugin instance, plus one aggregator connection for each plugin instance
func buildConnectionConfigMap(counts connectionCounts) connection.ConnectionConfigMap {
	configMap := make(connection.ConnectionConfigMap)
	for pluginInstance, count := range counts {
		var children []string
		for i := 1; i <= count; i++ {
			name := fmt.Sprintf("%s_%d", pluginInstance, i)
			configMap[name] = newTestConnectionConfig(pluginInstance, pluginInstance, name)
			children = append(children, name)
		}
		aggregator := newTestConnectionConfig(pluginInstance, pluginInstance, pluginInstance+"_all")
		aggregator.ChildConnections = children
		configMap[aggregator.Connection] = aggregator
	}
	return configMap
}

func newCacheTestPluginManager(t *testing.T, counts connectionCounts) *PluginManager {
	t.Helper()
	pm := newTestPluginManager(t)
	pm.connectionConfigMap = buildConnectionConfigMap(counts)
	pm.populatePluginConnectionConfigs()
	return pm
}

// cacheSizesMb starts a fake plugin for each plugin instance and returns the cache size each one received
func cacheSizesMb(t *testing.T, pm *PluginManager, supportsSetCacheOptions bool, pluginInstances ...string) map[string]int64 {
	t.Helper()
	sizes := make(map[string]int64)
	for _, pluginInstance := range pluginInstances {
		sizes[pluginInstance] = startFakePlugin(t, pm, pluginInstance, supportsSetCacheOptions).cacheSizeMb(t)
	}
	return sizes
}

// cacheSizeTestPaths are the two ways the plugin manager sends a plugin its cache size
var cacheSizeTestPaths = []struct {
	name                    string
	supportsSetCacheOptions bool
}{
	{name: "SetCacheOptions", supportsSetCacheOptions: true},
	{name: "SetAllConnectionConfigs (plugin without SetCacheOptions)", supportsSetCacheOptions: false},
}

// Worked example: a total of 900 MB, with aws 10 connections, gcp 5 and net 1 (16 in all).
// aws gets 900 x 10/16 = 562, gcp 900 x 5/16 = 281, net 900 x 1/16 = 56.
// The aggregator connections do not count.
func TestInitializePlugin_CacheSizeIsShareOfConfigTotal(t *testing.T) {
	for _, path := range cacheSizeTestPaths {
		t.Run(path.name, func(t *testing.T) {
			total := 900
			setDatabaseOptions(t, &total)
			pm := newCacheTestPluginManager(t, connectionCounts{"aws": 10, "gcp": 5, "net": 1})

			sizes := cacheSizesMb(t, pm, path.supportsSetCacheOptions, "aws", "gcp", "net")

			assert.Equal(t, map[string]int64{"aws": 562, "gcp": 281, "net": 56}, sizes)
			assert.LessOrEqual(t, sizes["aws"]+sizes["gcp"]+sizes["net"], int64(total), "the shares must not add up to more than the total")
		})
	}
}

// The other cache options are sent unchanged
func TestInitializePlugin_CacheOptionsKeepTtlAndEnabled(t *testing.T) {
	total := 900
	setDatabaseOptions(t, &total)
	pm := newCacheTestPluginManager(t, connectionCounts{"aws": 1})

	_, cacheOptions := startFakePlugin(t, pm, "aws", true).received()

	require.NotNil(t, cacheOptions)
	assert.True(t, cacheOptions.Enabled)
	assert.Equal(t, int64(300), cacheOptions.Ttl)
}

// A share is never less than 1 MB, because 0 means no limit.
// A total of 5 MB, with aws 9 connections and net 1: aws gets 5 x 9/10 = 4, net gets 5 x 1/10 = 0, so 1.
func TestInitializePlugin_CacheShareIsAtLeastOneMb(t *testing.T) {
	for _, path := range cacheSizeTestPaths {
		t.Run(path.name, func(t *testing.T) {
			total := 5
			setDatabaseOptions(t, &total)
			pm := newCacheTestPluginManager(t, connectionCounts{"aws": 9, "net": 1})

			sizes := cacheSizesMb(t, pm, path.supportsSetCacheOptions, "aws", "net")

			assert.Equal(t, map[string]int64{"aws": 4, "net": 1}, sizes)
		})
	}
}

// With no total, or a total of 0, there is no limit: every plugin gets 0
func TestInitializePlugin_NoCacheTotalMeansNoLimit(t *testing.T) {
	zero := 0
	totals := map[string]*int{"cache_max_size_mb not set": nil, "cache_max_size_mb = 0": &zero}
	for totalName, total := range totals {
		for _, path := range cacheSizeTestPaths {
			t.Run(totalName+"/"+path.name, func(t *testing.T) {
				setDatabaseOptions(t, total)
				pm := newCacheTestPluginManager(t, connectionCounts{"aws": 10, "net": 1})

				sizes := cacheSizesMb(t, pm, path.supportsSetCacheOptions, "aws", "net")

				assert.Equal(t, map[string]int64{"aws": 0, "net": 0}, sizes)
			})
		}
	}
}

// A plugin gets its share from the connections that exist when it starts.
// The plugin manager starts with aws 1 connection and net 1, and starts aws: 900 x 1/2 = 450.
// Then 2 aws connections are added (aws 3, net 1), and net starts: 900 x 1/4 = 225.
func TestInitializePlugin_CacheShareUsesConnectionsWhenPluginStarts(t *testing.T) {
	total := 900
	setDatabaseOptions(t, &total)
	pm := newCacheTestPluginManager(t, connectionCounts{"aws": 1, "net": 1})

	aws := startFakePlugin(t, pm, "aws", true)
	assert.Equal(t, int64(450), aws.cacheSizeMb(t))

	pm.OnConnectionConfigChanged(context.Background(), buildConnectionConfigMap(connectionCounts{"aws": 3, "net": 1}), pm.plugins)

	net := startFakePlugin(t, pm, "net", true)
	assert.Equal(t, int64(225), net.cacheSizeMb(t))
}
