package list

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/api7/a6/internal/extension"
	"github.com/api7/a6/pkg/iostreams"
)

type mockListManager struct {
	exts []extension.Extension
	err  error
}

func (m *mockListManager) List() ([]extension.Extension, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.exts, nil
}

func TestListRunEmpty(t *testing.T) {
	io, _, out, _ := iostreams.Test()
	opts := &Options{IO: io, Manager: &mockListManager{}}
	err := listRun(opts)
	require.NoError(t, err)
	assert.Contains(t, out.String(), "No extensions installed.")
}

func TestListRunTable(t *testing.T) {
	io, _, out, _ := iostreams.Test()
	opts := &Options{IO: io, Manager: &mockListManager{exts: []extension.Extension{{Manifest: extension.Manifest{Name: "hello", Version: "1.0.0", Owner: "api7", Repo: "a6-hello"}}}}}
	err := listRun(opts)
	require.NoError(t, err)
	assert.Contains(t, out.String(), "NAME")
	assert.Contains(t, out.String(), "hello")
	assert.Contains(t, out.String(), "api7/a6-hello")
}

func TestListRunJSON(t *testing.T) {
	io, _, out, _ := iostreams.Test()
	opts := &Options{IO: io, Output: "json", Manager: &mockListManager{exts: []extension.Extension{{Manifest: extension.Manifest{Name: "hello", Version: "1.0.0", Owner: "api7", Repo: "a6-hello"}}}}}
	err := listRun(opts)
	require.NoError(t, err)
	assert.Contains(t, out.String(), `"name": "hello"`)
	assert.Contains(t, out.String(), `"owner_repo": "api7/a6-hello"`)
}

func TestListRunError(t *testing.T) {
	io, _, _, _ := iostreams.Test()
	opts := &Options{IO: io, Manager: &mockListManager{err: fmt.Errorf("boom")}}
	err := listRun(opts)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "boom")
}
