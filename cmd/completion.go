package cmd

import (
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/turbot/steampipe/pkg/cmdconfig"
	"github.com/turbot/steampipe/pkg/constants"
)

func generateCompletionScriptsCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:                   "completion [bash|zsh|fish]",
		Args:                  cobra.ArbitraryArgs,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish"},
		Run:                   runGenCompletionScriptsCmd,
		Short:                 "Generate completion scripts",
	}

	cmd.ResetFlags()

	cmd.SetHelpFunc(completionHelp)

	cmdconfig.OnCmd(cmd).AddBoolFlag(constants.ArgHelp, "h", false, "Help for completion")

	return cmd
}

func includeBashHelp(base string) string {
	buildUp := base
	buildUp = fmt.Sprintf(`%s
  Bash:`, buildUp)

	if runtime.GOOS == "darwin" {
		buildUp = fmt.Sprintf(`%s
    # Load for the current session:
    $ source <(steampipe completion bash)
		
    # Load for every session (requires shell restart):
    $ steampipe completion bash > $(brew --prefix)/etc/bash_completion.d/steampipe
`, buildUp)
	} else if runtime.GOOS == "linux" {
		buildUp = fmt.Sprintf(`%s
    # Load for the current session:
    $ source <(steampipe completion bash)

    # Load for every session (requires shell restart):
    $ steampipe completion bash > /etc/bash_completion.d/steampipe
	`, buildUp)
	}

	return buildUp
}

func includeZshHelp(base string) string {
	buildUp := base

	if runtime.GOOS == "darwin" {
		buildUp = fmt.Sprintf(`%s
  Zsh:
    # Load for every session (requires shell restart):
    $ steampipe completion zsh > "${fpath[1]}/steampipe"
`, buildUp)
	}

	return buildUp
}

func includeFishHelp(base string) string {
	buildUp := base

	buildUp = fmt.Sprintf(`%s
  fish:
    # Load for the current session:
    $ steampipe completion fish | source
    
    # Load for every session (requires shell restart):
    $ steampipe completion fish > ~/.config/fish/completions/steampipe.fish
	`, buildUp)

	return buildUp
}

func completionHelp(cmd *cobra.Command, args []string) {
	helpString := ""

	if runtime.GOOS == "darwin" {
		helpString = `
Note: Completions must be enabled in your environment. Please refer to: https://steampipe.io/docs/reference/cli-args#steampipe-completion
	
To load completions:
`
	} else if runtime.GOOS == "linux" {
		helpString = `
To load completions:
`
	}

	helpString = includeBashHelp(helpString)
	helpString = includeZshHelp(helpString)
	helpString = includeFishHelp(helpString)

	fmt.Println(helpString)
	fmt.Println(cmd.UsageString())
}

func runGenCompletionScriptsCmd(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		completionHelp(cmd, args)
		return
	}

	completionFor := args[0]

	switch completionFor {
	case "bash":
		cmd.Root().GenBashCompletionV2(os.Stdout, false)
	case "zsh":
		cmd.Root().GenZshCompletionNoDesc(os.Stdout)
	case "fish":
		cmd.Root().GenFishCompletion(os.Stdout, false)
	default:
		completionHelp(cmd, args)
	}
}
