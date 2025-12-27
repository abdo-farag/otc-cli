package resource

import (
	"encoding/json"
	"fmt"
	"github.com/abdo-farag/otc-cli/internal/config"
	"github.com/abdo-farag/otc-cli/internal/otc"

	"github.com/fatih/color"
	"github.com/rodaine/table"
)

// listProjects lists all OTC projects
func ListProjects(cfg *config.Config, client *otc.Client, unscopedToken string, raw bool) {
	domainToken, err := client.GetDomainScopedToken(unscopedToken)
	if err != nil {
		color.Red("✗ Failed to get domain token: %v", err)
		return
	}

	projects, err := client.ListProjects(domainToken)
	if err != nil {
		color.Red("✗ Failed to list projects: %v", err)
		return
	}

	if raw {
		// Output as JSON
		jsonData := map[string]interface{}{
			"projects": projects,
		}
		formatted, _ := json.MarshalIndent(jsonData, "", "  ")
		fmt.Println(string(formatted))
		return
	}

	// Formatted table output
	headerFmt := color.New(color.FgCyan, color.Bold).SprintfFunc()
	tbl := table.New("Name", "ID")
	tbl.WithHeaderFormatter(headerFmt)

	for _, p := range projects {
		tbl.AddRow(p.Name, p.ID)
	}

	fmt.Printf("\n")
	color.Cyan("OTC Projects")
	tbl.Print()
	fmt.Printf("\nTotal: %d projects\n", len(projects))
}
