package cli

import (
  "fmt"

  "github.com/spf13/cobra"
)

var (
  // Version information - set via ldflags during build
  version   = "dev"
  commit    = "none"
  buildDate = "unknown"
)

// Global flags
var (
  projectFlag string
  rawFlag     bool
  csvFlag     bool
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
  rootCmd.PersistentFlags().BoolVar(&csvFlag, "csv", false, "Output in CSV format")

  rootCmd.MarkFlagsMutuallyExclusive("raw", "csv")
  rootCmd.MarkFlagsMutuallyExclusive("json", "csv")
  
  // Add subcommands
  rootCmd.AddCommand(loginCmd)
  rootCmd.AddCommand(logoutCmd)
  rootCmd.AddCommand(docsCmd)
  rootCmd.AddCommand(listCmd)
  rootCmd.AddCommand(getCmd)
  rootCmd.AddCommand(consoleCmd)
  rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
  Use:   "version",
  Short: "Print the version number",
  Long:  "Display version information including build details",
  Run: func(cmd *cobra.Command, args []string) {
    fmt.Printf("otc-cli version %s\n", version)
    if commit != "none" {
      fmt.Printf("  commit: %s\n", commit)
    }
    if buildDate != "unknown" {
      fmt.Printf("  built:  %s\n", buildDate)
    }
  },
}
