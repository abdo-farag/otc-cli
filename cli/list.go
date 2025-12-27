package cli

import (
  "github.com/abdo-farag/otc-cli/internal/commands"
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
  otc-cli list projects`,
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
  Use:   "flavor",
  Aliases: []string{"flavors"},
  Short: "List server flavors with pricing",
  Args:    cobra.NoArgs,
  Example: `  otc-cli list flavor
  otc-cli list flavor --os windows`,
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
