package solc

import (
	"context"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAvailableVersions(t *testing.T) {
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
			name:         "Invalid Release URL",
			releasesPath: tempDir,
			expectedGOOS: runtime.GOOS,
			config: func() *Config {
				config, err := NewDefaultConfig()
				assert.NoError(t, err)
				assert.NotNil(t, config)

				config.SetHttpClientTimeout(1 * time.Second)
				config.releasesUrl = "https://api.github.com/repos/ethereum/solidity/releasesssss"

				return config
			}(),
			expectedConfig: &Config{
				releasesPath:      tempDir,
				releasesUrl:       "https://api.github.com/repos/ethereum/solidity/releasesssss",
				httpClientTimeout: httpClientTimeout,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.config != nil {
				err = tt.config.SetReleasesPath(tt.releasesPath)
				assert.NoError(t, err)
			}

			s, err := New(context.TODO(), tt.config)
			assert.NoError(t, err)
			assert.NotNil(t, s)

			versions, err := s.SyncReleases()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, versions)
				return
			}

			assert.NoError(t, err)
			assert.NotEmpty(t, versions)

			// Check the first version for non-empty fields as a basic validation
			assert.NotEmpty(t, versions[0].URL)
			assert.NotEmpty(t, versions[0].Name)
			assert.NotEmpty(t, versions[0].TagName)

			// Check that the config was not modified
			assert.Equal(t, tt.expectedConfig, s.GetConfig())

			// Check that the GOOS function was not modified
			assert.Equal(t, tt.expectedGOOS, s.gOOSFunc())

			// Check that the context was not modified
			assert.Equal(t, context.TODO(), s.GetContext())

			// Check that the http client was not modified
			assert.Equal(t, tt.config.GetHttpClientTimeout(), s.GetHTTPClient().Timeout)

			// Compare result with local releases
			localReleases, err := s.GetLocalReleases()
			assert.NoError(t, err)
			assert.NotNil(t, localReleases)
			assert.Equal(t, localReleases, versions)

			// Compare results with latest release
			latestRelease, err := s.GetLatestRelease()
			assert.NoError(t, err)
			assert.NotNil(t, latestRelease)
			assert.Equal(t, latestRelease.URL, versions[0].URL)
			assert.Equal(t, latestRelease.Name, versions[0].Name)
			assert.Equal(t, latestRelease.TagName, versions[0].TagName)
			assert.Equal(t, latestRelease.Prerelease, versions[0].Prerelease)

			// Compare results with latest release without local cache
			s.localReleases = nil
			latestRelease, err = s.GetLatestRelease()
			assert.NoError(t, err)
			assert.NotNil(t, latestRelease)
			assert.Equal(t, latestRelease.URL, versions[0].URL)
			assert.Equal(t, latestRelease.Name, versions[0].Name)
			assert.Equal(t, latestRelease.TagName, versions[0].TagName)
			assert.Equal(t, latestRelease.Prerelease, versions[0].Prerelease)

			// Get single release by tag name
			release, err := s.GetRelease(versions[0].TagName)
			assert.NoError(t, err)
			assert.NotNil(t, release)

			// Get single release by tag name without local cache
			s.localReleases = nil
			release, err = s.GetRelease(versions[0].TagName)
			assert.NoError(t, err)
			assert.NotNil(t, release)

			// Get simplified version info
			versionInfo, err := s.GetReleasesSimplified()
			assert.NoError(t, err)
			assert.NotEmpty(t, versionInfo)
			for i, v := range versionInfo {
				assert.Equal(t, v.TagName, versions[i].TagName)
				assert.Equal(t, v.IsLatest, versions[i].TagName == latestRelease.TagName)
				assert.Equal(t, v.IsPrerelease, versions[i].Prerelease)
			}

		})
	}
}
