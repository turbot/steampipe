package interactive

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/turbot/steampipe/db/db_common"

	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/utils"
)

var initTimeout = 40 * time.Second

func (c *InteractiveClient) readInitDataStream() {
	defer func() {
		if r := recover(); r != nil {
			c.interactivePrompt.ClearScreen()
			utils.ShowError(helpers.ToError(r))

		}
	}()
	initData := <-*(c.initDataChan)
	c.initData = initData

	if initData.Result.Error != nil {
		c.initResultChan <- initData.Result
		return
	}

	// now create prepared statements
	log.Printf("[TRACE] readInitDataStream - data has arrived")

	// start the workspace file watcher
	if viper.GetBool(constants.ArgWatch) {
		// provide an explicit error handler which re-renders the prompt after displaying the error
		initData.Result.Error = c.initData.Workspace.SetupWatcher(c.initData.Client, c.workspaceWatcherErrorHandler)

	}
	c.initResultChan <- initData.Result
}

func (c *InteractiveClient) workspaceWatcherErrorHandler(err error) {
	fmt.Println()
	utils.ShowError(err)
	c.interactivePrompt.Render()
}

func (c *InteractiveClient) isInitialised() bool {
	return c.initData != nil
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
func (c *InteractiveClient) workspace() db_common.WorkspaceResourceProvider {
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
