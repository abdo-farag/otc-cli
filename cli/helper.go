package cli

import (
	"fmt"
	"time"

	"github.com/abdo-farag/otc-cli/internal/auth"
	"github.com/abdo-farag/otc-cli/internal/cache"
	"github.com/abdo-farag/otc-cli/internal/config"
	"github.com/abdo-farag/otc-cli/internal/otc"

	"github.com/fatih/color"
)

// ensureAuthenticated checks for a valid cached token or performs authentication
func ensureAuthenticated(cfg *config.Config) (*cache.TokenCache, error) {
	tokenCache, err := cache.LoadToken()
	if err == nil {
		// Check if token is still valid (with 5 minute buffer)
		if time.Now().Add(5 * time.Minute).Before(tokenCache.ExpiresAt) {
			color.Green("✓ Using cached token (expires: %s)", tokenCache.ExpiresAt.Format("2006-01-02 15:04"))
			return tokenCache, nil
		}
		color.Yellow("⚠ Cached token expired or expiring soon")
	}

	color.Yellow("⚠ No valid cached token found. Running login...")

	authClient := auth.NewClient(cfg)
	tokenResp, handler, err := authClient.GetOIDCToken()
	if err != nil {
		if handler != nil {
			handler.SetValidationStatus("failed", "Authentication failed: "+err.Error())
			time.Sleep(2 * time.Second)
			handler.Close()
		}
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	if handler != nil {
		handler.SetValidationStatus("success", "Authentication successful! Validating OTC access...")
	}
	color.Green("✓ Authorization code received")

	otcClient := otc.NewClient(cfg)
	unscopedToken, err := otcClient.GetUnscopedToken(tokenResp.IDToken, handler)
	if err != nil {
		if handler != nil {
			handler.SetValidationStatus("failed", "Failed to validate OTC access: "+err.Error())
			time.Sleep(2 * time.Second)
			handler.Close()
		}
		return nil, fmt.Errorf("failed to get unscoped token: %w", err)
	}

	if handler != nil {
		handler.SetValidationStatus("success", "All validations passed! Check your terminal.")
		color.Green("✓ OTC access validated")
		time.Sleep(1 * time.Second)
		handler.Close()
	}

	// Calculate proper expiry time
	// OIDC tokens typically last 1 hour, but OpenStack unscoped tokens last 24 hours
	// Use the shorter of the two to be safe
	expiresAt := time.Now().Add(1 * time.Hour) // Conservative: use OIDC token expiry
	if tokenResp.ExpiresIn > 0 {
		expiresAt = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	}

	tokenCache = &cache.TokenCache{
		UnscopedToken: unscopedToken,
		IDToken:       tokenResp.IDToken,
		RefreshToken:  tokenResp.RefreshToken,
		ExpiresAt:     expiresAt,
		Domain:        cfg.DomainName,
		Region:        cfg.Region,
	}
	cache.SaveToken(tokenCache)
	color.Green("✓ Authenticated and cached token")

	return tokenCache, nil
}

// resolveProject resolves a project name or ID to a project ID
func resolveProject(cfg *config.Config, unscopedToken, projectID string) string {
	otcClient := otc.NewClient(cfg)
	domainToken, err := otcClient.GetDomainScopedToken(unscopedToken)
	if err != nil {
		color.Yellow("⚠ Failed to get domain token for project resolution")
		return projectID
	}

	projects, err := otcClient.ListProjects(domainToken)
	if err != nil {
		color.Yellow("⚠ Failed to list projects")
		return projectID
	}

	// If projectID is provided, try to resolve it
	if projectID != "" {
		for _, p := range projects {
			if p.ID == projectID || p.Name == projectID {
				if p.Name != projectID {
					color.Cyan("✓ Resolved project '%s' to ID: %s", projectID, p.ID)
				}
				return p.ID
			}
		}
		color.Yellow("⚠ Project '%s' not found, using as-is", projectID)
		return projectID
	}

	// Use default (first) project if none specified
	if len(projects) > 0 {
		color.Cyan("✓ Using default project: %s (%s)", projects[0].Name, projects[0].ID)
		return projects[0].ID
	}

	color.Yellow("⚠ No projects found")
	return ""
}