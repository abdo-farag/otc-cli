package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
)

type PKCEChallenge struct {
	Verifier  string
	Challenge string
	Method    string // "S256" or "plain"
}

// GeneratePKCE generates PKCE challenge with support for both S256 and plain methods
// S256 (default): Uses SHA256 hash of the verifier
// plain: Uses the verifier directly as the challenge (less secure, only for development)
func GeneratePKCE() (*PKCEChallenge, error) {
	// Generate 32 random bytes
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}

	// Create code verifier
	verifier := base64.RawURLEncoding.EncodeToString(b)

	// Create code challenge using S256 (SHA256 hash of verifier)
	// Note: The actual method (S256 vs plain) is determined by config.CodeChallengeMethod
	h := sha256.New()
	h.Write([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(h.Sum(nil))

	return &PKCEChallenge{
		Verifier:  verifier,
		Challenge: challenge,
		Method:    "S256",
	}, nil
}

// GeneratePKCEWithMethod generates PKCE challenge with specified method
// method: "S256" (recommended, uses SHA256) or "plain" (development only, uses verifier directly)
func GeneratePKCEWithMethod(method string) (*PKCEChallenge, error) {
	// Generate 32 random bytes
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}

	// Create code verifier
	verifier := base64.RawURLEncoding.EncodeToString(b)

	var challenge string

	switch method {
	case "S256":
		// SHA256 hash of the verifier (recommended)
		h := sha256.New()
		h.Write([]byte(verifier))
		challenge = base64.RawURLEncoding.EncodeToString(h.Sum(nil))
	case "plain":
		// Use verifier directly as challenge (less secure, development only)
		challenge = verifier
	default:
		// Default to S256 if invalid method
		h := sha256.New()
		h.Write([]byte(verifier))
		challenge = base64.RawURLEncoding.EncodeToString(h.Sum(nil))
		method = "S256"
	}

	return &PKCEChallenge{
		Verifier:  verifier,
		Challenge: challenge,
		Method:    method,
	}, nil
}

// GenerateState generates a random state parameter for CSRF protection
func GenerateState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
