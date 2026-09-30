package updater

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var (
	ErrDevVersion = errors.New("cannot check updates for development builds")
)

type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

type Release struct {
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	HTMLURL     string    `json:"html_url"`
	PublishedAt time.Time `json:"published_at"`
	Assets      []Asset   `json:"assets"`
}

type VersionManifest struct {
	Desktop     string `json:"desktop"`
	Android     string `json:"android"`
	PublishedAt string `json:"published_at"`
}

type Checker struct {
	RepoOwner      string
	RepoName       string
	CurrentVersion string
	HTTPClient     *http.Client
	BaseURL        string // Optional override for testing
}

func NewChecker(currentVersion string) *Checker {
	return &Checker{
		RepoOwner:      "obliviousorion",
		RepoName:       "Kawaii-wify",
		CurrentVersion: currentVersion,
		HTTPClient:     &http.Client{Timeout: 10 * time.Second},
	}
}

// CheckLatest checks GitHub releases for a newer version than CurrentVersion.
// It prioritizes checking versions.json attached to the release assets. If only Android
// was bumped in versions.json, Desktop remains silent.
// Returns (release, isNewer, error).
func (c *Checker) CheckLatest(ctx context.Context) (*Release, bool, error) {
	curr := strings.TrimPrefix(c.CurrentVersion, "v")
	if curr == "" || curr == "dev" {
		return nil, false, ErrDevVersion
	}

	// 1. Direct CDN Check via versions.json (Rate-Limit Immune)
	// Fetching versions.json directly from GitHub CDN bypasses api.github.com's
	// 60 req/hr unauthenticated IP rate limit, which is frequently exhausted on campus Wi-Fi.
	if c.BaseURL == "" {
		cdnURL := fmt.Sprintf("https://github.com/%s/%s/releases/latest/download/versions.json", c.RepoOwner, c.RepoName)
		cdnReq, err := http.NewRequestWithContext(ctx, http.MethodGet, cdnURL, nil)
		if err == nil {
			cdnReq.Header.Set("User-Agent", "Kawaii-Wify/"+c.CurrentVersion)
			if cdnResp, err := c.HTTPClient.Do(cdnReq); err == nil && cdnResp.StatusCode == http.StatusOK {
				var manifest VersionManifest
				decErr := json.NewDecoder(cdnResp.Body).Decode(&manifest)
				cdnResp.Body.Close()
				if decErr == nil {
					if manifest.Desktop == "" {
						return &Release{TagName: "v" + curr}, false, nil
					}
					latestDesktop := strings.TrimSpace(strings.TrimPrefix(manifest.Desktop, "v"))
					tagName := "v" + latestDesktop
					rel := &Release{
						TagName: tagName,
						Name:    tagName,
						HTMLURL: fmt.Sprintf("https://github.com/%s/%s/releases/tag/%s", c.RepoOwner, c.RepoName, tagName),
						Assets: []Asset{
							{
								Name:               "kawaii-wify-windows-amd64.exe",
								BrowserDownloadURL: fmt.Sprintf("https://github.com/%s/%s/releases/download/%s/kawaii-wify-windows-amd64.exe", c.RepoOwner, c.RepoName, tagName),
							},
							{
								Name:               "kawaii-wify-linux-amd64",
								BrowserDownloadURL: fmt.Sprintf("https://github.com/%s/%s/releases/download/%s/kawaii-wify-linux-amd64", c.RepoOwner, c.RepoName, tagName),
							},
							{
								Name:               "kawaii-wify-linux-arm64",
								BrowserDownloadURL: fmt.Sprintf("https://github.com/%s/%s/releases/download/%s/kawaii-wify-linux-arm64", c.RepoOwner, c.RepoName, tagName),
							},
							{
								Name:               "kawaii-wify-darwin-amd64",
								BrowserDownloadURL: fmt.Sprintf("https://github.com/%s/%s/releases/download/%s/kawaii-wify-darwin-amd64", c.RepoOwner, c.RepoName, tagName),
							},
							{
								Name:               "kawaii-wify-darwin-arm64",
								BrowserDownloadURL: fmt.Sprintf("https://github.com/%s/%s/releases/download/%s/kawaii-wify-darwin-arm64", c.RepoOwner, c.RepoName, tagName),
							},
						},
					}
					if t, err := time.Parse(time.RFC3339, manifest.PublishedAt); err == nil {
						rel.PublishedAt = t
					}
					return rel, isNewerSemver(latestDesktop, curr), nil
				}
			} else if cdnResp != nil {
				cdnResp.Body.Close()
			}
		}
	}

	// 2. Fallback to GitHub REST API (used if custom BaseURL is set for tests or CDN download fails)
	url := c.BaseURL
	if url == "" {
		url = fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", c.RepoOwner, c.RepoName)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, false, fmt.Errorf("failed to create update request: %w", err)
	}
	req.Header.Set("User-Agent", "Kawaii-Wify/"+c.CurrentVersion)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, false, fmt.Errorf("update request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("unexpected status code from release API: %d", resp.StatusCode)
	}

	var rel Release
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, false, fmt.Errorf("failed to parse release JSON: %w", err)
	}

	// 1. Try to check versions.json in release assets
	for _, asset := range rel.Assets {
		if asset.Name == "versions.json" && asset.BrowserDownloadURL != "" {
			manifestReq, err := http.NewRequestWithContext(ctx, http.MethodGet, asset.BrowserDownloadURL, nil)
			if err == nil {
				manifestReq.Header.Set("User-Agent", "Kawaii-Wify/"+c.CurrentVersion)
				if mResp, err := c.HTTPClient.Do(manifestReq); err == nil {
					if mResp.StatusCode == http.StatusOK {
						var manifest VersionManifest
						decErr := json.NewDecoder(mResp.Body).Decode(&manifest)
						mResp.Body.Close()
						if decErr == nil {
							if manifest.Desktop != "" {
								latestDesktop := strings.TrimSpace(strings.TrimPrefix(manifest.Desktop, "v"))
								return &rel, isNewerSemver(latestDesktop, curr), nil
							}
							// Manifest is present but desktop version is omitted or empty (e.g. android-only release).
							// Desktop must NOT update.
							return &rel, false, nil
						}
					} else {
						mResp.Body.Close()
					}
				}
			}
		}
	}

	// 2. Fallback to release.TagName if versions.json is not attached
	latest := strings.TrimSpace(strings.TrimPrefix(rel.TagName, "v"))
	if latest == "" {
		return nil, false, nil
	}

	return &rel, isNewerSemver(latest, curr), nil
}

// isNewerSemver compares two semver strings (e.g. "0.2.1" vs "0.2.0").
// Returns true if remote is strictly greater than local.
// Strips pre-release labels (e.g. "0.2.1-rc1" -> "0.2.1") and trims whitespace.
func isNewerSemver(remote, local string) bool {
	remote = strings.TrimSpace(strings.TrimPrefix(remote, "v"))
	local = strings.TrimSpace(strings.TrimPrefix(local, "v"))

	cleanPart := func(p string) string {
		if idx := strings.IndexAny(p, "-+"); idx != -1 {
			return p[:idx]
		}
		return p
	}

	rParts := strings.Split(remote, ".")
	lParts := strings.Split(local, ".")

	for i := 0; i < len(rParts) && i < len(lParts); i++ {
		rNum, rErr := strconv.Atoi(cleanPart(rParts[i]))
		lNum, lErr := strconv.Atoi(cleanPart(lParts[i]))
		if rErr == nil && lErr == nil {
			if rNum > lNum {
				return true
			}
			if rNum < lNum {
				return false
			}
		} else {
			if rParts[i] > lParts[i] {
				return true
			}
			if rParts[i] < lParts[i] {
				return false
			}
		}
	}
	return len(rParts) > len(lParts)
}

// StartBackgroundChecker launches a non-blocking background routine that checks
// for updates periodically after an initial delay.
func StartBackgroundChecker(ctx context.Context, currentVersion string, interval time.Duration, onUpdate func(rel *Release)) {
	curr := strings.TrimPrefix(currentVersion, "v")
	if curr == "" || curr == "dev" {
		return
	}

	checker := NewChecker(currentVersion)

	go func() {
		// Wait 2 minutes after launch before performing initial check
		select {
		case <-ctx.Done():
			return
		case <-time.After(2 * time.Minute):
		}

		if rel, newer, err := checker.CheckLatest(ctx); err == nil && newer {
			onUpdate(rel)
		}

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if rel, newer, err := checker.CheckLatest(ctx); err == nil && newer {
					onUpdate(rel)
				}
			}
		}
	}()
}
