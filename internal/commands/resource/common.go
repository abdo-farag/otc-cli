package resource

import (
	"fmt"
	"io"
	"net/http"
	"github.com/abdo-farag/otc-cli/internal/config"
	"github.com/abdo-farag/otc-cli/internal/otc"
	"time"

	"github.com/fatih/color"
)

// GetProjectToken gets or resolves a project token
func GetProjectToken(cfg *config.Config, client *otc.Client, unscopedToken, projectID string, raw bool) (string, string, error) {
	if projectID == "" {
		domainToken, err := client.GetDomainScopedToken(unscopedToken)
		if err != nil {
			return "", "", fmt.Errorf("failed to get domain token: %w", err)
		}

		projects, err := client.ListProjects(domainToken)
		if err != nil || len(projects) == 0 {
			return "", "", fmt.Errorf("no projects found")
		}

		projectID = projects[0].ID
		if !raw {
			color.Cyan("✓ Using default project: %s (%s)", projects[0].Name, projectID)
		}
	}

	projectToken, err := client.GetProjectScopedToken(unscopedToken, projectID)
	if err != nil {
		return "", "", fmt.Errorf("failed to get project token: %w", err)
	}

	return projectID, projectToken, nil
}

// MakeRequest makes an authenticated HTTP GET request
func MakeRequest(url, token string) ([]byte, int, error) {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("X-Auth-Token", token)
	req.Header.Set("Content-Type", "application/json")

	httpClient := &http.Client{Timeout: 30 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	return body, resp.StatusCode, nil
}
