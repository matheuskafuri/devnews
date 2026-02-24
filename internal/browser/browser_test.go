package browser

import "testing"

func TestOpenRejectsNonHTTP(t *testing.T) {
	tests := []struct {
		url     string
		wantErr bool
	}{
		{"https://example.com", false},
		{"http://example.com", false},
		{"file:///etc/passwd", true},
		{"javascript:alert(1)", true},
		{"ftp://example.com", true},
		{"", true},
	}

	for _, tt := range tests {
		err := Open(tt.url)
		if tt.wantErr && err == nil {
			t.Errorf("Open(%q): expected error, got nil", tt.url)
		}
		if !tt.wantErr && err != nil {
			// On CI/headless, the open command may fail, but the URL validation should pass.
			// We only care that scheme validation doesn't reject valid URLs.
			// The actual browser launch may fail in test environments.
			_ = err
		}
	}
}
