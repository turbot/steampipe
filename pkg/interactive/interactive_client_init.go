package interactive

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/v2/pkg/db/db_common"
	"github.com/turbot/steampipe/v2/pkg/error_helpers"
	"github.com/turbot/steampipe/v2/pkg/statushooks"
)

// init data has arrived, handle any errors/warnings/messages
func (c *InteractiveClient) handleInitResult(ctx context.Context, initResult *db_common.InitResult) {
	// whatever happens, set initialisationComplete
	defer func() {
		c.initialisationComplete.Store(true)
	}()

	if initResult.Error != nil {
		c.ClosePrompt(AfterPromptCloseExit)
		// add newline to ensure error is not printed at end of current prompt line
		fmt.Println()
		c.promptResult.PromptErr = initResult.Error
		return
	}

	if error_helpers.IsContextCanceled(ctx) {
		c.ClosePrompt(AfterPromptCloseExit)
		// add newline to ensure error is not printed at end of current prompt line
		fmt.Println()
		error_helpers.ShowError(ctx, initResult.Error)
		log.Printf("[TRACE] prompt context has been cancelled - not handling init result")
		return
	}

	if initResult.HasMessages() {
		c.showMessages(ctx, initResult.DisplayMessages)
	}

	// initialise autocomplete suggestions
	//nolint:golint,errcheck // worst case is we won't have autocomplete - this is not a failure
	c.initialiseSuggestions(ctx)

}

func (c *InteractiveClient) showMessages(ctx context.Context, showMessages func()) {
	statushooks.Done(ctx)
	// clear the prompt
	// NOTE: this must be done BEFORE setting hidePrompt
	// otherwise the cursor calculations in go-prompt do not work and multi-line test is not cleared
	c.interactivePrompt.ClearLine()
	// set the flag hide the prompt prefix in the next prompt render cycle
	c.hidePrompt = true
	// call ClearLine to render the empty prefix
	c.interactivePrompt.ClearLine()

	// call the passed in func to display the messages
	showMessages()

	// show the prompt again
	c.hidePrompt = false

	// We need to render the prompt here to make sure that it comes back
	// after the messages have been displayed (only if there's no execution)
	//
	// We check for query execution by TRYING to acquire the same lock that
	// execution locks on
	//
	// If we can acquire a lock, that means that there's no
	// query execution underway - and it is safe to render the prompt
	//
	// otherwise, that query execution is waiting for this init to finish
	// and as such will be out of the prompt - in which case, we shouldn't
	// re-render the prompt
	//
	// the prompt will be re-rendered when the query execution finished
	if c.executionLock.TryLock() {
		c.interactivePrompt.Render()
		// release the lock
		c.executionLock.Unlock()
	}
}

func (c *InteractiveClient) readInitDataStream(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			c.interactivePrompt.ClearScreen()
			error_helpers.ShowError(ctx, helpers.ToError(r))
		}
	}()

	<-c.initData.Loaded

	defer func() { c.initResultChan <- c.initData.Result }()

	if c.initData.Result.Error != nil {
		return
	}
	statushooks.SetStatus(ctx, "Load plugin schemas…")
	//  fetch the schema
	// TODO make this async https://github.com/turbot/steampipe/issues/3400
	// NOTE: we would like to do this asyncronously, but we are currently limited to a single Db connection in our
	// as the client cache settings are set per connection so we rely on only having a single connection
	// This means that the schema load would block other queries anyway so there is no benefit right not in making asyncronous

	if err := c.loadSchema(); err != nil {
		c.initData.Result.Error = err
		return
	}

	log.Printf("[TRACE] SetupWatcher")

	statushooks.SetStatus(ctx, "Start file watcher…")

	statushooks.SetStatus(ctx, "Start notifications listener…")
	log.Printf("[TRACE] Start notifications listener")

	// subscribe to postgres notifications
	statushooks.SetStatus(ctx, "Subscribe to postgres notifications…")

	c.listenToPgNotifications(ctx)
}

// return whether the client is initialises
// there are 3 conditions>
func (c *InteractiveClient) isInitialised() bool {
	return c.initialisationComplete.Load()
}

func (c *InteractiveClient) waitForInitData(ctx context.Context) error {
	var initTimeout = 40 * time.Second
	ticker := time.NewTicker(20 * time.Millisecond)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if c.isInitialised() {
				// if there was an error in initialisation, return it
				return c.initData.Result.Error
			}
		case <-time.After(initTimeout):
			return fmt.Errorf("timed out waiting for initialisation to complete")
		}
	}
}

// return the client, or nil if not yet initialised
func (c *InteractiveClient) client() db_common.Client {
	if c.initData == nil {
		return nil
	}
	return c.initData.Client
}
