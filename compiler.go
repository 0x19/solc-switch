package solc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"go.uber.org/zap"
)

// Compiler represents a Solidity compiler instance.
type Compiler struct {
	ctx    context.Context // The context for the compiler.
	source string          // The Solidity sources to compile.
	solc   *Solc           // The solc instance.
	config *CompilerConfig // The configuration for the compiler.
}

// NewCompiler creates a new Compiler instance with the given context, configuration, and source.
func NewCompiler(ctx context.Context, solc *Solc, config *CompilerConfig, source string) (*Compiler, error) {
	if config == nil {
		return nil, fmt.Errorf("config must be provided to create new compiler")
	}

	if solc == nil {
		return nil, fmt.Errorf("solc instance must be provided to create new compiler")
	}

	if source == "" {
		return nil, fmt.Errorf("source code must be provided to create new compiler")
	}

	if config.JsonConfig == nil {
		if err := config.Validate(); err != nil {
			return nil, fmt.Errorf("invalid compiler configuration: %w", err)
		}
	}

	return &Compiler{
		ctx:    ctx,
		source: source,
		config: config,
		solc:   solc,
	}, nil
}

// SetCompilerVersion sets the version of the solc compiler to use.
func (v *Compiler) SetCompilerVersion(version string) {
	v.config.SetCompilerVersion(version)
}

// GetCompilerVersion returns the currently set version of the solc compiler.
func (v *Compiler) GetCompilerVersion() string {
	return v.config.GetCompilerVersion()
}

// GetContext returns the context associated with the compiler.
func (v *Compiler) GetContext() context.Context {
	return v.ctx
}

// GetSources returns the Solidity sources associated with the compiler.
func (v *Compiler) GetSources() string {
	return v.source
}

// Compile compiles the Solidity sources using the configured compiler version and arguments.
func (v *Compiler) Compile() ([]*CompilerResults, error) {
	compilerVersion := v.GetCompilerVersion()
	if compilerVersion == "" {
		return nil, fmt.Errorf("no compiler version specified")
	}

	binaryPath, err := v.solc.GetBinary(compilerVersion)
	if err != nil {
		return nil, err
	}

	args := []string{}
	sanitizedArgs, err := v.config.SanitizeArguments(v.config.Arguments)
	if err != nil {
		return nil, err
	}
	args = append(args, sanitizedArgs...)

	if v.config.JsonConfig == nil {
		if err := v.config.Validate(); err != nil {
			return nil, err
		}
	}

	// #nosec G204
	// G204 (CWE-78): Subprocess launched with variable (Confidence: HIGH, Severity: MEDIUM)
	// We did sanitization and verification of the arguments above, so we are safe to use them.
	cmd := exec.Command(binaryPath, args...)

	cmd.Stdin = strings.NewReader(v.source)

	// Capture the output
	var out bytes.Buffer
	cmd.Stdout = &out
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		zap.L().Error(
			"Failed to compile Solidity sources",
			zap.String("version", compilerVersion),
			zap.String("stderr", stderr.String()),
		)
		var errors []string
		var warnings []string

		// Parsing the error message to extract line and column information.
		errorMessage := stderr.String()
		if strings.Contains(errorMessage, "Error:") {
			errors = append(errors, errorMessage)
		} else if strings.HasPrefix(errorMessage, "Warning:") {
			warnings = append(warnings, errorMessage)
		}

		// Construct the CompilerResults structure with errors and warnings.
		results := &CompilerResults{
			RequestedVersion: compilerVersion,
			Errors:           errors,
			Warnings:         warnings,
		}
		return []*CompilerResults{results}, err
	}

	// Parse the output
	var compilationOutput struct {
		Contracts map[string]map[string]struct {
			Abi interface{} `json:"abi"`
			Evm struct {
				Bytecode struct {
					GeneratedSources []interface{}          `json:"generatedSources"`
					LinkReferences   map[string]interface{} `json:"linkReferences"`
					Object           string                 `json:"object"`
					Opcodes          string                 `json:"opcodes"`
					SourceMap        string                 `json:"sourceMap"`
				} `json:"bytecode"`
			} `json:"evm"`
			Metadata string `json:"metadata"`
		} `json:"contracts"`
		Errors  []string `json:"errors"`
		Version string   `json:"version"`
	}

	err = json.Unmarshal(out.Bytes(), &compilationOutput)
	if err != nil {
		return nil, err
	}

	// Separate errors and warnings
	var errors, warnings []string
	for _, msg := range compilationOutput.Errors {
		if strings.Contains(msg, "Warning:") {
			warnings = append(warnings, msg)
		} else {
			errors = append(errors, msg)
		}
	}

	var results []*CompilerResults

	for key := range compilationOutput.Contracts {
		for key, output := range compilationOutput.Contracts[key] {
			isEntryContract := false
			if v.config.GetEntrySourceName() != "" && key == v.config.GetEntrySourceName() {
				isEntryContract = true
			}

			abi, err := json.Marshal(output.Abi)
			if err != nil {
				return nil, err
			}

			results = append(results, &CompilerResults{
				IsEntryContract:  isEntryContract,
				RequestedVersion: compilerVersion,
				Bytecode:         output.Evm.Bytecode.Object,
				ABI:              string(abi),
				Opcodes:          output.Evm.Bytecode.Opcodes,
				ContractName:     key,
				Errors:           errors,
				Warnings:         warnings,
				Metadata:         output.Metadata,
			})
		}
	}

	return results, nil
}

// CompilerResults represents the results of a solc compilation.
type CompilerResults struct {
	IsEntryContract  bool     `json:"is_entry_contract"`
	RequestedVersion string   `json:"requested_version"`
	CompilerVersion  string   `json:"compiler_version"`
	ContractName     string   `json:"contract_name"`
	Bytecode         string   `json:"bytecode"`
	ABI              string   `json:"abi"`
	Opcodes          string   `json:"opcodes"`
	Metadata         string   `json:"metadata"`
	Errors           []string `json:"errors"`
	Warnings         []string `json:"warnings"`
}

// HasErrors returns true if there are compilation errors.
func (v *CompilerResults) HasErrors() bool {
	if v == nil {
		return false
	}

	return len(v.Errors) > 0
}

// HasWarnings returns true if there are compilation warnings.
func (v *CompilerResults) HasWarnings() bool {
	if v == nil {
		return false
	}

	return len(v.Warnings) > 0
}

// GetErrors returns the compilation errors.
func (v *CompilerResults) GetErrors() []string {
	return v.Errors
}

// GetWarnings returns the compilation warnings.
func (v *CompilerResults) GetWarnings() []string {
	return v.Warnings
}

// GetABI returns the compiled contract's ABI (Application Binary Interface) in JSON format.
func (v *CompilerResults) GetABI() string {
	return v.ABI
}

// GetBytecode returns the compiled contract's bytecode.
func (v *CompilerResults) GetBytecode() string {
	return v.Bytecode
}

// GetContractName returns the name of the compiled contract.
func (v *CompilerResults) GetContractName() string {
	return v.ContractName
}

// GetRequestedVersion returns the requested compiler version used for compilation.
func (v *CompilerResults) GetRequestedVersion() string {
	return v.RequestedVersion
}

// GetCompilerVersion returns the actual compiler version used for compilation.
func (v *CompilerResults) GetCompilerVersion() string {
	return v.CompilerVersion
}
