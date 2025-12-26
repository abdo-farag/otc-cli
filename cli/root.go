package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	version = "1.0.0"
)

// Global flags
var (
	projectFlag string
	rawFlag     bool
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "otc-cli",
	Short: "OTC CLI - Manage your Open Telekom Cloud resources",
	Long: `A command-line tool for managing Open Telekom Cloud (OTC) resources.
Supports authentication, resource management, and automation.`,
	Version: version,
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Persistent flags available to all subcommands
	rootCmd.PersistentFlags().StringVarP(&projectFlag, "project", "p", "", "Project ID or name")
	rootCmd.PersistentFlags().BoolVar(&rawFlag, "raw", false, "Output raw JSON response")
	rootCmd.PersistentFlags().BoolVar(&rawFlag, "json", false, "Output raw JSON response (alias)")

	// Add subcommands
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(logoutCmd)
	rootCmd.AddCommand(docsCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("otc-cli v%s\n", version)
	},
}
