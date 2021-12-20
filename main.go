package main

import (
	"database/sql"
	"fmt"
	"os"

	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/go-kit/helpers"

	"github.com/hashicorp/go-hclog"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/turbot/steampipe/cmd"
	"github.com/turbot/steampipe/utils"
)

var Logger hclog.Logger
var exitCode int

func main() {
	utils.LogTime("main start")
	exitCode := 0
	defer func() {
		if r := recover(); r != nil {
			utils.ShowError(helpers.ToError(r))
		}
		utils.LogTime("main end")
		utils.DisplayProfileData()
		os.Exit(exitCode)
	}()

	// ensure steampipe is not being run as root
	checkRoot()

	// increase the soft ULIMIT to match the hard limit
	err := setULimit()
	utils.FailOnErrorWithMessage(err, "failed to increase the file limit")

	cmd.InitCmd()

	// execute the command
	exitCode = cmd.Execute()
	// trySimple()
}

// func trySimple() {
// 	baseCtx := context.Background()
// 	connection, _ := db_client.EstablishConnection(baseCtx, "postgres://steampipe:4cbe-4bc2-9c18@localhost:9193/")
// 	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
// 	defer cancel()

// 	c, _ := connection.Conn(ctx)
// 	fmt.Println("Sending query")
// 	c.ExecContext(ctx, "select pg_sleep(30)")
// 	c.Close()

// 	count := getCount(connection)
// 	fmt.Println(count)
// }

func getCount(db *sql.DB) int {
	clientCount := 0
	// get the total number of connected clients
	row := db.QueryRow("select count(*) from pg_stat_activity where client_port IS NOT NULL and backend_type='client backend';")
	row.Scan(&clientCount)
	return clientCount - db.Stats().OpenConnections
}

// set the current to the max to avoid any file handle shortages
func setULimit() error {
	ulimit, err := filehelpers.GetULimit()
	if err != nil {
		return err
	}

	// set the current ulimit to 8192 (or the max, if less)
	// this is to ensure we do not run out of file handler when watching files
	var newULimit uint64 = 8192
	if newULimit > ulimit.Max {
		newULimit = ulimit.Max
	}
	err = filehelpers.SetULimit(newULimit)
	return err
}

// this is to replicate the user security mechanism of out underlying
// postgresql engine.
func checkRoot() {
	if os.Geteuid() == 0 {
		exitCode = 1
		utils.ShowError(fmt.Errorf(`Steampipe cannot be run as the "root" user.
To reduce security risk, use an unprivileged user account instead.`))
		os.Exit(exitCode)
	}

	/*
	 * Also make sure that real and effective uids are the same. Executing as
	 * a setuid program from a root shell is a security hole, since on many
	 * platforms a nefarious subroutine could setuid back to root if real uid
	 * is root.  (Since nobody actually uses postgres as a setuid program,
	 * trying to actively fix this situation seems more trouble than it's
	 * worth; we'll just expend the effort to check for it.)
	 */

	if os.Geteuid() != os.Getuid() {
		exitCode = 1
		utils.ShowError(fmt.Errorf("real and effective user IDs must match."))
		os.Exit(exitCode)
	}
}
