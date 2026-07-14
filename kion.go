// Package kion provides shared configuration and helpers for the Kion
// Go SDK's per-version sub-packages.
//
// The SDK is organized as one sub-package per supported Kion API version
// (e.g. generated/v3_15, generated/v3_14, generated/master). Each
// sub-package exposes its own typed Client constructed with
// `<version>.New(baseURL, opts...)`, where `opts` are shared Option values
// from this package:
//
//	import (
//	    kion "github.com/kionsoftware/kion-sdk-go"
//	    v315 "github.com/kionsoftware/kion-sdk-go/generated/v3_15"
//	)
//
//	client, err := v315.New("https://kion.example.com",
//	    kion.WithAPIKey("..."),
//	    kion.WithSkipVerify(true),
//	)
//
// A program can import multiple versions simultaneously and talk to
// different Kion instances running different versions.
//
// This root package contains only version-agnostic code: it has no
// dependency on any generated package, so it never drifts.
package kion

import (
	"crypto/tls"
	"net/http"
	"strings"
	"time"
)

// ConfigFor resolves a set of Options into a Config, applying defaults.
// It is exported for use by per-version sub-packages' New constructors.
func ConfigFor(opts ...Option) *Config {
	cfg := &Config{
		Timeout: 30 * time.Second,
	}
	for _, o := range opts {
		o(cfg)
	}
	return cfg
}

// NormalizeServerURL ensures the base URL ends with the "/api" prefix that
// the Kion API expects. Trailing slashes are stripped. It is exported for
// use by per-version sub-packages' New constructors.
func NormalizeServerURL(baseURL string) string {
	serverURL := strings.TrimRight(baseURL, "/")
	if !strings.HasSuffix(serverURL, "/api") {
		serverURL += "/api"
	}
	return serverURL
}

// BuildHTTPClient constructs an *http.Client honoring the SkipVerify and
// Timeout settings from a Config. It is exported for use by per-version
// sub-packages' New constructors.
func BuildHTTPClient(skipVerify bool, timeout time.Duration) *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if skipVerify {
		transport.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true, //nolint:gosec // user-requested
		}
	}
	return &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}
}
