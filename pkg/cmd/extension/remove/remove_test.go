package remove

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/api7/a6/pkg/iostreams"
)

type mockRemoveManager struct {
	err error
}

func (m *mockRemoveManager) Remove(name string) error {
	return m.err
}

func TestRemoveRunSuccess(t *testing.T) {
	io, _, out, _ := iostreams.Test()
	opts := &Options{IO: io, Manager: &mockRemoveManager{}, Force: true}
	err := removeRun(opts, "hello")
	require.NoError(t, err)
	assert.Contains(t, out.String(), "Removed extension hello")
}

func TestRemoveRunError(t *testing.T) {
	io, _, _, _ := iostreams.Test()
	opts := &Options{IO: io, Manager: &mockRemoveManager{err: fmt.Errorf("not found")}, Force: true}
	err := removeRun(opts, "hello")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}
