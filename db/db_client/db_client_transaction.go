package db_client

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/turbot/steampipe-plugin-sdk/instrument"
	"github.com/turbot/steampipe/utils"
)

// createTransaction , with a timeout - this may be required if the db client becomes unresponsive
func (c *DbClient) createTransaction(ctx context.Context, session *sql.Conn, retryOnTimeout bool) (tx *sql.Tx, err error) {
	traceCtx, span := instrument.StartSpan(ctx, "DbClient.createTransaction")
	defer span.End()

	doneChan := make(chan bool)
	go func() {
		tx, err = session.BeginTx(traceCtx, nil)
		if err != nil {
			err = utils.PrefixError(err, "error creating transaction")
		}
		close(doneChan)
	}()

	select {
	case <-doneChan:
	case <-time.After(time.Second * 5):
		log.Printf("[TRACE] timed out creating a transaction")
		if retryOnTimeout {
			log.Printf("[TRACE] refresh the client and retry")
			// refresh the db client to try to fix the issue
			c.refreshDbClient(ctx)

			// recurse back into this function, passing 'retryOnTimeout=false' to prevent further recursion
			return c.createTransaction(traceCtx, session, false)
		}
		err = fmt.Errorf("error creating transaction - please restart Steampipe")
	}
	return
}
