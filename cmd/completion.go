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
		Short:                 "Generate completion script",
		Long: `To load completions:

  Bash:
  
    $ source <(steampipe completion bash)
    
    # To load completions for each session, execute once:
    # Linux:
    $ steampipe completion bash > /etc/bash_completion.d/steampipe
    # macOS:
    $ steampipe completion bash > /usr/local/etc/bash_completion.d/steampipe
  
  Zsh:

	# If shell completion is not already enabled in your environment,
	# you will need to enable it.  You can execute the following once:

	$ echo "autoload -U compinit; compinit" >> ~/.zshrc

	# To load completions for each session, execute once:
	$ yourprogram completion zsh > "${fpath[1]}/_yourprogram"

	# You will need to start a new shell for this setup to take effect.
  
  fish:
  
    $ steampipe completion fish | source
    
    # To load completions for each session, execute once:
    $ steampipe completion fish > ~/.config/fish/completions/steampipe.fish
`,
	}

	cmdconfig.
		OnCmd(cmd)

	return cmd
}

func runGenCompletionScriptsCmd(cmd *cobra.Command, args []string) {
	completionFor := args[0]

	switch completionFor {
	case "bash":
		cmd.Root().GenBashCompletion(os.Stdout)
	case "zsh":
		cmd.Root().GenZshCompletion(os.Stdout)
	case "fish":
		cmd.Root().GenFishCompletion(os.Stdout, true)
	default:
		utils.ShowError(fmt.Errorf("'%s' is not a recognized shell", completionFor))
	}
}
