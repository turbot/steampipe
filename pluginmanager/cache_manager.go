package pluginmanager

import (
	"io"
	"log"

	"github.com/hashicorp/go-plugin"
	sdkgrpc "github.com/turbot/steampipe-plugin-sdk/v3/grpc"
	"github.com/turbot/steampipe-plugin-sdk/v3/grpc/proto"
)

type CacheManager struct {
	pluginManager *PluginManager
}

func NewCacheManager(pluginManager *PluginManager) *CacheManager {
	return &CacheManager{
		pluginManager: pluginManager,
	}
}

func (m CacheManager) AddConnection(client *plugin.Client, connection string) error {
	cacheStream, err := m.openCacheStream(client, connection)
	if err != nil {
		return err
	}
	// if no cache stream was returned, this plugin cannot support cache streams
	if cacheStream == nil {
		return nil
	}
	// todo - heartbeat for these connections?
	go m.runCacheListener(cacheStream, connection)
	return nil
}

func (m *CacheManager) openCacheStream(rawClient *plugin.Client, connection string) (proto.WrapperPlugin_EstablishCacheConnectionClient, error) {
	log.Printf("[TRACE] openCacheStream for connection '%s'", connection)

	plugin := m.pluginManager.connectionConfig[connection].Plugin
	client, err := sdkgrpc.NewPluginClient(rawClient, plugin)
	if err != nil {
		return nil, err
	}

	// does this plugin support streaming cache
	supportedOperations, err := client.GetSupportedOperations()
	if err != nil {
		return nil, err
	}
	if !supportedOperations.CacheStream {
		log.Printf("[TRACE] plugin '%s' does not support streamed cache", m.pluginManager.connectionConfig[connection].Plugin)
		return nil, nil
	}
	cacheStream, err := client.EstablishCacheConnection()
	return cacheStream, nil
}

func (m *CacheManager) runCacheListener(stream proto.WrapperPlugin_EstablishCacheConnectionClient, connection string) {
	defer stream.CloseSend()

	log.Printf("[WARN] runCacheListener connection '%s'", connection)
	for {
		request, err := stream.Recv()
		if err != nil {
			m.logReceiveError(err, connection)
			// signal error and reestablish connection?
			return
		}
		log.Printf("[WARN] runCacheListener got request for connection '%s': %v", connection, request)
		if result, err := m.handleCacheRequest(request); err != nil {
			// TODO WHAT TO DO?
			log.Printf("[ERROR] error handling cache request for connection '%s': %v", connection, err)
		} else {
			if err := stream.Send(result); err != nil {
				// TODO WHAT TO DO?
				log.Printf("[ERROR] error sending cache result for connection '%s': %v", connection, err)

			}
		}
	}
}

func (m *CacheManager) logReceiveError(err error, connection string) {
	log.Printf("[TRACE] receive error for connection '%s': %v", connection, err)

	switch {
	case err == io.EOF:
		log.Printf("[TRACE] cache listener received EOF for connection '%s', returning", connection)
	case sdkgrpc.IsNotImplementedError(err):
		log.Printf("[TRACE] connection '%s' does not support centralised cache", connection)
	default:
		log.Printf("[ERROR] error in runCacheListener for connection '%s': %v", connection, err)
	}
}

func (m CacheManager) handleCacheRequest(request *proto.CacheRequest) (*proto.CacheResult, error) {
	log.Printf("[WARN] HANDLE CACHE REQUEST %v", request)
	return &proto.CacheResult{CacheHit: false}, nil
}
