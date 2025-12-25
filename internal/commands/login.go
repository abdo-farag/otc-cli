package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/abdo-farag/otc-cli/internal/auth"
	"github.com/abdo-farag/otc-cli/internal/cache"
	"github.com/abdo-farag/otc-cli/internal/config"
	"github.com/abdo-farag/otc-cli/internal/otc"

	"github.com/fatih/color"
	"golang.org/x/term"
)

func Login(cfg *config.Config) error {
	// Step 1: Auth
	authClient := auth.NewClient(cfg)
	tokenResp, handler, err := authClient.GetOIDCToken()
	if err != nil {
		return err
	}
	defer handler.Close()
	color.Green("✓ Authenticated")

	// Step 2: Unscoped Token
	otcClient := otc.NewClient(cfg)
	unscopedToken, err := otcClient.GetUnscopedToken(tokenResp.IDToken, handler)
	if err != nil {
		return err
	}
	color.Green("✓ OTC validated")

	// Save token cache
	tokenCache := &cache.TokenCache{
		UnscopedToken: unscopedToken,
		IDToken:       tokenResp.IDToken,
		RefreshToken:  tokenResp.RefreshToken,
		ExpiresAt:     time.Now().Add(23 * time.Hour),
		Domain:        cfg.DomainName,
		Region:        cfg.Region,
	}

	if err := cache.SaveToken(tokenCache); err != nil {
		color.Yellow("⚠ Warning: Failed to save token cache: %v", err)
	} else {
		color.Green("✓ Token cached at %s", cache.GetTokenPath())
	}

	// Step 3: Domain Token
	domainToken, err := otcClient.GetDomainScopedToken(unscopedToken)
	if err != nil {
		return fmt.Errorf("failed to get domain token: %w", err)
	}
	color.Green("✓ Domain token obtained")

	// Step 4: Project
	color.Yellow("⏳ Listing projects...")
	projects, err := otcClient.ListProjects(domainToken)
	if err != nil {
		return fmt.Errorf("failed to list projects: %w", err)
	}

	if len(projects) == 0 {
		return fmt.Errorf("no projects found for domain %s", cfg.DomainName)
	}

	color.Green("✓ Found %d project(s)", len(projects))
	color.Cyan("  Using project: %s (%s)", projects[0].Name, projects[0].ID)

	projectToken, err := otcClient.GetProjectScopedToken(unscopedToken, projects[0].ID)
	if err != nil {
		return fmt.Errorf("failed to get project token: %w", err)
	}

	// Step 5: Credentials
	color.Yellow("⏳ Creating temporary credentials...")
	creds, err := otcClient.CreateTemporaryCredentials(projectToken, 86400)
	if err != nil {
		return fmt.Errorf("failed to create credentials: %w", err)
	}

	// Save
	if err := creds.SaveShellScript(cfg.OutputFile+".sh", cfg.Region); err != nil {
		return fmt.Errorf("failed to save credentials: %w", err)
	}

	color.Green("\n✓ Credentials saved to %s.sh", cfg.OutputFile)
	color.Yellow("⏰ Expires: %s (24h)", creds.ExpiresAt)
	color.Cyan("\nLoad credentials:")
	fmt.Printf("  source %s.sh\n", cfg.OutputFile)
	return nil
}

// LoginIAM handles IAM username/password authentication
func LoginIAM(cfg *config.Config, username, password string) error {
	// Step 1: Get unscoped token
	color.Yellow("⏳ Step 1: Authenticating with IAM...")
	iamClient := auth.NewIAMClient(cfg)
	unscopedToken, err := iamClient.GetIAMToken(username, password)
	if err != nil {
		return fmt.Errorf("IAM authentication failed: %w", err)
	}
	color.Green("✓ IAM authentication successful")

	// Step 2: Get domain token
	color.Yellow("⏳ Step 2: Getting domain token...")
	otcClient := otc.NewClient(cfg)
	domainToken, err := otcClient.GetDomainScopedToken(unscopedToken)
	if err != nil {
		return fmt.Errorf("failed to get domain token: %w", err)
	}
	color.Green("✓ Domain token obtained")

	// Step 3: List projects
	color.Yellow("⏳ Step 3: Listing projects...")
	projects, err := otcClient.ListProjects(domainToken)
	if err != nil {
		return fmt.Errorf("failed to list projects: %w", err)
	}

	if len(projects) == 0 {
		return fmt.Errorf("no projects found for domain %s", cfg.DomainName)
	}

	color.Green("✓ Found %d project(s)", len(projects))
	color.Cyan("  Using project: %s (%s)", projects[0].Name, projects[0].ID)

	// Step 4: Get project-scoped token
	color.Yellow("⏳ Step 4: Getting project token...")
	projectToken, err := otcClient.GetProjectScopedToken(unscopedToken, projects[0].ID)
	if err != nil {
		return fmt.Errorf("failed to get project token: %w", err)
	}
	color.Green("✓ Project token obtained")

	// Step 5: Create temporary credentials
	color.Yellow("⏳ Step 5: Creating temporary credentials...")
	creds, err := otcClient.CreateTemporaryCredentials(projectToken, 86400)
	if err != nil {
		return fmt.Errorf("failed to create credentials: %w", err)
	}
	color.Green("✓ Temporary credentials created")

	// Step 6: Save credentials script
	scriptPath := cfg.OutputFile + ".sh"
	if err := creds.SaveShellScript(scriptPath, cfg.Region); err != nil {
		return fmt.Errorf("failed to save credentials: %w", err)
	}

	// Cache token
	tokenCache := &cache.TokenCache{
		UnscopedToken: unscopedToken,
		IDToken:       unscopedToken,
		RefreshToken:  "",
		ExpiresAt:     time.Now().Add(24 * time.Hour),
		Domain:        cfg.DomainName,
		Region:        cfg.Region,
	}
	cache.SaveToken(tokenCache)

	// Success message
	color.Green("\n✓ Credentials saved to %s", scriptPath)
	color.Yellow("⏰ Expires: %s (24h)", creds.ExpiresAt)
	color.Cyan("\nLoad credentials:")
	fmt.Printf("  source %s\n", scriptPath)

	return nil
}

// PromptUsername prompts for username with env var fallback
func PromptUsername() string {
	username := os.Getenv("OS_USERNAME")
	if username != "" {
		return username
	}

	fmt.Print("IAM Username: ")
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

// PromptPassword prompts for password (hidden) with env var fallback
func PromptPassword() string {
	password := os.Getenv("OS_PASSWORD")
	if password != "" {
		return password
	}

	fmt.Print("IAM Password: ")
	bytePwd, _ := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	return strings.TrimSpace(string(bytePwd))
}
