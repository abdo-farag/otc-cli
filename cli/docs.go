package cli

import (
	"bytes"
	"os"

	"github.com/fatih/color"
	"github.com/gavv/cobradoc"
	"github.com/spf13/cobra"
)

var docsCmd = &cobra.Command{
	Use:   "docs",
	Short: "Generate CLI documentation",
	Long:  `Generate comprehensive markdown documentation for all CLI commands.`,
	Example: `  # Generate documentation in current directory
  otc-cli docs

  # Generate with custom output file
  otc-cli docs --output ./documentation/otc-cli.md`,
	RunE:   runDocs,
}

var docsOutput string

func init() {
	docsCmd.Flags().StringVarP(&docsOutput, "output", "o", "otc-cli.md", "Output file path")
}

func runDocs(cmd *cobra.Command, args []string) error {
	buf := bytes.NewBufferString("")
	
	// Generate markdown documentation
	if err := cobradoc.WriteDocument(buf, rootCmd, cobradoc.Markdown, cobradoc.Options{}); err != nil {
		return err
	}

	// Write to file
	if err := os.WriteFile(docsOutput, buf.Bytes(), 0644); err != nil {
		return err
	}

	color.Green("✓ Documentation generated: %s", docsOutput)
	return nil
}