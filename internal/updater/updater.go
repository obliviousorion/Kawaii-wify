package updater

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

var (
	ErrDevVersion = errors.New("cannot check updates for development builds")
)

type Release struct {
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	HTMLURL     string    `json:"html_url"`
	PublishedAt time.Time `json:"published_at"`
}

type Checker struct {
	RepoOwner      string
	RepoName       string
	CurrentVersion string
	HTTPClient     *http.Client
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
// Returns (release, isNewer, error).
func (c *Checker) CheckLatest(ctx context.Context) (*Release, bool, error) {
	curr := strings.TrimPrefix(c.CurrentVersion, "v")
	if curr == "" || curr == "dev" {
		return nil, false, ErrDevVersion
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", c.RepoOwner, c.RepoName)
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

	latest := strings.TrimPrefix(rel.TagName, "v")
	if latest == "" {
		return nil, false, nil
	}

	if latest != curr {
		return &rel, true, nil
	}

	return &rel, false, nil
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
		// Wait 1 minute after launch before performing initial check
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
