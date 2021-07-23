package interactive

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/turbot/steampipe/db"

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

	log.Printf("[TRACE] readInitDataStream - data has arrived")

	// start the workspace file watcher
	if viper.GetBool(constants.ArgWatch) {
		err := c.initData.Workspace.SetupWatcher(c.initData.Client)
		initData.Result.Error = err
	}
	c.initResultChan <- initData.Result
}

func (c *InteractiveClient) getInitError() error {
	if c.initData == nil {
		return nil
	}
	return c.initData.Result.Error
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
				return nil
			}
		case <-time.After(initTimeout):
			return fmt.Errorf("timed out waiting for initialisation to complete")
		}
	}
}

// return the workspace, or nil if not yet initialised
func (c *InteractiveClient) workspace() db.WorkspaceResourceProvider {
	if c.initData == nil {
		return nil
	}
	return c.initData.Workspace
}

// return the client, or nil if not yet initialised
func (c *InteractiveClient) client() *db.Client {
	if c.initData == nil {
		return nil
	}
	return c.initData.Client
}
