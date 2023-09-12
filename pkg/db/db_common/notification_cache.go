package db_common

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/turbot/steampipe/pkg/error_helpers"
	"log"
	"sync"
)

type NotificationCache struct {
	notifications  []*pgconn.Notification
	conn           *pgx.Conn
	doneChan       chan struct{}
	onNotification func(*pgconn.Notification)
	mut            sync.Mutex
}

func NewNotificationCache(ctx context.Context, connection *pgx.Conn) *NotificationCache {
	res := &NotificationCache{conn: connection,
		doneChan: make(chan struct{}),
	}

	res.listenToPgNotifications(ctx)

	return res
}

func (c *NotificationCache) Stop() {
	if c.doneChan != nil {
		close(c.doneChan)
		c.doneChan = nil
	}
}
func (c *NotificationCache) RegisterListener(onNotification func(*pgconn.Notification)) {
	c.mut.Lock()
	defer c.mut.Unlock()

	c.onNotification = onNotification
	// send any notifications we have already collected
	for _, n := range c.notifications {
		onNotification(n)
	}
	// clear notifications
	c.notifications = nil
}

func (c *NotificationCache) listenToPgNotifications(ctx context.Context) {
	log.Printf("[INFO] NotificationCache listenToPgNotifications")
	notificationCtx, cancel := context.WithCancel(ctx)

	go func() {
		go func() {
			for notificationCtx.Err() == nil {
				log.Printf("[INFO] Wait for notification")
				notification, err := c.conn.WaitForNotification(notificationCtx)
				log.Printf("[INFO] DONE WAITING")
				if err != nil && !error_helpers.IsContextCancelledError(err) {
					log.Printf("[WARN] Error waiting for notification: %s", err)
					return
				}

				if notification != nil {
					log.Printf("[INFO] GOT NOTIFICATION")
					c.mut.Lock()
					// if we have a callback, call it
					if c.onNotification != nil {
						c.onNotification(notification)
					} else {
						// otherwise cache the notification
						c.notifications = append(c.notifications, notification)
					}
					c.mut.Unlock()
					log.Printf("[WARN] Handled notification")
				}
			}
		}()

		select {
		case <-ctx.Done():
			log.Printf("[INFO] NotificationCache context cancelklked - returning")
		case <-c.doneChan:
			// cancel the notificationCtx
			cancel()
		}

		c.conn.Close(ctx)

	}()
	log.Printf("[TRACE] InteractiveClient listenToPgNotifications DONE")
}
