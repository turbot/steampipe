package pluginmanager

import (
	"context"
	"fmt"
	"github.com/allegro/bigcache/v3"
	"github.com/eko/gocache/v3/cache"
	"github.com/eko/gocache/v3/store"
	"github.com/hashicorp/go-plugin"
	sdkgrpc "github.com/turbot/steampipe-plugin-sdk/v3/grpc"
	sdkproto "github.com/turbot/steampipe-plugin-sdk/v3/grpc/proto"
	"google.golang.org/protobuf/proto"
	"io"
	"log"
	"sync"
	"time"
)

type CacheData interface {
	proto.Message
	*sdkproto.QueryResult | *sdkproto.IndexBucket
}

type CacheServer struct {
	pluginManager *PluginManager
	cache         *cache.Cache[[]byte]
	// map of ongoing request
	setRequests map[string]*sdkproto.CacheRequest
	setLock     sync.Mutex
}

func NewCacheServer(maxCacheStorageMb int, pluginManager *PluginManager) (*CacheServer, error) {
	cacheStore, err := createCacheStore(maxCacheStorageMb)
	if err != nil {
		return nil, err
	}
	res := &CacheServer{
		pluginManager: pluginManager,
		cache:         cache.New[[]byte](cacheStore),
		setRequests:   make(map[string]*sdkproto.CacheRequest),
	}
	return res, nil
}

func createCacheStore(maxCacheStorageMb int) (store.StoreInterface, error) {
	//ristrettoCache, err := ristretto.NewCache(&ristretto.Config{
	//	NumCounters: 1000,
	//	MaxCost:     100000,
	//	BufferItems: 64,
	//})
	//if err != nil {
	//	return nil, err
	//}
	//ristrettoStore := store.NewRistretto(ristrettoCache)
	//return ristrettoStore, nil
	//
	//gocacheClient := gocache.New(5*time.Minute, 10*time.Minute)
	//return store.NewGoCache(gocacheClient), nil

	config := bigcache.DefaultConfig(5 * time.Minute)
	//config.HardMaxCacheSize = maxCacheStorageMb
	//config.Shards = 10

	// max entry size is HardMaxCacheSize/1000
	//config.MaxEntrySize = (maxCacheStorageMb) * 1024 * 1024

	bigcacheClient, _ := bigcache.NewBigCache(config)
	bigcacheStore := store.NewBigcache(bigcacheClient)

	return bigcacheStore, nil
}

func (m *CacheServer) AddConnection(client *plugin.Client, connection string) error {
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

func (m *CacheServer) openCacheStream(rawClient *plugin.Client, connection string) (sdkproto.WrapperPlugin_EstablishCacheConnectionClient, error) {
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

func (m *CacheServer) runCacheListener(stream sdkproto.WrapperPlugin_EstablishCacheConnectionClient, connection string) {
	defer stream.CloseSend()

	log.Printf("[WARN] runCacheListener connection '%s'", connection)
	for {
		request, err := stream.Recv()
		if err != nil {
			m.logReceiveError(err, connection)
			// signal error and reestablish connection?
			//return
			continue
		}
		m.handleCacheRequest(stream, request, connection)
		log.Printf("[WARN] handleCacheRequest DONE")
		if request.CallId == "" {
			log.Printf("[ERROR] no callId provided")
			continue
		}
		log.Printf("[WARN] runCacheListener got request for connection '%s': %s, callId: '%s'", connection, request.Command, request.CallId)

	}
}

func (m *CacheServer) logReceiveError(err error, connection string) {
	log.Printf("[TRACE] receive error for connection '%s': %v", connection, err)

	switch {
	case err == io.EOF:
		log.Printf("[TRACE] cache listener received EOF for connection '%s', returning", connection)
	case sdkgrpc.IsNotImplementedError(err):
		// should not be possible
		log.Printf("[TRACE] connection '%s' does not support centralised cache", connection)
	default:
		log.Printf("[ERROR] error in runCacheListener for connection '%s': %v", connection, err)
	}
}

func (m *CacheServer) handleCacheRequest(stream sdkproto.WrapperPlugin_EstablishCacheConnectionClient, request *sdkproto.CacheRequest, connection string) {
	ctx := stream.Context()

	switch request.Command {
	case sdkproto.CacheCommand_GET_RESULT:
		log.Printf("[WARN] GET RESULT")
		res := doGet[*sdkproto.QueryResult](ctx, request.Key, m.cache)
		log.Printf("[WARN] streamQueryResults")
		// stream 'get' results a row at a time
		m.streamQueryResults(stream, res, connection, request.CallId)
		log.Printf("[WARN] GET RESULT DONE")
		return

	case sdkproto.CacheCommand_SET_RESULT_START:
		log.Printf("[WARN] handleCacheRequest: CacheCommand_SET_RESULT_START")
		m.startSet(ctx, request)

	case sdkproto.CacheCommand_SET_RESULT_ITERATE:
		log.Printf("[WARN] handleCacheRequest: CacheCommand_SET_RESULT_ITERATE")
		m.iterateSet(ctx, request)

	case sdkproto.CacheCommand_SET_RESULT_END:
		log.Printf("[WARN] handleCacheRequest: CacheCommand_SET_RESULT_ITERATE")
		res := m.endSet(ctx, request)

		m.streamResponse(stream, res, connection, request.CallId)

	case sdkproto.CacheCommand_DELETE_RESULT, sdkproto.CacheCommand_DELETE_INDEX:
		log.Printf("[WARN] handleCacheRequest: CacheCommand_DELETE_RESULT")
		res := doDelete(ctx, request.Key, m.cache)
		m.streamResponse(stream, res, connection, request.CallId)

	case sdkproto.CacheCommand_GET_INDEX:
		log.Printf("[WARN] handleCacheRequest: CacheCommand_GET_INDEX")
		res := doGet[*sdkproto.IndexBucket](ctx, request.Key, m.cache)
		m.streamResponse(stream, res, connection, request.CallId)

	case sdkproto.CacheCommand_SET_INDEX:
		log.Printf("[WARN] handleCacheRequest: CacheCommand_SET_INDEX")
		res := doSet(ctx, request.Key, request.IndexBucket, request.Ttl, m.cache)
		m.streamResponse(stream, res, connection, request.CallId)
	}
}

func doGet[T CacheData](ctx context.Context, key string, cache *cache.Cache[[]byte]) *sdkproto.CacheResponse {
	log.Printf("[WARN] doGet key %s", key)

	// get the bytes from the cache
	getRes, err := cache.Get(ctx, key)
	if err != nil {
		log.Printf("[WARN] cache.Get returned error %s", err.Error())
		if isCacheMiss(err) {
			log.Printf("[WARN] cache miss - return empty response")
			// return response with success false
			return &sdkproto.CacheResponse{}
		}
		// otherwise just return the error
		return &sdkproto.CacheResponse{Error: err.Error()}
	}

	res := &sdkproto.CacheResponse{Success: true}

	// unmarshall into the correct type
	var t T
	if _, ok := any(t).(*sdkproto.QueryResult); ok {
		target := &sdkproto.QueryResult{}
		err = proto.Unmarshal(getRes, target)
		res.QueryResult = target
	} else if _, ok := any(t).(*sdkproto.IndexBucket); ok {
		target := &sdkproto.IndexBucket{}
		err = proto.Unmarshal(getRes, target)
		res.IndexBucket = target
	}
	if err != nil {
		log.Printf("[WARN] error unmarshalling result: %s", err.Error())
		return &sdkproto.CacheResponse{Error: err.Error()}
	}

	return res
}

func isCacheMiss(err error) bool {
	// NOTE: this is the error returned from BigCache
	return err.Error() == "Entry not found"
}

func (m *CacheServer) startSet(_ context.Context, req *sdkproto.CacheRequest) {
	// add entry into map
	m.setLock.Lock()
	m.setRequests[req.CallId] = req
	m.setLock.Unlock()

	return
}

func (m *CacheServer) iterateSet(_ context.Context, req *sdkproto.CacheRequest) {
	// find the entry for the in-progress et operation
	m.setLock.Lock()
	defer m.setLock.Unlock()
	inProgress, ok := m.setRequests[req.CallId]
	if !ok {
		log.Printf("[WARN] iterateSet could not find in-progress Set operation for call id '%s'", req.CallId)
	}

	if req.Result == nil {
		log.Printf("[WARN] iterateSet called with nil result")
	}

	inProgress.Result.Rows = append(inProgress.Result.Rows, req.Result.Rows...)
}

func (m *CacheServer) endSet(ctx context.Context, req *sdkproto.CacheRequest) *sdkproto.CacheResponse {
	// find the entry for the in-progress et operation
	m.setLock.Lock()
	defer m.setLock.Unlock()
	inProgress, ok := m.setRequests[req.CallId]
	if !ok {
		return &sdkproto.CacheResponse{
			Error: fmt.Sprintf("endSet could not find in-progress set operastion for call id '%s'", req.CallId),
		}
	}
	if req.Result != nil {
		// no result should be passed with end set
		return &sdkproto.CacheResponse{Error: "endSet called with non-nil result"}
	}
	// remove from in progress map
	delete(m.setRequests, req.CallId)

	// now do the actual set
	return doSet(ctx, inProgress.Key, inProgress.Result, inProgress.Ttl, m.cache)
}

func (m *CacheServer) streamQueryResults(stream sdkproto.WrapperPlugin_EstablishCacheConnectionClient, res *sdkproto.CacheResponse, connection string, callId string) {
	log.Printf("[WARN] streamQueryResults callId %s", callId)

	// stream, a row at a time
	rowResult := &sdkproto.CacheResponse{
		Success:     res.Success,
		QueryResult: &sdkproto.QueryResult{},
		Error:       res.Error,
	}
	if res.QueryResult != nil {
		for _, row := range res.QueryResult.Rows {
			// TODO chunk into N rows per row
			rowResult.QueryResult.Rows = []*sdkproto.Row{row}
			//log.Printf("[WARN] rowResult: %v", rowResult)
			m.streamResponse(stream, rowResult, connection, callId)
		}
		rowResult.QueryResult.Rows = nil
	}
	// now stream empty row to indicate end of data
	m.streamResponse(stream, rowResult, connection, callId)
}

// attempt to stream a result
func (m *CacheServer) streamResponse(stream sdkproto.WrapperPlugin_EstablishCacheConnectionClient, response *sdkproto.CacheResponse, connection string, callId string) {
	response.CallId = callId
	log.Printf("[WARN] streamResponse, call id: %s", callId)
	if err := stream.Send(response); err != nil {
		// TODO WHAT TO DO?
		log.Printf("[ERROR] error sending cache result for connection '%s': %v", connection, err)
	}
}

func doSet[T CacheData](ctx context.Context, key string, value T, ttl int64, cache *cache.Cache[[]byte]) *sdkproto.CacheResponse {
	res := &sdkproto.CacheResponse{
		Success: true,
	}

	bytes, err := proto.Marshal(value)
	if err != nil {
		log.Printf("[WARN] marshal failed")
		return &sdkproto.CacheResponse{Error: err.Error()}
	}

	expiration := time.Duration(ttl) * time.Second
	err = cache.Set(ctx,
		key,
		bytes,
		store.WithExpiration(expiration),
	)
	if err != nil {
		log.Printf("[WARN] startSet failed: %v", err)
		res.Success = false
		res.Error = err.Error()
	}
	return res
}

func doDelete(ctx context.Context, key string, cache *cache.Cache[[]byte]) *sdkproto.CacheResponse {
	res := &sdkproto.CacheResponse{
		Success: true,
	}
	err := cache.Delete(ctx, key)
	if err != nil {
		res.Success = false
		res.Error = err.Error()
	}
	return res
}
