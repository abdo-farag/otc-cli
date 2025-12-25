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
  otc-cli list images --visibility private
  otc-cli list projects`,
	SuggestionsMinimumDistance: 2,
}

var listFlavorCmd = &cobra.Command{
	Use:   "flavor",
	Aliases: []string{"flavors"},
	Short: "List server flavors with pricing",
	Args:    cobra.NoArgs, // Add this
	Example: `  otc-cli list flavor
  otc-cli list flavor --os windows`,
	RunE: runListFlavor,
}

var listImageCmd = &cobra.Command{
	Use:     "image",
	Aliases: []string{"images"},
	Short:   "List system and custom images",
	Args:    cobra.NoArgs, // Add this
	Example: `  otc-cli list image
  otc-cli list image --visibility private
  otc-cli list image --platform ubuntu --name 22.04`,
	RunE: runListImage,
}

var listProjectsCmd = &cobra.Command{
	Use:     "projects",
	Aliases: []string{"project"},
	Short:   "List OTC projects",
	Args:    cobra.NoArgs, // Add this
	RunE:    runListProjects,
}

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
	listCmd.AddCommand(listFlavorCmd)
	listCmd.AddCommand(listImageCmd)

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
