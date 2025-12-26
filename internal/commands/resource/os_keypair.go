package resource

import (
	"encoding/json"
	"fmt"
	"github.com/abdo-farag/otc-cli/internal/config"
	"github.com/abdo-farag/otc-cli/internal/otc"

	"github.com/fatih/color"
	"github.com/rodaine/table"
)

// ListKeypairs lists all available SSH keypairs
func ListKeypairs(cfg *config.Config, client *otc.Client, unscopedToken, projectID string, raw bool) {
	projectID, projectToken, err := GetProjectToken(cfg, client, unscopedToken, projectID, raw)
	if err != nil {
		color.Red("✗ %v", err)
		return
	}

	// Keypairs endpoint - includes project ID
	keypairURL := fmt.Sprintf("https://ecs.%s.otc.t-systems.com/v2.1/%s/os-keypairs", cfg.Region, projectID)

	body, statusCode, err := MakeRequest(keypairURL, projectToken)
	if err != nil {
		color.Red("✗ Request failed: %v", err)
		return
	}

	if statusCode != 200 {
		color.Red("✗ API error (status %d): %s", statusCode, string(body))
		return
	}

	if raw {
		var prettyJSON map[string]interface{}
		json.Unmarshal(body, &prettyJSON)
		formatted, _ := json.MarshalIndent(prettyJSON, "", "  ")
		fmt.Println(string(formatted))
		return
	}

var result struct {
  Keypairs []struct {
    Keypair struct {
      Name        string `json:"name"`
      Fingerprint string `json:"fingerprint"`
      PublicKey   string `json:"public_key"`
    } `json:"keypair"`
  } `json:"keypairs"`
}

if err := json.Unmarshal(body, &result); err != nil {
  color.Red("✗ Failed to parse response: %v", err)
  return
}

if len(result.Keypairs) == 0 {
  color.Cyan("No keypairs found")
  return
}

headerFmt := color.New(color.FgCyan, color.Bold).SprintfFunc()
tbl := table.New("Name", "Fingerprint")
tbl.WithHeaderFormatter(headerFmt)

for _, item := range result.Keypairs {
  tbl.AddRow(item.Keypair.Name, item.Keypair.Fingerprint)
}

	fmt.Printf("\n")
	color.Cyan("Available Keypairs")
	tbl.Print()
	fmt.Printf("\nTotal: %d keypairs\n", len(result.Keypairs))
}
