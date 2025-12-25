package resource

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"github.com/abdo-farag/otc-cli/internal/config"
	"github.com/abdo-farag/otc-cli/internal/otc"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/rodaine/table"
)

// parsePrice extracts numeric value from price string
func parsePrice(priceStr string) float64 {
	parts := strings.Fields(priceStr)
	if len(parts) > 0 {
		if val, err := strconv.ParseFloat(parts[0], 64); err == nil {
			return val
		}
	}
	return 0.0
}

// extractVCPUCount extracts numeric vCPU value for sorting
func extractVCPUCount(vCPUStr string) int {
	val, _ := strconv.Atoi(strings.TrimSpace(vCPUStr))
	return val
}

// extractRAMSize extracts numeric RAM value in GiB for sorting
func extractRAMSize(ramStr string) float64 {
	parts := strings.Fields(ramStr)
	if len(parts) > 0 {
		if val, err := strconv.ParseFloat(parts[0], 64); err == nil {
			return val
		}
	}
	return 0.0
}

// PricingInfo holds pricing data from OTC Price API
type PricingInfo struct {
	OpiFlavour         string
	vCPU               string
	RAM                string
	HourlyCost         float64
	MonthlyCost        float64
	OSUnit             string
	ProductIdParameter string
}

// FetchFlavorPricing fetches pricing from OTC price API with real specs
func FetchFlavorPricing(region string, osType string) (map[string]PricingInfo, error) {
	pricing := make(map[string]PricingInfo)

	// Query all service types: ecs, ecsnoc, gpu, deh
	serviceNames := []string{"ecs", "ecsnoc", "memo", "uhio", "hps", "gpu", "deh", "dehl"}

	for _, serviceName := range serviceNames {
		pricingURL := fmt.Sprintf("https://calculator.otc-service.com/en/open-telekom-price-api/?serviceName=%s&region=%s&limitMax=1000", serviceName, region)

		req, _ := http.NewRequest("GET", pricingURL, nil)
		req.Header.Set("Content-Type", "application/json")

		httpClient := &http.Client{Timeout: 15 * time.Second}
		resp, err := httpClient.Do(req)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)

		// Generic response structure that works for all service types
		var apiResponse struct {
			Response struct {
				Result map[string][]struct {
					OpiFlavour         string `json:"opiFlavour"`
					VCPU               string `json:"vCpu"`
					RAM                string `json:"ram"`
					PriceAmount        string `json:"priceAmount"`
					R12                string `json:"R12"`
					OSUnit             string `json:"osUnit"`
					Region             string `json:"region"`
					ProductIdParameter string `json:"productIdParameter"`
				} `json:"result"`
			} `json:"response"`
		}

		if err := json.Unmarshal(body, &apiResponse); err != nil {
			continue
		}

		// Parse pricing data - iterate through result map
		flavorSeen := make(map[string]bool)

		for _, items := range apiResponse.Response.Result {
			for _, item := range items {
				if item.Region != region {
					continue
				}

				// Match selected OS type
				osMatches := false
				osLower := strings.ToLower(item.OSUnit)
				osTypeLower := strings.ToLower(osType)

				switch osTypeLower {
				case "openlinux":
					osMatches = strings.Contains(osLower, "open") ||
						(strings.Contains(osLower, "linux") &&
							!strings.Contains(osLower, "suse") &&
							!strings.Contains(osLower, "redhat") &&
							!strings.Contains(osLower, "oracle") &&
							!strings.Contains(osLower, "windows"))
				case "suse", "suse linux":
					osMatches = strings.Contains(osLower, "suse")
				case "redhat":
					osMatches = strings.Contains(osLower, "redhat")
				case "oracle":
					osMatches = strings.Contains(osLower, "oracle")
				case "windows":
					osMatches = strings.Contains(osLower, "windows")
				default:
					osMatches = strings.Contains(osLower, "open") ||
						(strings.Contains(osLower, "linux") &&
							!strings.Contains(osLower, "suse") &&
							!strings.Contains(osLower, "redhat") &&
							!strings.Contains(osLower, "oracle") &&
							!strings.Contains(osLower, "windows"))
				}

				// Use first matching OS variant for each flavor
				if osMatches && !flavorSeen[item.OpiFlavour] {
					hourly := parsePrice(item.PriceAmount)
					monthly := hourly * 730

					pricing[item.OpiFlavour] = PricingInfo{
						OpiFlavour:         item.OpiFlavour,
						vCPU:               item.VCPU,
						RAM:                item.RAM,
						HourlyCost:         hourly,
						MonthlyCost:        monthly,
						OSUnit:             item.OSUnit,
						ProductIdParameter: item.ProductIdParameter,
					}
					flavorSeen[item.OpiFlavour] = true
				}
			}
		}
	}

	return pricing, nil
}

// ListPricing lists pricing for all servers from OTC API - sorted
func ListFlavors(cfg *config.Config, client *otc.Client, unscopedToken, projectID string, raw bool, osType string) {
	// Default to OpenLinux if not specified
	if osType == "" {
		osType = "openlinux"
	}

	// Fetch pricing data
	pricing, err := FetchFlavorPricing(cfg.Region, osType)
	if err != nil {
		color.Red("✗ Failed to fetch pricing: %v", err)
		return
	}

	if len(pricing) == 0 {
		color.Red("✗ No pricing data found for region: %s with OS: %s", cfg.Region, osType)
		return
	}

	if raw {
		formatted, _ := json.MarshalIndent(pricing, "", "  ")
		fmt.Println(string(formatted))
		return
	}

	// Convert map to slice for sorting
	var pricingList []PricingInfo
	for _, p := range pricing {
		pricingList = append(pricingList, p)
	}

	// Sort by vCPU count, then by RAM size, then by hourly cost
	sort.Slice(pricingList, func(i, j int) bool {
		vCPUI := extractVCPUCount(pricingList[i].vCPU)
		vCPUJ := extractVCPUCount(pricingList[j].vCPU)

		if vCPUI != vCPUJ {
			return vCPUI < vCPUJ
		}

		ramI := extractRAMSize(pricingList[i].RAM)
		ramJ := extractRAMSize(pricingList[j].RAM)

		if ramI != ramJ {
			return ramI < ramJ
		}

		return pricingList[i].HourlyCost < pricingList[j].HourlyCost
	})

	// Create table with header formatter
	headerFmt := color.New(color.FgCyan, color.Bold).SprintfFunc()
	tbl := table.New("Flavor ID", "Type", "vCPUs", "RAM", "Cost/Hour", "Cost/Month")
	tbl.WithHeaderFormatter(headerFmt)

	// Add rows
	for _, p := range pricingList {
		// Skip zero CPU/RAM entries
		vCPUCount := extractVCPUCount(p.vCPU)
		ramSize := extractRAMSize(p.RAM)
		if vCPUCount == 0 || ramSize == 0 {
			continue
		}

		// Format price, show N/A if zero
		priceStr := fmt.Sprintf("€%.4f", p.HourlyCost)
		if p.HourlyCost == 0 {
			priceStr = "N/A"
		}
		monthlyStr := fmt.Sprintf("€%.2f", p.MonthlyCost)
		if p.MonthlyCost == 0 {
			monthlyStr = "N/A"
		}

		tbl.AddRow(
			p.OpiFlavour,
			p.ProductIdParameter,
			p.vCPU,
			p.RAM,
			priceStr,
			monthlyStr,
		)
	}

	// Print table
	fmt.Printf("\n")
	color.Cyan("Server Pricing (Region: %s, OS: %s)", cfg.Region, osType)
	tbl.Print()
	fmt.Printf("Total: %d servers | Sorted by vCPUs → RAM → Cost\n", len(pricingList))
	fmt.Printf("Pricing based on hourly rates (730 hours/month)\n\n")
}
