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
	"github.com/turbot/steampipe/pkg/utils"
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
		error_helpers.ShowError(ctx, initResult.Error)
		return
	}

	if utils.IsContextCancelled(ctx) {
		c.ClosePrompt(AfterPromptCloseExit)
		// add newline to ensure error is not printed at end of current prompt line
		fmt.Println()
		error_helpers.ShowError(ctx, initResult.Error)
		log.Printf("[TRACE] prompt context has been cancelled - not handling init result")
		return
	}

	if initResult.HasMessages() {
		c.interactivePrompt.ClearLine()
		c.hidePrompt = true
		fmt.Println()
		initResult.DisplayMessages()
		c.hidePrompt = false
		// We need to render the prompt here to make sure that it comes back
		// after the messages have been displayed
		c.interactivePrompt.Render()
	}

	c.initialiseSuggestions()
	// tell the workspace to reset the prompt after displaying async filewatcher messages
	c.initData.Workspace.SetOnFileWatcherEventMessages(func() {
		c.initialiseSuggestions()
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

	// Trigger a re-render of the prompt, so that the prompt actually shows up,
	// since the prompt may have been removed by the installation spinner
	c.interactivePrompt.Render()

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
