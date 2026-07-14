package master

import "context"

// bearerAuth implements the ogen-generated SecuritySource interface
// for this Kion API version. It supplies a Bearer token for both API key
// and bearer token authentication.
type bearerAuth struct {
	apiKey    string
	authToken string
}

// Token returns the security token for the given operation. The auth
// token takes precedence over the API key if both are set.
func (s *bearerAuth) Token(_ context.Context, _ OperationName) (Token, error) {
	key := s.authToken
	if key == "" {
		key = s.apiKey
	}
	return Token{APIKey: "Bearer " + key}, nil
}
