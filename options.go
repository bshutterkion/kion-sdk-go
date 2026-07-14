package kion

import (
	"net/http"
	"time"
)

// Config holds settings for constructing a Kion SDK client.
type Config struct {
	BaseURL    string
	APIKey     string
	AuthToken  string
	SkipVerify bool
	Timeout    time.Duration
	HTTPClient *http.Client
}

// Option configures a Config.
type Option func(*Config)

// WithAPIKey sets the API key for authentication.
func WithAPIKey(key string) Option {
	return func(c *Config) { c.APIKey = key }
}

// WithBearerToken sets the auth token for authentication.
func WithBearerToken(tok string) Option {
	return func(c *Config) { c.AuthToken = tok }
}

// WithSkipVerify disables TLS certificate verification.
func WithSkipVerify(skip bool) Option {
	return func(c *Config) { c.SkipVerify = skip }
}

// WithTimeout sets the HTTP client timeout.
func WithTimeout(d time.Duration) Option {
	return func(c *Config) { c.Timeout = d }
}

// WithHTTPClient sets a custom HTTP client, overriding SkipVerify and Timeout.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Config) { c.HTTPClient = hc }
}
