package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	IdpURL              string
	IdpClientID         string
	IDPProviderName     string
	IdpProtocol         string
	DomainName          string
	AUTHURL             string
	Region              string
	RedirectPort        int
	OutputFile          string
	NoBrowser           bool
	CodeChallengeMethod string // "S256" (default) or "plain"
	Scope               string // OIDC scopes (default: "openid email profile roles groups organization offline_access")
}

func New() *Config {
	region := getEnv("OS_REGION", "eu-de")
	return &Config{
		IdpURL:              getEnv("IDP_URL", ""),
		IdpClientID:         getEnv("IDP_CLIENT_ID", ""),
		IDPProviderName:     getEnv("IDP_PROVIDER_NAME", ""),
		IdpProtocol:         getEnv("IDP_PROTOCOL", "oidc"),
		DomainName:          getEnv("OS_DOMAIN_NAME", ""),
		AUTHURL:             getEnv("OS_AUTH_URL", getIAMEndpoint(region)),
		Region:              region,
		RedirectPort:        getEnvInt("REDIRECT_PORT", 9197),
		OutputFile:          getEnv("OUTPUT_FILE", "otc-credentials"),
		NoBrowser:           getEnvBool("NO_BROWSER", false),
		CodeChallengeMethod: getEnv("CODE_CHALLENGE_METHOD", "S256"),
		Scope:               getEnv("OIDC_SCOPE", ""),
	}
}

func getIAMEndpoint(region string) string {
	switch region {
	case "eu-ch2":
		return "https://iam-pub.eu-ch2.sc.otc.t-systems.com"
	default:
		return fmt.Sprintf("https://iam.%s.otc.t-systems.com", region)
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return defaultVal
}

func getEnvBool(key string, defaultVal bool) bool {
	if val := os.Getenv(key); val != "" {
		return val == "true" || val == "1" || val == "yes"
	}
	return defaultVal
}

// ValidateCodeChallengeMethod validates the code challenge method
func (c *Config) ValidateCodeChallengeMethod() error {
	validMethods := []string{"S256", "plain"}
	for _, method := range validMethods {
		if c.CodeChallengeMethod == method {
			return nil
		}
	}
	return fmt.Errorf("invalid code_challenge_method: %s (must be S256 or plain)", c.CodeChallengeMethod)
}

// GetScopes returns the scopes as a space-separated string with + for URL encoding
func (c *Config) GetScopes() string {
	// Replace spaces with + for URL encoding
	return strings.ReplaceAll(c.Scope, " ", "+")
}
