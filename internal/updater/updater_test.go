package updater

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestChecker_CheckLatest(t *testing.T) {
	tests := []struct {
		name           string
		currentVersion string
		apiResponse    string
		manifestJSON   string
		statusCode     int
		wantNewer      bool
		wantErr        bool
	}{
		{
			name:           "dev version returns ErrDevVersion",
			currentVersion: "dev",
			wantNewer:      false,
			wantErr:        true,
		},
		{
			name:           "empty version returns ErrDevVersion",
			currentVersion: "",
			wantNewer:      false,
			wantErr:        true,
		},
		{
			name:           "same version tag fallback",
			currentVersion: "v1.0.0",
			apiResponse:    `{"tag_name": "v1.0.0", "html_url": "https://github.com/..."}`,
			statusCode:     http.StatusOK,
			wantNewer:      false,
			wantErr:        false,
		},
		{
			name:           "newer version tag fallback",
			currentVersion: "v1.0.0",
			apiResponse:    `{"tag_name": "v1.1.0", "html_url": "https://github.com/..."}`,
			statusCode:     http.StatusOK,
			wantNewer:      true,
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.currentVersion == "dev" || tt.currentVersion == "" {
				c := NewChecker(tt.currentVersion)
				_, _, err := c.CheckLatest(context.Background())
				if (err != nil) != tt.wantErr {
					t.Fatalf("expected error: %v, got: %v", tt.wantErr, err)
				}
				return
			}

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(tt.apiResponse))
			}))
			defer server.Close()

			c := &Checker{
				CurrentVersion: tt.currentVersion,
				HTTPClient:     server.Client(),
				BaseURL:        server.URL,
			}

			rel, newer, err := c.CheckLatest(context.Background())
			if (err != nil) != tt.wantErr {
				t.Fatalf("unexpected error state: %v", err)
			}
			if newer != tt.wantNewer {
				t.Fatalf("expected newer=%v, got: %v (rel=%+v)", tt.wantNewer, newer, rel)
			}
		})
	}
}

func TestChecker_VersionsManifest(t *testing.T) {
	// Manifest server provides both the releases API response and the versions.json asset
	var manifestServer *httptest.Server
	manifestServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/releases/latest":
			resp := fmt.Sprintf(`{
				"tag_name": "v0.3.0",
				"html_url": "https://github.com/obliviousorion/Kawaii-wify/releases/tag/v0.3.0",
				"assets": [
					{
						"name": "versions.json",
						"browser_download_url": "%s/assets/versions.json"
					}
				]
			}`, manifestServer.URL)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(resp))

		case "/assets/versions.json":
			w.WriteHeader(http.StatusOK)
			// Manifest shows Android is 0.3.0, but Desktop is still 0.2.1
			_, _ = w.Write([]byte(`{
				"desktop": "0.2.1",
				"android": "0.3.0"
			}`))

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer manifestServer.Close()

	// 1. When desktop is on 0.2.1 and manifest desktop is 0.2.1 -> should NOT notify even though tag is v0.3.0!
	checker := &Checker{
		CurrentVersion: "0.2.1",
		HTTPClient:     manifestServer.Client(),
		BaseURL:        manifestServer.URL + "/releases/latest",
	}

	_, newer, err := checker.CheckLatest(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if newer {
		t.Fatalf("expected desktop to remain silent when only android was bumped, but got newer=true")
	}

	// 2. When desktop is on 0.2.0 and manifest desktop is 0.2.1 -> should notify!
	checkerOld := &Checker{
		CurrentVersion: "0.2.0",
		HTTPClient:     manifestServer.Client(),
		BaseURL:        manifestServer.URL + "/releases/latest",
	}

	_, newerOld, err := checkerOld.CheckLatest(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !newerOld {
		t.Fatalf("expected desktop to notify when manifest desktop is newer")
	}

	// 3. When manifest exists but has no "desktop" field at all (empty) -> should NOT notify or fall back to tag
	noDesktopServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/releases/latest":
			resp := fmt.Sprintf(`{
				"tag_name": "v0.9.9",
				"assets": [{"name": "versions.json", "browser_download_url": "%s/assets/versions.json"}]
			}`, manifestServer.URL)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(resp))
		case "/assets/versions.json":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"android": "0.9.9"}`))
		}
	}))
	defer noDesktopServer.Close()

	checkerNoDesktop := &Checker{
		CurrentVersion: "0.2.1",
		HTTPClient:     noDesktopServer.Client(),
		BaseURL:        noDesktopServer.URL + "/releases/latest",
	}
	_, newerNoDesk, err := checkerNoDesktop.CheckLatest(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if newerNoDesk {
		t.Fatalf("expected desktop to stay false when manifest has no desktop entry, but got newer=true")
	}
}

func TestIsNewerSemver(t *testing.T) {
	cases := []struct {
		remote string
		local  string
		want   bool
	}{
		{"0.2.1", "0.2.0", true},
		{"0.2.0", "0.2.1", false},
		{"0.2.1", "0.2.1", false},
		{"1.0.0", "0.9.9", true},
		{"0.3.0", "0.2.9", true},
		{"0.2.0", "0.2.0.1", false},
		{"v0.2.1-rc1", "v0.2.0", true},
		{" 0.2.1 ", "0.2.0", true},
		{"0.2.1+build123", "0.2.1", false},
	}

	for _, c := range cases {
		got := isNewerSemver(c.remote, c.local)
		if got != c.want {
			t.Errorf("isNewerSemver(%q, %q) = %v; want %v", c.remote, c.local, got, c.want)
		}
	}
}
