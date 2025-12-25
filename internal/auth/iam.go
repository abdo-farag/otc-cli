package auth

import (
	"fmt"
	"time"

	"github.com/abdo-farag/otc-cli/internal/config"

	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/identity/v3/tokens"
)

// IAMClient handles direct OTC IAM authentication using gophertelekomcloud
type IAMClient struct {
	cfg *config.Config
}

// NewIAMClient creates a new IAM authentication client
func NewIAMClient(cfg *config.Config) *IAMClient {
	return &IAMClient{cfg: cfg}
}

// GetIAMToken authenticates using OTC IAM user credentials (username/email and password)
// Returns the unscoped token ID that can be used for further API calls
func (ic *IAMClient) GetIAMToken(username, password string) (string, error) {
	if username == "" || password == "" {
		return "", fmt.Errorf("username and password are required")
	}

	if ic.cfg.DomainName == "" {
		return "", fmt.Errorf("OS_DOMAIN_NAME or --domain-name is not configured")
	}

	if ic.cfg.AUTHURL == "" {
		return "", fmt.Errorf("AUTH_URL or --auth-url is not configured")
	}

	// Build auth options for OTC IAM
	authOpts := golangsdk.AuthOptions{
		DomainName:       ic.cfg.DomainName,
		Username:         username,
		Password:         password,
		IdentityEndpoint: ic.cfg.AUTHURL,
	}

	// Get authenticated provider
	provider, err := openstack.AuthenticatedClient(authOpts)
	if err != nil {
		return "", fmt.Errorf("authentication failed: %w", err)
	}

	// Get identity v3 client
	identityClient, err := openstack.NewIdentityV3(provider, golangsdk.EndpointOpts{})
	if err != nil {
		return "", fmt.Errorf("failed to create identity client: %w", err)
	}

	// Create token request
	tokenResult := tokens.Create(identityClient, &authOpts)

	// Extract token
	token, err := tokenResult.ExtractToken()
	if err != nil {
		return "", fmt.Errorf("failed to extract token: %w", err)
	}

	return token.ID, nil
}

// GetScopedToken gets a project-scoped token using an unscoped token
func (ic *IAMClient) GetScopedToken(unscopedToken string, projectID string) (string, error) {
	if unscopedToken == "" {
		return "", fmt.Errorf("unscoped token is required")
	}

	if projectID == "" {
		return "", fmt.Errorf("project ID is required")
	}

	if ic.cfg.AUTHURL == "" {
		return "", fmt.Errorf("AUTH_URL is not configured")
	}

	// Build auth options using token ID and project
	authOpts := golangsdk.AuthOptions{
		IdentityEndpoint: ic.cfg.AUTHURL,
		TokenID:          unscopedToken,
		TenantID:         projectID,
	}

	// Get authenticated provider
	provider, err := openstack.AuthenticatedClient(authOpts)
	if err != nil {
		return "", fmt.Errorf("failed to authenticate with token: %w", err)
	}

	// Get identity v3 client
	identityClient, err := openstack.NewIdentityV3(provider, golangsdk.EndpointOpts{})
	if err != nil {
		return "", fmt.Errorf("failed to create identity client: %w", err)
	}

	// Create scoped token request
	tokenResult := tokens.Create(identityClient, &authOpts)

	// Extract token
	token, err := tokenResult.ExtractToken()
	if err != nil {
		return "", fmt.Errorf("failed to extract scoped token: %w", err)
	}

	return token.ID, nil
}

// RefreshIAMToken validates if a token is still valid
func (ic *IAMClient) RefreshIAMToken(token string) (bool, error) {
	if token == "" {
		return false, fmt.Errorf("token is required")
	}

	if ic.cfg.AUTHURL == "" {
		return false, fmt.Errorf("AUTH_URL is not configured")
	}

	// Try to use the token - if it's valid, the call succeeds
	authOpts := golangsdk.AuthOptions{
		IdentityEndpoint: ic.cfg.AUTHURL,
		TokenID:          token,
	}

	provider, err := openstack.AuthenticatedClient(authOpts)
	if err != nil {
		return false, nil // Token is invalid
	}

	_, err = openstack.NewIdentityV3(provider, golangsdk.EndpointOpts{})
	if err != nil {
		return false, nil // Token is invalid
	}

	return true, nil
}

// IAMTokenResponse represents the IAM token response structure
type IAMTokenResponse struct {
	Token struct {
		Methods   []string        `json:"methods"`
		User      IAMTokenUser    `json:"user"`
		AuditIds  []string        `json:"audit_ids"`
		Roles     []IAMTokenRole  `json:"roles"`
		ExpiresAt time.Time       `json:"expires_at"`
		IssuedAt  time.Time       `json:"issued_at"`
		Project   IAMTokenProject `json:"project"`
		Catalog   []interface{}   `json:"catalog"`
		Domain    IAMTokenDomain  `json:"domain"`
	} `json:"token"`
}

type IAMTokenUser struct {
	Domain struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"domain"`
	ID                string `json:"id"`
	Name              string `json:"name"`
	PasswordExpiresAt string `json:"password_expires_at"`
}

type IAMTokenRole struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type IAMTokenProject struct {
	Domain struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"domain"`
	ID   string `json:"id"`
	Name string `json:"name"`
}

type IAMTokenDomain struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
