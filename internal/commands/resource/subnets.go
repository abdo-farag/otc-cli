package resource

import (
	"encoding/json"
	"fmt"
	"github.com/abdo-farag/otc-cli/internal/config"
	"github.com/abdo-farag/otc-cli/internal/otc"

	"github.com/fatih/color"
	"github.com/rodaine/table"
)

// ListSubnet lists all subnets
func ListSubnet(cfg *config.Config, client *otc.Client, unscopedToken, projectID string, raw bool) {
	projectID, projectToken, err := GetProjectToken(cfg, client, unscopedToken, projectID, raw)
	if err != nil {
		color.Red("✗ %v", err)
		return
	}

	subnetURL := fmt.Sprintf("https://vpc.%s.otc.t-systems.com/v1/%s/subnets", cfg.Region, projectID)

	body, statusCode, err := MakeRequest(subnetURL, projectToken)
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
		Subnets []struct {
			ID               string `json:"id"`
			Name             string `json:"name"`
			CIDR             string `json:"cidr"`
			GatewayIP        string `json:"gateway_ip"`
			VpcID            string `json:"vpc_id"`
			AvailabilityZone string `json:"availability_zone"`
			Status           string `json:"status"`
		} `json:"subnets"`
	}

	json.Unmarshal(body, &result)

	headerFmt := color.New(color.FgCyan, color.Bold).SprintfFunc()
	tbl := table.New("ID", "Name", "CIDR", "Gateway", "VPC ID", "Status")
	tbl.WithHeaderFormatter(headerFmt)

	for _, s := range result.Subnets {
		tbl.AddRow(s.ID, s.Name, s.CIDR, s.GatewayIP, s.VpcID, s.Status)
	}

	fmt.Printf("\n")
	color.Cyan("Project: %s", projectID)
	tbl.Print()
	fmt.Printf("\nTotal: %d subnets\n", len(result.Subnets))
}

// GetSubnet gets a specific subnet
func GetSubnet(cfg *config.Config, client *otc.Client, unscopedToken, projectID, resourceID string, raw bool) {
	projectID, projectToken, err := GetProjectToken(cfg, client, unscopedToken, projectID, raw)
	if err != nil {
		color.Red("✗ %v", err)
		return
	}

	subnetURL := fmt.Sprintf("https://vpc.%s.otc.t-systems.com/v1/%s/subnets/%s", cfg.Region, projectID, resourceID)

	body, statusCode, err := MakeRequest(subnetURL, projectToken)
	if err != nil {
		color.Red("✗ Request failed: %v", err)
		return
	}

	if statusCode != 200 {
		color.Red("✗ API error (status %d): %s", statusCode, string(body))
		return
	}

	var prettyJSON map[string]interface{}
	json.Unmarshal(body, &prettyJSON)
	formatted, _ := json.MarshalIndent(prettyJSON, "", "  ")
	fmt.Println(string(formatted))
}
