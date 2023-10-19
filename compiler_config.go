package solc

import (
	"fmt"
	"regexp"
	"strings"
)

// allowedArgs defines a list of allowed arguments for solc.
var allowedArgs = map[string]bool{
	"--combined-json":     true,
	"-":                   true,
	"--optimize":          true,
	"--optimize-runs":     true,
	"--evm-version":       true,
	"--overwrite":         true,
	"--libraries":         true,
	"--standard-json":     true,
	"--allow-paths":       true,
	"--base-path":         true,
	"--ignore-missing":    true,
	"--ast":               true,
	"--ast-json":          true,
	"--include-path":      true,
	"--output-dir":        true,
	"--asm":               true,
	"--bin":               true,
	"--abi":               true,
	"--asm-json":          true,
	"--bin-runtime":       true,
	"--ir":                true,
	"--opcodes":           true,
	"--ir-optimized":      true,
	"--ewasm":             true,
	"--ewasm-ir":          true,
	"--no-optimize-yul":   true,
	"--yul-optimizations": true,
	"--yul":               true,
	"--assemble":          true,
	"--lsp":               true,
	"--hashes":            true,
	"--userdoc":           true,
	"--devdoc":            true,
	"--metadata":          true,
	"--storage-layout":    true,
	"--gas":               true,
	"--metadata-hash":     true,
	"--metadata-literal":  true,
	"--error-recovery":    true,
}

// requiredArgs defines a list of required arguments for solc.
var requiredArgs = map[string]bool{
	"--overwrite":     true,
	"--combined-json": true,
	"-":               true,
}

// CompilerConfig represents the compiler configuration for the solc binaries.
type CompilerConfig struct {
	CompilerVersion string              // The version of the compiler to use.
	EntrySourceName string              // The name of the entry source file.
	Arguments       []string            // Arguments to pass to the solc tool.
	JsonConfig      *CompilerJsonConfig // The json config to pass to the solc tool.
}

// NewDefaultCompilerConfig creates and returns a default CompilerConfiguration for compiler to use.
func NewDefaultCompilerConfig(compilerVersion string) (*CompilerConfig, error) {
	toReturn := &CompilerConfig{
		CompilerVersion: compilerVersion,
		Arguments: []string{
			"--overwrite", "--combined-json", "bin,abi", "-", // Output to stdout.
		},
	}

	if _, err := toReturn.SanitizeArguments(toReturn.Arguments); err != nil {
		return nil, err
	}

	if err := toReturn.Validate(); err != nil {
		return nil, err
	}

	return toReturn, nil
}

// NewDefaultCompilerConfig creates and returns a default CompilerConfiguration for compiler to use with provided JSON settings.
func NewCompilerConfigFromJSON(compilerVersion string, entrySourceName string, config *CompilerJsonConfig) (*CompilerConfig, error) {
	toReturn := &CompilerConfig{
		EntrySourceName: entrySourceName,
		CompilerVersion: compilerVersion,
		Arguments:       []string{"--standard-json"},
		JsonConfig:      config,
	}

	if _, err := toReturn.SanitizeArguments(toReturn.Arguments); err != nil {
		return nil, err
	}

	/*
		TODO: Validation at this point for the JSON config is not done.
		It's assumed that the caller has already validated the JSON config.
		if err := toReturn.Validate(); err != nil {
			return nil, err
		} */

	return toReturn, nil
}

// SetJsonConfig sets the json config to pass to the solc tool.
func (c *CompilerConfig) SetJsonConfig(config *CompilerJsonConfig) {
	c.JsonConfig = config
}

// GetJsonConfig returns the json config to pass to the solc tool.
func (c *CompilerConfig) GetJsonConfig() *CompilerJsonConfig {
	return c.JsonConfig
}

// SetEntrySourceName sets the name of the entry source file.
func (c *CompilerConfig) SetEntrySourceName(name string) {
	c.EntrySourceName = name
}

// GetEntrySourceName returns the name of the entry source file.
func (c *CompilerConfig) GetEntrySourceName() string {
	return c.EntrySourceName
}

// SetCompilerVersion sets the version of the solc compiler to use.
func (c *CompilerConfig) SetCompilerVersion(version string) {
	c.CompilerVersion = version
}

// GetCompilerVersion returns the currently set version of the solc compiler.
func (c *CompilerConfig) GetCompilerVersion() string {
	return c.CompilerVersion
}

// SanitizeArguments sanitizes the provided arguments against a list of allowed arguments.
// Returns an error if any of the provided arguments are not in the allowed list.
func (c *CompilerConfig) SanitizeArguments(args []string) ([]string, error) {
	var sanitizedArgs []string
	for _, arg := range args {
		if strings.Contains(arg, "-") {
			if _, ok := allowedArgs[arg]; !ok {
				return nil, fmt.Errorf("invalid argument: %s", arg)
			}
		}
		sanitizedArgs = append(sanitizedArgs, arg)
	}
	return sanitizedArgs, nil
}

// Validate checks if the current CompilerConfiguration's arguments are valid.
// It ensures that all required arguments are present.
func (c *CompilerConfig) Validate() error {
	sanitized, err := c.SanitizeArguments(c.Arguments)
	if err != nil {
		return err
	}

	// Convert the sanitized slice into a map for easier lookup.
	sanitizedMap := make(map[string]bool)
	for _, arg := range sanitized {
		sanitizedMap[arg] = true
	}

	for arg := range requiredArgs {
		if _, ok := sanitizedMap[arg]; !ok {
			return fmt.Errorf("missing required argument: %s", arg)
		}
	}

	matched, _ := regexp.MatchString(`^(\d+\.\d+\.\d+)$`, c.CompilerVersion)
	if !matched {
		return fmt.Errorf("invalid compiler version: %s", c.CompilerVersion)
	}

	return nil
}

// SetArguments sets the arguments to be passed to the solc tool.
func (c *CompilerConfig) SetArguments(args []string) {
	c.Arguments = args
}

// AppendArguments appends new arguments to the existing set of arguments.
func (c *CompilerConfig) AppendArguments(args ...string) {
	c.Arguments = append(c.Arguments, args...)
}

// GetArguments returns the arguments to be passed to the solc tool.
func (c *CompilerConfig) GetArguments() []string {
	return c.Arguments
}
