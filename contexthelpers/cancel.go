package contexthelpers

import (
	"context"
	"log"
	"os"
	"os/signal"
)

func StartCancelHandler(cancel context.CancelFunc) chan os.Signal {
	sigIntChannel := make(chan os.Signal, 1)
	signal.Notify(sigIntChannel, os.Interrupt)
	go func() {
		<-sigIntChannel
		log.Println("[TRACE] got SIGINT")
		// call context cancellation function
		cancel()
		// leave the channel open - any subsequent interrupts hits will be ignored
	}()
	return sigIntChannel
}
