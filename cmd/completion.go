package cmd

import (
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/turbot/steampipe/cmdconfig"
	"github.com/turbot/steampipe/utils"
)

func generateCompletionScriptsCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:                   "completion [bash|zsh|fish]",
		Args:                  cobra.ExactValidArgs(1),
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish"},
		Run:                   runGenCompletionScriptsCmd,
		Short:                 "Generate completion scripts",
	}

	cmd.ResetFlags()

	cmd.SetHelpFunc(completionHelp)

	cmdconfig.
		OnCmd(cmd)

	return cmd
}

func includeBashHelp(base string) string {
	buildUp := base
	buildUp = fmt.Sprintf(`%s
  Bash:
    # To load for the current session, execute:
    $ source <(steampipe completion bash)
`, buildUp)

	if runtime.GOOS == "darwin" {
		buildUp = fmt.Sprintf(`%s

		# This script depends on the 'bash-completion' package.
		# If it is not installed already, you can install it via your OS’s package manager.
		
		# To install with 'homebrew':
		$ brew install bash-completion
		
		# Once installed, to edit your '.bash_profile' file, execute the following:
		$ echo "[ -f $(brew --prefix)/etc/bash_completion ] && . $(brew --prefix)/etc/bash_completion" >> ~/.bash_profile

		$ steampipe completion bash > $(brew --prefix)/etc/bash_completion.d/steampipe
`, buildUp)
	} else if runtime.GOOS == "linux" {
		buildUp = fmt.Sprintf(`%s
    # To load completions for every session, execute once:
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
    # If shell completion is not already enabled in your environment, you will need to enable it.

    # To enable completion in your environment, execute:
    $ echo "autoload -U compinit; compinit" >> ~/.zshrc
    
    # To load completions for each session, execute once:
    $ steampipe completion zsh > "${fpath[1]}/steampipe"
    
    # You will need to start a new shell for this setup to take effect.
`, buildUp)
	}

	return buildUp
}

func includeFishHelp(base string) string {
	buildUp := base

	buildUp = fmt.Sprintf(`%s
  fish:
    # To enable completion for the current session:
    $ steampipe completion fish | source
    
    # To load completions for each session, execute once:
    $ steampipe completion fish > ~/.config/fish/completions/steampipe.fish
	`, buildUp)

	return buildUp
}

func completionHelp(cmd *cobra.Command, args []string) {
	fmt.Println(runtime.GOOS)

	helpString := "To load completions:"
	helpString = includeBashHelp(helpString)
	helpString = includeZshHelp(helpString)
	helpString = includeFishHelp(helpString)

	fmt.Println(helpString)
	fmt.Println(cmd.UsageString())
}

func runGenCompletionScriptsCmd(cmd *cobra.Command, args []string) {
	completionFor := args[0]

	switch completionFor {
	case "bash":
		cmd.Root().GenBashCompletionV2(os.Stdout, false)
	case "zsh":
		cmd.Root().GenZshCompletionNoDesc(os.Stdout)
	case "fish":
		cmd.Root().GenFishCompletion(os.Stdout, false)
	default:
		utils.ShowError(fmt.Errorf("completion for '%s' is not supported yet", completionFor))
	}
}
