package solc

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

// GetLocalReleasesPath returns the path to the local releases.json file.
func (s *Solc) GetLocalReleasesPath() string {
	return filepath.Join(s.config.GetReleasesPath(), "releases.json")
}

// GetLocalReleases fetches the Solidity versions saved locally in releases.json.
func (s *Solc) GetLocalReleases() ([]Version, error) {
	data, err := os.ReadFile(s.GetLocalReleasesPath())
	if err != nil {

		// If the file does not exist, fetch the available releases from GitHub
		if errors.Is(err, os.ErrNotExist) {
			releases, err := s.SyncReleases()
			if err != nil {
				return nil, err
			}
			s.localReleases = releases
			return releases, nil
		}

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
func (s *Solc) GetCachedReleases() []Version {
	return s.localReleases
}

// GetLatestRelease reads the memory cache or local releases.json file and returns the latest Solidity version.
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

// GetVersion returns the memory cache or local releases.json file and returns the Solidity version matching the given tag name.
func (s *Solc) GetRelease(tagName string) (*Version, error) {
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

	// Find the version matching the given tag name
	for _, version := range versions {
		if version.TagName == tagName {
			return &version, nil
		}
	}

	return nil, errors.New("version not found")
}

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
