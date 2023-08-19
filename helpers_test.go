package solc

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidatePath(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	tempFile, err := os.CreateTemp("", "test")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tempFile.Close()
	defer os.Remove(tempFile.Name())

	// Create a directory with no read permissions
	unreadableDir, err := os.MkdirTemp("", "test_unreadable")
	if err != nil {
		t.Fatalf("Failed to create unreadable directory: %v", err)
	}
	defer os.RemoveAll(unreadableDir)
	os.Chmod(unreadableDir, 0222) // Write-only permissions

	tests := []struct {
		name    string
		path    string
		wantErr string
	}{
		{
			name:    "Valid Directory Path",
			path:    tempDir,
			wantErr: "",
		},
		{
			name:    "Invalid Path",
			path:    "/path/that/does/not/exist",
			wantErr: "path does not exist",
		},
		{
			name:    "Path is a File",
			path:    tempFile.Name(),
			wantErr: "path is not a directory",
		},
		{
			name:    "Unreadable Directory",
			path:    unreadableDir,
			wantErr: "directory is not readable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePath(tt.path)
			if tt.wantErr == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, err.Error(), fmt.Sprintf("%s: %s", tt.wantErr, tt.path))
			}
		})
	}
}
