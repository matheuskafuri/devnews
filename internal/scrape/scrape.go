package scrape

import (
	"fmt"
	"html"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

var (
	reScript = regexp.MustCompile(`(?is)<script[^>]*>.*?</script>`)
	reStyle  = regexp.MustCompile(`(?is)<style[^>]*>.*?</style>`)
	reBlock  = regexp.MustCompile(`(?i)</?(p|div|br|h[1-6]|li|tr|blockquote)[^>]*>`)
	reTag    = regexp.MustCompile(`<[^>]+>`)
	reSpaces = regexp.MustCompile(`\s+`)
)

const maxBodySize = 512 * 1024 // 512KB max download
const maxTextLen = 4000

var httpClient = &http.Client{Timeout: 15 * time.Second}

// StripHTML removes HTML tags, scripts, styles and returns plain text.
func StripHTML(s string) string {
	s = reScript.ReplaceAllString(s, "")
	s = reStyle.ReplaceAllString(s, "")
	s = reBlock.ReplaceAllString(s, " ")
	s = reTag.ReplaceAllString(s, "")
	s = html.UnescapeString(s)
	s = reSpaces.ReplaceAllString(s, " ")
	return strings.TrimSpace(s)
}

// Fetch downloads a URL, strips HTML, and returns plain text truncated to ~4000 chars.
func Fetch(url string) (string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "devnews/1.0 (article summarizer)")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetching %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("fetching %s: status %d", url, resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBodySize))
	if err != nil {
		return "", fmt.Errorf("reading %s: %w", url, err)
	}

	text := StripHTML(string(body))
	if len(text) > maxTextLen {
		text = text[:maxTextLen]
	}
	return text, nil
}
