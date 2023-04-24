package interactive

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/db/db_common"
	"github.com/turbot/steampipe/pkg/error_helpers"
	"github.com/turbot/steampipe/pkg/statushooks"
	"github.com/turbot/steampipe/pkg/workspace"
)

// init data has arrived, handle any errors/warnings/messages
func (c *InteractiveClient) handleInitResult(ctx context.Context, initResult *db_common.InitResult) {
	// whatever happens, set initialisationComplete
	defer func() {
		c.initialisationComplete = true
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
		statushooks.Done(ctx)
		// clear the prompt
		// NOTE: this must be done BEFORE setting hidePrompt
		// otherwise the cursor calculations in go-prompt do not work and multi-line test is not cleared
		c.interactivePrompt.ClearLine()
		// set the flag hide the prompt prefix in the next prompt render cycle
		c.hidePrompt = true
		// call ClearLine to render the empty prefix
		c.interactivePrompt.ClearLine()

		// display messages
		initResult.DisplayMessages()
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

	// initialise autocomplete suggestions
	c.initialiseSuggestions(ctx)
	// tell the workspace to reset the prompt after displaying async filewatcher messages
	c.initData.Workspace.SetOnFileWatcherEventMessages(func() {
		c.initialiseQuerySuggestions(ctx)
		c.interactivePrompt.Render()
	})

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

	// asyncronously fetch the schema
	if err := c.loadSchema(); err != nil {
		c.initData.Result.Error = err
		return
	}

	log.Printf("[TRACE] readInitDataStream - data has arrived")

	// start the workspace file watcher
	if viper.GetBool(constants.ArgWatch) {
		// provide an explicit error handler which re-renders the prompt after displaying the error
		if err := c.initData.Workspace.SetupWatcher(ctx, c.initData.Client, c.workspaceWatcherErrorHandler); err != nil {
			c.initData.Result.Error = err
		}
	}

	// create a cancellation context used to cancel the listen thread when we exit
	listenCtx, cancel := context.WithCancel(ctx)
	go c.listenToPgNotifications(listenCtx)
	c.cancelNotificationListener = cancel
}

func (c *InteractiveClient) workspaceWatcherErrorHandler(ctx context.Context, err error) {
	fmt.Println()
	error_helpers.ShowError(ctx, err)
	c.interactivePrompt.Render()
}

// return whether the client is initialises
// there are 3 conditions>
func (c *InteractiveClient) isInitialised() bool {
	return c.initialisationComplete
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

// return the workspace, or nil if not yet initialised
func (c *InteractiveClient) workspace() *workspace.Workspace {
	if c.initData == nil {
		return nil
	}
	return c.initData.Workspace
}

// return the client, or nil if not yet initialised
func (c *InteractiveClient) client() db_common.Client {
	if c.initData == nil {
		return nil
	}
	return c.initData.Client
}
