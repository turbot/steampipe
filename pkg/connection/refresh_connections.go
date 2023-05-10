package connection

import (
	"context"
	"github.com/turbot/steampipe/pkg/steampipeconfig"
	"log"
	"sync"
)

// only allow one execution of refresh connections
var executeLock sync.Mutex

// only allow one queued execution
var queueLock sync.Mutex

func RefreshConnections(ctx context.Context, forceUpdateConnectionNames ...string) *steampipeconfig.RefreshConnectionResult {
	log.Printf("[TRACE] Refreshing connections")

	//time.Sleep(10 * time.Second)
	// first grab the queue lock
	if !queueLock.TryLock() {
		// someone has it - they will execute so we have nothing to do
		log.Printf("[INFO] RefreshConnections - another execution is already queued - returning")
		return &steampipeconfig.RefreshConnectionResult{}
	}

	log.Printf("[INFO] RefreshConnections acquired refreshQueueLock, try to acquire refreshExecuteLock")

	// so we have the queue lock, now wait on the execute lock
	executeLock.Lock()
	defer func() {
		executeLock.Unlock()
		log.Printf("[INFO] RefreshConnections  released refreshExecuteLock")
	}()

	// we have the execute-lock, release the queue-lock so someone else can queue
	queueLock.Unlock()
	log.Printf("[INFO] RefreshConnections acquired refreshExecuteLock, released refreshQueueLock")

	// now refresh connections
	// package up all necessary data into a state object6
	state, err := newRefreshConnectionState(ctx, forceUpdateConnectionNames)
	if err != nil {
		return steampipeconfig.NewErrorRefreshConnectionResult(err)
	}
	defer state.close()

	// now do the refresh
	state.refreshConnections(ctx)

	return state.res
}
