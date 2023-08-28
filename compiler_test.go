package solc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestCompiler(t *testing.T) {
	logger, err := GetDevelopmentLogger(zapcore.DebugLevel)
	assert.NoError(t, err)
	assert.NotNil(t, logger)
	zap.ReplaceGlobals(logger)

	// Replace the global logger.
	zap.ReplaceGlobals(logger)

	solcConfig, err := NewDefaultConfig()
	assert.NoError(t, err)
	assert.NotNil(t, solcConfig)

	solc, err := New(context.TODO(), solcConfig)
	assert.NoError(t, err)
	assert.NotNil(t, solc)

	testCases := []struct {
		name           string
		source         string
		wantErr        bool
		wantCompileErr bool
		compilerConfig *CompilerConfig
		sync           bool
		solc           *Solc
	}{
		{
			name: "Valid Source",
			source: `// SPDX-License-Identifier: MIT
			pragma solidity ^0.8.0;
			
			contract SimpleStorage {
				uint256 private storedData;
			
				function set(uint256 x) public {
					storedData = x;
				}
			
				function get() public view returns (uint256) {
					return storedData;
				}
			}`,
			wantCompileErr: false,
			compilerConfig: func() *CompilerConfig {
				config, err := NewDefaultCompilerConfig("0.8.0")
				assert.NoError(t, err)
				assert.NotNil(t, config)
				return config
			}(),
			solc: solc,
			sync: true,
		},
		{
			name: "Invalid Source",
			source: `// SPDX-License-Identifier: MIT
			pragma solidity ^0.8.0;
			
			contract SimpleStorage {
				uint256 private storedData;
			
				function set(uint256 x) public {
				
			}`,
			wantCompileErr: true,
			compilerConfig: func() *CompilerConfig {
				config, err := NewDefaultCompilerConfig("0.8.0")
				assert.NoError(t, err)
				assert.NotNil(t, config)
				return config
			}(),
			solc: solc,
			sync: false,
		},
		{
			name: "Invalid Compiler Version",
			source: `// SPDX-License-Identifier: MIT
			pragma solidity ^0.8.0;
			
			contract SimpleStorage {
				uint256 private storedData;
			
				function set(uint256 x) public {
					storedData = x;
				}
			
				function get() public view returns (uint256) {
					return storedData;
				}
			}`,
			wantCompileErr: true,
			compilerConfig: func() *CompilerConfig {
				config, err := NewDefaultCompilerConfig("0.4.11")
				assert.NoError(t, err)
				assert.NotNil(t, config)
				return config
			}(),
			solc: solc,
			sync: false,
		},
		{
			name: "Invalid Compiler Config Version",
			source: `// SPDX-License-Identifier: MIT
			pragma solidity ^0.8.0;
			
			contract SimpleStorage {
				uint256 private storedData;
			
				function set(uint256 x) public {
					storedData = x;
				}
			
				function get() public view returns (uint256) {
					return storedData;
				}
			}`,
			wantErr:        true,
			wantCompileErr: true,
			compilerConfig: func() *CompilerConfig {
				return nil
			}(),
			solc: solc,
			sync: false,
		},
		{
			name: "Invalid Compiler Solc Instance",
			source: `// SPDX-License-Identifier: MIT
			pragma solidity ^0.8.0;
			
			contract SimpleStorage {
				uint256 private storedData;
			
				function set(uint256 x) public {
					storedData = x;
				}
			
				function get() public view returns (uint256) {
					return storedData;
				}
			}`,
			wantErr:        true,
			wantCompileErr: true,
			compilerConfig: func() *CompilerConfig {
				config, err := NewDefaultCompilerConfig("0.4.11")
				assert.NoError(t, err)
				assert.NotNil(t, config)
				return config
			}(),
			solc: nil,
			sync: false,
		},
		{
			name:           "Invalid Compiler source",
			wantErr:        true,
			wantCompileErr: true,
			compilerConfig: func() *CompilerConfig {
				config, err := NewDefaultCompilerConfig("0.4.11")
				assert.NoError(t, err)
				assert.NotNil(t, config)
				return config
			}(),
			solc: solc,
			sync: false,
		},
		{
			name:           "Invalid Compiler config",
			wantErr:        true,
			wantCompileErr: true,
			compilerConfig: func() *CompilerConfig {
				return nil
			}(),
			solc: solc,
			sync: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			compiler, err := NewCompiler(context.Background(), testCase.solc, testCase.compilerConfig, testCase.source)
			if testCase.wantErr {
				assert.Error(t, err)
				assert.Nil(t, compiler)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, compiler)
			assert.NotNil(t, compiler.GetContext())
			assert.NotNil(t, compiler.GetSources())

			// In case we drop releases path ability to test that it syncs successfully prior
			// to compiling.
			if testCase.sync {
				err := solc.Sync()
				assert.NoError(t, err)

				// Just so the function is tested, nothing else...
				compiler.SetCompilerVersion(compiler.GetCompilerVersion())
				currentVersion := compiler.GetCompilerVersion()
				assert.NotEmpty(t, currentVersion)
			}

			compilerResults, err := compiler.Compile()
			if testCase.wantCompileErr {
				for _, result := range compilerResults.GetResults() {
					assert.True(t, result.HasErrors())
					assert.False(t, result.HasWarnings())
					assert.GreaterOrEqual(t, len(result.GetWarnings()), 0)
					assert.GreaterOrEqual(t, len(result.GetErrors()), 1)
				}

				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, compilerResults)

			for _, result := range compilerResults.GetResults() {
				assert.NotEmpty(t, result.GetRequestedVersion())
				assert.NotEmpty(t, result.GetCompilerVersion())
				assert.NotEmpty(t, result.GetBytecode())
				assert.NotEmpty(t, result.GetABI())
				assert.NotEmpty(t, result.GetContractName())
				assert.GreaterOrEqual(t, len(result.GetWarnings()), 0)
				assert.GreaterOrEqual(t, len(result.GetErrors()), 0)
			}
		})
	}
}

func TestCompilerFromSolc(t *testing.T) {
	logger, err := GetDevelopmentLogger(zapcore.ErrorLevel)
	assert.NoError(t, err)
	assert.NotNil(t, logger)
	zap.ReplaceGlobals(logger)

	zap.ReplaceGlobals(logger)

	solcConfig, err := NewDefaultConfig()
	assert.NoError(t, err)
	assert.NotNil(t, solcConfig)

	solc, err := New(context.TODO(), solcConfig)
	assert.NoError(t, err)
	assert.NotNil(t, solc)

	testCases := []struct {
		name           string
		source         string
		wantErr        bool
		wantCompileErr bool
		compilerConfig *CompilerConfig
		sync           bool
		solc           *Solc
	}{
		{
			name: "Valid Source",
			source: `// SPDX-License-Identifier: MIT
			pragma solidity ^0.8.0;
			
			contract SimpleStorage {
				uint256 private storedData;
			
				function set(uint256 x) public {
					storedData = x;
				}
			
				function get() public view returns (uint256) {
					return storedData;
				}
			}`,
			wantCompileErr: false,
			compilerConfig: func() *CompilerConfig {
				config, err := NewDefaultCompilerConfig("0.8.0")
				assert.NoError(t, err)
				assert.NotNil(t, config)
				return config
			}(),
			solc: solc,
			sync: true,
		},
		{
			name: "Invalid Source",
			source: `// SPDX-License-Identifier: MIT
			pragma solidity ^0.8.0;

			contract SimpleStorage {
				uint256 private storedData;

				function set(uint256 x) public {

			}`,
			wantCompileErr: true,
			compilerConfig: func() *CompilerConfig {
				config, err := NewDefaultCompilerConfig("0.8.0")
				assert.NoError(t, err)
				assert.NotNil(t, config)
				return config
			}(),
			solc: solc,
			sync: false,
		},
		{
			name: "Invalid Compiler Version",
			source: `// SPDX-License-Identifier: MIT
			pragma solidity ^0.8.0;

			contract SimpleStorage {
				uint256 private storedData;

				function set(uint256 x) public {
					storedData = x;
				}

				function get() public view returns (uint256) {
					return storedData;
				}
			}`,
			wantCompileErr: true,
			compilerConfig: func() *CompilerConfig {
				config, err := NewDefaultCompilerConfig("0.4.11")
				assert.NoError(t, err)
				assert.NotNil(t, config)
				return config
			}(),
			solc: solc,
			sync: false,
		},
		{
			name: "Invalid Compiler Config Version",
			source: `// SPDX-License-Identifier: MIT
			pragma solidity ^0.8.0;

			contract SimpleStorage {
				uint256 private storedData;

				function set(uint256 x) public {
					storedData = x;
				}

				function get() public view returns (uint256) {
					return storedData;
				}
			}`,
			wantErr:        true,
			wantCompileErr: true,
			compilerConfig: func() *CompilerConfig {
				return nil
			}(),
			solc: solc,
			sync: false,
		},
		{
			name: "Invalid Compiler Solc Instance",
			source: `// SPDX-License-Identifier: MIT
			pragma solidity ^0.8.0;

			contract SimpleStorage {
				uint256 private storedData;

				function set(uint256 x) public {
					storedData = x;
				}

				function get() public view returns (uint256) {
					return storedData;
				}
			}`,
			wantErr:        true,
			wantCompileErr: true,
			compilerConfig: func() *CompilerConfig {
				config, err := NewDefaultCompilerConfig("0.4.11")
				assert.NoError(t, err)
				assert.NotNil(t, config)
				return config
			}(),
			solc: nil,
			sync: false,
		},
		{
			name:           "Invalid Compiler source",
			wantErr:        true,
			wantCompileErr: true,
			compilerConfig: func() *CompilerConfig {
				config, err := NewDefaultCompilerConfig("0.4.11")
				assert.NoError(t, err)
				assert.NotNil(t, config)
				return config
			}(),
			solc: solc,
			sync: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {

			// In case we drop releases path ability to test that it syncs successfully prior
			// to compiling.
			if testCase.sync {
				err := solc.Sync()
				assert.NoError(t, err)
			}

			compilerResults, err := solc.Compile(context.TODO(), testCase.source, testCase.compilerConfig)
			if testCase.wantCompileErr {
				assert.Nil(t, compilerResults)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, compilerResults)

			for _, result := range compilerResults.GetResults() {
				assert.NotEmpty(t, result.GetRequestedVersion())
				assert.NotEmpty(t, result.GetCompilerVersion())
				assert.NotEmpty(t, result.GetBytecode())
				assert.NotEmpty(t, result.GetABI())
				assert.NotEmpty(t, result.GetContractName())
				assert.GreaterOrEqual(t, len(result.GetWarnings()), 0)
				assert.GreaterOrEqual(t, len(result.GetErrors()), 0)
			}
		})
	}
}

func TestCompilerWithJSON(t *testing.T) {
	logger, err := GetDevelopmentLogger(zapcore.DebugLevel)
	assert.NoError(t, err)
	assert.NotNil(t, logger)
	zap.ReplaceGlobals(logger)

	// Replace the global logger.
	zap.ReplaceGlobals(logger)

	solcConfig, err := NewDefaultConfig()
	assert.NoError(t, err)
	assert.NotNil(t, solcConfig)

	solc, err := New(context.TODO(), solcConfig)
	assert.NoError(t, err)
	assert.NotNil(t, solc)

	testCases := []struct {
		name           string
		wantErr        bool
		wantCompileErr bool
		compilerConfig *CompilerConfig
		sync           bool
		solc           *Solc
	}{
		{
			name:           "Valid Source",
			wantCompileErr: false,
			compilerConfig: func() *CompilerConfig {
				jsonConfig := &CompilerJsonConfig{
					Language: "Solidity",
					Sources: map[string]Source{
						"SimpleStorage.sol": {
							Content: `// SPDX-License-Identifier: MIT
							pragma solidity ^0.8.0;
							
							contract SimpleStorage {
								uint256 private storedData;
							
								function set(uint256 x) public {
									storedData = x;
								}
							
								function get() public view returns (uint256) {
									return storedData;
								}
							}`,
						},
					},
					Settings: Settings{
						Optimizer: Optimizer{
							Enabled: false,
							Runs:    200,
						},
						OutputSelection: map[string]map[string][]string{
							"*": {
								"*": []string{
									"abi",
									"evm.bytecode",
									"evm.runtimeBytecode",
									"metadata",
									"evm.deployedBytecode*",
								},
							},
						},
					},
				}

				config, err := NewCompilerConfigFromJSON("0.8.0", "SimpleStorage", jsonConfig)
				assert.NoError(t, err)
				assert.NotNil(t, config)
				return config
			}(),
			solc: solc,
			sync: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			source, err := testCase.compilerConfig.GetJsonConfig().ToJSON()
			assert.NoError(t, err)
			assert.NotNil(t, source)

			compiler, err := NewCompiler(context.Background(), testCase.solc, testCase.compilerConfig, string(source))
			if testCase.wantErr {
				assert.Error(t, err)
				assert.Nil(t, compiler)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, compiler)
			assert.NotNil(t, compiler.GetContext())
			assert.NotNil(t, compiler.GetSources())

			// In case we drop releases path ability to test that it syncs successfully prior
			// to compiling.
			if testCase.sync {
				err := solc.Sync()
				assert.NoError(t, err)

				// Just so the function is tested, nothing else...
				compiler.SetCompilerVersion(compiler.GetCompilerVersion())
				currentVersion := compiler.GetCompilerVersion()
				assert.NotEmpty(t, currentVersion)
			}

			compilerResults, err := compiler.Compile()
			if testCase.wantCompileErr {
				for _, result := range compilerResults.GetResults() {
					assert.True(t, result.HasErrors())
					assert.False(t, result.HasWarnings())
					assert.GreaterOrEqual(t, len(result.GetWarnings()), 0)
					assert.GreaterOrEqual(t, len(result.GetErrors()), 1)
				}

				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, compilerResults)

			for _, result := range compilerResults.GetResults() {
				assert.NotNil(t, result.IsEntry())
				assert.NotEmpty(t, result.GetRequestedVersion())
				assert.NotEmpty(t, result.GetBytecode())
				assert.NotEmpty(t, result.GetABI())
				assert.NotEmpty(t, result.GetContractName())
				assert.NotEmpty(t, result.GetOpcodes())
				assert.NotEmpty(t, result.GetMetadata())
				assert.GreaterOrEqual(t, len(result.GetWarnings()), 0)
				assert.GreaterOrEqual(t, len(result.GetErrors()), 0)
			}
		})
	}
}
