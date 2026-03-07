package completion

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestRoot() *cobra.Command {
	root := &cobra.Command{Use: "a6"}
	root.AddCommand(NewCmdCompletion())
	return root
}

func TestCompletion_Bash(t *testing.T) {
	root := newTestRoot()
	var stdout bytes.Buffer
	root.SetOut(&stdout)
	root.SetArgs([]string{"completion", "bash"})

	err := root.Execute()
	require.NoError(t, err)

	output := stdout.String()
	assert.Contains(t, output, "bash completion", "should contain bash completion script markers")
	assert.Contains(t, output, "a6", "should reference the root command name")
}

func TestCompletion_Zsh(t *testing.T) {
	root := newTestRoot()
	var stdout bytes.Buffer
	root.SetOut(&stdout)
	root.SetArgs([]string{"completion", "zsh"})

	err := root.Execute()
	require.NoError(t, err)

	output := stdout.String()
	assert.Contains(t, output, "compdef", "should contain zsh completion directives")
}

func TestCompletion_Fish(t *testing.T) {
	root := newTestRoot()
	var stdout bytes.Buffer
	root.SetOut(&stdout)
	root.SetArgs([]string{"completion", "fish"})

	err := root.Execute()
	require.NoError(t, err)

	output := stdout.String()
	assert.Contains(t, output, "complete", "should contain fish completion commands")
}

func TestCompletion_PowerShell(t *testing.T) {
	root := newTestRoot()
	var stdout bytes.Buffer
	root.SetOut(&stdout)
	root.SetArgs([]string{"completion", "powershell"})

	err := root.Execute()
	require.NoError(t, err)

	output := stdout.String()
	assert.Contains(t, output, "Register-ArgumentCompleter", "should contain PowerShell completion registration")
}

func TestCompletion_InvalidShell(t *testing.T) {
	root := newTestRoot()
	root.SetArgs([]string{"completion", "invalid"})

	err := root.Execute()
	assert.Error(t, err)
}

func TestCompletion_NoArgs(t *testing.T) {
	root := newTestRoot()
	root.SetArgs([]string{"completion"})

	err := root.Execute()
	assert.Error(t, err)
}
