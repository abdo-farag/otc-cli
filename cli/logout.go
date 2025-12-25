package cli

import (
	"github.com/abdo-farag/otc-cli/internal/cache"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Clear cached authentication token",
	Long:  "Remove the cached authentication token from local storage.",
	RunE:  runLogout,
}

func runLogout(cmd *cobra.Command, args []string) error {
	if err := cache.ClearToken(); err != nil {
		return err
	}
	color.Green("✓ Logged out successfully")
	return nil
}