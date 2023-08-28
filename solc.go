package solc

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"time"
)

// Solc represents the main structure for interacting with the Solidity compiler.
// It holds the configuration, context, and other necessary components to perform operations like compilation.
type Solc struct {
	ctx           context.Context
	config        *Config
	client        *http.Client
	gOOSFunc      func() string
	localReleases []Version
	lastSync      time.Time
}

// New initializes and returns a new instance of the Solc structure.
func New(ctx context.Context, config *Config) (*Solc, error) {
	if config == nil {
		return nil, fmt.Errorf("config needs to be provided")
	}

	if err := config.Validate(); err != nil {
		return nil, err
	}

	return &Solc{
		ctx:      ctx,
		config:   config,
		gOOSFunc: func() string { return runtime.GOOS },
		client: &http.Client{
			Timeout: config.GetHttpClientTimeout(),
		},
	}, nil
}

// GetContext retrieves the context associated with the Solc instance.
func (s *Solc) GetContext() context.Context {
	return s.ctx
}

// LastSyncTime retrieves the last time the Solc instance was synced.
func (s *Solc) LastSyncTime() time.Time {
	return s.lastSync
}

// GetConfig retrieves the configuration associated with the Solc instance.
func (s *Solc) GetConfig() *Config {
	return s.config
}

// GetHTTPClient retrieves the HTTP client associated with the Solc instance.
func (s *Solc) GetHTTPClient() *http.Client {
	return s.client
}

// Compile compiles the provided Solidity source code using the specified compiler configuration.
func (s *Solc) Compile(ctx context.Context, source string, config *CompilerConfig) (*CompilerResults, error) {
	compiler, err := NewCompiler(ctx, s, config, source)
	if err != nil {
		return nil, err
	}

	compilerResults, err := compiler.Compile()
	if err != nil {
		return nil, err
	}

	return compilerResults, nil
}
