package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turbot/steampipe/constants"
)

func daemonCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "daemon",
		Run:    runDaemonCmd,
		Hidden: true,
	}
	return cmd
}

func runDaemonCmd(cmd *cobra.Command, args []string) {
	// get the location of the currently running steampipe process
	executable, err := os.Executable()
	if err != nil {
		fmt.Printf("[WARN] plugin manager start() - failed to get steampipe executable path: %s", err)
		os.Exit(1)
	}
	// create command which will run steampipe plugin-manager
	pluginManagerCmd := exec.Command(executable, "plugin-manager", "--install-dir", viper.GetString(constants.ArgInstallDir))
	pluginManagerCmd.Stdout = os.Stdout
	pluginManagerCmd.Start()

	// wait to be killed
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan,
		syscall.SIGINT,
		syscall.SIGKILL,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	<-sigchan

	// kill our child
	// NOTE we will not do this if kill -9 is run
	pluginManagerCmd.Process.Kill()
}
