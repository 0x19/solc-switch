package solc

import (
	"context"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// TestSyncer tests the Syncer but as well builds the releases in the releases path.
func TestSyncer(t *testing.T) {
	logger, err := GetDevelopmentLogger(zapcore.DebugLevel)
	assert.NoError(t, err)
	assert.NotNil(t, logger)
	zap.ReplaceGlobals(logger)

	tests := []struct {
		name         string
		releasesPath string
		expectedGOOS string
		config       *Config
		wantSyncErr  bool
	}{
		{
			name:         "Download Binaries Successfully",
			expectedGOOS: runtime.GOOS,
			config: func() *Config {
				config, err := NewDefaultConfig()
				assert.NoError(t, err)
				assert.NotNil(t, config)
				return config
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// We need to set timeout to 3 minutes so we can debug the HTTP requests.
			tt.config.SetHttpClientTimeout(180 * time.Second)

			s, err := New(context.TODO(), tt.config)
			assert.NoError(t, err)
			assert.NotNil(t, s)

			err = s.Sync()
			if tt.wantSyncErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

		})
	}
}

func TestSyncOnce(t *testing.T) {
	logger, err := GetDevelopmentLogger(zapcore.DebugLevel)
	assert.NoError(t, err)
	assert.NotNil(t, logger)
	zap.ReplaceGlobals(logger)

	tests := []struct {
		name         string
		releasesPath string
		expectedGOOS string
		config       *Config
		wantSyncErr  bool
	}{
		{
			name:         "Download Binaries Successfully",
			expectedGOOS: runtime.GOOS,
			config: func() *Config {
				config, err := NewDefaultConfig()
				assert.NoError(t, err)
				assert.NotNil(t, config)
				return config
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// We need to set timeout to 1.5 minutes so we can debug the HTTP requests.
			tt.config.SetHttpClientTimeout(90 * time.Second)

			s, err := New(context.TODO(), tt.config)
			assert.NoError(t, err)
			assert.NotNil(t, s)

			latestRelease, err := s.GetLatestRelease()
			assert.NoError(t, err)
			assert.NotNil(t, latestRelease)

			// Remove release from path so we can test the SyncOne method.
			err = s.RemoveBinary(latestRelease.TagName)
			assert.NoError(t, err)

			err = s.SyncOne(latestRelease)
			if tt.wantSyncErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

		})
	}
}
