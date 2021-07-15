package db

import (
	"log"
	"time"

	"github.com/spf13/viper"

	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/utils"
)

var initTimeout = 20 * time.Second

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

func (c *InteractiveClient) waitForInitData() {
	startWait := time.Now()
	for !c.isInitialised() {
		time.Sleep(20 * time.Millisecond)
		if time.Since(startWait) > initTimeout {
			// TODO is panic right?
			panic("timed out waiting for initialisation to complete")
		}
	}
}

func (c *InteractiveClient) waitForWorkspace() WorkspaceResourceProvider {
	c.waitForInitData()
	return c.initData.Workspace
}

func (c *InteractiveClient) waitForClient() *Client {
	c.waitForInitData()

	return c.initData.Client
}
