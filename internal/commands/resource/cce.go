package resource

import (
	"encoding/json"
	"fmt"
	"github.com/abdo-farag/otc-cli/internal/config"
	"github.com/abdo-farag/otc-cli/internal/otc"

	"github.com/fatih/color"
	"github.com/rodaine/table"
)

// ListCCE lists all CCE clusters
func ListCCE(cfg *config.Config, client *otc.Client, unscopedToken, projectID string, raw bool) {
	projectID, projectToken, err := GetProjectToken(cfg, client, unscopedToken, projectID, raw)
	if err != nil {
		color.Red("✗ %v", err)
		return
	}

	cceURL := fmt.Sprintf("https://cce.%s.otc.t-systems.com/api/v3/projects/%s/clusters", cfg.Region, projectID)

	body, statusCode, err := MakeRequest(cceURL, projectToken)
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
		Clusters []struct {
			Metadata struct {
				UID  string `json:"uid"`
				Name string `json:"name"`
			} `json:"metadata"`
			Spec struct {
				Type    string `json:"type"`
				Flavor  string `json:"flavor"`
				Version string `json:"version"`
			} `json:"spec"`
			Status struct {
				Phase string `json:"phase"`
			} `json:"status"`
		} `json:"items"`
	}

	json.Unmarshal(body, &result)

	headerFmt := color.New(color.FgCyan, color.Bold).SprintfFunc()
	tbl := table.New("Name", "Status", "Type", "Flavor", "Version", "ID")
	tbl.WithHeaderFormatter(headerFmt)

	for _, c := range result.Clusters {
		tbl.AddRow(c.Metadata.Name, c.Status.Phase, c.Spec.Type, c.Spec.Flavor, c.Spec.Version, c.Metadata.UID)
	}

	fmt.Printf("\n")
	color.Cyan("Project: %s", projectID)
	tbl.Print()
	fmt.Printf("\nTotal: %d clusters\n", len(result.Clusters))
}
