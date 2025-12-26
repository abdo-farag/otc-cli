package resource

import (
	"encoding/json"
	"fmt"
	"github.com/abdo-farag/otc-cli/internal/config"
	"github.com/abdo-farag/otc-cli/internal/otc"

	"github.com/fatih/color"
	"github.com/rodaine/table"
)

// ListVPC lists all VPCs
func ListVPC(cfg *config.Config, client *otc.Client, unscopedToken, projectID string, raw bool) {
	projectID, projectToken, err := GetProjectToken(cfg, client, unscopedToken, projectID, raw)
	if err != nil {
		color.Red("✗ %v", err)
		return
	}

	vpcURL := fmt.Sprintf("https://vpc.%s.otc.t-systems.com/v1/%s/vpcs", cfg.Region, projectID)

	body, statusCode, err := MakeRequest(vpcURL, projectToken)
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
		VPCs []struct {
			ID     string `json:"id"`
			Name   string `json:"name"`
			CIDR   string `json:"cidr"`
			Status string `json:"status"`
		} `json:"vpcs"`
	}

	json.Unmarshal(body, &result)

	headerFmt := color.New(color.FgCyan, color.Bold).SprintfFunc()
	tbl := table.New("Name", "ID", "CIDR", "Status")
	tbl.WithHeaderFormatter(headerFmt)

	for _, v := range result.VPCs {
		tbl.AddRow(v.Name, v.ID, v.CIDR, v.Status)
	}

	fmt.Printf("\n")
	color.Cyan("Project: %s", projectID)
	tbl.Print()
	fmt.Printf("\nTotal: %d VPCs\n", len(result.VPCs))
}
