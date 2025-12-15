package pluginmanager_service

import (
	"github.com/turbot/steampipe-plugin-sdk/v5/error_helpers"
	sdkgrpc "github.com/turbot/steampipe-plugin-sdk/v5/grpc"
	sdkproto "github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"log"
)

type PluginMessageServer struct {
	pluginManager *PluginManager
}

func NewPluginMessageServer(pluginManager *PluginManager) (*PluginMessageServer, error) {
	res := &PluginMessageServer{
		pluginManager: pluginManager,
	}
	return res, nil
}

func (m *PluginMessageServer) AddConnection(pluginClient *sdkgrpc.PluginClient, pluginName string, connectionNames ...string) error {
	log.Printf("[TRACE] PluginMessageServer AddConnection for connections %v", connectionNames)

	for _, connection := range connectionNames {
		cacheStream, err := m.openMessageStream(pluginClient, pluginName, connection)
		if err != nil {
			return err
		}

		// if no cache stream was returned, this plugin cannot support cache streams
		if cacheStream == nil {
			return nil
		}
		go m.runMessageListener(cacheStream, connection)
	}
	return nil
}

func (m *PluginMessageServer) openMessageStream(pluginClient *sdkgrpc.PluginClient, pluginName, connection string) (sdkproto.WrapperPlugin_EstablishMessageStreamClient, error) {
	log.Printf("[TRACE] openMessageStream for connection '%s'", connection)

	// does this plugin support streaming cache
	supportedOperations, err := pluginClient.GetSupportedOperations()
	if err != nil {
		return nil, err
	}
	if !supportedOperations.MessageStream {
		log.Printf("[WARN] plugin '%s' does not support message stream", pluginName)
		return nil, nil
	}

	log.Printf("[TRACE] calling EstablishMessageStream")

	stream, err := pluginClient.EstablishMessageStream()
	return stream, err

}

func (m *PluginMessageServer) runMessageListener(stream sdkproto.WrapperPlugin_EstablishMessageStreamClient, connection string) {
	defer stream.CloseSend()

	log.Printf("[TRACE] runMessageListener connection '%s'", connection)
	for {
		message, err := stream.Recv()
		if err != nil {
			m.logReceiveError(err, connection)
			return
		}
		m.handleMessage(stream, message, connection)
	}
}

func (m *PluginMessageServer) logReceiveError(err error, connection string) {
	if err == nil {
		return
	}
	log.Printf("[TRACE] receive error for connection '%s': %v", connection, err)
	switch {
	case sdkgrpc.IsEOFError(err):
		log.Printf("[TRACE] cache listener received EOF for connection '%s', returning", connection)
	case sdkgrpc.IsNotImplementedError(err):
		// should not be possible
		log.Printf("[TRACE] connection '%s' does not support centralised cache", connection)
	case error_helpers.IsContextCancelledError(err):
		// ignore
	default:
		log.Printf("[WARN] error in PluginMessageServer runMessageListener for connection '%s': %v", connection, err)
	}
}

func (m *PluginMessageServer) handleMessage(stream sdkproto.WrapperPlugin_EstablishMessageStreamClient, message *sdkproto.PluginMessage, connection string) {
	ctx := stream.Context()

	switch message.MessageType {
	case sdkproto.PluginMessageType_SCHEMA_UPDATED:
		log.Printf("[INFO] PluginMessageServer.handleMessage: PluginMessageType_SCHEMA_UPDATED for connection: %s", message.Connection)
		m.pluginManager.updateConnectionSchema(ctx, message.Connection)
	}
}
