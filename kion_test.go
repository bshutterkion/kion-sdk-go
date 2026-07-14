package kion

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/ogen-go/ogen/validate"
)

func TestConfigForDefaults(t *testing.T) {
	cfg := ConfigFor()
	if cfg.Timeout != 30*time.Second {
		t.Errorf("default Timeout = %v, want 30s", cfg.Timeout)
	}
	if cfg.APIKey != "" {
		t.Errorf("default APIKey = %q, want empty", cfg.APIKey)
	}
	if cfg.AuthToken != "" {
		t.Errorf("default AuthToken = %q, want empty", cfg.AuthToken)
	}
	if cfg.SkipVerify {
		t.Error("default SkipVerify = true, want false")
	}
	if cfg.HTTPClient != nil {
		t.Error("default HTTPClient should be nil")
	}
}

func TestConfigForOptions(t *testing.T) {
	hc := &http.Client{}
	cfg := ConfigFor(
		WithAPIKey("my-key"),
		WithBearerToken("my-token"),
		WithSkipVerify(true),
		WithTimeout(60*time.Second),
		WithHTTPClient(hc),
	)
	if cfg.APIKey != "my-key" {
		t.Errorf("APIKey = %q, want my-key", cfg.APIKey)
	}
	if cfg.AuthToken != "my-token" {
		t.Errorf("AuthToken = %q, want my-token", cfg.AuthToken)
	}
	if !cfg.SkipVerify {
		t.Error("SkipVerify should be true")
	}
	if cfg.Timeout != 60*time.Second {
		t.Errorf("Timeout = %v, want 60s", cfg.Timeout)
	}
	if cfg.HTTPClient != hc {
		t.Error("HTTPClient not set")
	}
}

func TestNormalizeServerURL(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"https://kion.example.com", "https://kion.example.com/api"},
		{"https://kion.example.com/", "https://kion.example.com/api"},
		{"https://kion.example.com/api", "https://kion.example.com/api"},
		{"https://kion.example.com/api/", "https://kion.example.com/api"},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			if got := NormalizeServerURL(tt.in); got != tt.want {
				t.Errorf("NormalizeServerURL(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestBuildHTTPClient(t *testing.T) {
	hc := BuildHTTPClient(false, 10*time.Second)
	if hc == nil {
		t.Fatal("BuildHTTPClient returned nil")
	}
	if hc.Timeout != 10*time.Second {
		t.Errorf("Timeout = %v, want 10s", hc.Timeout)
	}

	hcSkip := BuildHTTPClient(true, 5*time.Second)
	tr, ok := hcSkip.Transport.(*http.Transport)
	if !ok {
		t.Fatal("Transport is not *http.Transport")
	}
	if tr.TLSClientConfig == nil || !tr.TLSClientConfig.InsecureSkipVerify {
		t.Error("SkipVerify=true should set InsecureSkipVerify")
	}
}

func TestErrorHelpers(t *testing.T) {
	// Nil error cases.
	if code := StatusCode(nil); code != 0 {
		t.Errorf("StatusCode(nil) = %d, want 0", code)
	}
	if IsNotFound(nil) {
		t.Error("IsNotFound(nil) = true, want false")
	}
	if IsAuthError(nil) {
		t.Error("IsAuthError(nil) = true, want false")
	}
	if IsConflict(nil) {
		t.Error("IsConflict(nil) = true, want false")
	}

	// Status code extraction from wrapped ogen validate error.
	wrapped := &validate.UnexpectedStatusCodeError{StatusCode: http.StatusNotFound}
	if code := StatusCode(wrapped); code != http.StatusNotFound {
		t.Errorf("StatusCode(404 wrapped) = %d, want 404", code)
	}
	if !IsNotFound(wrapped) {
		t.Error("IsNotFound on 404 wrapped = false, want true")
	}

	// Wrapped via errors.Join / fmt.Errorf.
	joined := errors.Join(wrapped)
	if !IsNotFound(joined) {
		t.Error("IsNotFound on joined 404 = false, want true")
	}
}
