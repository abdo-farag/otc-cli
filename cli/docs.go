package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var docsCmd = &cobra.Command{
	Use:   "docs",
	Short: "Generate CLI documentation",
	Long:  `Generate comprehensive markdown documentation for all CLI commands.`,
	Example: `  # Generate documentation
  otc-cli docs

  # Generates:
  # - otc-cli.md (main docs in root)
  # - docs/*.md (subcommand docs)`,
	RunE: runDocs,
}

func runDocs(cmd *cobra.Command, args []string) error {
	// Create docs subdirectory if it doesn't exist
	docsSubDir := "./docs"
	if err := os.MkdirAll(docsSubDir, 0755); err != nil {
		return fmt.Errorf("failed to create docs directory: %w", err)
	}

	// Custom link handler to fix relative paths
	linkHandler := func(name string) string {
		base := strings.TrimSuffix(name, filepath.Ext(name))
		// Main doc stays in root, subcommands go to docs/
		if base == "otc-cli" {
			return name
		}
		return "docs/" + name
	}

	// Generate documentation with custom link handler
	if err := doc.GenMarkdownTreeCustom(rootCmd, ".", filePrepender, linkHandler); err != nil {
		return fmt.Errorf("failed to generate docs: %w", err)
	}

	// Move subcommand docs to docs/ directory, keep root doc in root
	entries, err := os.ReadDir(".")
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	var mainDocCount, subDocCount int
	for _, entry := range entries {
		// Skip directories and non-markdown files
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" {
			continue
		}

		// If it's the main otc-cli.md, keep it in root
		if entry.Name() == "otc-cli.md" {
			mainDocCount++
			continue
		}

		// If it's a subcommand doc (starts with otc-cli_), move to docs/
		if strings.HasPrefix(entry.Name(), "otc-cli_") {
			oldPath := entry.Name()
			newPath := filepath.Join(docsSubDir, entry.Name())

			if err := os.Rename(oldPath, newPath); err != nil {
				color.Yellow("Warning: could not move %s: %v", entry.Name(), err)
			} else {
				subDocCount++
			}
		}
	}

	color.Green("\n✓ Documentation generated successfully!\n")
	color.Cyan("Main documentation:")
	fmt.Printf("  - %s (root directory)\n\n", "otc-cli.md")

	color.Cyan("Subcommand documentation (%d files in docs/):", subDocCount)

	// List files in docs/
	docEntries, _ := os.ReadDir(docsSubDir)
	for _, entry := range docEntries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".md" {
			fmt.Printf("  - docs/%s\n", entry.Name())
		}
	}

	fmt.Println()
	return nil
}

// filePrepender adds a header to generated files
func filePrepender(filename string) string {
	return ""
}
