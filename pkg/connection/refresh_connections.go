package connection

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/v2/pkg/steampipeconfig"
)

// only allow one execution of refresh connections
var executeLock sync.Mutex

// only allow one queued execution
var queueLock sync.Mutex

func RefreshConnections(ctx context.Context, pluginManager pluginManager, forceUpdateConnectionNames ...string) (res *steampipeconfig.RefreshConnectionResult) {
	log.Println("[INFO] RefreshConnections start")
	defer log.Println("[INFO] RefreshConnections end")

	// TODO KAI if we, for example, access a nil map, this does not seem to catch it and startup hangs
	defer func() {
		if r := recover(); r != nil {
			res = steampipeconfig.NewErrorRefreshConnectionResult(helpers.ToError(r))
		}
	}()

	t := time.Now()
	defer log.Printf("[INFO] refreshConnections completion time (%fs)", time.Since(t).Seconds())

	// first grab the queue lock
	if !queueLock.TryLock() {
		// someone has it - they will execute so we have nothing to do
		log.Printf("[INFO] another execution is already queued - returning")
		return &steampipeconfig.RefreshConnectionResult{}
	}

	log.Printf("[INFO] acquired refreshQueueLock, try to acquire refreshExecuteLock")

	// so we have the queue lock, now wait on the execute lock
	executeLock.Lock()
	defer func() {
		executeLock.Unlock()
		log.Printf("[INFO] released refreshExecuteLock")
	}()

	// we have the execute-lock, release the queue-lock so someone else can queue
	queueLock.Unlock()
	log.Printf("[INFO] acquired refreshExecuteLock, released refreshQueueLock")

	// now refresh connections

	// package up all necessary data into a state object
	state, err := newRefreshConnectionState(ctx, pluginManager, forceUpdateConnectionNames)
	if err != nil {
		return steampipeconfig.NewErrorRefreshConnectionResult(err)
	}

	// now do the refresh
	state.refreshConnections(ctx)

	return state.res
}
