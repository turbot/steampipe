package interactive

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

func (c *InteractiveClient) startCancelHandler() chan os.Signal {
	interruptSignalChannel := make(chan os.Signal, 10)
	signal.Notify(interruptSignalChannel, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for range interruptSignalChannel {
			c.cancelActiveQueryIfAny()
		}
	}()
	return interruptSignalChannel
}

// create a cancel context for the interactive prompt, and set c.cancelFunc
func (c *InteractiveClient) createPromptContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	c.cancelPrompt = cancel
	return ctx
}

func (c *InteractiveClient) createQueryContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	c.cancelActiveQuery = cancel
	return ctx
}

func (c *InteractiveClient) cancelActiveQueryIfAny() {
	if c.cancelActiveQuery != nil {
		c.cancelActiveQuery()
		c.cancelActiveQuery = nil
	}
}
