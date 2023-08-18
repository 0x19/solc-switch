package solc

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetDistribution(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test")
	assert.NoError(t, err)
	assert.NotEmpty(t, tempDir)

	defer os.RemoveAll(tempDir)

	tests := []struct {
		name     string
		goos     string
		expected Distribution
	}{
		{
			name:     "Windows OS",
			goos:     Windows.String(),
			expected: Windows,
		},
		{
			name:     "MacOS",
			goos:     MacOS.String(),
			expected: MacOS,
		},
		{
			name:     "Linux OS",
			goos:     Linux.String(),
			expected: Linux,
		},
		{
			name:     "Unknown OS",
			goos:     "solaris", // Just an example of an OS that's not in our main switch case
			expected: Unknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := NewDefaultConfig()
			assert.NoError(t, err)
			assert.NotNil(t, config)

			err = config.SetReleasesPath(tempDir)
			assert.NoError(t, err)

			assert.Equal(t, tempDir, config.GetReleasesPath())

			s, err := New(context.TODO(), config)
			assert.NoError(t, err)
			assert.NotNil(t, s)

			s.gOOSFunc = func() string { return tt.goos }
			assert.Equal(t, tt.expected, s.GetDistribution())
		})
	}
}
