package interactive

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/db/db_common"
	"github.com/turbot/steampipe/utils"
	"github.com/turbot/steampipe/workspace"
)

var initTimeout = 40 * time.Second

// init data has arrived, handle any errors/warnings/messages
func (c *InteractiveClient) handleInitResult(ctx context.Context, initResult *db_common.InitResult) {
	// try to take an execution lock, so that we don't end up showing warnings and errors
	// while an execution is underway
	c.executionLock.Lock()
	defer c.executionLock.Unlock()

	if utils.IsContextCancelled(ctx) {
		log.Printf("[TRACE] prompt context has been cancelled - not handling init result")
		return
	}

	if initResult.Error != nil {
		c.ClosePrompt(AfterPromptCloseExit)
		// add newline to ensure error is not printed at end of current prompt line
		fmt.Println()
		utils.ShowError(ctx, initResult.Error)
		return
	}

	if initResult.HasMessages() {
		fmt.Println()
		initResult.DisplayMessages()
	}

	// We need to render the prompt here to make sure that it comes back
	// after the messages have been displayed
	c.interactivePrompt.Render()

	// tell the workspace to reset the prompt after displaying async filewatcher messages
	c.initData.Workspace.SetOnFileWatcherEventMessages(func() { c.interactivePrompt.Render() })
}

func (c *InteractiveClient) readInitDataStream(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			c.interactivePrompt.ClearScreen()
			utils.ShowError(ctx, helpers.ToError(r))

		}
	}()
	<-c.initData.Loaded

	if c.initData.Result.Error != nil {
		c.initResultChan <- c.initData.Result
		return
	}

	// asyncronously fetch the schema
	go c.loadSchema()

	log.Printf("[TRACE] readInitDataStream - data has arrived")

	// start the workspace file watcher
	if viper.GetBool(constants.ArgWatch) {
		// provide an explicit error handler which re-renders the prompt after displaying the error
		if err := c.initData.Workspace.SetupWatcher(ctx, c.initData.Client, c.workspaceWatcherErrorHandler); err != nil {
			c.initData.Result.Error = err
		}
	}
	c.initResultChan <- c.initData.Result
}

func (c *InteractiveClient) workspaceWatcherErrorHandler(ctx context.Context, err error) {
	fmt.Println()
	utils.ShowError(ctx, err)
	c.interactivePrompt.Render()
}

// return whether the client is initialises
// there are 3 conditions>
func (c *InteractiveClient) isInitialised() bool {
	return c.initData != nil && c.schemaMetadata != nil
}

func (c *InteractiveClient) waitForInitData(ctx context.Context) error {
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
