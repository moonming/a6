package cmdutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeLabel(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"env=test", "env"},
		{"env=prod", "env"},
		{"team=backend", "team"},
		{"env:test", "env:test"},
		{"env", "env"},
		{"key=value=extra", "key"},
	}
	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			assert.Equal(t, tc.expected, NormalizeLabel(tc.input))
		})
	}
}

func TestParseLabel(t *testing.T) {
	tests := []struct {
		input         string
		expectedKey   string
		expectedValue string
	}{
		{"", "", ""},
		{"env=test", "env", "test"},
		{"env=prod", "env", "prod"},
		{"team=backend", "team", "backend"},
		{"env", "env", ""},
		{"key=value=extra", "key", "value=extra"},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			key, value := ParseLabel(tc.input)
			assert.Equal(t, tc.expectedKey, key)
			assert.Equal(t, tc.expectedValue, value)
		})
	}
}
