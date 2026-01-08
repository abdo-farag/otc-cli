package config

import (
	"fmt"
)

// ServiceEndpoints defines all service endpoints for different regions
type ServiceEndpoints struct {
	// Identity and Access Management
	IAM string

	// Compute Services
	ECS string // Elastic Cloud Server
	IMS string // Image Management Service
	AS  string // Auto Scaling
	FGS string // FunctionGraph

	// Storage Services
	EVS  string // Elastic Volume Service
	OBS  string // Object Storage Service
	SFS  string // Scalable File Service
	CBR  string // Cloud Backup and Recovery
	SDRS string // Storage Disaster Recovery Service

	// Network Services
	VPC   string // Virtual Private Cloud
	EIP   string // Elastic IP
	NAT   string // NAT Gateway
	VPN   string // VPN
	ELB   string // Elastic Load Balancer
	DNS   string // Domain Name Service
	VPCEP string // VPC Endpoint

	// Container Services
	CCE string // Cloud Container Engine
	SWR string // Software Repository for Container

	// Database Services
	RDS     string // Relational Database Service
	DDS     string // Document Database Service
	GaussDB string // GaussDB Database
	DCS     string // Distributed Cache Service
	DDM     string // Distributed Database Middleware

	// Application Services
	APIG          string // API Gateway
	DMS           string // Distributed Message Service
	SMN           string // Simple Message Notification
	DIS           string // Data Ingestion Service
	FunctionGraph string

	// Security Services
	WAF      string // Web Application Firewall
	AntiDDoS string // Anti-DDoS
	CFW      string // Cloud Firewall

	// Management & Deployment
	CTS string // Cloud Trace Service
	LTS string // Log Tank Service
	AOM string // Application Operations Management
	CES string // Cloud Eye Service
	TMS string // Tag Management Service
	RMS string // Resource Management Service

	// Big Data & AI
	DWS       string // Data Warehouse Service
	MRS       string // MapReduce Service
	DLI       string // Data Lake Insight
	ModelArts string

	// DevOps
	CodeArts     string
	ServiceStage string

	// Legacy/Alias
	Compute string // Alias for ECS
}

// RegionConfig holds region-specific configuration
type RegionConfig struct {
	Name      string
	Domain    string
	IsSwiss   bool
	IAMPrefix string // Special IAM prefix for some regions
}

var regionConfigs = map[string]RegionConfig{
	"eu-de": {
		Name:      "eu-de",
		Domain:    "otc.t-systems.com",
		IsSwiss:   false,
		IAMPrefix: "", // IAM for eu-de/eu-nl doesn't use region prefix
	},
	"eu-nl": {
		Name:      "eu-nl",
		Domain:    "otc.t-systems.com",
		IsSwiss:   false,
		IAMPrefix: "", // IAM for eu-de/eu-nl doesn't use region prefix
	},
	"eu-ch2": {
		Name:      "eu-ch2",
		Domain:    "sc.otc.t-systems.com",
		IsSwiss:   true,
		IAMPrefix: "iam-pub.eu-ch2", // Special IAM endpoint for Swiss region
	},
}

// GetEndpoints returns service endpoints for a given region
func GetEndpoints(region string) (*ServiceEndpoints, error) {
	regionCfg, exists := regionConfigs[region]
	if !exists {
		return nil, fmt.Errorf("unsupported region: %s (supported: eu-de, eu-nl, eu-ch2)", region)
	}

	// IAM endpoint - special handling
	var iamEndpoint string
	if regionCfg.IAMPrefix != "" {
		// Swiss region with custom IAM prefix
		iamEndpoint = fmt.Sprintf("https://%s.%s", regionCfg.IAMPrefix, regionCfg.Domain)
	} else {
		// eu-de and eu-nl use global IAM endpoint
		iamEndpoint = fmt.Sprintf("https://iam.%s.%s", region, regionCfg.Domain)
	}

	// OBS includes region in subdomain
	obsEndpoint := buildEndpoint("obs", region, regionCfg.Domain)

	return &ServiceEndpoints{
		// Identity
		IAM: iamEndpoint,

		// Compute
		ECS: buildEndpoint("ecs", region, regionCfg.Domain),
		IMS: buildEndpoint("ims", region, regionCfg.Domain),
		AS:  buildEndpoint("as", region, regionCfg.Domain),
		FGS: buildEndpoint("functiongraph", region, regionCfg.Domain),

		// Storage
		EVS:  buildEndpoint("evs", region, regionCfg.Domain),
		OBS:  obsEndpoint,
		SFS:  buildEndpoint("sfs", region, regionCfg.Domain),
		CBR:  buildEndpoint("cbr", region, regionCfg.Domain),
		SDRS: buildEndpoint("sdrs", region, regionCfg.Domain),

		// Network
		VPC:   buildEndpoint("vpc", region, regionCfg.Domain),
		EIP:   buildEndpoint("vpc", region, regionCfg.Domain), // EIP uses VPC endpoint
		NAT:   buildEndpoint("nat", region, regionCfg.Domain),
		VPN:   buildEndpoint("vpn", region, regionCfg.Domain),
		ELB:   buildEndpoint("elb", region, regionCfg.Domain),
		DNS:   buildEndpoint("dns", region, regionCfg.Domain),
		VPCEP: buildEndpoint("vpcep", region, regionCfg.Domain),

		// Container
		CCE: buildEndpoint("cce", region, regionCfg.Domain),
		SWR: buildEndpoint("swr", region, regionCfg.Domain),

		// Database
		RDS:     buildEndpoint("rds", region, regionCfg.Domain),
		DDS:     buildEndpoint("dds", region, regionCfg.Domain),
		GaussDB: buildEndpoint("gaussdb", region, regionCfg.Domain),
		DCS:     buildEndpoint("dcs", region, regionCfg.Domain),
		DDM:     buildEndpoint("ddm", region, regionCfg.Domain),

		// Application
		APIG:          buildEndpoint("apig", region, regionCfg.Domain),
		DMS:           buildEndpoint("dms", region, regionCfg.Domain),
		SMN:           buildEndpoint("smn", region, regionCfg.Domain),
		DIS:           buildEndpoint("dis", region, regionCfg.Domain),
		FunctionGraph: buildEndpoint("functiongraph", region, regionCfg.Domain),

		// Security
		WAF:      buildEndpoint("waf", region, regionCfg.Domain),
		AntiDDoS: buildEndpoint("antiddos", region, regionCfg.Domain),
		CFW:      buildEndpoint("cfw", region, regionCfg.Domain),

		// Management
		CTS: buildEndpoint("cts", region, regionCfg.Domain),
		LTS: buildEndpoint("lts", region, regionCfg.Domain),
		AOM: buildEndpoint("aom", region, regionCfg.Domain),
		CES: buildEndpoint("ces", region, regionCfg.Domain),
		TMS: buildEndpoint("tms", region, regionCfg.Domain),
		RMS: buildEndpoint("rms", region, regionCfg.Domain),

		// Big Data & AI
		DWS:       buildEndpoint("dws", region, regionCfg.Domain),
		MRS:       buildEndpoint("mrs", region, regionCfg.Domain),
		DLI:       buildEndpoint("dli", region, regionCfg.Domain),
		ModelArts: buildEndpoint("modelarts", region, regionCfg.Domain),

		// DevOps
		CodeArts:     buildEndpoint("codearts", region, regionCfg.Domain),
		ServiceStage: buildEndpoint("servicestage", region, regionCfg.Domain),

		// Aliases
		Compute: buildEndpoint("ecs", region, regionCfg.Domain),
	}, nil
}

// buildEndpoint constructs service endpoint URL
func buildEndpoint(service, region, domain string) string {
	return fmt.Sprintf("https://%s.%s.%s", service, region, domain)
}

// GetRegionConfig returns configuration for a specific region
func GetRegionConfig(region string) (*RegionConfig, error) {
	cfg, exists := regionConfigs[region]
	if !exists {
		return nil, fmt.Errorf("unsupported region: %s", region)
	}
	return &cfg, nil
}

// SupportedRegions returns list of supported regions
func SupportedRegions() []string {
	return []string{"eu-de", "eu-nl", "eu-ch2"}
}

// IsValidRegion checks if a region is supported
func IsValidRegion(region string) bool {
	_, exists := regionConfigs[region]
	return exists
}

// GetRegionDomain returns the domain for a region
func GetRegionDomain(region string) string {
	if cfg, exists := regionConfigs[region]; exists {
		return cfg.Domain
	}
	return "otc.t-systems.com" // default
}
