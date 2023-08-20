package solc

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// validatePath checks the validity of a given path.
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

	file, err := os.Open(filepath.Clean(path))
	if err != nil {
		return fmt.Errorf("directory is not readable: %s", path)
	}

	if err := file.Close(); err != nil {
		return err
	}

	return nil
}

// getCleanedVersionTag removes the "v" prefix from a version tag.
func getCleanedVersionTag(versionTag string) string {
	return strings.ReplaceAll(versionTag, "v", "")
}
