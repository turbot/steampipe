package cmd

import (
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/turbot/steampipe/constants"

	"github.com/spf13/viper"

	"github.com/spf13/cobra"
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
	// create command which will run steampipe in plugin-manager mode
	pluginManagerCmd := exec.Command("steampipe", "plugin-manager", "--install-dir", viper.GetString(constants.ArgInstallDir))
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
