package solc

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// GetLocalReleasesPath returns the path to the local releases.json file.
func (s *Solc) GetLocalReleasesPath() string {
	return filepath.Join(s.config.GetReleasesPath(), "releases.json")
}

// GetLocalReleases fetches the Solidity versions saved locally in releases.json.
//
// Returns:
// - A slice of Version representing all the fetched Solidity versions.
// - An error if there's any issue during the fetch process.
func (s *Solc) GetLocalReleases() ([]Version, error) {
	data, err := os.ReadFile(s.GetLocalReleasesPath())
	if err != nil {
		return nil, err
	}

	var releases []Version
	if err := json.Unmarshal(data, &releases); err != nil {
		return nil, err
	}

	s.localReleases = releases
	return releases, nil
}

// GetCachedReleases returns the cached releases from memory.
//
// Returns:
// - A slice of Version representing all the cached Solidity versions.
func (s *Solc) GetCachedReleases() []Version {
	return s.localReleases
}

// GetLatestRelease reads the memory cache or local releases.json file and returns the latest Solidity version.
//
// Returns:
// - A pointer to the latest Version.
// - An error if there's any issue during the fetch process.
func (s *Solc) GetLatestRelease() (*Version, error) {
	var versions []Version

	if s.GetCachedReleases() == nil {
		localReleases, err := s.GetLocalReleases()
		if err != nil {
			return nil, err
		}
		versions = localReleases
	} else {
		versions = s.localReleases
	}

	// Check if there are any versions available
	if len(versions) == 0 {
		return nil, errors.New("no versions found in releases.json")
	}

	// Return the first version as the latest release (assuming the list is sorted by release date)
	return &versions[0], nil
}

// GetRelease reads the memory cache or local releases.json file and returns the Solidity version matching the given tag name.
//
// Parameters:
// - tagName: A string representing the tag name of the desired Solidity version.
//
// Returns:
// - A pointer to the matching Version.
// - An error if there's any issue during the fetch process or if the version is not found.
func (s *Solc) GetRelease(tagName string) (*Version, error) {
	var versions []Version

	tagName = getCleanedVersionTag(tagName)

	if s.GetCachedReleases() == nil {
		localReleases, err := s.GetLocalReleases()
		if err != nil {
			return nil, err
		}
		versions = localReleases
	} else {
		versions = s.localReleases
	}

	// Check if there are any versions available
	if len(versions) == 0 {
		return nil, errors.New("no versions found in available releases")
	}

	// Find the version matching the given tag name
	for _, version := range versions {
		if getCleanedVersionTag(version.TagName) == tagName {
			return &version, nil
		}
	}

	return nil, errors.New("version not found")
}

// GetReleasesSimplified fetches the Solidity versions saved locally in releases.json and returns a simplified version info.
//
// Returns:
// - A slice of VersionInfo representing the simplified version information.
// - An error if there's any issue during the fetch process.
func (s *Solc) GetReleasesSimplified() ([]VersionInfo, error) {
	var versions []Version

	versions, err := s.GetLocalReleases()
	if err != nil {
		return nil, err
	}

	// Return the first version as the latest release (assuming the list is sorted by release date)
	var versionsInfo []VersionInfo
	for _, version := range versions {
		versionsInfo = append(versionsInfo, version.GetVersionInfo(versions[0].TagName))
	}

	return versionsInfo, nil
}

// GetBinary returns the path to the binary of the specified version.
//
// Parameters:
// - version: A string representing the desired Solidity version.
//
// Returns:
// - A string representing the path to the binary.
// - An error if there's any issue during the fetch process or if the binary is not found.
func (s *Solc) GetBinary(version string) (string, error) {
	version = getCleanedVersionTag(version)
	_, err := s.GetRelease(version)
	if err != nil {
		return "", err
	}

	filename := fmt.Sprintf("solc-%s", version)
	distribution := s.GetDistributionForAsset()
	if distribution == "solc-windows" {
		filename += ".exe"
	}

	binaryPath := filepath.Join(s.config.GetReleasesPath(), filename)

	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		return "", fmt.Errorf("binary for version %s not found", version)
	}

	return binaryPath, nil
}

// RemoveBinary removes the binary file of the specified version.
//
// Parameters:
// - version: A string representing the Solidity version whose binary should be removed.
//
// Returns:
// - An error if there's any issue during the removal process or if the binary is not found.
func (s *Solc) RemoveBinary(version string) error {
	version = getCleanedVersionTag(version)
	_, err := s.GetRelease(version)
	if err != nil {
		return err
	}

	filename := fmt.Sprintf("solc-%s", version)
	distribution := s.GetDistributionForAsset()
	if distribution == "solc-windows" {
		filename += ".exe"
	}

	binaryPath := filepath.Join(s.config.GetReleasesPath(), filename)

	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		return fmt.Errorf("binary for version %s not found", version)
	}

	if err := os.Remove(binaryPath); err != nil {
		return err
	}

	return nil
}
