package solc

import "encoding/json"

// Source represents the content of a Solidity source file.
type Source struct {
	Content string `json:"content"` // The content of the Solidity source file.
}

// Settings defines the configuration settings for the Solidity compiler.
type Settings struct {
	Optimizer       Optimizer                      `json:"optimizer"`            // Configuration for the optimizer.
	EVMVersion      string                         `json:"evmVersion,omitempty"` // The version of the Ethereum Virtual Machine to target. Optional.
	Remappings      []string                       `json:"remappings,omitempty"` // List of remappings for library addresses. Optional.
	OutputSelection map[string]map[string][]string `json:"outputSelection"`      // Specifies the type of information to output (e.g., ABI, AST).
}

// Optimizer represents the configuration for the Solidity compiler's optimizer.
type Optimizer struct {
	Enabled bool `json:"enabled"` // Indicates whether the optimizer is enabled.
	Runs    int  `json:"runs"`    // Specifies the number of optimization runs.
}

// CompilerJsonConfig represents the JSON configuration for the Solidity compiler.
type CompilerJsonConfig struct {
	Language string            `json:"language"` // Specifies the language version (e.g., "Solidity").
	Sources  map[string]Source `json:"sources"`  // Map of source file names to their content.
	Settings Settings          `json:"settings"` // Compiler settings.
}

// ToJSON converts the CompilerJsonConfig to its JSON representation.
// It returns the JSON byte array or an error if the conversion fails.
func (c *CompilerJsonConfig) ToJSON() ([]byte, error) {
	return json.Marshal(c)
}
