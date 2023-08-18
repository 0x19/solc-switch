[![Build Status](https://github.com/0x19/solc-switch/actions/workflows/test.yml/badge.svg)](https://github.com/0x19/solc-switch/actions/workflows/test.yml)
[![Security Status](https://github.com/0x19/solc-switch/actions/workflows/gosec.yml/badge.svg)](https://github.com/0x19/solc-switch/actions/workflows/gosec.yml)
[![Coverage Status](https://coveralls.io/repos/github/0x19/solc-switch/badge.svg?branch=main)](https://coveralls.io/github/0x19/solc-switch?branch=main)
[![Go Report Card](https://goreportcard.com/badge/github.com/0x19/solc-switch)](https://goreportcard.com/report/github.com/0x19/solc-switch)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/0x19/solc-switch)](https://pkg.go.dev/github.com/0x19/solc-switch)

# Solc-Switch: A Concurrent Solidity Compiler Manager in Go

**solc-switch** is more than just a tool; it's your partner in streamlining the Solidity development process. Built with Go, it's designed to let developers seamlessly switch between different Solidity compiler versions, ensuring optimal compatibility and a frictionless development experience.

While **solc-switch** is lightweight, fast, and intuitive, it's primarily tailored for APIs and other tool integrations. It's not primarily intended for direct end-user interactions. If you're on the hunt for a CLI-based solution, consider exploring [solc-select](https://github.com/crytic/solc-select).

## Highlighted Features

- **Speedy Downloads:** Parallel downloading of multiple Solidity compiler versions ensures you're not waiting around. In fact, all releases are already pre-downloaded for you. You can see them in [releases](./releases/) directory.
- **Smooth Version Switching:** Navigate between different Solidity compiler versions without a hitch.
- **Always in the Loop:** With automatic updates, always have access to the latest releases from the official Solidity GitHub repository.
- **Broad Compatibility:** Crafted with MacOS and Linux in mind, and potentially adaptable for Windows.

📢 **Note for Windows Enthusiasts:** Testing on a Windows environment hasn't been done yet. If you're a Windows user, your feedback on its functionality would be invaluable. If you're passionate about contributing, your PRs are more than welcome!

## Documentation

For a comprehensive overview of `solc-switch`, its functions, and detailed usage instructions, please refer to the [official documentation](https://pkg.go.dev/github.com/0x19/solc-switch).

## Installation

To use solc-switch in your project, you don't need a separate installation. Simply import it directly:

```go
import "github.com/0x19/solc-switch"
```

### Setting Up Environment Variable

Before you start, ensure you've set up the required environment variable for the GitHub personal access token:

```bash
export SOLC_SWITCH_GITHUB_TOKEN="{your_github_token_here}"
```

Replace **{your_github_token_here}** with your actual GitHub personal access token. If you don't have one, you can create it by following the instructions [here](https://docs.github.com/en/github/authenticating-to-github/creating-a-personal-access-token).

It is used to fetch the latest Solidity compiler releases from the official Solidity GitHub repository. If you don't set it up, you'll get an rate limit error quite quickly.

## Example Usage

Below is a simple example demonstrating how to use **solc-switch** to compile a Solidity smart contract:

```go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/0x19/solc-switch"
	"github.com/davecgh/go-spew/spew"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	// Sample Solidity contract for demonstration purposes
	sourceCode := `
	// SPDX-License-Identifier: MIT
	pragma solidity ^0.8.0;

	contract SimpleStorage {
		uint256 private storedData;

		function set(uint256 x) public {
			storedData = x;
		}

		function get() public view returns (uint256) {
			return storedData;
		}
	}
	`

	// Create a context with cancellation capabilities
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize a development logger for solc-switch
	logger, err := solc.GetDevelopmentLogger(zapcore.ErrorLevel)
	if err != nil {
		panic(err)
	}
	zap.ReplaceGlobals(logger)

	// Create a default configuration for solc-switch
	solcConfig, err := solc.NewDefaultConfig()
	if err != nil {
		panic(err)
	}

	// Initialize the solc-switch manager with the provided configuration
	solcManager, err := solc.New(ctx, solcConfig)
	if err != nil {
		panic(err)
	}

	// Increase the default HTTP client timeout to ensure successful fetching of compiler versions
	solcConfig.SetHttpClientTimeout(60 * time.Second)

	// Sync with the official Solidity GitHub repository to ensure the latest compiler versions are available locally
	if err := solcManager.Sync(); err != nil {
		panic(err)
	}

	// Revert the HTTP client timeout to a shorter duration post-sync
	solcConfig.SetHttpClientTimeout(10 * time.Second)

	// Create a default compiler configuration targeting Solidity version "0.8.0"
	compilerConfig, err := solc.NewDefaultCompilerConfig("0.8.0")
	if err != nil {
		panic(err)
	}

	// Start the timer to measure the compilation time
	startTime := time.Now()

	// Compile the provided Solidity source code using the specified compiler configuration
	results, err := solcManager.Compile(ctx, sourceCode, compilerConfig)
	if err != nil {
		panic(err)
	}

	// Calculate the time taken for the compilation
	timeTaken := time.Since(startTime)

	// Display the compilation results for review
	spew.Dump(results)

	// Print the time taken for the compilation
	fmt.Printf("Time took: %s\n", timeTaken)
}

```

### Compilation Results

- **Requested Compiler Version:** 0.8.0
- **Compiler Version:** 0.8.0+commit.c7dfd78e.Linux.g++
- **Contract Name:** SimpleStorage

```shell
(*solc.CompilerResults)(0xc0002faa80)({
 RequestedVersion: (string) (len=5) "0.8.0",
 CompilerVersion: (string) (len=31) "0.8.0+commit.c7dfd78e.Linux.g++",
 Bytecode: (string) (len=670) "608060405234801561001057600080fd5b5061012f806100206000396000f3fe6080604052348015600f57600080fd5b506004361060325760003560e01c806360fe47b11460375780636d4ce63c14604f575b600080fd5b604d600480360381019060499190608f565b6069565b005b60556073565b6040516060919060c2565b60405180910390f35b8060008190555050565b60008054905090565b60008135905060898160e5565b92915050565b60006020828403121560a057600080fd5b600060ac84828501607c565b91505092915050565b60bc8160db565b82525050565b600060208201905060d5600083018460b5565b92915050565b6000819050919050565b60ec8160db565b811460f657600080fd5b5056fea26469706673582212202155718f049537eeb007a6f90d13320064bd4989f47c4ff85c9042cc2257550164736f6c63430008000033",
 ABI: (string) (len=280) "[{\"inputs\":[],\"name\":\"get\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"x\",\"type\":\"uint256\"}],\"name\":\"set\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
 ContractName: (string) (len=13) "SimpleStorage",
 Errors: ([]string) <nil>,
 Warnings: ([]string) <nil>
})
```

### Compilation Time

```shell
Time took: 3.309062ms
```

## Contributing

We welcome contributions from the community! Whether it's bug reports, feature requests, or code contributions, your involvement is highly appreciated.

- **Reporting Bugs:** Before creating a bug report, please check if the issue has already been reported in the Issues section. If not, create a new one with a clear description and steps to reproduce the issue.
- **Feature Requests:** If you have an idea to enhance the tool, feel free to share it in the [Issues](https://github.com/0x19/solc-switch/issues) section.
- **Code Contributions:** If you wish to contribute code, please fork the repository, make your changes, and submit a pull request.


## License

**solc-switch** is licensed under the Apache License 2.0. You can read the full license [here](./LICENSE).