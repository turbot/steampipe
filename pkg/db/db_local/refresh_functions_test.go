package db_local

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/turbot/pipe-fittings/v2/app_specific"
	"github.com/turbot/steampipe/v2/pkg/constants"
)

// test used for debug purposes to replicate `tuple concurrently updated` DB error
func TestConcurrentPerms(t *testing.T) {
	t.Skip()
	app_specific.InstallDir = "/users/kai/.steampipe"

	ctx := context.Background()
	res := StartServices(ctx, []string{"localhost"}, constants.DatabaseDefaultPort, "query")
	if res.Error != nil {
		t.Fatal(res.Error)
	}
	//defer StopServices(ctx, false, "query")

	queries := []string{
		//"lock table pg_namespace",
		//"lock table pg_user",
		//"lock table pg_authid",
		//
		//fmt.Sprintf(`create schema if not exists %s;`, constants.FunctionSchema),
		//fmt.Sprintf(`grant usage on schema %s to %s`, constants.FunctionSchema, constants.DatabaseUsersRole),
		"lock table pg_user",
		//"lock pg_authid",
		fmt.Sprintf(`alter user steampipe with password '%s'`, "3da8_4e46_8301"),
	}
	count := 100
	errchan := make(chan error, count)
	var wg sync.WaitGroup
	wg.Add(count)

	for i := 1; i <= count; i++ {
		runQueriesAsync(queries, &wg, errchan)
	}

	var doneChan = make(chan struct{})
	go func() {
		wg.Wait()
		close(doneChan)
	}()

	for {
		select {
		case err := <-errchan:
			fmt.Println("ERROR ", err)
		case <-doneChan:
			fmt.Println("DONE!")
			return
		}
	}
}

func runQueriesAsync(queries []string, wg *sync.WaitGroup, errChan chan error) {

	go func() {
		_, err := executeSqlAsRoot(context.Background(), queries...)
		if err != nil {
			errChan <- err
		}
		wg.Done()
	}()
}
