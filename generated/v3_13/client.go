package v3_13

import (
	"net/http"
	"net/url"
	"strings"

	kion "github.com/kionsoftware/kion-sdk-go"
)

// New constructs a typed client for the Kion 3.13
// API version.
//
// Options come from the root kion package and are shared across all
// SDK versions:
//
//	import (
//	    kion "github.com/kionsoftware/kion-sdk-go"
//	    master "github.com/kionsoftware/kion-sdk-go/generated/v3_13"
//	)
//
//	client, err := v3_13.New("https://kion.example.com",
//	    kion.WithAPIKey("..."),
//	    kion.WithSkipVerify(true),
//	)
//
// For a specific stable version, import a versioned sub-package instead
// (e.g. github.com/kionsoftware/kion-sdk-go/generated/v3_15).
func New(baseURL string, opts ...kion.Option) (*Client, error) {
	cfg := kion.ConfigFor(opts...)
	serverURL := kion.NormalizeServerURL(baseURL)

	sec := &bearerAuth{
		apiKey:    cfg.APIKey,
		authToken: cfg.AuthToken,
	}

	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = kion.BuildHTTPClient(cfg.SkipVerify, cfg.Timeout)
	}

	// Wrap the HTTP client with the synthetic-query-string rewriter so
	// operations whose Swagger 2.0 paths embedded query strings (e.g.
	// "POST /v3/account?account-type=aws") get sent to the right URL.
	// fixspec promotes those into synthetic paths under /__qs__/ that
	// this transport rewrites at request time.
	httpClient = wrapWithQueryStringRewriter(httpClient)

	return NewClient(serverURL, sec, WithClient(httpClient))
}

// queryStringPathMarker MUST match the same constant in cmd/fixspec/main.go.
// fixspec rewrites paths like "/v3/account?account-type=aws" to
// "/v3/account/__qs__/account-type/aws", and this transport rewrites them
// back at request time.
const queryStringPathMarker = "/__qs__/"

// wrapWithQueryStringRewriter clones the input *http.Client and installs
// a transport that detects synthetic /__qs__/ path markers and rewrites
// them to real query-string URLs before delegating to the original
// transport. The original client's Timeout is preserved; only the
// transport is wrapped.
func wrapWithQueryStringRewriter(c *http.Client) *http.Client {
	if c == nil {
		c = http.DefaultClient
	}
	inner := c.Transport
	if inner == nil {
		inner = http.DefaultTransport
	}
	clone := &http.Client{
		Transport:     &queryStringRewriter{inner: inner},
		CheckRedirect: c.CheckRedirect,
		Jar:           c.Jar,
		Timeout:       c.Timeout,
	}
	return clone
}

// queryStringRewriter is an http.RoundTripper that detects fixspec's
// synthetic /__qs__/key/value path markers and rewrites them to real
// query-string URLs before delegating to the inner transport.
//
// Why this exists: Swagger 2.0 (which portal generates) allows multiple
// operations on the same HTTP method+path discriminated by a query
// parameter value (e.g. POST /v3/account?account-type=aws vs azure).
// OpenAPI 3.0 (which ogen consumes) does not — paths are unique. fixspec
// preserves all variants by promoting each one to a synthetic distinct
// path. Without this transport, requests would go to the synthetic
// (non-existent) URL and Kion would return 404.
type queryStringRewriter struct {
	inner http.RoundTripper
}

// RoundTrip implements http.RoundTripper.
func (r *queryStringRewriter) RoundTrip(req *http.Request) (*http.Response, error) {
	idx := strings.Index(req.URL.Path, queryStringPathMarker)
	if idx < 0 {
		return r.inner.RoundTrip(req)
	}

	// Clone the request so we don't mutate the caller's value (RoundTrippers
	// must not modify their input per the http.RoundTripper contract).
	req2 := req.Clone(req.Context())
	rewriteSyntheticPath(req2.URL, idx)
	return r.inner.RoundTrip(req2)
}

// rewriteSyntheticPath edits u in place: it splits the path at idx
// (where queryStringPathMarker begins), parses the suffix as repeated
// key/value path segments, and merges them into u.RawQuery.
func rewriteSyntheticPath(u *url.URL, idx int) {
	suffix := u.Path[idx+len(queryStringPathMarker):]
	u.Path = u.Path[:idx]
	if u.Path == "" {
		u.Path = "/"
	}
	parts := strings.Split(suffix, "/")

	// Parse existing query so we can merge instead of overwrite.
	existing := u.Query()
	for i := 0; i+1 < len(parts); i += 2 {
		existing.Set(parts[i], parts[i+1])
	}
	u.RawQuery = existing.Encode()
}
