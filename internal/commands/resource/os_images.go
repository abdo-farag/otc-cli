package resource

import (
	"encoding/json"
	"fmt"
	"strings"
	"github.com/abdo-farag/otc-cli/internal/config"
	"github.com/abdo-farag/otc-cli/internal/otc"

	"github.com/fatih/color"
	"github.com/rodaine/table"
)

// ListImages lists all available ECS images
func ListImages(cfg *config.Config, client *otc.Client, unscopedToken, projectID string, options map[string]interface{}, raw bool) {
	projectID, projectToken, err := GetProjectToken(cfg, client, unscopedToken, projectID, raw)
	if err != nil {
		color.Red("✗ %v", err)
		return
	}

	// Using IMS (Image Management Service) API for images
	imageURL := fmt.Sprintf("https://ims.%s.otc.t-systems.com/v2/cloudimages", cfg.Region)

	// Build query parameters for filtering
	queryParams := []string{}
	
	// Filter by visibility (private/public/shared)
	if visibility, ok := options["visibility"].(string); ok && visibility != "" {
		if strings.EqualFold(visibility, "public") {
			queryParams = append(queryParams, "__imagetype=gold")
		} else if strings.EqualFold(visibility, "private") {
			queryParams = append(queryParams, "__imagetype=private")
		} else if strings.EqualFold(visibility, "shared") {
			queryParams = append(queryParams, "__imagetype=shared")
		}
	}

	// Filter by OS type
	if osType, ok := options["os"].(string); ok && osType != "" {
		queryParams = append(queryParams, fmt.Sprintf("__os_type=%s", osType))
	}

	// Filter by platform (case-insensitive)
	if platform, ok := options["platform"].(string); ok && platform != "" {
		platform = strings.Title(strings.ToLower(platform))
		queryParams = append(queryParams, fmt.Sprintf("__platform=%s", platform))
	}

	// Filter by status
	if status, ok := options["status"].(string); ok && status != "" {
		queryParams = append(queryParams, fmt.Sprintf("status=%s", status))
	}

	if len(queryParams) > 0 {
		imageURL = imageURL + "?" + strings.Join(queryParams, "&")
	}

	body, statusCode, err := MakeRequest(imageURL, projectToken)
	if err != nil {
		color.Red("✗ Request failed: %v", err)
		return
	}

	if statusCode != 200 {
		color.Red("✗ API error (status %d): %s", statusCode, string(body))
		return
	}

	var result struct {
		Images []struct {
			ID              string `json:"id"`
			Name            string `json:"name"`
			Status          string `json:"status"`
			MinDisk         int    `json:"min_disk"`
			MinRAM          int    `json:"min_ram"`
			DiskSize        int    `json:"disk_size"`
			OsType          string `json:"__os_type"`
			Platform        string `json:"__platform"`
			ImageType       string `json:"__imagetype"`
			Visibility      string `json:"visibility"`
			OsVersion       string `json:"__os_version"`
			SupportKvm      string `json:"__support_kvm"`
			VirtualEnvType  string `json:"virtual_env_type"`
		} `json:"images"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		color.Red("✗ Failed to parse response: %v", err)
		return
	}

	// Apply client-side name filtering (case-insensitive contains)
	images := result.Images
	if nameFilter, ok := options["name"].(string); ok && nameFilter != "" {
		var filtered []struct {
			ID              string `json:"id"`
			Name            string `json:"name"`
			Status          string `json:"status"`
			MinDisk         int    `json:"min_disk"`
			MinRAM          int    `json:"min_ram"`
			DiskSize        int    `json:"disk_size"`
			OsType          string `json:"__os_type"`
			Platform        string `json:"__platform"`
			ImageType       string `json:"__imagetype"`
			Visibility      string `json:"visibility"`
			OsVersion       string `json:"__os_version"`
			SupportKvm      string `json:"__support_kvm"`
			VirtualEnvType  string `json:"virtual_env_type"`
		}
		for _, img := range images {
			if strings.Contains(strings.ToLower(img.Name), strings.ToLower(nameFilter)) {
				filtered = append(filtered, img)
			}
		}
		images = filtered
	}

	if raw {
		jsonData, _ := json.MarshalIndent(images, "", "  ")
		fmt.Println(string(jsonData))
		return
	}

	if len(images) == 0 {
		color.Yellow("⚠ No images found matching the filters")
		
		// If filtered and no results, suggest trying without filters
		if visibility, ok := options["visibility"].(string); ok && visibility != "" {
			color.Cyan("\nTip: Try listing all images first to see what's available:")
			color.Cyan("  otc-cli list images")
		}
		return
	}

	headerFmt := color.New(color.FgCyan, color.Bold).SprintfFunc()
	tbl := table.New("Name", "Platform", "Type", "Status", "ID")
	tbl.WithHeaderFormatter(headerFmt)

	for _, img := range images {
		platform := img.Platform
		if platform == "" {
			platform = img.OsType
		}
		if platform == "" {
			platform = "-"
		}

		imageType := img.ImageType
		if imageType == "" {
			imageType = img.Visibility
		}

		tbl.AddRow(img.Name, platform, imageType, img.Status, img.ID)
	}

	fmt.Printf("\n")
	color.Cyan("Project: %s", projectID)
	
	// Show active filters
	if visibility, ok := options["visibility"].(string); ok && visibility != "" {
		color.Yellow("Filter: Visibility = %s", visibility)
	}
	if osType, ok := options["os"].(string); ok && osType != "" {
		color.Yellow("Filter: OS Type = %s", osType)
	}
	if platform, ok := options["platform"].(string); ok && platform != "" {
		color.Yellow("Filter: Platform = %s", platform)
	}
	if name, ok := options["name"].(string); ok && name != "" {
		color.Yellow("Filter: Name contains '%s'", name)
	}
	
	tbl.Print()
	fmt.Printf("\nTotal: %d images\n", len(images))
}