package resource

import (
	"encoding/json"
	"fmt"
	"github.com/abdo-farag/otc-cli/internal/config"
	"github.com/abdo-farag/otc-cli/internal/otc"

	"github.com/fatih/color"
	"github.com/rodaine/table"
)

// ListVolume lists all volumes
func ListVolume(cfg *config.Config, client *otc.Client, unscopedToken, projectID string, raw bool) {
	projectID, projectToken, err := GetProjectToken(cfg, client, unscopedToken, projectID, raw)
	if err != nil {
		color.Red("✗ %v", err)
		return
	}

	volumeURL := fmt.Sprintf("https://evs.%s.otc.t-systems.com/v2/%s/volumes/detail", cfg.Region, projectID)

	body, statusCode, err := MakeRequest(volumeURL, projectToken)
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
		Volumes []struct {
			ID               string `json:"id"`
			Name             string `json:"name"`
			Status           string `json:"status"`
			Size             int    `json:"size"`
			VolumeType       string `json:"volume_type"`
			AvailabilityZone string `json:"availability_zone"`
			Attachments      []struct {
				ServerID string `json:"server_id"`
				Device   string `json:"device"`
			} `json:"attachments"`
		} `json:"volumes"`
	}

	json.Unmarshal(body, &result)

	headerFmt := color.New(color.FgCyan, color.Bold).SprintfFunc()
	tbl := table.New("Name", "Size (GB)", "Type", "Status", "Attached To", "Device", "AZ")
	tbl.WithHeaderFormatter(headerFmt)

	for _, v := range result.Volumes {
		serverID := ""
		device := ""
		if len(v.Attachments) > 0 {
			serverID = v.Attachments[0].ServerID
			device = v.Attachments[0].Device
		}
		tbl.AddRow(v.Name, v.Size, v.VolumeType, v.Status, serverID, device, v.AvailabilityZone)
	}

	fmt.Printf("\n")
	color.Cyan("Project: %s", projectID)
	tbl.Print()
	fmt.Printf("\nTotal: %d volumes\n", len(result.Volumes))
}

// GetVolume gets a specific volume
func GetVolume(cfg *config.Config, client *otc.Client, unscopedToken, projectID, resourceID string, raw bool) {
	projectID, projectToken, err := GetProjectToken(cfg, client, unscopedToken, projectID, raw)
	if err != nil {
		color.Red("✗ %v", err)
		return
	}

	volumeURL := fmt.Sprintf("https://evs.%s.otc.t-systems.com/v2/%s/volumes/%s", cfg.Region, projectID, resourceID)

	body, statusCode, err := MakeRequest(volumeURL, projectToken)
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
