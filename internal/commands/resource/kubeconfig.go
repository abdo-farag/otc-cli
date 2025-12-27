package resource

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"github.com/abdo-farag/otc-cli/internal/config"
	"github.com/abdo-farag/otc-cli/internal/otc"
	"time"

	"github.com/fatih/color"
)

func GetKubeconfig(cfg *config.Config, client *otc.Client, unscopedToken, projectID, clusterNameOrID, outputPath string) {
	// Get project token
	projectID, projectToken, err := GetProjectToken(cfg, client, unscopedToken, projectID, false)
	if err != nil {
		color.Red("✗ %v", err)
		return
	}

	color.Yellow("⏳ Finding cluster...")

	// First, list clusters to find by name if needed
	cceURL := fmt.Sprintf("https://cce.%s.otc.t-systems.com/api/v3/projects/%s/clusters", cfg.Region, projectID)

	req, _ := http.NewRequest("GET", cceURL, nil)
	req.Header.Set("X-Auth-Token", projectToken)
	req.Header.Set("Content-Type", "application/json")

	httpClient := &http.Client{Timeout: 30 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		color.Red("✗ Request failed: %v", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		color.Red("✗ API error (status %d): %s", resp.StatusCode, string(body))
		return
	}

	var clusterList struct {
		Clusters []struct {
			Metadata struct {
				UID  string `json:"uid"`
				Name string `json:"name"`
			} `json:"metadata"`
		} `json:"items"`
	}

	json.Unmarshal(body, &clusterList)

	// Find cluster by name or ID
	var clusterID string
	for _, c := range clusterList.Clusters {
		if c.Metadata.UID == clusterNameOrID || c.Metadata.Name == clusterNameOrID {
			clusterID = c.Metadata.UID
			color.Cyan("✓ Found cluster: %s (%s)", c.Metadata.Name, clusterID)
			break
		}
	}

	if clusterID == "" {
		color.Red("✗ Cluster not found: %s", clusterNameOrID)
		return
	}

	// Get kubeconfig
	color.Yellow("⏳ Downloading kubeconfig...")
	kubeconfigURL := fmt.Sprintf("https://cce.%s.otc.t-systems.com/api/v3/projects/%s/clusters/%s/clustercert",
		cfg.Region, projectID, clusterID)

	req2, _ := http.NewRequest("GET", kubeconfigURL, nil)
	req2.Header.Set("X-Auth-Token", projectToken)
	req2.Header.Set("Content-Type", "application/json")

	resp2, err := httpClient.Do(req2)
	if err != nil {
		color.Red("✗ Request failed: %v", err)
		return
	}
	defer resp2.Body.Close()

	body2, _ := io.ReadAll(resp2.Body)

	if resp2.StatusCode != 200 {
		color.Red("✗ API error (status %d): %s", resp2.StatusCode, string(body2))
		return
	}

	var kubeconfigResp struct {
		Kubeconfig string `json:"kubeconfig"`
	}

	if err := json.Unmarshal(body2, &kubeconfigResp); err != nil {
		color.Red("✗ Failed to parse kubeconfig: %v", err)
		return
	}

	// Save to file
	if err := os.WriteFile(outputPath, []byte(kubeconfigResp.Kubeconfig), 0600); err != nil {
		color.Red("✗ Failed to save kubeconfig: %v", err)
		return
	}

	color.Green("✓ Kubeconfig saved to: %s", outputPath)
	color.Cyan("\nUsage:")
	fmt.Printf("  export KUBECONFIG=%s\n", outputPath)
	fmt.Printf("  kubectl get nodes\n")
}
