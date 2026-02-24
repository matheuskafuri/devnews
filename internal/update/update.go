package update

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

// Result holds the outcome of a version check.
type Result struct {
	LatestVersion string
}

type ghRelease struct {
	TagName string `json:"tag_name"`
}

// Check queries the GitHub Releases API to see if a newer version is available.
// Returns nil on any error (non-fatal).
func Check(ctx context.Context, currentVersion string) *Result {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/repos/matheuskafuri/devnews/releases/latest", nil)
	if err != nil {
		return nil
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	var release ghRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil
	}

	latest := strings.TrimPrefix(release.TagName, "v")
	current := strings.TrimPrefix(currentVersion, "v")

	if latest == "" || latest == current {
		return nil
	}

	return &Result{LatestVersion: latest}
}
