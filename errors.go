package kion

import (
	"errors"
	"net/http"

	"github.com/ogen-go/ogen/validate"
)

// IsNotFound returns true if the error represents an HTTP 404 response.
func IsNotFound(err error) bool {
	return hasStatusCode(err, http.StatusNotFound)
}

// IsAuthError returns true if the error represents an HTTP 401 or 403 response.
func IsAuthError(err error) bool {
	return hasStatusCode(err, http.StatusUnauthorized) || hasStatusCode(err, http.StatusForbidden)
}

// IsConflict returns true if the error represents an HTTP 409 response.
func IsConflict(err error) bool {
	return hasStatusCode(err, http.StatusConflict)
}

// StatusCode extracts the HTTP status code from an error, if available.
// Returns 0 if the error does not contain status code information.
func StatusCode(err error) int {
	var unexpected *validate.UnexpectedStatusCodeError
	if errors.As(err, &unexpected) {
		return unexpected.StatusCode
	}
	return 0
}

func hasStatusCode(err error, code int) bool {
	return StatusCode(err) == code
}
