package completion

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewCmdCompletion creates the completion command for generating shell scripts.
func NewCmdCompletion() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion <shell>",
		Short: "Generate shell completion scripts",
		Long: `Generate shell completion scripts for a6.

To load completions:

Bash:
  $ source <(a6 completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ a6 completion bash > /etc/bash_completion.d/a6
  # macOS:
  $ a6 completion bash > $(brew --prefix)/etc/bash_completion.d/a6

Zsh:
  # If shell completion is not already enabled in your environment,
  # you will need to enable it.  You can execute the following once:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ a6 completion zsh > "${fpath[1]}/_a6"

  # You will need to start a new shell for this setup to take effect.

Fish:
  $ a6 completion fish | source

  # To load completions for each session, execute once:
  $ a6 completion fish > ~/.config/fish/completions/a6.fish

PowerShell:
  PS> a6 completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> a6 completion powershell > a6.ps1
  # and source this file from your PowerShell profile.
`,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			switch args[0] {
			case "bash":
				return cmd.Root().GenBashCompletionV2(out, true)
			case "zsh":
				return cmd.Root().GenZshCompletion(out)
			case "fish":
				return cmd.Root().GenFishCompletion(out, true)
			case "powershell":
				return cmd.Root().GenPowerShellCompletionWithDesc(out)
			default:
				return fmt.Errorf("unsupported shell: %s", args[0])
			}
		},
	}

	return cmd
}
