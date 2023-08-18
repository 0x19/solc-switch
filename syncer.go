package solc

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// SyncReleases fetches the available Solidity versions from GitHub, saves them to releases.json, and reloads the local cache.
//
// Returns:
// - A slice of Version representing all the fetched Solidity versions.
// - An error if there's any issue during the synchronization process.
func (s *Solc) SyncReleases() ([]Version, error) {
	var allVersions []Version
	page := 1

	for {
		url := fmt.Sprintf("%s?page=%d", s.config.GetReleasesUrl(), page)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}

		req.Header.Add("Authorization", fmt.Sprintf("token %s", s.config.personalAccessToken))
		req = req.WithContext(s.ctx)

		resp, err := s.GetHTTPClient().Do(req)
		if err != nil {
			return nil, err
		}

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			if err := resp.Body.Close(); err != nil {
				return nil, err
			}
			return nil, err
		}

		if err := resp.Body.Close(); err != nil {
			return nil, err
		}

		var versions []Version
		if err := json.Unmarshal(bodyBytes, &versions); err != nil {
			return nil, err
		}

		// If the current page has no releases, break out of the loop
		if len(versions) == 0 {
			break
		}

		allVersions = append(allVersions, versions...)
		page++
	}

	allVersionsBytes, err := json.Marshal(allVersions)
	if err != nil {
		return nil, err
	}

	if err := os.WriteFile(s.GetLocalReleasesPath(), allVersionsBytes, 0600); err != nil {
		return nil, err
	}

	s.localReleases = allVersions
	return allVersions, nil
}

// SyncBinaries downloads all the binaries for the specified versions in parallel.
//
// Parameters:
// - versions: A slice of Version representing the Solidity versions for which binaries should be downloaded.
// - limitVersion: A string representing a specific version to limit the download to. If empty, binaries for all versions will be downloaded.
//
// Returns:
// - An error if there's any issue during the download process.
func (s *Solc) SyncBinaries(versions []Version, limitVersion string) error {
	var wg sync.WaitGroup
	errorsCh := make(chan error, len(versions))
	progressCh := make(chan int, len(versions))
	totalDownloads := 0
	completedDownloads := 0

	for _, version := range versions {
		versionTag := getCleanedVersionTag(version.TagName)
		if limitVersion != "" && versionTag != limitVersion {
			continue
		}

		for _, asset := range version.Assets {
			distribution := s.GetDistributionForAsset()
			if strings.Contains(asset.Name, distribution) {
				filename := fmt.Sprintf("%s/solc-%s", s.config.GetReleasesPath(), versionTag)
				if distribution == "solc-windows" {
					filename += ".exe"
				}

				zap.L().Debug(
					"Checking if solc asset needs to be downloaded",
					zap.String("version", versionTag),
					zap.String("asset_name", asset.Name),
				)

				if _, err := os.Stat(filename); os.IsNotExist(err) {
					totalDownloads++
					zap.L().Debug(
						"Downloading solc asset",
						zap.String("version", versionTag),
						zap.String("asset_name", asset.Name),
						zap.String("asset_local_filename", filepath.Base(filename)),
					)

					wg.Add(1)

					// Just a bit of the time because we could receive 503 from GitHub so we don't want to spam them
					time.Sleep(100 * time.Millisecond)

					go func(v Version, a Asset, fName string) {
						defer wg.Done()
						select {
						case <-s.ctx.Done():
							zap.L().Debug(
								"Context cancelled. Stopping the download",
								zap.String("version", versionTag),
								zap.String("asset_name", asset.Name),
								zap.String("asset_local_filename", filepath.Base(filename)),
							)
							errorsCh <- fmt.Errorf("context cancelled")
							return
						default:
							err := s.downloadFile(fName, a.BrowserDownloadURL)
							if err != nil {
								errorsCh <- fmt.Errorf("error downloading binary for version %s: %v", getCleanedVersionTag(v.TagName), err)
							}
							progressCh <- 1
						}
					}(version, asset, filename)
				}
				break
			}
		}
	}

	// Progress ticker
	ticker := time.NewTicker(1 * time.Second)
	go func() {
		for range ticker.C {
			select {
			case <-s.ctx.Done():
				return
			default:
				zap.L().Debug(fmt.Sprintf(
					"Downloaded %d out of %d binaries\n", completedDownloads, totalDownloads,
				))
			}
		}
	}()

	go func() {
		for range progressCh {
			completedDownloads++
		}
	}()

	wg.Wait()
	close(errorsCh)
	close(progressCh)
	ticker.Stop()

	// One error is really enough. Could potentially troll the user with multiple errors but heck...
	for err := range errorsCh {
		if err != nil {
			return err
		}
	}

	return nil
}

// Sync fetches the available Solidity versions from GitHub, saves them to releases.json, reloads the local cache,
// and downloads all the binaries for the distribution for future use.
//
// Returns:
// - An error if there's any issue during the synchronization process.
func (s *Solc) Sync() error {
	versions, err := s.SyncReleases()
	if err != nil {
		return err
	}

	zap.L().Debug("Syncing solc binaries...", zap.Int("versions_count", len(versions)))

	if err := s.SyncBinaries(versions, ""); err != nil {
		return err
	}

	return nil
}

// SyncOne fetches a specific Solidity version from GitHub, saves it to releases.json, reloads the local cache,
// and downloads the binary for the distribution for future use.
//
// Parameters:
// - version: A pointer to a Version representing the specific Solidity version to be synchronized.
//
// Returns:
// - An error if there's any issue during the synchronization process.
func (s *Solc) SyncOne(version *Version) error {
	if version == nil {
		return fmt.Errorf("version must be provided to synchronize one version")
	}

	versions, err := s.SyncReleases()
	if err != nil {
		return err
	}

	zap.L().Debug(
		"Attempt to synchronize solc release", zap.Int("versions_count", len(versions)),
		zap.String("version", getCleanedVersionTag(version.TagName)),
	)

	if err := s.SyncBinaries(versions, version.TagName); err != nil {
		return err
	}

	return nil
}

// downloadFile downloads a file from the provided URL and saves it to the specified path.
//
// Parameters:
// - filepath: A string representing the path where the downloaded file should be saved.
// - url: A string representing the URL from which the file should be downloaded.
//
// Returns:
// - An error if there's any issue during the download process.
func (s *Solc) downloadFile(file string, url string) error {
	rand.Seed(time.Now().UnixNano())

	// Just a bit of the time because we could receive 503 from GitHub so we don't want to spam them
	time.Sleep(time.Duration((rand.Intn(1001) + 500)) * time.Millisecond)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", fmt.Sprintf("token %s", s.config.personalAccessToken))
	req = req.WithContext(s.ctx)

	resp, err := s.GetHTTPClient().Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download file: %s", resp.Status)
	}

	out, err := os.Create(filepath.Clean(file))
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err = io.Copy(out, resp.Body); err != nil {
		return err
	}

	// #nosec G302
	// G302 (CWE-276): Expect file permissions to be 0600 or less (Confidence: HIGH, Severity: MEDIUM)
	// We want executable files to be executable by the user running the program so we can't use 0600.
	if err := os.Chmod(file, 0777); err != nil {
		return fmt.Errorf("failed to set file as executable: %v", err)
	}

	return nil
}
