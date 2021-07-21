package db

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func (c *InteractiveClient) startCancelHandler() chan os.Signal {
	interruptSignalChannel := make(chan os.Signal, 10)
	signal.Notify(interruptSignalChannel, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for range interruptSignalChannel {
			log.Printf("[WARN] InteractiveClient cancel handler")
			if c.hasActiveCancel() {
				log.Printf("[WARN] hasActiveCancel")
				c.cancelFunc()
				c.clearCancelFunction()
			}
		}
	}()
	return interruptSignalChannel

}

// create a cancel context for the interactive prompt, and set c.cancelFunc
func (c *InteractiveClient) createCancelContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	c.cancelFunc = cancel
	return ctx
}

func (c *InteractiveClient) hasActiveCancel() bool {
	return c.cancelFunc != nil
}

func (c *InteractiveClient) clearCancelFunction() {
	c.cancelFunc = nil
}
