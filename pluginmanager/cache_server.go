package pluginmanager

import (
	"context"
	"io"
	"log"
	"time"

	"github.com/eko/gocache/v3/cache"
	"github.com/eko/gocache/v3/store"
	"github.com/hashicorp/go-plugin"
	gocache "github.com/patrickmn/go-cache"
	sdkcache "github.com/turbot/steampipe-plugin-sdk/v3/cache"
	sdkgrpc "github.com/turbot/steampipe-plugin-sdk/v3/grpc"
	"github.com/turbot/steampipe-plugin-sdk/v3/grpc/proto"
)

type CacheData interface {
	*sdkcache.QueryCacheResult | *sdkcache.IndexBucket
}

type CacheServer struct {
	pluginManager *PluginManager
	indexCache    *cache.Cache[*sdkcache.IndexBucket]
	resultCache   *cache.Cache[*sdkcache.QueryCacheResult]
}

func NewCacheServer(pluginManager *PluginManager) (*CacheServer, error) {
	cacheStore, err := createCacheStore()
	if err != nil {
		return nil, err
	}
	resultCache := cache.New[*sdkcache.QueryCacheResult](cacheStore)
	indexCache := cache.New[*sdkcache.IndexBucket](cacheStore)

	res := &CacheServer{
		pluginManager: pluginManager,
		resultCache:   resultCache,
		indexCache:    indexCache,
	}
	return res, nil
}

func createCacheStore() (store.StoreInterface, error) {
	//ristrettoCache, err := ristretto.NewCache(&ristretto.Config{
	//	NumCounters: 1000,
	//	MaxCost:     100,
	//	BufferItems: 64,
	//})
	//if err != nil {
	//	return nil, err
	//}
	//ristrettoStore := store.NewRistretto(ristrettoCache)
	//return ristrettoStore, nil
	gocacheClient := gocache.New(5*time.Minute, 10*time.Minute)
	return store.NewGoCache(gocacheClient), nil

}

func (m CacheServer) AddConnection(client *plugin.Client, connection string) error {
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

func (m *CacheServer) openCacheStream(rawClient *plugin.Client, connection string) (proto.WrapperPlugin_EstablishCacheConnectionClient, error) {
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

func (m *CacheServer) runCacheListener(stream proto.WrapperPlugin_EstablishCacheConnectionClient, connection string) {
	defer stream.CloseSend()

	log.Printf("[WARN] runCacheListener connection '%s'", connection)
	for {
		request, err := stream.Recv()
		if err != nil {
			m.logReceiveError(err, connection)
			// signal error and reestablish connection?
			return
		}
		log.Printf("[WARN] runCacheListener got request for connection '%s': %s", connection, request.Command)
		result := m.handleCacheRequest(stream.Context(), request)

		if err := stream.Send(result); err != nil {
			// TODO WHAT TO DO?
			log.Printf("[ERROR] error sending cache result for connection '%s': %v", connection, err)

		}

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

func (m CacheServer) handleCacheRequest(ctx context.Context, request *proto.CacheRequest) *proto.CacheResult {

	var res *proto.CacheResult
	switch request.Command {
	case proto.CacheCommand_GET_RESULT:
		log.Printf("[WARN] GET RESULT")
		res = doGet(ctx, request.Key, m.resultCache)

	case proto.CacheCommand_SET_RESULT:
		log.Printf("[WARN] SET RESULT")
		data := sdkcache.QueryCacheResultFromProto(request.Result)
		res = doSet(ctx, request.Key, data, request.Cost, request.Ttl, m.resultCache)

	case proto.CacheCommand_DELETE_RESULT:
		log.Printf("[WARN] DELETE RESULT")
		res = doDelete(ctx, request.Key, m.resultCache)

	case proto.CacheCommand_GET_INDEX:
		log.Printf("[WARN] GET INDEX")
		res = doGet(ctx, request.Key, m.indexCache)

	case proto.CacheCommand_SET_INDEX:
		log.Printf("[WARN] SET INDEX")
		data := sdkcache.IndexBucketfromProto(request.IndexBucket)
		res = doSet(ctx, request.Key, data, request.Cost, request.Ttl, m.indexCache)

	case proto.CacheCommand_DELETE_INDEX:
		log.Printf("[WARN] DELETE INDEX")
		res = doDelete(ctx, request.Key, m.indexCache)
	}
	return res
}

func doGet[T CacheData](ctx context.Context, key string, cache *cache.Cache[T]) *proto.CacheResult {
	log.Printf("[WARN] doGet key %s", key)
	res := &proto.CacheResult{
		Success: true,
	}

	getRes, err := cache.Get(ctx, key)
	if err != nil {
		res.Success = false
		res.Error = err.Error()
		return res
	}

	if queryResult, ok := any(getRes).(*sdkcache.QueryCacheResult); ok {
		res.QueryResult = queryResult.AsProto()
	} else if indexBucket, ok := any(getRes).(*sdkcache.IndexBucket); ok {
		res.IndexBucket = indexBucket.AsProto()
	}
	return res
}
func doSet[T CacheData](ctx context.Context, key string, value T, cost int64, ttl int64, cache *cache.Cache[T]) *proto.CacheResult {
	res := &proto.CacheResult{
		Success: true,
	}

	expiration := time.Duration(ttl) * time.Second
	err := cache.Set(ctx,
		key,
		value,
		store.WithCost(cost),
		store.WithExpiration(expiration),
	)
	if err != nil {
		log.Printf("[WARN] doSet failed: %v", err)
		res.Success = false
		res.Error = err.Error()
	}
	return res
}

func doDelete[T CacheData](ctx context.Context, key string, cache *cache.Cache[T]) *proto.CacheResult {
	res := &proto.CacheResult{
		Success: true,
	}
	err := cache.Delete(ctx, key)
	if err != nil {
		res.Success = false
		res.Error = err.Error()
	}
	return res
	return res
}
