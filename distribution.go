package solc

// Distribution represents the type of operating system.
type Distribution string

// String returns the string representation of the Distribution.
//
// Returns:
// - "windows" if the distribution is Windows.
// - "macos" if the distribution is MacOS.
// - "linux" if the distribution is Linux.
// - "unknown" if the distribution is Unknown or not recognized.
func (d Distribution) String() string {
	return string(d)
}

const (
	// Windows represents the Microsoft Windows operating system.
	Windows Distribution = "windows"

	// MacOS represents the Apple macOS operating system.
	MacOS Distribution = "darwin"

	// Linux represents the Linux operating system.
	Linux Distribution = "linux"

	// Unknown represents an unrecognized operating system.
	Unknown Distribution = "unknown"
)

// GetDistribution determines the operating system type on which the code is running.
// It returns one of the predefined Distribution constants: Windows, MacOS, Linux, or Unknown.
//
// Returns:
// - Windows if the operating system is Microsoft Windows.
// - MacOS if the operating system is Apple macOS.
// - Linux if the operating system is Linux.
// - Unknown if the operating system is not recognized.
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
