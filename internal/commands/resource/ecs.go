package resource

import (
  "encoding/json"
  "fmt"
  "strings"

  "github.com/abdo-farag/otc-cli/internal/config"
  "github.com/abdo-farag/otc-cli/internal/otc"

  "github.com/fatih/color"
  "github.com/rodaine/table"
	
	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack"
  "github.com/opentelekomcloud/gophertelekomcloud/openstack/ecs/v1/cloudservers"
)

// ListECS lists all ECS instances with optional filters
func ListECS(cfg *config.Config, client *otc.Client, unscopedToken, projectID string, options map[string]interface{}, raw bool) {
  projectID, projectToken, err := GetProjectToken(cfg, client, unscopedToken, projectID, raw)
  if err != nil {
    color.Red("✗ %v", err)
    return
  }

  // Use ECS v1 API endpoint directly (SDK doesn't have List method)
  computeURL := fmt.Sprintf("https://ecs.%s.otc.t-systems.com/v1/%s/cloudservers/detail", cfg.Region, projectID)

  body, statusCode, err := MakeRequest(computeURL, projectToken)
  if err != nil {
    color.Red("✗ Request failed: %v", err)
    return
  }

  if statusCode != 200 {
    color.Red("✗ API error (status %d): %s", statusCode, string(body))
    return
  }

  // Use SDK's CloudServer struct for type safety
  var result struct {
    Servers []cloudservers.CloudServer `json:"servers"`
  }

  if err := json.Unmarshal(body, &result); err != nil {
    color.Red("✗ Failed to parse response: %v", err)
    return
  }

  // Apply client-side filters
  servers := result.Servers

  // Filter by availability zone
  if azFilter, ok := options["az"].(string); ok && azFilter != "" {
    var filtered []cloudservers.CloudServer
    for _, s := range servers {
      if s.AvailabilityZone == azFilter {
        filtered = append(filtered, s)
      }
    }
    servers = filtered
  }

  // Filter by status
  if statusFilter, ok := options["status"].(string); ok && statusFilter != "" {
    var filtered []cloudservers.CloudServer
    for _, s := range servers {
      if strings.EqualFold(s.Status, statusFilter) {
        filtered = append(filtered, s)
      }
    }
    servers = filtered
  }

  // Filter by name (partial match)
  if nameFilter, ok := options["name"].(string); ok && nameFilter != "" {
    var filtered []cloudservers.CloudServer
    for _, s := range servers {
      if strings.Contains(strings.ToLower(s.Name), strings.ToLower(nameFilter)) {
        filtered = append(filtered, s)
      }
    }
    servers = filtered
  }

  // Filter by tag (tags are array of "key=value" strings)
  if tagFilter, ok := options["tag"].(string); ok && tagFilter != "" {
    var filtered []cloudservers.CloudServer
    parts := strings.SplitN(tagFilter, "=", 2)

    for _, s := range servers {
      matched := false

      if len(parts) == 2 {
        // Full key=value match
        searchTag := tagFilter
        for _, tag := range s.Tags {
          if tag == searchTag {
            matched = true
            break
          }
        }
      } else {
        // Key only match
        searchKey := tagFilter + "="
        for _, tag := range s.Tags {
          if strings.HasPrefix(tag, searchKey) {
            matched = true
            break
          }
        }
      }

      if matched {
        filtered = append(filtered, s)
      }
    }
    servers = filtered
  }

  if raw {
    jsonData, _ := json.MarshalIndent(servers, "", "  ")
    fmt.Println(string(jsonData))
    return
  }

  // Display table
  displayServersTable(servers, projectID, options)
}

// extractIPv4FromAddresses extracts IPv4 addresses from server addresses
func extractIPv4FromAddresses(addresses map[string][]cloudservers.Address) []string {
  var ipv4s []string

  for _, addrs := range addresses {
    for _, addr := range addrs {
      if addr.Version == "4" {
        ipv4s = append(ipv4s, addr.Addr)
      }
    }
  }

  return ipv4s
}

func displayServersTable(servers []cloudservers.CloudServer, projectID string, options map[string]interface{}) {
  headerFmt := color.New(color.FgCyan, color.Bold).SprintfFunc()
  tbl := table.New("Name", "Status", "IPv4", "Flavor", "AZ", "ID")
  tbl.WithHeaderFormatter(headerFmt)

  for _, s := range servers {
    flavorID := s.Flavor.ID
    if flavorID == "" {
      flavorID = "-"
    }

    az := s.AvailabilityZone
    if az == "" {
      az = "-"
    }

    // Extract IPv4 addresses
    ipv4s := extractIPv4FromAddresses(s.Addresses)
    ips := "-"
    if len(ipv4s) > 0 {
      ips = strings.Join(ipv4s, ", ")
    }

    tbl.AddRow(s.Name, s.Status, ips, flavorID, az, s.ID)
  }

  fmt.Printf("\n")
  color.Cyan("Project: %s", projectID)

  // Show active filters
  if az, ok := options["az"].(string); ok && az != "" {
    color.Yellow("Filter: AZ = %s", az)
  }
  if status, ok := options["status"].(string); ok && status != "" {
    color.Yellow("Filter: Status = %s", status)
  }
  if name, ok := options["name"].(string); ok && name != "" {
    color.Yellow("Filter: Name contains '%s'", name)
  }
  if tag, ok := options["tag"].(string); ok && tag != "" {
    color.Yellow("Filter: Tag = %s", tag)
  }

  tbl.Print()
  fmt.Printf("\nTotal: %d instances\n", len(servers))
}

// GetECS gets a specific ECS instance using SDK
func GetECS(cfg *config.Config, client *otc.Client, unscopedToken, projectID, resourceID string, raw bool) {
	projectID, projectToken, err := GetProjectToken(cfg, client, unscopedToken, projectID, raw)
	if err != nil {
		color.Red("✗ %v", err)
		return
	}

	// Create authenticated client
	authOpts := golangsdk.AuthOptions{
		IdentityEndpoint: cfg.AUTHURL,
		TokenID:          projectToken,
		TenantID:         projectID,
	}

	provider, err := openstack.AuthenticatedClient(authOpts)
	if err != nil {
		color.Red("✗ Failed to create authenticated client: %v", err)
		return
	}

	// Create ECS client
	ecsClient, err := openstack.NewComputeV1(provider, golangsdk.EndpointOpts{
		Region: cfg.Region,
	})
	if err != nil {
		color.Red("✗ Failed to create ECS client: %v", err)
		return
	}

	// Use SDK's Get method
	server, err := cloudservers.Get(ecsClient, resourceID).Extract()
	if err != nil {
		color.Red("✗ Failed to get server: %v", err)
		return
	}

	// Output as JSON
	jsonData, err := json.MarshalIndent(server, "", "  ")
	if err != nil {
		color.Red("✗ Failed to format output: %v", err)
		return
	}

	fmt.Println(string(jsonData))
}
