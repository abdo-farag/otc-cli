package cli

import (
	"github.com/abdo-farag/otc-cli/internal/commands/resource"
	"github.com/abdo-farag/otc-cli/internal/config"
	"github.com/abdo-farag/otc-cli/internal/otc"

	"github.com/spf13/cobra"
)

var (
	consoleNoBrowser bool
)

var consoleCmd = &cobra.Command{
	Use:   "console",
	Short: "Get remote console access to resources",
	Long:  `Get VNC console URLs for accessing ECS instances remotely and open in browser.`,
}

var consoleEcsCmd = &cobra.Command{
	Use:     "ecs [server-name-or-id]",
	Aliases: []string{"server"},
	Short:   "Open VNC console for an ECS instance",
	Args:    cobra.ExactArgs(1),
	Example: `  # Open VNC console by server name
  otc-cli console ecs my-server

  # Open VNC console by server ID
  otc-cli console ecs 0000000-11111-22222-abcd...

  # Get URL without opening browser
  otc-cli console ecs my-server --no-browser

  # With specific project
  otc-cli console ecs my-server -p eu-de_PROJECT`,
	RunE: runConsoleEcs,
}

func init() {
	consoleCmd.AddCommand(consoleEcsCmd)

	// Console flags
	consoleEcsCmd.Flags().BoolVar(&consoleNoBrowser, "no-browser", false, "Don't open browser automatically")
}

func runConsoleEcs(cmd *cobra.Command, args []string) error {
	cfg := config.New()

	tokenCache, err := ensureAuthenticated(cfg)
	if err != nil {
		return err
	}

	selectedProjectID := projectFlag
	if selectedProjectID != "" {
		selectedProjectID = resolveProject(cfg, tokenCache.UnscopedToken, selectedProjectID)
	}

	otcClient := otc.NewClient(cfg)
	resource.GetConsoleECS(cfg, otcClient, tokenCache.UnscopedToken, selectedProjectID, args[0], consoleNoBrowser)

	return nil
}
