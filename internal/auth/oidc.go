package auth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/abdo-farag/otc-cli/internal/config"
	"github.com/abdo-farag/otc-cli/web"
	
	"github.com/fatih/color"
	"github.com/pkg/browser"
)

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	IDToken      string `json:"id_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

type Client struct {
	cfg *config.Config
}

func NewClient(cfg *config.Config) *Client {
	return &Client{cfg: cfg}
}

func (c *Client) GetOIDCToken() (*TokenResponse, *CallbackHandler, error) {
	// Validate code challenge method
	if err := c.cfg.ValidateCodeChallengeMethod(); err != nil {
		return nil, nil, err
	}

	// Generate PKCE challenge with configured method
	pkce, err := GeneratePKCEWithMethod(c.cfg.CodeChallengeMethod)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate PKCE: %w", err)
	}

	state, err := GenerateState()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate state: %w", err)
	}

	redirectURI := fmt.Sprintf("http://localhost:%d/oidc/auth", c.cfg.RedirectPort)

	// Build auth URL with configurable code challenge method and scopes
	authURLParams := url.Values{}
	authURLParams.Set("client_id", c.cfg.IdpClientID)
	authURLParams.Set("response_type", "code")
	authURLParams.Set("scope", c.cfg.Scope)
	authURLParams.Set("redirect_uri", redirectURI)
	authURLParams.Set("state", state)
	authURLParams.Set("code_challenge", pkce.Challenge)
	authURLParams.Set("code_challenge_method", c.cfg.CodeChallengeMethod)

	authURL := fmt.Sprintf("%s/protocol/openid-connect/auth?%s",
		c.cfg.IdpURL,
		authURLParams.Encode(),
	)

	callbackHandler := NewCallbackHandler(c.cfg.RedirectPort, web.Content)

	if err := callbackHandler.StartServer(); err != nil {
		return nil, nil, fmt.Errorf("failed to start callback server: %w", err)
	}

	if !c.cfg.NoBrowser {
		color.Cyan("🌐 Opening browser for authentication...")
		if err := browser.OpenURL(authURL); err != nil {
			color.Yellow("⚠ Could not open browser automatically")
			fmt.Printf("Please visit: %s\n", authURL)
		}
	} else {
		color.Cyan("🌐 Please visit this URL in your browser:")
		fmt.Println(authURL)
	}

	color.Yellow("⏳ Waiting for authentication...")

	authCode, err := callbackHandler.WaitForCode(300 * time.Second)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get authorization code: %w", err)
	}

	color.Green("✓ Authorization code received")
	color.Yellow("⏳ Exchanging code for token...")

	tokenResp, err := c.exchangeCodeForToken(authCode, redirectURI, pkce.Verifier)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to exchange token: %w", err)
	}

	return tokenResp, callbackHandler, nil
}

func (c *Client) exchangeCodeForToken(code, redirectURI, codeVerifier string) (*TokenResponse, error) {
	tokenURL := fmt.Sprintf("%s/protocol/openid-connect/token", c.cfg.IdpURL)

	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("client_id", c.cfg.IdpClientID)
	data.Set("code", code)
	data.Set("redirect_uri", redirectURI)
	data.Set("code_verifier", codeVerifier)

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token exchange failed: %s", string(body))
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, err
	}

	return &tokenResp, nil
}
