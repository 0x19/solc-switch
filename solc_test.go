package solc

import (
	"context"
	"os"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestSolc(t *testing.T) {
	logger, err := GetProductionLogger(zap.InfoLevel)
	assert.NoError(t, err)
	assert.NotNil(t, logger)
	zap.ReplaceGlobals(logger)

	tempDir, err := os.MkdirTemp("", "test")
	assert.NoError(t, err)
	assert.NotEmpty(t, tempDir)
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name           string
		releasesPath   string
		expectedGOOS   string
		config         *Config
		expectedConfig *Config
		wantErr        bool
		wantConfigErr  bool
	}{
		{
			name:         "Valid Config",
			releasesPath: tempDir,
			expectedGOOS: runtime.GOOS,
			config: func() *Config {
				config, err := NewDefaultConfig()
				assert.NoError(t, err)
				assert.NotNil(t, config)
				return config
			}(),
			expectedConfig: &Config{
				releasesPath:        tempDir,
				releasesUrl:         "https://api.github.com/repos/ethereum/solidity/releases",
				httpClientTimeout:   httpClientTimeout,
				personalAccessToken: os.Getenv("SOLC_SWITCH_GITHUB_TOKEN"),
			},
			wantErr: false,
		},
		{
			name:         "Invalid Path",
			releasesPath: "/invalid/path/that/does/not/exist",
			config: func() *Config {
				config, err := NewDefaultConfig()
				assert.NoError(t, err)
				assert.NotNil(t, config)
				return config
			}(),
			wantErr: true,
		},
		{
			name:         "Empty Path",
			releasesPath: "",
			config: func() *Config {
				config, err := NewDefaultConfig()
				assert.NoError(t, err)
				assert.NotNil(t, config)
				return config
			}(),
			wantErr: true,
		},
		{
			name:           "No config",
			releasesPath:   tempDir,
			expectedGOOS:   runtime.GOOS,
			config:         nil,
			expectedConfig: nil,
			wantErr:        false,
			wantConfigErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.config != nil {
				err = tt.config.SetReleasesPath(tt.releasesPath)
				if tt.wantErr {
					assert.Error(t, err)
					return
				}
				assert.NoError(t, err)
			}

			s, err := New(context.TODO(), tt.config)
			if tt.wantConfigErr {
				assert.Error(t, err)
				assert.Nil(t, s)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, s)

			assert.NoError(t, err)
			assert.Equal(t, context.TODO(), s.GetContext())
			assert.Equal(t, tt.expectedConfig, s.GetConfig())
			assert.Equal(t, tt.expectedGOOS, s.gOOSFunc())
			assert.Equal(t, tt.expectedConfig.GetHttpClientTimeout(), s.GetHTTPClient().Timeout)
			assert.Equal(t, tt.expectedConfig.GetReleasesPath(), s.GetConfig().GetReleasesPath())
			assert.Equal(t, tt.expectedConfig.GetReleasesUrl(), s.GetConfig().GetReleasesUrl())
		})
	}
}
