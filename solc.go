package solc

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
)

// Solc represents the main structure for interacting with the Solidity compiler.
// It holds the configuration, context, and other necessary components to perform operations like compilation.
type Solc struct {
	ctx           context.Context
	config        *Config
	client        *http.Client
	gOOSFunc      func() string
	localReleases []Version
}

// New initializes and returns a new instance of the Solc structure.
// It requires a context and a configuration to be provided.
//
// Parameters:
// - ctx: The context for the Solc instance.
// - config: The configuration settings for the Solc instance.
//
// Returns:
// - A pointer to the initialized Solc instance.
// - An error if there's any issue during initialization.
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
//
// Returns:
// - The context.Context of the Solc instance.
func (s *Solc) GetContext() context.Context {
	return s.ctx
}

// GetConfig retrieves the configuration associated with the Solc instance.
//
// Returns:
// - A pointer to the Config of the Solc instance.
func (s *Solc) GetConfig() *Config {
	return s.config
}

// GetHTTPClient retrieves the HTTP client associated with the Solc instance.
//
// Returns:
// - A pointer to the http.Client of the Solc instance.
func (s *Solc) GetHTTPClient() *http.Client {
	return s.client
}

// Compile compiles the provided Solidity source code using the specified compiler configuration.
//
// Parameters:
// - ctx: The context for the compilation process.
// - source: The Solidity source code to be compiled.
// - config: The configuration settings for the compiler.
//
// Returns:
// - A pointer to the CompilerResults containing the results of the compilation.
// - An error if there's any issue during the compilation process.
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
