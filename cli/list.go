package cli

import (
	"strings"

	"github.com/abdo-farag/otc-cli/internal/commands"
	"github.com/abdo-farag/otc-cli/internal/commands/resource"
	"github.com/abdo-farag/otc-cli/internal/config"
	"github.com/abdo-farag/otc-cli/internal/otc"

	"github.com/spf13/cobra"
)

// Main list command - acts as parent
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List resources",
	Long:  `List OTC resources such as servers, VPCs, volumes, and more.`,
	Example: `  # List all available resource types
  otc-cli list

  # List specific resources
  otc-cli list ecs --az eu-de-01
  otc-cli list images --visibility private
  otc-cli list projects
  
  # List service pricing
  otc-cli list pricing --services
  otc-cli list pricing ecs
  otc-cli list pricing obs`,
	SuggestionsMinimumDistance: 2,
}

// Subcommands for each resource type
var listVpcCmd = &cobra.Command{
	Use:     "vpc",
	Aliases: []string{"vpcs"},
	Short:   "List Virtual Private Clouds",
	Args:    cobra.NoArgs,
	RunE:    runListVpc,
}

var listSubnetCmd = &cobra.Command{
	Use:     "subnet",
	Aliases: []string{"subnets"},
	Short:   "List VPC subnets",
	Args:    cobra.NoArgs,
	RunE:    runListSubnet,
}

var listVolumeCmd = &cobra.Command{
	Use:     "volume",
	Aliases: []string{"volumes"},
	Short:   "List volumes",
	Args:    cobra.NoArgs,
	RunE:    runListVolume,
}

var listCceCmd = &cobra.Command{
	Use:     "cce",
	Aliases: []string{"clusters", "cluster"},
	Short:   "List Kubernetes clusters",
	Args:    cobra.NoArgs,
	RunE:    runListCce,
}

var listFlavorCmd = &cobra.Command{
	Use:     "flavor",
	Aliases: []string{"flavors"},
	Short:   "List server flavors with pricing (consolidated view)",
	Long: `List server flavors with pricing across all compute services.

This command provides a consolidated view of:
- ECS (Elastic Cloud Server)
- Compute-optimized (ecsnoc)
- Memory-optimized (memo)
- GPU instances (gpu)
- High-performance (hps)
- Dedicated hosts (deh)

For service-specific pricing with all variants, use:
  otc-cli list pricing <service-name>`,
	Args: cobra.NoArgs,
	Example: `  otc-cli list flavor
  otc-cli list flavor --os windows
  otc-cli list flavor --os redhat
  
  # For detailed ECS-only pricing
  otc-cli list pricing ecs`,
	RunE: runListFlavor,
}

var listImageCmd = &cobra.Command{
	Use:     "image",
	Aliases: []string{"images"},
	Short:   "List system and custom images",
	Args:    cobra.NoArgs,
	Example: `  otc-cli list image
  otc-cli list image --visibility private
  otc-cli list image --platform ubuntu --name 22.04`,
	RunE: runListImage,
}

var listProjectsCmd = &cobra.Command{
	Use:     "projects",
	Aliases: []string{"project", "p"},
	Short:   "List OTC projects",
	Args:    cobra.NoArgs,
	RunE:    runListProjects,
}

var listEcsCmd = &cobra.Command{
	Use:     "ecs",
	Aliases: []string{"servers", "server"},
	Short:   "List Elastic Cloud Servers",
	Args:    cobra.NoArgs,
	Example: `  otc-cli list ecs
  otc-cli list ecs --az eu-de-01
  otc-cli list ecs --status ACTIVE --tag Environment=production`,
	RunE: runListEcs,
}

var listKeypairCmd = &cobra.Command{
	Use:     "keypair",
	Aliases: []string{"keypairs", "key", "ssh", "ssh-key"},
	Short:   "List SSH keypairs",
	Args:    cobra.NoArgs,
	Example: `  otc-cli list keypair
  otc-cli list keypair --raw`,
	RunE: runListKeypair,
}

var listPricingCmd = &cobra.Command{
	Use:     "pricing [service-name]",
	Aliases: []string{"price", "cost", "costs"},
	Short:   "List pricing for OTC services (detailed view)",
	Long: `List detailed pricing for a specific OTC service with all variants.

This shows ALL pricing records including different:
- OS types (Linux, Windows, RedHat, etc.)
- Billing modes (hourly, monthly)
- Regions and availability zones
- License types

For a consolidated compute pricing view, use:
  otc-cli list flavors`,
	Example: `  # List all available services
  otc-cli list pricing --services
  otc-cli list pricing --discover

  # Service-specific pricing
  otc-cli list pricing ecs
  otc-cli list pricing gpu --csv
  otc-cli list pricing rds
  
  # Filter by OS type
  otc-cli list pricing ecs --os linux
  otc-cli list pricing gpu --os windows --csv
  otc-cli list pricing ecs -o redhat
  
  # CSV output for Excel/spreadsheets
  otc-cli list pricing ecs --os linux --csv > ecs-linux-pricing.csv
  otc-cli list pricing obs --csv > obs-pricing.csv
  
  # With filters
  otc-cli list pricing obs --filter productIdParameter=Standard
  otc-cli list pricing ecs --os linux --filter billingMode=hourly --csv`,
	Args: cobra.MaximumNArgs(1),
	RunE: runListPricing,
}

// ECS-specific flags
var (
	ecsAZ     string
	ecsStatus string
	ecsName   string
	ecsTag    string
)

// Image-specific flags
var (
	imageVisibility string
	imagePlatform   string
	imageName       string
	imageStatus     string
)

// Flavor-specific flags
var (
	flavorOS string
)

// Pricing-specific flags
var (
	pricingServices bool
	pricingDiscover bool
	pricingFilter   string
	pricingOS       string
)

func init() {
	// Add subcommands
	listCmd.AddCommand(listProjectsCmd)
	listCmd.AddCommand(listEcsCmd)
	listCmd.AddCommand(listVpcCmd)
	listCmd.AddCommand(listSubnetCmd)
	listCmd.AddCommand(listVolumeCmd)
	listCmd.AddCommand(listCceCmd)
	listCmd.AddCommand(listFlavorCmd)
	listCmd.AddCommand(listImageCmd)
	listCmd.AddCommand(listKeypairCmd)
	listCmd.AddCommand(listPricingCmd)

	// ECS flags
	listEcsCmd.Flags().StringVar(&ecsAZ, "az", "", "Filter by availability zone (e.g., eu-de-01)")
	listEcsCmd.Flags().StringVar(&ecsStatus, "status", "", "Filter by status (ACTIVE, SHUTOFF, etc.)")
	listEcsCmd.Flags().StringVar(&ecsName, "name", "", "Filter by server name (partial match)")
	listEcsCmd.Flags().StringVar(&ecsTag, "tag", "", "Filter by tag (key=value)")

	// Image flags
	listImageCmd.Flags().StringVar(&imageVisibility, "visibility", "", "Filter by visibility (private, public, shared)")
	listImageCmd.Flags().StringVar(&imagePlatform, "platform", "", "Filter by platform (Ubuntu, CentOS, Windows, etc.)")
	listImageCmd.Flags().StringVar(&imageName, "name", "", "Filter by image name (partial match)")
	listImageCmd.Flags().StringVar(&imageStatus, "status", "", "Filter by status (active, queued, etc.)")

	// Flavor flags
	listFlavorCmd.Flags().StringVarP(&flavorOS, "os", "o", "openlinux", "OS type for pricing (openlinux, redhat, oracle, windows)")

	// Pricing flags
	listPricingCmd.Flags().BoolVar(&pricingServices, "services", false, "List all available services with descriptions")
	listPricingCmd.Flags().BoolVar(&pricingDiscover, "discover", false, "Discover services with actual pricing data in the region")
	listPricingCmd.Flags().StringVar(&pricingFilter, "filter", "", "Filter results (format: key=value, e.g., productIdParameter=Standard)")
	listPricingCmd.Flags().StringVarP(&pricingOS, "os", "o", "", "Filter by OS type (linux, windows, redhat, suse, oracle, ubuntu)")
}

// RunE functions for each resource
func runListProjects(cmd *cobra.Command, args []string) error {
	return runListResource("projects", map[string]interface{}{})
}

func runListEcs(cmd *cobra.Command, args []string) error {
	options := map[string]interface{}{
		"az":     ecsAZ,
		"status": ecsStatus,
		"name":   ecsName,
		"tag":    ecsTag,
	}
	return runListResource("ecs", options)
}

func runListVpc(cmd *cobra.Command, args []string) error {
	return runListResource("vpc", map[string]interface{}{})
}

func runListSubnet(cmd *cobra.Command, args []string) error {
	return runListResource("subnet", map[string]interface{}{})
}

func runListVolume(cmd *cobra.Command, args []string) error {
	return runListResource("volume", map[string]interface{}{})
}

func runListCce(cmd *cobra.Command, args []string) error {
	return runListResource("cce", map[string]interface{}{})
}

func runListFlavor(cmd *cobra.Command, args []string) error {
	options := map[string]interface{}{
		"os": flavorOS,
	}
	return runListResource("flavor", options)
}

func runListImage(cmd *cobra.Command, args []string) error {
	options := map[string]interface{}{
		"visibility": imageVisibility,
		"platform":   imagePlatform,
		"name":       imageName,
		"status":     imageStatus,
	}
	return runListResource("image", options)
}

func runListKeypair(cmd *cobra.Command, args []string) error {
	return runListResource("keypair", map[string]interface{}{})
}

func runListPricing(cmd *cobra.Command, args []string) error {
	cfg := config.New()

	// If --discover flag is set, fetch actual available services
	if pricingDiscover {
		return resource.DiscoverAvailableServices(cfg.Region)
	}

	// If --services flag is set, list documented services
	if pricingServices {
		resource.ListAvailableServices(cfg.Region)
		return nil
	}

	// Service name is required if not listing services
	if len(args) == 0 {
		resource.ListAvailableServices(cfg.Region)
		return nil
	}

	serviceName := args[0]

	// Parse filters
	filters := make(map[string]string)
	if pricingFilter != "" {
		parts := strings.SplitN(pricingFilter, "=", 2)
		if len(parts) == 2 {
			filters[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}

	// Add OS filter if specified
	osFilter := pricingOS

	// Pricing API is public - no authentication required
	otcClient := otc.NewClient(cfg)
	resource.ListServicePricing(cfg, otcClient, "", "", serviceName, rawFlag, csvFlag, filters, osFilter)

	return nil
}

// Common list logic
func runListResource(resourceType string, options map[string]interface{}) error {
	cfg := config.New()

	// Authenticate
	tokenCache, err := ensureAuthenticated(cfg)
	if err != nil {
		return err
	}

	// Resolve project
	selectedProjectID := projectFlag
	if selectedProjectID != "" && resourceType != "projects" {
		selectedProjectID = resolveProject(cfg, tokenCache.UnscopedToken, selectedProjectID)
	}

	// Execute list command
	otcClient := otc.NewClient(cfg)
	return commands.ListCommand(cfg, otcClient, tokenCache.UnscopedToken, resourceType, selectedProjectID, options, rawFlag)
}
