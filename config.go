package solc

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

const (
	// Define a reasonable timeout for the HTTP client.
	httpClientTimeout = 10 * time.Second
)

type Config struct {
	releasesPath        string
	releasesUrl         string
	httpClientTimeout   time.Duration
	personalAccessToken string
}

func (c *Config) Validate() error {
	if err := validatePath(c.releasesPath); err != nil {
		return err
	}

	if c.releasesUrl == "" {
		return fmt.Errorf("releases url is empty")
	}

	return nil
}

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

func (c *Config) SetReleasesPath(path string) error {
	if err := validatePath(path); err != nil {
		return err
	}

	c.releasesPath = path
	return nil
}

func (c *Config) GetReleasesPath() string {
	return c.releasesPath
}

func (c *Config) GetReleasesUrl() string {
	return c.releasesUrl
}

func (c *Config) SetHttpClientTimeout(timeout time.Duration) {
	c.httpClientTimeout = timeout
}

func (c *Config) GetHttpClientTimeout() time.Duration {
	return c.httpClientTimeout
}
