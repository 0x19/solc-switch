package solc

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// SyncReleases fetches the available Solidity versions from GitHub, saves them to releases.json, and reloads the local cache.
func (s *Solc) SyncReleases() ([]Version, error) {
	var allVersions []Version
	page := 1

	// Sync maximum 4 times per day in order to increase the speed of the sync process when there's really
	// no need to sync more often than that.
	if time.Since(s.lastSync) < time.Duration(6*time.Hour) {
		return s.localReleases, nil
	}

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
	s.lastSync = time.Now()
	return allVersions, nil
}

// SyncBinaries downloads all the binaries for the specified versions in parallel.
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

				if _, err := os.Stat(filename); os.IsNotExist(err) {
					totalDownloads++
					zap.L().Info(
						"Downloading missing solc release",
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
func (s *Solc) downloadFile(file string, url string) error {
	// Just a bit of the time because we could receive 503 from GitHub so we don't want to spam them
	randomDelayBetween500And1500()

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
	if err := os.Chmod(file, 0700); err != nil {
		return fmt.Errorf("failed to set file as executable: %v", err)
	}

	return nil
}

// randomDelayBetween500And1500 sleeps for a random amount of time between 500 and 1500 milliseconds.
func randomDelayBetween500And1500() {
	n, err := rand.Int(rand.Reader, big.NewInt(1001))
	if err != nil {
		panic(err)
	}
	delay := n.Int64() + 500
	time.Sleep(time.Duration(delay) * time.Millisecond)
}
