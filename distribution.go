package solc

// Distribution represents the type of operating system.
type Distribution string

// String returns the string representation of the Distribution.
// Possible return values include:
// - "windows" for Windows.
// - "macos" for MacOS.
// - "linux" for Linux.
// - "unknown" for unrecognized or unknown distributions.
func (d Distribution) String() string {
	return string(d)
}

const (
	// Windows denotes the Microsoft Windows operating system.
	Windows Distribution = "windows"

	// MacOS denotes the Apple macOS operating system.
	MacOS Distribution = "darwin"

	// Linux denotes the Linux operating system.
	Linux Distribution = "linux"

	// Unknown denotes an unrecognized or unknown operating system.
	Unknown Distribution = "unknown"
)

// GetDistribution determines the operating system type on which the code is running.
// It returns one of the predefined Distribution constants: Windows, MacOS, Linux, or Unknown.
func (s *Solc) GetDistribution() Distribution {
	switch s.gOOSFunc() {
	case "windows":
		return Windows
	case "darwin":
		return MacOS
	case "linux":
		return Linux
	default:
		return Unknown
	}
}

// GetDistributionForAsset determines the appropriate asset name based on the operating system.
// This is useful for fetching the correct compiler binaries or assets.
// Possible return values include:
// - "solc-windows" for Windows.
// - "solc-macos" for MacOS.
// - "solc-static-linux" for Linux.
// - "unknown" for unrecognized or unknown distributions.
func (s *Solc) GetDistributionForAsset() string {
	switch s.gOOSFunc() {
	case "windows":
		return "solc-windows"
	case "darwin":
		return "solc-macos"
	case "linux":
		return "solc-static-linux"
	default:
		return "unknown"
	}
}
