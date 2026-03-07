package version

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/api7/a6/pkg/cmd"
	"github.com/api7/a6/pkg/iostreams"
)

func TestVersion_Plain(t *testing.T) {
	ios, _, stdout, _ := iostreams.Test()

	f := &cmd.Factory{IOStreams: ios}
	c := NewCmdVersion(f)
	c.SetArgs([]string{})
	err := c.Execute()

	require.NoError(t, err)
	output := stdout.String()
	assert.Contains(t, output, "a6 version")
	assert.Contains(t, output, "commit:")
	assert.Contains(t, output, "built:")
	assert.Contains(t, output, "go:")
	assert.Contains(t, output, "platform:")
}

func TestVersion_JSON(t *testing.T) {
	ios, _, stdout, _ := iostreams.Test()

	f := &cmd.Factory{IOStreams: ios}
	c := NewCmdVersion(f)
	c.SetArgs([]string{"--output", "json"})
	err := c.Execute()

	require.NoError(t, err)
	var info Info
	err = json.Unmarshal(stdout.Bytes(), &info)
	require.NoError(t, err)
	assert.NotEmpty(t, info.Version)
	assert.NotEmpty(t, info.GoVersion)
	assert.NotEmpty(t, info.Platform)
}

func TestVersion_YAML(t *testing.T) {
	ios, _, stdout, _ := iostreams.Test()

	f := &cmd.Factory{IOStreams: ios}
	c := NewCmdVersion(f)
	c.SetArgs([]string{"--output", "yaml"})
	err := c.Execute()

	require.NoError(t, err)
	output := stdout.String()
	assert.Contains(t, output, "version:")
	assert.Contains(t, output, "goVersion:")
	assert.Contains(t, output, "platform:")
}
