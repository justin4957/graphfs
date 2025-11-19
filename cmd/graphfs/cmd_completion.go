/*
# Module: cmd/graphfs/cmd_completion.go
Shell completion generation command.

Generates shell completion scripts for Bash, Zsh, and Fish shells.

## Linked Modules
- [root](./root.go) - Root command

## Tags
cli, completion, shells

## Exports
completionCmd

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#cmd_completion.go> a code:Module ;
    code:name "cmd/graphfs/cmd_completion.go" ;
    code:description "Shell completion generation command" ;
    code:language "go" ;
    code:layer "cli" ;
    code:linksTo <./root.go> ;
    code:exports <#completionCmd> ;
    code:tags "cli", "completion", "shells" .
<!-- End LinkedDoc RDF -->
*/

package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish]",
	Short: "Generate shell completion scripts",
	Long: `Generate shell completion scripts for GraphFS.

IMPORTANT: Use the completion script that matches your shell!
- Check your shell: echo $SHELL
- Zsh users: use 'graphfs completion zsh' (not bash)
- Bash users: use 'graphfs completion bash'
- Fish users: use 'graphfs completion fish'

To load completions:

Bash:
  $ source <(graphfs completion bash)

  # Add to ~/.bashrc for persistence:
  $ graphfs completion bash > ~/.graphfs-completion.bash
  $ echo "source ~/.graphfs-completion.bash" >> ~/.bashrc

Zsh:
  $ source <(graphfs completion zsh)

  # Add to ~/.zshrc for persistence:
  $ mkdir -p ~/.zsh/completion
  $ graphfs completion zsh > ~/.zsh/completion/_graphfs
  $ echo 'fpath=(~/.zsh/completion $fpath)' >> ~/.zshrc
  $ echo 'autoload -Uz compinit && compinit' >> ~/.zshrc

Fish:
  $ graphfs completion fish | source

  # Add to ~/.config/fish/completions/ for persistence:
  $ mkdir -p ~/.config/fish/completions
  $ graphfs completion fish > ~/.config/fish/completions/graphfs.fish

Examples:
  # Generate bash completions
  graphfs completion bash

  # Install completions for current shell
  source <(graphfs completion bash)  # Bash
  source <(graphfs completion zsh)   # Zsh
  graphfs completion fish | source   # Fish
`,
	ValidArgs:             []string{"bash", "zsh", "fish"},
	Args:                  cobra.ExactArgs(1),
	DisableFlagsInUseLine: true,
	RunE:                  runCompletion,
}

func init() {
	rootCmd.AddCommand(completionCmd)
}

func runCompletion(cmd *cobra.Command, args []string) error {
	shell := args[0]

	switch shell {
	case "bash":
		return cmd.Root().GenBashCompletionV2(os.Stdout, true)
	case "zsh":
		return cmd.Root().GenZshCompletion(os.Stdout)
	case "fish":
		return cmd.Root().GenFishCompletion(os.Stdout, true)
	default:
		return fmt.Errorf("unsupported shell: %s (supported: bash, zsh, fish)", shell)
	}
}
