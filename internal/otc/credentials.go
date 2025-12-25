package otc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type Credentials struct {
	Access        string `json:"access"`
	Secret        string `json:"secret"`
	SecurityToken string `json:"securitytoken"`
	ExpiresAt     string `json:"expires_at"`
}

func (c *Client) CreateTemporaryCredentials(projectToken string, durationSeconds int) (*Credentials, error) {
	url := fmt.Sprintf("%s/v3.0/OS-CREDENTIAL/securitytokens", c.cfg.AUTHURL)

	payload := map[string]interface{}{
		"auth": map[string]interface{}{
			"identity": map[string]interface{}{
				"methods": []string{"token"},
				"token": map[string]int{
					"duration_seconds": durationSeconds,
				},
			},
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Auth-Token", projectToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to create credentials: %s", string(body))
	}

	var result struct {
		Credential Credentials `json:"credential"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result.Credential, nil
}

func (c *Credentials) SaveShellScript(filename, region string) error {
	script := fmt.Sprintf(`#!/bin/bash
# CloudAstro SSO - OTC Temporary Credentials
# Generated: %s
# Expires: %s
# Valid for: 24 hours

export OS_REGION_NAME=%s
export OS_ACCESS_KEY="%s"
export OS_SECRET_KEY="%s"
export OS_SECURITY_TOKEN="%s"

export AWS_ACCESS_KEY_ID="$OS_ACCESS_KEY"
export AWS_SECRET_ACCESS_KEY="$OS_SECRET_KEY"
export AWS_SESSION_TOKEN="$OS_SECURITY_TOKEN"

echo "✓ OTC Temporary Credentials loaded"
echo "  Expires: %s"
echo "  Region: %s"
`,
		time.Now().Format("2006-01-02 15:04:05"),
		c.ExpiresAt,
		region,
		c.Access,
		c.Secret,
		c.SecurityToken,
		c.ExpiresAt,
		region,
	)

	if err := os.WriteFile(filename, []byte(script), 0755); err != nil {
		return err
	}

	return nil
}
