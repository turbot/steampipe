package main

import (
	"fmt"
	"os"

	"github.com/hashicorp/go-hclog"
	_ "github.com/lib/pq"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/cmd"
	"github.com/turbot/steampipe/utils"
)

var Logger hclog.Logger

func main() {
	utils.LogTime("main start")
	defer func() {
		utils.LogTime("main end")
		if r := recover(); r != nil {
			utils.ShowError(helpers.ToError(r))
		}
	}()

	// ensure steampipe is not being run as root
	checkRoot()
	cmd.InitCmd()

	// execute the command
	cmd.Execute()

	utils.LogTime("end")
	utils.DisplayProfileData()
}

// this is to replicate the user security mechanism of out underlying
// postgresql engine.
func checkRoot() {
	if os.Geteuid() == 0 {
		panic(fmt.Errorf(`Steampipe cannot be run as the "root" user.
To reduce security risk, use an unprivileged user account instead.`))
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
		panic(fmt.Errorf("real and effective user IDs must match."))
	}
}
