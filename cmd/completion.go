package cmd

import (
	"fmt"
	"os"

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
		Long: `
To load completions:
    Bash:
        # To load for the current session
        $ source <(steampipe completion bash)
        
        # To load completions for every session, execute once:
        # Linux:
        $ steampipe completion bash > /etc/bash_completion.d/steampipe
        
        # MacOS:
        $ steampipe completion bash > $(brew --prefix)/etc/bash_completion.d/steampipe
		
        # This script depends on the 'bash-completion' package. If it is not installed already, you can install it via your OS’s package manager.
        
        # To install with 'homebrew':
        $ brew install bash-completion
        
        # Once installed, to edit your '.bash_profile' file, execute the following:
        $ echo "[ -f $(brew --prefix)/etc/bash_completion ] && . $(brew --prefix)/etc/bash_completion" >> ~/.bash_profile

    Zsh:
        # If shell completion is not already enabled in your environment,
        # you will need to enable it.  You can execute the following once:
        $ echo "autoload -U compinit; compinit" >> ~/.zshrc
        
        # To load completions for each session, execute once:
        $ steampipe completion zsh > "${fpath[1]}/steampipe"
        
        # You will need to start a new shell for this setup to take effect.
        
    fish:
        # To enable completion for the current session
        $ steampipe completion fish | source
        
        # To load completions for each session, execute once:
        $ steampipe completion fish > ~/.config/fish/completions/steampipe.fish
`,
	}

	cmd.ResetFlags()

	cmdconfig.
		OnCmd(cmd)

	return cmd
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
