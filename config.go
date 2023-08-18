// Package solc provides utilities for managing and interacting with the Solidity compiler.
package solc

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

const (
	// httpClientTimeout defines a default timeout duration for the HTTP client.
	httpClientTimeout = 10 * time.Second
)

// Config represents the configuration settings for solc-switch.
type Config struct {
	releasesPath        string
	releasesUrl         string
	httpClientTimeout   time.Duration
	personalAccessToken string
}

// Validate checks the validity of the configuration settings.
func (c *Config) Validate() error {
	if err := validatePath(c.releasesPath); err != nil {
		return err
	}

	if c.releasesUrl == "" {
		return fmt.Errorf("releases url is empty")
	}

	return nil
}

// NewDefaultConfig initializes a new Config with default settings.
func NewDefaultConfig() (*Config, error) {
	// Get the current file's path
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return nil, fmt.Errorf("error while discovering path")
	}

	execDir := filepath.Dir(filename)

	return &Config{
		releasesPath:        filepath.Join(execDir, "releases"),
		releasesUrl:         "https://api.github.com/repos/ethereum/solidity/releases",
		httpClientTimeout:   httpClientTimeout,
		personalAccessToken: os.Getenv("SOLC_SWITCH_GITHUB_TOKEN"),
	}, nil
}

// SetReleasesPath sets the path for the releases.
func (c *Config) SetReleasesPath(path string) error {
	if err := validatePath(path); err != nil {
		return err
	}

	c.releasesPath = path
	return nil
}

// GetReleasesPath returns the path where releases are stored.
func (c *Config) GetReleasesPath() string {
	return c.releasesPath
}

// GetReleasesUrl returns the URL from which releases are fetched.
func (c *Config) GetReleasesUrl() string {
	return c.releasesUrl
}

// SetHttpClientTimeout sets the timeout duration for the HTTP client.
func (c *Config) SetHttpClientTimeout(timeout time.Duration) {
	c.httpClientTimeout = timeout
}

// GetHttpClientTimeout returns the timeout duration set for the HTTP client.
func (c *Config) GetHttpClientTimeout() time.Duration {
	return c.httpClientTimeout
}
