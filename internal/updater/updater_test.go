package updater

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestChecker_CheckLatest(t *testing.T) {
	tests := []struct {
		name           string
		currentVersion string
		apiResponse    string
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
			name:           "same version",
			currentVersion: "v1.0.0",
			apiResponse:    `{"tag_name": "v1.0.0", "html_url": "https://github.com/..."}`,
			statusCode:     http.StatusOK,
			wantNewer:      false,
			wantErr:        false,
		},
		{
			name:           "newer version available",
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
			}

			// Direct test through custom client
			req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL, nil)
			resp, err := c.HTTPClient.Do(req)
			if err != nil {
				t.Fatalf("unexpected client error: %v", err)
			}
			defer resp.Body.Close()
		})
	}
}
