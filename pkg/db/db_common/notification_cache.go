package db_common

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/turbot/steampipe-plugin-sdk/v5/sperr"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/turbot/steampipe/v2/pkg/constants"
	"github.com/turbot/steampipe/v2/pkg/error_helpers"
)

type NotificationListener struct {
	notifications []*pgconn.Notification
	conn          *pgx.Conn

	onNotification func(*pgconn.Notification)
	mut            sync.Mutex
	cancel         context.CancelFunc
}

func NewNotificationListener(ctx context.Context, conn *pgx.Conn) (*NotificationListener, error) {
	if conn == nil {
		return nil, sperr.New("nil connection passed to NewNotificationListener")
	}

	listener := &NotificationListener{conn: conn}

	// tell the connection to listen to notifications
	listenSql := fmt.Sprintf("listen %s", constants.PostgresNotificationChannel)
	_, err := conn.Exec(ctx, listenSql)
	if err != nil {
		log.Printf("[INFO] Error listening to notification channel: %s", err)
		conn.Close(ctx)
		return nil, err
	}

	// create cancel context to shutdown the listener
	cancelCtx, cancel := context.WithCancel(ctx)
	listener.cancel = cancel

	// start the goroutine to listen
	listener.listenToPgNotificationsAsync(cancelCtx)

	return listener, nil
}

func (c *NotificationListener) Stop(ctx context.Context) {
	c.conn.Close(ctx)
	// stop the listener goroutine
	c.cancel()
}

func (c *NotificationListener) RegisterListener(onNotification func(*pgconn.Notification)) {
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

func (c *NotificationListener) listenToPgNotificationsAsync(ctx context.Context) {
	log.Printf("[INFO] notificationListener listenToPgNotificationsAsync")

	go func() {
		for ctx.Err() == nil {
			log.Printf("[INFO] Wait for notification")
			notification, err := c.conn.WaitForNotification(ctx)
			if err != nil && !error_helpers.IsContextCancelledError(err) {
				log.Printf("[WARN] Error waiting for notification: %s", err)
				return
			}

			if notification != nil {
				log.Printf("[INFO] got notification")
				c.mut.Lock()
				// if we have a callback, call it
				if c.onNotification != nil {
					log.Printf("[INFO] call notification handler")
					c.onNotification(notification)
				} else {
					// otherwise cache the notification
					log.Printf("[INFO] cache notification")
					c.notifications = append(c.notifications, notification)
				}
				c.mut.Unlock()
				log.Printf("[INFO] Handled notification")
			}
		}
	}()

	log.Printf("[TRACE] InteractiveClient listenToPgNotificationsAsync DONE")
}
