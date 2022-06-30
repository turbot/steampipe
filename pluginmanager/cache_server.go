package pluginmanager

import (
	"context"
	"github.com/allegro/bigcache/v3"
	"github.com/eko/gocache/v3/cache"
	"github.com/eko/gocache/v3/store"
	"github.com/hashicorp/go-plugin"
	sdkgrpc "github.com/turbot/steampipe-plugin-sdk/v3/grpc"
	sdkproto "github.com/turbot/steampipe-plugin-sdk/v3/grpc/proto"
	"google.golang.org/protobuf/proto"
	"io"
	"log"
	"time"
)

type CacheData interface {
	proto.Message
	*sdkproto.QueryResult | *sdkproto.IndexBucket
}

type CacheServer struct {
	pluginManager *PluginManager
	cache         *cache.Cache[[]byte]
}

func NewCacheServer(maxCacheStorageMb int, pluginManager *PluginManager) (*CacheServer, error) {
	cacheStore, err := createCacheStore(maxCacheStorageMb)
	if err != nil {
		return nil, err
	}
	res := &CacheServer{
		pluginManager: pluginManager,
		cache:         cache.New[[]byte](cacheStore),
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

func (m CacheServer) handleCacheRequest(ctx context.Context, request *sdkproto.CacheRequest) *sdkproto.CacheResult {
	var res *sdkproto.CacheResult
	switch request.Command {
	case sdkproto.CacheCommand_GET_RESULT:
		log.Printf("[WARN] GET RESULT")
		res = doGet[*sdkproto.QueryResult](ctx, request.Key, m.cache)

	case sdkproto.CacheCommand_SET_RESULT:
		log.Printf("[WARN] SET RESULT")
		res = doSet(ctx, request.Key, request.Result, request.Cost, request.Ttl, m.cache)

	case sdkproto.CacheCommand_DELETE_RESULT, sdkproto.CacheCommand_DELETE_INDEX:
		log.Printf("[WARN] DELETE RESULT")
		res = doDelete(ctx, request.Key, m.cache)

	case sdkproto.CacheCommand_GET_INDEX:
		log.Printf("[WARN] GET INDEX")
		res = doGet[*sdkproto.IndexBucket](ctx, request.Key, m.cache)

	case sdkproto.CacheCommand_SET_INDEX:
		log.Printf("[WARN] SET INDEX")
		res = doSet(ctx, request.Key, request.IndexBucket, request.Cost, request.Ttl, m.cache)
	}
	return res
}

func doGet[T CacheData](ctx context.Context, key string, cache *cache.Cache[[]byte]) *sdkproto.CacheResult {
	log.Printf("[WARN] doGet key %s", key)

	// get the bytes from the cache
	getRes, err := cache.Get(ctx, key)
	if err != nil {
		return &sdkproto.CacheResult{Error: err.Error()}
	}
	res := &sdkproto.CacheResult{
		Success: true,
	}

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
		return &sdkproto.CacheResult{Error: err.Error()}
	}

	return res
}

func doSet[T CacheData](ctx context.Context, key string, value T, cost int64, ttl int64, cache *cache.Cache[[]byte]) *sdkproto.CacheResult {
	res := &sdkproto.CacheResult{
		Success: true,
	}

	bytes, err := proto.Marshal(value)
	if err != nil {
		log.Printf("[WARN] marshal failed")
		return &sdkproto.CacheResult{Error: err.Error()}
	}

	expiration := time.Duration(ttl) * time.Second
	err = cache.Set(ctx,
		key,
		bytes,
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

func doDelete(ctx context.Context, key string, cache *cache.Cache[[]byte]) *sdkproto.CacheResult {
	res := &sdkproto.CacheResult{
		Success: true,
	}
	err := cache.Delete(ctx, key)
	if err != nil {
		res.Success = false
		res.Error = err.Error()
	}
	return res
}
