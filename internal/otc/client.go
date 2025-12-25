package otc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"github.com/abdo-farag/otc-cli/internal/config"
	"time"

	"github.com/fatih/color"
)

type Client struct {
	cfg        *config.Config
	httpClient *http.Client
}

type Project struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ValidationStatusSetter is an interface for setting validation status
type ValidationStatusSetter interface {
	SetValidationStatus(status, message string)
}

func NewClient(cfg *config.Config) *Client {
	return &Client{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) GetUnscopedToken(idToken string, statusSetter ValidationStatusSetter) (string, error) {
	// Determine protocol (default to oidc if not set)
	protocol := c.cfg.IdpProtocol
	if protocol == "" {
		protocol = "oidc"
	}

	url := fmt.Sprintf("%s/v3/OS-FEDERATION/identity_providers/%s/protocols/%s/auth",
		c.cfg.AUTHURL, c.cfg.IDPProviderName, protocol)

	color.Yellow("⏳ Validating organization access with OTC (%s)...", protocol)

	// Set validation status for callback page
	if statusSetter != nil {
		statusSetter.SetValidationStatus("pending", "Validating...")
	}

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return "", err
	}

	// Set authentication header based on protocol
	if protocol == "saml" {
		// For SAML, the assertion goes in X-Auth-Token header
		req.Header.Set("X-Auth-Token", idToken)
	} else {
		// For OIDC, use Bearer token in Authorization header
		req.Header.Set("Authorization", "Bearer "+idToken)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		if statusSetter != nil {
			statusSetter.SetValidationStatus("failed", "Network error")
		}
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)

		var errResp struct {
			Error struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		json.Unmarshal(body, &errResp)

		errMsg := "Access denied"
		if errResp.Error.Message != "" {
			errMsg = errResp.Error.Message
		}

		if statusSetter != nil {
			statusSetter.SetValidationStatus("failed", errMsg)
		}
		time.Sleep(2 * time.Second)

		return "", fmt.Errorf("OTC authorization failed: %s", errMsg)
	}

	if statusSetter != nil {
		statusSetter.SetValidationStatus("success", "Your organization has been validated successfully!")
	}

	token := resp.Header.Get("X-Subject-Token")
	if token == "" {
		return "", fmt.Errorf("no X-Subject-Token in response")
	}

	return token, nil
}

func (c *Client) GetDomainScopedToken(unscopedToken string) (string, error) {
	url := fmt.Sprintf("%s/v3/auth/tokens", c.cfg.AUTHURL)

	payload := map[string]interface{}{
		"auth": map[string]interface{}{
			"identity": map[string]interface{}{
				"methods": []string{"token"},
				"token": map[string]string{
					"id": unscopedToken,
				},
			},
			"scope": map[string]interface{}{
				"domain": map[string]string{
					"name": c.cfg.DomainName,
				},
			},
		},
	}

	token, err := c.scopedTokenRequest(url, payload)
	if err != nil {
		return "", fmt.Errorf("failed to get domain-scoped token: %w", err)
	}

	return token, nil
}

func (c *Client) ListProjects(domainToken string) ([]Project, error) {
	url := fmt.Sprintf("%s/v3/auth/projects", c.cfg.AUTHURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Auth-Token", domainToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to list projects: %s", string(body))
	}

	var result struct {
		Projects []Project `json:"projects"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Projects, nil
}

func (c *Client) GetProjects(unscopedToken string) ([]Project, error) {
	// First get domain token to list projects
	domainToken, err := c.GetDomainScopedToken(unscopedToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get domain token: %w", err)
	}

	// Now list projects with domain token
	return c.ListProjects(domainToken)
}

func (c *Client) GetProjectScopedToken(unscopedToken, projectID string) (string, error) {
	url := fmt.Sprintf("%s/v3/auth/tokens", c.cfg.AUTHURL)

	payload := map[string]interface{}{
		"auth": map[string]interface{}{
			"identity": map[string]interface{}{
				"methods": []string{"token"},
				"token": map[string]string{
					"id": unscopedToken,
				},
			},
			"scope": map[string]interface{}{
				"project": map[string]string{
					"id": projectID,
				},
			},
		},
	}

	token, err := c.scopedTokenRequest(url, payload)
	if err != nil {
		return "", fmt.Errorf("failed to get project-scoped token: %w", err)
	}

	return token, nil
}

func (c *Client) scopedTokenRequest(url string, payload map[string]interface{}) (string, error) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("request failed: %s", string(body))
	}

	token := resp.Header.Get("X-Subject-Token")
	if token == "" {
		return "", fmt.Errorf("no X-Subject-Token in response")
	}

	return token, nil
}
