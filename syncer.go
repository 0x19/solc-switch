package solc

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// SyncReleases fetches the available Solidity versions from GitHub and saves them to releases.json and reloads local cache
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
			resp.Body.Close()
			return nil, err
		}
		resp.Body.Close()

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

	// Save all versions to releases.json
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

// SyncBinaries downloads all the binaries for the distribution in parallel.
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

// Sync fetches the available Solidity versions from GitHub and saves them to releases.json and reloads local cache
// and downloads all the binaries for the distribution for future use...
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

// Sync fetches specific Solidity version from GitHub and saves them to releases.json and reloads local cache
// and downloads all the binaries for the distribution for future use...
func (s *Solc) SyncOne(version Version) error {
	versions, err := s.SyncReleases()
	if err != nil {
		return err
	}

	zap.L().Debug("Syncing solc binaries...", zap.Int("versions_count", len(versions)))

	if err := s.SyncBinaries(versions, version.TagName); err != nil {
		return err
	}

	return nil
}

// downloadFile downloads a file from the given URL and saves it to the specified path.
func (s *Solc) downloadFile(filepath string, url string) error {
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

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
