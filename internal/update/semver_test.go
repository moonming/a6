package update

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseSemver_Valid(t *testing.T) {
	tests := []struct {
		input string
		want  Semver
	}{
		{input: "v1.2.3", want: Semver{Major: 1, Minor: 2, Patch: 3}},
		{input: "1.2.3", want: Semver{Major: 1, Minor: 2, Patch: 3}},
		{input: "v1.2.3-rc1", want: Semver{Major: 1, Minor: 2, Patch: 3, Pre: "rc1"}},
		{input: "dev", want: Semver{Major: -1, Minor: -1, Patch: -1, Pre: "dev"}},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseSemver(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseSemver_Invalid(t *testing.T) {
	inputs := []string{"", "v", "1", "1.2", "1.2.3.4", "v1.2.x", "v-1.2.3", "v1.2.3-"}
	for _, input := range inputs {
		input := input
		t.Run(input, func(t *testing.T) {
			_, err := ParseSemver(input)
			require.Error(t, err)
		})
	}
}

func TestSemverCompare_Equal(t *testing.T) {
	a := Semver{Major: 1, Minor: 2, Patch: 3}
	b := Semver{Major: 1, Minor: 2, Patch: 3}
	assert.Equal(t, 0, a.Compare(b))
}

func TestSemverCompare_MajorMinorPatch(t *testing.T) {
	assert.Equal(t, -1, Semver{Major: 1, Minor: 0, Patch: 0}.Compare(Semver{Major: 2, Minor: 0, Patch: 0}))
	assert.Equal(t, -1, Semver{Major: 1, Minor: 1, Patch: 9}.Compare(Semver{Major: 1, Minor: 2, Patch: 0}))
	assert.Equal(t, -1, Semver{Major: 1, Minor: 2, Patch: 3}.Compare(Semver{Major: 1, Minor: 2, Patch: 4}))
	assert.Equal(t, 1, Semver{Major: 2, Minor: 0, Patch: 0}.Compare(Semver{Major: 1, Minor: 9, Patch: 9}))
}

func TestSemverCompare_PreReleaseOrdering(t *testing.T) {
	release := Semver{Major: 1, Minor: 0, Patch: 0}
	rc := Semver{Major: 1, Minor: 0, Patch: 0, Pre: "rc1"}
	beta := Semver{Major: 1, Minor: 0, Patch: 0, Pre: "beta.1"}

	assert.Equal(t, 1, release.Compare(rc))
	assert.Equal(t, -1, rc.Compare(release))
	assert.Equal(t, 1, rc.Compare(beta))
	assert.Equal(t, -1, beta.Compare(rc))
}

func TestSemverCompare_DevHandling(t *testing.T) {
	dev := Semver{Major: -1, Minor: -1, Patch: -1, Pre: "dev"}
	v1 := Semver{Major: 1, Minor: 0, Patch: 0}

	assert.Equal(t, -1, dev.Compare(v1))
	assert.Equal(t, 1, v1.Compare(dev))
	assert.Equal(t, 0, dev.Compare(dev))
}

func TestSemverIsNewer(t *testing.T) {
	current := Semver{Major: 1, Minor: 2, Patch: 3}
	latest := Semver{Major: 1, Minor: 3, Patch: 0}
	assert.True(t, current.IsNewer(latest))
	assert.False(t, latest.IsNewer(current))
}
