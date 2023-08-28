package solc

import "encoding/json"

type Source struct {
	Content string `json:"content"`
}

type Settings struct {
	Optimizer       Optimizer                      `json:"optimizer"`
	EVMVersion      string                         `json:"evmVersion,omitempty"`
	Remappings      []string                       `json:"remappings,omitempty"`
	OutputSelection map[string]map[string][]string `json:"outputSelection"`
}

type Optimizer struct {
	Enabled bool `json:"enabled"`
	Runs    int  `json:"runs"`
}

type CompilerJsonConfig struct {
	Language string            `json:"language"`
	Sources  map[string]Source `json:"sources"`
	Settings Settings          `json:"settings"`
}

func (c *CompilerJsonConfig) ToJSON() ([]byte, error) {
	return json.Marshal(c)
}
