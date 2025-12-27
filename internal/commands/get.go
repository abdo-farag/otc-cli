package commands

import (
	"fmt"
	"github.com/abdo-farag/otc-cli/internal/commands/resource"
	"github.com/abdo-farag/otc-cli/internal/config"
	"github.com/abdo-farag/otc-cli/internal/otc"
)

// GetCommand handles all get operations
func GetCommand(cfg *config.Config, client *otc.Client, unscopedToken, resourceType, resourceID, projectID string, options map[string]interface{}, raw bool) error {
	switch resourceType {
	case "ecs", "server", "instance", "servers", "instances":
		resource.GetECS(cfg, client, unscopedToken, projectID, resourceID, raw)
	case "vpc", "vpcs":
		resource.GetVPC(cfg, client, unscopedToken, projectID, resourceID, raw)
	case "subnet", "subnets":
		resource.GetSubnet(cfg, client, unscopedToken, projectID, resourceID, raw)
	case "volume", "volumes":
		resource.GetVolume(cfg, client, unscopedToken, projectID, resourceID, raw)
	case "cce", "cluster", "clusters":
		resource.GetCCE(cfg, client, unscopedToken, projectID, resourceID, raw)
 case "kubeconfig":
    // Get output path from options
    outputPath := "~/.kube"
    if path, ok := options["output"].(string); ok && path != "" {
      outputPath = path
    }
    resource.GetKubeconfig(cfg, client, unscopedToken, projectID, resourceID, outputPath)
	default:
		return fmt.Errorf("unknown resource type: %s", resourceType)
	}
	return nil
}
