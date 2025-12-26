package commands

import (
	"fmt"
	"github.com/abdo-farag/otc-cli/internal/commands/resource"
	"github.com/abdo-farag/otc-cli/internal/config"
	"github.com/abdo-farag/otc-cli/internal/otc"
)

// ListCommand handles all list operations
func ListCommand(cfg *config.Config, client *otc.Client, unscopedToken, resourceType, projectID string, options map[string]interface{}, raw bool) error {
	osType, _ := options["os"].(string)
	if osType == "" {
		osType = "openlinux"
	}
	switch resourceType {
	case "projects":
		listProjects(cfg, client, unscopedToken, raw)
	case "ecs", "server", "instance", "servers", "instances":
		resource.ListECS(cfg, client, unscopedToken, projectID, options, raw)
	case "vpc", "vpcs":
		resource.ListVPC(cfg, client, unscopedToken, projectID, raw)
	case "subnet", "subnets":
		resource.ListSubnet(cfg, client, unscopedToken, projectID, raw)
	case "volume", "volumes":
		resource.ListVolume(cfg, client, unscopedToken, projectID, raw)
	case "cce", "cluster", "clusters":
		resource.ListCCE(cfg, client, unscopedToken, projectID, raw)
	case "image", "images":
		resource.ListImages(cfg, client, unscopedToken, projectID, options, raw)
	case "keypair", "keypairs":
		resource.ListKeypairs(cfg, client, unscopedToken, projectID, raw)
	case "flavor", "flavors":
		resource.ListFlavors(cfg, client, unscopedToken, projectID, raw, osType)
	default:
		return fmt.Errorf("unknown resource type: %s", resourceType)
	}
	return nil
}
