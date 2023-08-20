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
				if compilerResults != nil {
					assert.True(t, compilerResults.HasErrors())
					assert.False(t, compilerResults.HasWarnings())
					assert.GreaterOrEqual(t, len(compilerResults.GetWarnings()), 0)
					assert.GreaterOrEqual(t, len(compilerResults.GetErrors()), 1)
				}

				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, compilerResults)

			assert.NotEmpty(t, compilerResults.GetRequestedVersion())
			assert.NotEmpty(t, compilerResults.GetCompilerVersion())
			assert.NotEmpty(t, compilerResults.GetBytecode())
			assert.NotEmpty(t, compilerResults.GetABI())
			assert.NotEmpty(t, compilerResults.GetContractName())
			assert.GreaterOrEqual(t, len(compilerResults.GetWarnings()), 0)
			assert.GreaterOrEqual(t, len(compilerResults.GetErrors()), 0)
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
				if compilerResults != nil {
					assert.True(t, compilerResults.HasErrors())
					assert.False(t, compilerResults.HasWarnings())
					assert.GreaterOrEqual(t, len(compilerResults.GetWarnings()), 0)
					assert.GreaterOrEqual(t, len(compilerResults.GetErrors()), 1)
				}

				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, compilerResults)

			assert.NotEmpty(t, compilerResults.GetRequestedVersion())
			assert.NotEmpty(t, compilerResults.GetCompilerVersion())
			assert.NotEmpty(t, compilerResults.GetBytecode())
			assert.NotEmpty(t, compilerResults.GetABI())
			assert.NotEmpty(t, compilerResults.GetContractName())
			assert.GreaterOrEqual(t, len(compilerResults.GetWarnings()), 0)
			assert.GreaterOrEqual(t, len(compilerResults.GetErrors()), 0)
		})
	}
}
