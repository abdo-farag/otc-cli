package cli

import (
  "github.com/abdo-farag/otc-cli/internal/commands"
  "github.com/abdo-farag/otc-cli/internal/config"
  "github.com/abdo-farag/otc-cli/internal/otc"

  "github.com/spf13/cobra"
)

var (
  getOutputPath string
)

var getCmd = &cobra.Command{
  Use:   "get",
  Short: "Get a specific resource",
  Long:  `Get detailed information about a specific OTC resource.`,
  Example: `  # Get ECS instance details
  otc-cli get ecs my-server

  # Get kubeconfig for cluster
  otc-cli get kubeconfig my-cluster --output ~/.kube/config`,
}

var getEcsCmd = &cobra.Command{
  Use:     "ecs [server-id-or-name]",
  Aliases: []string{"server"},
  Short:   "Get ECS instance details",
  Args:    cobra.ExactArgs(1),
  Example: `  otc-cli get ecs abc123
  otc-cli get ecs my-server --project Production
  otc-cli get ecs abc123 --json`,
  RunE: runGetEcs,
}

var getVpcCmd = &cobra.Command{
  Use:     "vpc [vpc-id-or-name]",
  Aliases: []string{"vpcs"},
  Short:   "Get VPC details",
  Args:    cobra.ExactArgs(1),
  Example: `  otc-cli get vpc vpc-xyz
  otc-cli get vpc my-vpc --project Production`,
  RunE: runGetVpc,
}

var getSubnetCmd = &cobra.Command{
  Use:     "subnet [subnet-id-or-name]",
  Aliases: []string{"subnets"},
  Short:   "Get subnet details",
  Args:    cobra.ExactArgs(1),
  Example: `  otc-cli get subnet subnet-123`,
  RunE: runGetSubnet,
}

var getVolumeCmd = &cobra.Command{
  Use:     "volume [volume-id-or-name]",
  Aliases: []string{"volumes"},
  Short:   "Get volume details",
  Args:    cobra.ExactArgs(1),
  Example: `  otc-cli get volume vol-123`,
  RunE: runGetVolume,
}

var getCceCmd = &cobra.Command{
  Use:     "cce [cluster-id-or-name]",
  Aliases: []string{"cluster"},
  Short:   "Get CCE cluster details",
  Args:    cobra.ExactArgs(1),
  Example: `  otc-cli get cce my-cluster`,
  RunE: runGetCce,
}

var getKubeconfigCmd = &cobra.Command{
  Use:     "kubeconfig [cluster-id-or-name]",
  Aliases: []string{"kube"},
  Short:   "Download kubeconfig for CCE cluster",
  Args:    cobra.ExactArgs(1),
  Example: `  # Get kubeconfig by cluster name
  otc-cli get kubeconfig my-cluster

  # Get kubeconfig by cluster ID
  otc-cli get kubeconfig c8198b6d-7633-4afc-9ec5-ab97bcd94ab8

  # Save to custom path
  otc-cli get kubeconfig my-cluster --output ~/.kube/otc-config`,
  RunE: runGetKubeconfig,
}

func init() {
  // Add subcommands
  getCmd.AddCommand(getEcsCmd)
  getCmd.AddCommand(getVpcCmd)
  getCmd.AddCommand(getSubnetCmd)
  getCmd.AddCommand(getVolumeCmd)
  getCmd.AddCommand(getCceCmd)
  getCmd.AddCommand(getKubeconfigCmd)

  // Kubeconfig-specific flags
  getKubeconfigCmd.Flags().StringVarP(&getOutputPath, "output", "o", "./kubeconfig", "Output path for kubeconfig file")
}

func runGetEcs(cmd *cobra.Command, args []string) error {
  return runGetResource("ecs", args[0], map[string]interface{}{})
}

func runGetVpc(cmd *cobra.Command, args []string) error {
  return runGetResource("vpc", args[0], map[string]interface{}{})
}

func runGetSubnet(cmd *cobra.Command, args []string) error {
  return runGetResource("subnet", args[0], map[string]interface{}{})
}

func runGetVolume(cmd *cobra.Command, args []string) error {
  return runGetResource("volume", args[0], map[string]interface{}{})
}

func runGetCce(cmd *cobra.Command, args []string) error {
  return runGetResource("cce", args[0], map[string]interface{}{})
}

func runGetKubeconfig(cmd *cobra.Command, args []string) error {
  options := map[string]interface{}{
    "output": getOutputPath,
  }
  return runGetResource("kubeconfig", args[0], options)
}

func runGetResource(resourceType, resourceID string, options map[string]interface{}) error {
  cfg := config.New()

  // Authenticate
  tokenCache, err := ensureAuthenticated(cfg)
  if err != nil {
    return err
  }

  // Resolve project
  selectedProjectID := projectFlag
  if selectedProjectID != "" {
    selectedProjectID = resolveProject(cfg, tokenCache.UnscopedToken, selectedProjectID)
  }

  // Execute get command
  otcClient := otc.NewClient(cfg)
  return commands.GetCommand(cfg, otcClient, tokenCache.UnscopedToken, resourceType, resourceID, selectedProjectID, options, rawFlag)
}
