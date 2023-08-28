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
// It returns an error if the provided configuration, solc instance, or source is invalid.
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
// It returns the compilation results or an error if the compilation fails.
func (v *Compiler) Compile() (*CompilerResults, error) {
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
		var errors []CompilationError
		var warnings []CompilationError

		// Parsing the error message to extract line and column information.
		errorMessage := stderr.String()
		if strings.Contains(errorMessage, "Error:") {
			errors = append(errors, CompilationError{Message: errorMessage})
		} else if strings.HasPrefix(errorMessage, "Warning:") {
			warnings = append(warnings, CompilationError{Message: errorMessage})
		}

		// Construct the CompilerResults structure with errors and warnings.
		results := &CompilerResult{
			RequestedVersion: compilerVersion,
			Errors:           errors,
			Warnings:         warnings,
		}
		return &CompilerResults{Results: []*CompilerResult{results}}, err
	}

	if v.config.JsonConfig != nil {
		return v.resultsFromJson(compilerVersion, out)
	}

	return v.resultsFromSimple(compilerVersion, out)
}

// resultsFromSimple parses the output from the solc compiler when the output is in a simple format.
// It extracts the compilation details such as bytecode, ABI, and any errors or warnings.
// The method returns a slice of CompilerResults or an error if the output cannot be parsed.
func (v *Compiler) resultsFromSimple(compilerVersion string, out bytes.Buffer) (*CompilerResults, error) {
	// Parse the output
	var compilationOutput struct {
		Contracts map[string]struct {
			Bin string      `json:"bin"`
			Abi interface{} `json:"abi"`
		} `json:"contracts"`
		Errors  []string `json:"errors"`
		Version string   `json:"version"`
	}

	if err := json.Unmarshal(out.Bytes(), &compilationOutput); err != nil {
		return nil, err
	}

	// Separate errors and warnings
	var errors, warnings []CompilationError
	for _, msg := range compilationOutput.Errors {
		if strings.Contains(msg, "Warning:") {
			warnings = append(warnings, CompilationError{Message: msg})
		} else {
			errors = append(errors, CompilationError{Message: msg})
		}
	}

	var results []*CompilerResult

	for key, output := range compilationOutput.Contracts {
		isEntryContract := false
		if v.config.GetEntrySourceName() != "" && key == "<stdin>:"+v.config.GetEntrySourceName() {
			isEntryContract = true
		}

		abi, err := json.Marshal(output.Abi)
		if err != nil {
			return nil, err
		}

		results = append(results, &CompilerResult{
			IsEntryContract:  isEntryContract,
			RequestedVersion: compilerVersion,
			CompilerVersion:  compilationOutput.Version,
			Bytecode:         output.Bin,
			ABI:              string(abi),
			ContractName:     strings.TrimLeft(key, "<stdin>:"),
			Errors:           errors,
			Warnings:         warnings,
		})
	}

	return &CompilerResults{Results: results}, nil
}

// resultsFromJson parses the output from the solc compiler when the output is in a JSON format.
// It extracts detailed compilation information including bytecode, ABI, opcodes, and metadata.
// Additionally, it separates any errors and warnings from the compilation process.
// The method returns a slice of CompilerResults or an error if the output cannot be parsed.
func (v *Compiler) resultsFromJson(compilerVersion string, out bytes.Buffer) (*CompilerResults, error) {
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
				DeployedBytecode struct {
					GeneratedSources []interface{}          `json:"generatedSources"`
					LinkReferences   map[string]interface{} `json:"linkReferences"`
					Object           string                 `json:"object"`
					Opcodes          string                 `json:"opcodes"`
					SourceMap        string                 `json:"sourceMap"`
				} `json:"deployedBytecode"`
			} `json:"evm"`
			Metadata string `json:"metadata"`
		} `json:"contracts"`
		Errors  []CompilationError `json:"errors"`
		Version string             `json:"version"`
	}

	if err := json.Unmarshal(out.Bytes(), &compilationOutput); err != nil {
		return nil, err
	}

	var results []*CompilerResult

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

			results = append(results, &CompilerResult{
				IsEntryContract:  isEntryContract,
				RequestedVersion: compilerVersion,
				Bytecode:         output.Evm.Bytecode.Object,
				DeployedBytecode: output.Evm.DeployedBytecode.Object,
				ABI:              string(abi),
				Opcodes:          output.Evm.Bytecode.Opcodes,
				ContractName:     key,
				Errors:           compilationOutput.Errors,
				Metadata:         output.Metadata,
			})
		}
	}

	if len(compilationOutput.Errors) > 0 {
		results = append(results, &CompilerResult{
			RequestedVersion: compilerVersion,
			Errors:           compilationOutput.Errors,
		})
	}

	return &CompilerResults{Results: results}, nil
}

// CompilationError represents a compilation error.
type CompilationError struct {
	Component string `json:"component"`
	Formatted string `json:"formattedMessage"`
	Message   string `json:"message"`
	Severity  string `json:"severity"`
	Type      string `json:"type"`
}

type CompilerResults struct {
	Results []*CompilerResult `json:"results"`
}

func (cr *CompilerResults) GetResults() []*CompilerResult {
	return cr.Results
}

func (cr *CompilerResults) GetEntryContract() *CompilerResult {
	for _, result := range cr.Results {
		if result.IsEntry() {
			return result
		}
	}

	return nil
}

// CompilerResults represents the results of a solc compilation.
type CompilerResult struct {
	IsEntryContract  bool               `json:"is_entry_contract"`
	RequestedVersion string             `json:"requested_version"`
	CompilerVersion  string             `json:"compiler_version"`
	ContractName     string             `json:"contract_name"`
	Bytecode         string             `json:"bytecode"`
	DeployedBytecode string             `json:"deployedBytecode"`
	ABI              string             `json:"abi"`
	Opcodes          string             `json:"opcodes"`
	Metadata         string             `json:"metadata"`
	Errors           []CompilationError `json:"errors"`
	Warnings         []CompilationError `json:"warnings"`
}

// IsEntry returns true if the compiled contract is the entry contract.
func (v *CompilerResult) IsEntry() bool {
	return v.IsEntryContract
}

// GetOpcodes returns the compiled contract's opcodes.
func (v *CompilerResult) GetOpcodes() string {
	return v.Opcodes
}

// GetMetadata returns the compiled contract's metadata.
func (v *CompilerResult) GetMetadata() string {
	return v.Metadata
}

// HasErrors returns true if there are compilation errors.
func (v *CompilerResult) HasErrors() bool {
	if v == nil {
		return false
	}

	return len(v.Errors) > 0
}

// HasWarnings returns true if there are compilation warnings.
func (v *CompilerResult) HasWarnings() bool {
	if v == nil {
		return false
	}

	return len(v.Warnings) > 0
}

// GetErrors returns the compilation errors.
func (v *CompilerResult) GetErrors() []CompilationError {
	return v.Errors
}

// GetWarnings returns the compilation warnings.
func (v *CompilerResult) GetWarnings() []CompilationError {
	return v.Warnings
}

// GetABI returns the compiled contract's ABI (Application Binary Interface) in JSON format.
func (v *CompilerResult) GetABI() string {
	return v.ABI
}

// GetBytecode returns the compiled contract's bytecode.
func (v *CompilerResult) GetBytecode() string {
	return v.Bytecode
}

// GetDeployedBytecode returns the compiled contract's deployed bytecode.
func (v *CompilerResult) GetDeployedBytecode() string {
	return v.DeployedBytecode
}

// GetContractName returns the name of the compiled contract.
func (v *CompilerResult) GetContractName() string {
	return v.ContractName
}

// GetRequestedVersion returns the requested compiler version used for compilation.
func (v *CompilerResult) GetRequestedVersion() string {
	return v.RequestedVersion
}

// GetCompilerVersion returns the actual compiler version used for compilation.
func (v *CompilerResult) GetCompilerVersion() string {
	return v.CompilerVersion
}
