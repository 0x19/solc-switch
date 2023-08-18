package solc

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
)

type Solc struct {
	ctx           context.Context
	config        *Config
	client        *http.Client
	gOOSFunc      func() string
	localReleases []Version
}

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

func (s *Solc) GetContext() context.Context {
	return s.ctx
}

func (s *Solc) GetConfig() *Config {
	return s.config
}

func (s *Solc) GetHTTPClient() *http.Client {
	return s.client
}
