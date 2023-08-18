package solc

import (
	"fmt"
	"os"
	"strings"
)

// validatePath checks the validity of a given path.
// It ensures that:
// - The path exists.
// - The path points to a directory.
// - The directory is readable.
//
// Parameters:
// - path: The file system path to validate.
//
// Returns:
// - nil if the path is valid.
// - An error if the path is invalid or any other error occurs.
func validatePath(path string) error {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return fmt.Errorf("path does not exist: %s", path)
	}

	if err != nil {
		return err
	}

	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", path)
	}

	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("directory is not readable: %s", path)
	}
	file.Close()

	return nil
}

// getCleanedVersionTag removes the "v" prefix from a version tag.
func getCleanedVersionTag(versionTag string) string {
	return strings.ReplaceAll(versionTag, "v", "")
}
