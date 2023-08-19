package solc

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				releasesPath:      "/valid/path",
				releasesUrl:       "https://valid.url",
				httpClientTimeout: 10 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "invalid releases path",
			config: &Config{
				releasesPath:      "/invalid/path",
				releasesUrl:       "https://valid.url",
				httpClientTimeout: 10 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "empty releases url",
			config: &Config{
				releasesPath:      "/valid/path",
				releasesUrl:       "",
				httpClientTimeout: 10 * time.Second,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfig_SetReleasesPath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "valid path",
			path:    "/valid/path",
			wantErr: false,
		},
		{
			name:    "invalid path",
			path:    "/invalid/path",
			wantErr: true,
		},
	}

	config := &Config{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := config.SetReleasesPath(tt.path)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.path, config.GetReleasesPath())
			}
		})
	}
}

func TestConfig_SetHttpClientTimeout(t *testing.T) {
	config := &Config{}
	timeout := 5 * time.Second
	config.SetHttpClientTimeout(timeout)
	assert.Equal(t, timeout, config.GetHttpClientTimeout())
}
