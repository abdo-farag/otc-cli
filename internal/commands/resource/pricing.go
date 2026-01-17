package resource

import (
  "encoding/csv"
  "encoding/json"
  "fmt"
  "io"
  "net/http"
  "os"
  "sort"
  "strings"
  "time"

  "github.com/abdo-farag/otc-cli/internal/config"
  "github.com/abdo-farag/otc-cli/internal/otc"
  "github.com/fatih/color"
  "github.com/rodaine/table"
)

// ServiceInfo holds service metadata
type ServiceInfo struct {
	Code        string
	Name        string
	Description string
}

// Complete list of OTC services based on official documentation
var otcServices = []ServiceInfo{
  // Compute
  {"aom", "Application Operations Management", "Application monitoring and management"},
  {"bms", "Bare Metal Service", "Dedicated physical servers"},
  {"deh", "Dedicated Host", "Dedicated physical hosts"},
  {"dehl", "Dedicated Host License", "License for dedicated hosts"},
  {"dins", "Disk Intensive Service", "High disk I/O instances"},
  {"ecs", "Elastic Cloud Server", "Virtual machines"},
  {"ecsnoc", "ECS (Compute Optimized)", "Compute-optimized VMs"},
  {"gpu", "GPU Service", "GPU-accelerated computing"},
  {"hps", "High Performance Server", "High-performance computing"},
  {"lms", "Large Memory Service", "Large memory instances"},
  {"memo", "ECS (Memory Optimized)", "Memory-optimized VMs"},
  {"uhio", "ECS (Ultra High I/O)", "Ultra-high I/O VMs"},
  
  // Storage
  {"cbr", "Cloud Backup and Recovery", "Backup and disaster recovery"},
  {"coss", "Object Storage Service (Cold)", "Cold archive storage"},
  {"csbs", "Cloud Server Backup Service", "ECS backup service"},
  {"evs", "Elastic Volume Service", "Block storage volumes"},
  {"obs", "Object Storage Service (Standard)", "S3-compatible object storage"},
  {"sdrs", "Storage Disaster Recovery", "Disaster recovery for storage"},
  {"sfs", "Scalable File Service", "NFS file storage"},
  {"vbs", "Volume Backup Service", "Volume backup"},
  {"woss", "Object Storage Service (Warm)", "Infrequent access storage"},
  
  // Database
  {"das", "Data Admin Service", "Database administration"},
  {"dcs", "Distributed Cache Service", "Redis/Memcached cache"},
  {"dcsb", "Distributed Cache Service Backup", "DCS backup"},
  {"dds", "Document Database Service", "MongoDB-compatible database"},
  {"ddsrs", "Document Database Service RS", "DDS replica set"},
  {"ddssn", "Document Database SN", "DDS sharded cluster"},
  {"ddm", "Distributed Database Middleware", "Database sharding middleware"},
  {"drs", "Data Replication Service", "Database replication"},
  {"dws", "Data Warehouse Service", "Cloud data warehouse"},
  {"gaussdbforcassandra", "GaussDB for Cassandra", "Cassandra-compatible database"},
  {"gaussdbcassbak", "GaussDB Cassandra Backup", "Cassandra backup"},
  {"gaussdbcassvs", "GaussDB Cassandra Volume", "Cassandra storage"},
  {"gaussdbformysql", "GaussDB for MySQL", "MySQL-compatible database"},
  {"gaussdbmysqlbak", "GaussDB MySQL Backup", "MySQL backup"},
  {"gaussdbmysqlvs", "GaussDB MySQL Volume", "MySQL storage"},
  {"rds", "Relational Database Service", "MySQL/PostgreSQL/SQL Server"},
  {"rdss", "Relational Database Storage", "RDS storage"},
  
  // Networking
  {"bandwidth", "Bandwidth", "Internet bandwidth"},
  {"dc", "Direct Connect", "Dedicated network connection"},
  {"dcsetup", "Direct Connect Setup", "Direct Connect setup fees"},
  {"eip", "Elastic IP Service", "Public IP addresses"},
  {"elb", "Elastic Load Balancer", "Load balancing"},
  {"er", "Enterprise Router", "Enterprise routing service"},
  {"evpn", "Enterprise VPN", "Enterprise VPN connections"},
  {"ito", "Internet Traffic Outbound", "Internet bandwidth"},
  {"nat", "NAT Gateway", "Network address translation"},
  {"pnat", "Private NAT", "Private NAT gateway"},
  {"vpc", "Virtual Private Cloud", "Virtual private network"},
  {"vpcep", "VPC Endpoint", "Private service endpoints"},
  {"vpn", "Virtual Private Network", "VPN connections"},
  
  // Container & Orchestration
  {"cce", "Cloud Container Engine", "Kubernetes container orchestration"},
  {"cci", "Cloud Container Instance", "Serverless containers"},
  {"ms-man-cce", "Managed CCE", "Managed Kubernetes service"},
  {"swr", "Software Repository for Container", "Container image registry"},
  
  // Application Services
  {"apig", "API Gateway", "API management service"},
  {"cf", "Cloud Foundry", "PaaS platform"},
  {"dms", "Distributed Message Service", "Message queue service"},
  {"dmsip", "DMS Public IP", "DMS public IP addresses"},
  {"dmsk", "DMS Kafka", "Kafka message service"},
  {"dmsrmq", "DMS RabbitMQ", "RabbitMQ message service"},
  {"dmsvol", "DMS Volume", "DMS storage volume"},
  {"smn", "Simple Message Notification", "Pub/sub messaging"},
  
  // Big Data & Analytics
  {"css", "Cloud Search Service", "Elasticsearch-based search"},
  {"csscln", "CSS Client Node", "CSS client nodes"},
  {"csscon", "CSS Cold Node", "CSS cold storage nodes"},
  {"cssman", "CSS Master Node", "CSS master nodes"},
  {"dis", "Data Ingestion Service", "Real-time data ingestion"},
  {"dli", "Data Lake Insight", "Serverless data analytics"},
  {"mrs", "MapReduce Service", "Hadoop/Spark clusters"},
  
  // AI & Machine Learning
  {"modelarts", "ModelArts", "AI development platform"},
  {"ocr", "Optical Character Recognition", "OCR service"},
  
  // Security
  {"cwaf", "Cloud Web Application Firewall", "Cloud WAF service"},
  {"hss", "Host Security Service", "Host security"},
  {"kms", "Key Management Service", "Encryption key management"},
  {"waf", "Web Application Firewall", "Web application protection"},
  {"wafd", "WAF Dedicated", "Dedicated WAF instances"},
  
  // Management & Monitoring
  {"apm2", "Application Performance Management", "APM monitoring"},
  {"lts", "Log Tank Service", "Log management"},
  {"mas", "Monitoring as a Service", "Cloud monitoring"},
  {"smg", "Service Management Gateway", "Service management"},
  
  // DNS
  {"dnprq", "DNS Private Queries", "Private DNS queries"},
  {"dnq", "DNS Public Queries", "Public DNS queries"},
  {"phz", "DNS Service Public", "Public DNS zones"},
  {"prhz", "DNS Service Private", "Private DNS zones"},
  
  // Enterprise & Specialized
  {"buc", "Bucket", "Storage buckets"},
  {"cse", "Cloud Service Engine", "Microservices engine"},
  {"da", "Data Arts", "Data governance"},
  {"dss", "Data Security Service", "Data security"},
  {"ea", "Enterprise Agreement", "Enterprise contracts"},
  {"fug", "FunctionGraph", "Serverless functions"},
  {"mip", "Mail IP Service", "Email IP addresses"},
  {"mss", "Mobile Storage Solution", "Mobile device storage"},
  {"pefd", "Enterprise Financial Dashboard", "Financial reporting"},
  {"plas", "Private Line Access Service", "Private network access"},
  {"sap", "SAP", "SAP workloads"},
  {"sapls", "SAP License Service", "SAP licensing"},
  
  // Managed Services
  {"ms-base-fee", "Managed Service Base Fee", "Base management fee"},
  {"ms-man-app", "Managed Application", "Managed application services"},
  {"ms-man-db", "Managed Database", "Managed database services"},
  {"ms-man-os", "Managed OS", "Managed operating system"},
  {"ms-man-vpn", "Managed VPN", "Managed VPN service"},
}

// ServicePricing holds pricing data for any OTC service
type ServicePricing struct {
	ServiceName        string
	ResourceName       string
	Region             string
	PriceAmount        string
	OSUnit             string
	ProductIdParameter string
	Description        string
	BillingMode        string
	// Additional fields from API
	OpiFlavour string
	VCPU       string
	RAM        string
}

// ListAvailableServices lists all available OTC services
func ListAvailableServices(region string) {
	fmt.Printf("\n")
	color.Cyan("Available OTC Services for Pricing:")
	fmt.Printf("\n")

	// Group services by category
	categories := map[string][]ServiceInfo{
		"Compute":     {},
		"Containers":  {},
		"Storage":     {},
		"Database":    {},
		"Networking":  {},
		"Application": {},
		"Security":    {},
		"Monitoring":  {},
		"Big Data":    {},
		"AI/ML":       {},
		"Other":       {},
	}

	// Categorize services
	for _, svc := range otcServices {
		code := strings.ToLower(svc.Code)
		switch {
		case strings.Contains(code, "ecs") || strings.Contains(code, "bms") ||
			strings.Contains(code, "gpu") || strings.Contains(code, "hps") ||
			code == "deh" || code == "dehl" || code == "memo" || code == "uhio":
			categories["Compute"] = append(categories["Compute"], svc)
		case strings.Contains(code, "cce") || strings.Contains(code, "cci") || code == "swr":
			categories["Containers"] = append(categories["Containers"], svc)
		case strings.Contains(code, "obs") || strings.Contains(code, "evs") ||
			strings.Contains(code, "sfs") || strings.Contains(code, "oss"):
			categories["Storage"] = append(categories["Storage"], svc)
		case strings.Contains(code, "rds") || strings.Contains(code, "dds") ||
			strings.Contains(code, "dcs") || strings.Contains(code, "dws"):
			categories["Database"] = append(categories["Database"], svc)
		case strings.Contains(code, "vpc") || strings.Contains(code, "elb") ||
			strings.Contains(code, "eip") || strings.Contains(code, "nat") ||
			strings.Contains(code, "vpn") || code == "dc":
			categories["Networking"] = append(categories["Networking"], svc)
		case strings.Contains(code, "dms") || strings.Contains(code, "smn") ||
			strings.Contains(code, "kafka"):
			categories["Application"] = append(categories["Application"], svc)
		case strings.Contains(code, "waf") || strings.Contains(code, "kms"):
			categories["Security"] = append(categories["Security"], svc)
		case strings.Contains(code, "aom") || strings.Contains(code, "lts") ||
			strings.Contains(code, "mas"):
			categories["Monitoring"] = append(categories["Monitoring"], svc)
		case strings.Contains(code, "mrs") || strings.Contains(code, "dis") ||
			strings.Contains(code, "drs") || strings.Contains(code, "css"):
			categories["Big Data"] = append(categories["Big Data"], svc)
		case strings.Contains(code, "modelarts"):
			categories["AI/ML"] = append(categories["AI/ML"], svc)
		default:
			categories["Other"] = append(categories["Other"], svc)
		}
	}

	// Print categorized services
	categoryOrder := []string{"Compute", "Containers", "Storage", "Database", "Networking",
		"Application", "Security", "Monitoring", "Big Data", "AI/ML", "Other"}

	for _, category := range categoryOrder {
		services := categories[category]
		if len(services) == 0 {
			continue
		}

		color.Yellow("\n%s:", category)
		for _, svc := range services {
			fmt.Printf("  %-10s - %s\n", svc.Code, svc.Description)
		}
	}

	fmt.Printf("\n")
	color.Cyan("Usage:")
	fmt.Printf("  otc-cli list pricing <service-code>\n")
	fmt.Printf("  otc-cli list pricing ecs\n")
	fmt.Printf("  otc-cli list pricing obs\n")
	fmt.Printf("  otc-cli list pricing rds\n")
	if region != "" {
		fmt.Printf("  otc-cli list pricing ecs --region %s\n", region)
	}
  fmt.Printf("\n")
  color.Yellow("Note: Some services may not have pricing data in all regions.")
  color.Yellow("To see services with actual pricing data, use:")
  fmt.Printf("  otc-cli list pricing --discover\n\n")
  
  color.Cyan("Total Services: %d\n\n", len(otcServices))
}

// FetchAvailableServicesFromAPI dynamically fetches services from the API
func FetchAvailableServicesFromAPI(region string) ([]string, error) {
	// Make a dummy request to get the response structure
	pricingURL := fmt.Sprintf("https://calculator.otc-service.com/en/open-telekom-price-api/?region=%s&limitMax=1", region)

	req, _ := http.NewRequest("GET", pricingURL, nil)
	req.Header.Set("Content-Type", "application/json")

	httpClient := &http.Client{Timeout: 10 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var apiResponse struct {
		Response struct {
			Result map[string]interface{} `json:"result"`
		} `json:"response"`
	}

	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Extract service names from result keys
	var services []string
	for serviceName := range apiResponse.Response.Result {
		services = append(services, serviceName)
	}

	sort.Strings(services)
	return services, nil
}

// ListServicePricing lists pricing for a specific OTC service
func ListServicePricing(cfg *config.Config, client *otc.Client, unscopedToken, projectID string, serviceName string, raw bool, csvOutput bool, filters map[string]string, osFilter string) {
  if serviceName == "" {
    color.Red("✗ Service name is required")
    ListAvailableServices(cfg.Region)
    return
  }

  // Fetch pricing data
  pricing, err := FetchServicePricing(cfg.Region, serviceName, filters)
  if err != nil {
    color.Red("✗ Failed to fetch pricing: %v", err)
    return
  }

  if len(pricing) == 0 {
    color.Red("✗ No pricing data found for service: %s in region: %s", serviceName, cfg.Region)
    color.Yellow("\nTip: Check available services with: otc-cli list pricing --services")
    return
  }

  // Apply OS filter if specified
  if osFilter != "" {
    pricing = filterByOS(pricing, osFilter)
    if len(pricing) == 0 {
      color.Red("✗ No pricing data found for OS type: %s", osFilter)
      color.Yellow("\nAvailable OS types: linux, windows, redhat, suse, oracle, ubuntu, centos")
      return
    }
  }

  if raw {
    formatted, _ := json.MarshalIndent(pricing, "", "  ")
    fmt.Println(string(formatted))
    return
  }

  // Remove duplicates - keep unique combinations of resource + price + OS
  uniquePricing := deduplicatePricing(pricing)

  // Sort by price
  sort.Slice(uniquePricing, func(i, j int) bool {
    return parsePrice(uniquePricing[i].PriceAmount) < parsePrice(uniquePricing[j].PriceAmount)
  })

  // CSV output
  if csvOutput {
    outputCSV(uniquePricing, serviceName, cfg.Region, osFilter, filters)
    return
  }

  // Table output (existing code)
  outputTable(uniquePricing, serviceName, cfg.Region, osFilter, filters)
}

// outputCSV generates CSV format output
func outputCSV(pricing []ServicePricing, serviceName, region, osFilter string, filters map[string]string) {
  writer := csv.NewWriter(os.Stdout)
  defer writer.Flush()

  isCompute := isComputeService(serviceName)

  // Write headers
  if isCompute {
    writer.Write([]string{"Resource", "vCPUs", "RAM", "OS Type", "Price/Hour (EUR)", "Price/Month (EUR)"})
  } else {
    writer.Write([]string{"Resource", "Type", "OS/Spec", "Price/Hour (EUR)", "Price/Month (EUR)"})
  }

  // Write data rows
  for _, p := range pricing {
    hourly := parsePrice(p.PriceAmount)
    monthly := hourly * 730

    // Check if price is already monthly
    isMonthly := strings.Contains(strings.ToLower(p.BillingMode), "month")
    
    var priceStr, monthlyStr string
    if isMonthly {
      priceStr = "-"
      monthlyStr = fmt.Sprintf("%.2f", hourly)
    } else {
      priceStr = fmt.Sprintf("%.4f", hourly)
      monthlyStr = fmt.Sprintf("%.2f", monthly)
    }

    if isCompute {
      vcpu := p.VCPU
      ram := p.RAM
      if vcpu == "" || vcpu == "0" {
        vcpu = "-"
      }
      if ram == "" || ram == "0 GiB" {
        ram = "-"
      }
      
      osType := extractOSType(p.OSUnit)
      
      writer.Write([]string{
        p.ResourceName,
        vcpu,
        ram,
        osType,
        priceStr,
        monthlyStr,
      })
    } else {
      osOrSpec := p.OSUnit
      if osOrSpec == "" {
        osOrSpec = p.ProductIdParameter
      }
      if osOrSpec == "" {
        osOrSpec = "-"
      }
      
      writer.Write([]string{
        p.ResourceName,
        p.ProductIdParameter,
        osOrSpec,
        priceStr,
        monthlyStr,
      })
    }
  }
}

// outputTable generates table format output (existing code refactored)
func outputTable(uniquePricing []ServicePricing, serviceName, region, osFilter string, filters map[string]string) {
  // Create table
  headerFmt := color.New(color.FgCyan, color.Bold).SprintfFunc()
  
  // Dynamic table headers based on service type
  var tbl table.Table
  if isComputeService(serviceName) {
    tbl = table.New("Resource", "vCPUs", "RAM", "OS Type", "Price/Hour", "Price/Month")
  } else {
    tbl = table.New("Resource", "Type", "OS/Spec", "Price/Hour", "Price/Month")
  }
  tbl.WithHeaderFormatter(headerFmt)

  // Add rows
  for _, p := range uniquePricing {
    hourly := parsePrice(p.PriceAmount)
    monthly := hourly * 730

    // Check if price is already monthly
    isMonthly := strings.Contains(strings.ToLower(p.BillingMode), "month")
    
    var priceStr, monthlyStr string
    if isMonthly {
      priceStr = "-"
      monthlyStr = fmt.Sprintf("€%.2f", hourly)
    } else {
      priceStr = fmt.Sprintf("€%.4f", hourly)
      monthlyStr = fmt.Sprintf("€%.2f", monthly)
    }

    if isComputeService(serviceName) {
      vcpu := p.VCPU
      ram := p.RAM
      if vcpu == "" || vcpu == "0" {
        vcpu = "-"
      }
      if ram == "" || ram == "0 GiB" {
        ram = "-"
      }
      
      osType := extractOSType(p.OSUnit)
      
      tbl.AddRow(
        truncateString(p.ResourceName, 20),
        vcpu,
        ram,
        osType,
        priceStr,
        monthlyStr,
      )
    } else {
      osOrSpec := p.OSUnit
      if osOrSpec == "" {
        osOrSpec = p.ProductIdParameter
      }
      if osOrSpec == "" {
        osOrSpec = "-"
      }
      
      tbl.AddRow(
        truncateString(p.ResourceName, 25),
        truncateString(p.ProductIdParameter, 18),
        truncateString(osOrSpec, 18),
        priceStr,
        monthlyStr,
      )
    }
  }

  // Print table
  fmt.Printf("\n")
  
  serviceDesc := serviceName
  for _, svc := range otcServices {
    if strings.EqualFold(svc.Code, serviceName) {
      serviceDesc = fmt.Sprintf("%s (%s)", svc.Name, serviceName)
      break
    }
  }
  
  color.Cyan("Service Pricing: %s", serviceDesc)
  color.Cyan("Region: %s", region)
  
  if osFilter != "" {
    color.Yellow("OS Filter: %s", osFilter)
  }
  
  if len(filters) > 0 {
    color.Yellow("Filters applied:")
    for k, v := range filters {
      color.Yellow("  %s: %s", k, v)
    }
  }
  
  fmt.Printf("\n")
  tbl.Print()
  fmt.Printf("\nTotal: %d unique items | Sorted by price (lowest to highest)\n", len(uniquePricing))
  if osFilter != "" || len(filters) > 0 {
    color.Yellow("Note: Results filtered. Use without filters to see all variants.\n")
  }
  fmt.Printf("Pricing based on hourly rates (730 hours/month = 1 month)\n\n")
}

// filterByOS filters pricing by OS type
func filterByOS(pricing []ServicePricing, osFilter string) []ServicePricing {
  osLower := strings.ToLower(osFilter)
  var filtered []ServicePricing

  for _, p := range pricing {
    osUnit := strings.ToLower(p.OSUnit)
    
    match := false
    switch osLower {
    case "linux", "openlinux":
      match = (strings.Contains(osUnit, "open") || 
               strings.Contains(osUnit, "standard") ||
               strings.Contains(osUnit, "linux")) &&
              !strings.Contains(osUnit, "suse") &&
              !strings.Contains(osUnit, "redhat") &&
              !strings.Contains(osUnit, "oracle") &&
              !strings.Contains(osUnit, "windows")
    case "windows", "win":
      match = strings.Contains(osUnit, "windows")
    case "redhat", "rhel":
      match = strings.Contains(osUnit, "redhat")
    case "suse", "sles":
      match = strings.Contains(osUnit, "suse")
    case "oracle":
      match = strings.Contains(osUnit, "oracle")
    case "ubuntu":
      match = strings.Contains(osUnit, "ubuntu")
    case "centos":
      match = strings.Contains(osUnit, "centos")
    case "debian":
      match = strings.Contains(osUnit, "debian")
    case "euleros":
      match = strings.Contains(osUnit, "euleros")
    default:
      // Partial match for unknown OS types
      match = strings.Contains(osUnit, osLower)
    }

    if match {
      filtered = append(filtered, p)
    }
  }

  return filtered
}


// deduplicatePricing removes duplicate pricing entries
// Keeps unique combinations of: ResourceName + vCPU + RAM + Price + OSUnit
func deduplicatePricing(pricing []ServicePricing) []ServicePricing {
  seen := make(map[string]bool)
  var unique []ServicePricing

  for _, p := range pricing {
    // Create unique key based on resource specs and price
    key := fmt.Sprintf("%s|%s|%s|%s|%s", 
      p.ResourceName, 
      p.VCPU, 
      p.RAM, 
      p.PriceAmount,
      p.OSUnit,
    )

    if !seen[key] {
      seen[key] = true
      unique = append(unique, p)
    }
  }

  return unique
}

// extractOSType extracts a short OS type from OSUnit field
func extractOSType(osUnit string) string {
  osLower := strings.ToLower(osUnit)
  
  switch {
  case strings.Contains(osLower, "windows"):
    return "Windows"
  case strings.Contains(osLower, "redhat"):
    return "RedHat"
  case strings.Contains(osLower, "suse"):
    return "SUSE"
  case strings.Contains(osLower, "oracle"):
    return "Oracle"
  case strings.Contains(osLower, "ubuntu"):
    return "Ubuntu"
  case strings.Contains(osLower, "centos"):
    return "CentOS"
  case strings.Contains(osLower, "debian"):
    return "Debian"
  case strings.Contains(osLower, "euleros"):
    return "EulerOS"
  case strings.Contains(osLower, "open") || strings.Contains(osLower, "standard"):
    return "Linux"
  case osUnit == "":
    return "-"
  default:
    // Return first 12 chars if unknown
    if len(osUnit) > 12 {
      return osUnit[:12]
    }
    return osUnit
  }
}


// DiscoverAvailableServices fetches and displays services that actually have pricing data
func DiscoverAvailableServices(region string) error {
  fmt.Printf("\n")
  color.Cyan("Discovering services with pricing data for region: %s", region)
  fmt.Printf("\n\n")

  // Fetch ALL pricing data with no service filter
  pricingURL := fmt.Sprintf("https://calculator.otc-service.com/en/open-telekom-price-api/?region=%s&limitMax=999999", region)

  req, _ := http.NewRequest("GET", pricingURL, nil)
  req.Header.Set("Content-Type", "application/json")

  httpClient := &http.Client{Timeout: 30 * time.Second}
  resp, err := httpClient.Do(req)
  if err != nil {
    return fmt.Errorf("HTTP request failed: %w", err)
  }
  defer resp.Body.Close()

  body, _ := io.ReadAll(resp.Body)

  var apiResponse struct {
    Response struct {
      Result map[string]interface{} `json:"result"`
      Status struct {
        Count int `json:"count"`
      } `json:"status"`
    } `json:"response"`
  }

  if err := json.Unmarshal(body, &apiResponse); err != nil {
    return fmt.Errorf("failed to parse response: %w", err)
  }

  // Extract service names that have data
  var availableServices []string
  serviceDataCount := make(map[string]int)

  for serviceName, data := range apiResponse.Response.Result {
    // Check if service has actual pricing records
    var recordCount int
    switch v := data.(type) {
    case []interface{}:
      recordCount = len(v)
    case map[string]interface{}:
      for _, items := range v {
        if itemsArray, ok := items.([]interface{}); ok {
          recordCount += len(itemsArray)
        }
      }
    }

    if recordCount > 0 {
      availableServices = append(availableServices, serviceName)
      serviceDataCount[serviceName] = recordCount
    }
  }

  sort.Strings(availableServices)

  color.Green("✓ Found %d services with pricing data\n\n", len(availableServices))

  // Group by category
  categories := map[string][]string{
    "Compute":    {},
    "Storage":    {},
    "Database":   {},
    "Networking": {},
    "Container":  {},
    "Application": {},
    "Other":      {},
  }

  for _, svc := range availableServices {
    svcLower := strings.ToLower(svc)
    
    // Find description from our static list
    description := ""
    for _, s := range otcServices {
      if strings.EqualFold(s.Code, svc) {
        description = s.Description
        break
      }
    }
    if description == "" {
      description = "No description available"
    }

    // Categorize
    switch {
    case strings.Contains(svcLower, "ecs") || strings.Contains(svcLower, "bms") ||
         strings.Contains(svcLower, "gpu") || strings.Contains(svcLower, "hps") ||
         svcLower == "deh" || svcLower == "dehl" || svcLower == "memo" || svcLower == "uhio":
      categories["Compute"] = append(categories["Compute"], fmt.Sprintf("%-12s (%d items) - %s", svc, serviceDataCount[svc], description))
    case strings.Contains(svcLower, "obs") || strings.Contains(svcLower, "evs") ||
         strings.Contains(svcLower, "sfs") || strings.Contains(svcLower, "oss") || svcLower == "vbs":
      categories["Storage"] = append(categories["Storage"], fmt.Sprintf("%-12s (%d items) - %s", svc, serviceDataCount[svc], description))
    case strings.Contains(svcLower, "rds") || strings.Contains(svcLower, "dds") ||
         strings.Contains(svcLower, "dcs") || strings.Contains(svcLower, "dws"):
      categories["Database"] = append(categories["Database"], fmt.Sprintf("%-12s (%d items) - %s", svc, serviceDataCount[svc], description))
    case strings.Contains(svcLower, "vpc") || strings.Contains(svcLower, "elb") ||
         strings.Contains(svcLower, "eip") || strings.Contains(svcLower, "nat") ||
         strings.Contains(svcLower, "bandwidth") || svcLower == "vpn":
      categories["Networking"] = append(categories["Networking"], fmt.Sprintf("%-12s (%d items) - %s", svc, serviceDataCount[svc], description))
    case strings.Contains(svcLower, "cce") || strings.Contains(svcLower, "cci") || svcLower == "swr":
      categories["Container"] = append(categories["Container"], fmt.Sprintf("%-12s (%d items) - %s", svc, serviceDataCount[svc], description))
    case strings.Contains(svcLower, "dms") || strings.Contains(svcLower, "smn") ||
         strings.Contains(svcLower, "apig") || strings.Contains(svcLower, "kafka"):
      categories["Application"] = append(categories["Application"], fmt.Sprintf("%-12s (%d items) - %s", svc, serviceDataCount[svc], description))
    default:
      categories["Other"] = append(categories["Other"], fmt.Sprintf("%-12s (%d items) - %s", svc, serviceDataCount[svc], description))
    }
  }

  // Print categorized services
  categoryOrder := []string{"Compute", "Storage", "Database", "Networking", "Container", "Application", "Other"}

  for _, category := range categoryOrder {
    services := categories[category]
    if len(services) == 0 {
      continue
    }

    color.Yellow("\n%s (%d services):", category, len(services))
    for _, svc := range services {
      fmt.Printf("  %s\n", svc)
    }
  }

  fmt.Printf("\n")
  color.Cyan("Usage:")
  fmt.Printf("  otc-cli list pricing <service-code>\n")
  if len(availableServices) > 0 {
    // Show a meaningful example (prefer common services)
    exampleService := availableServices[0]
    preferredExamples := []string{"ecs", "obs", "rds", "dcs", "cce", "elb", "evs"}
    
    for _, preferred := range preferredExamples {
      for _, available := range availableServices {
        if available == preferred {
          exampleService = preferred
          break
        }
      }
      if exampleService != availableServices[0] {
        break
      }
    }
    
    fmt.Printf("  Example: otc-cli list pricing %s\n", exampleService)
  }
  fmt.Printf("\n")
  
  // Calculate total records
  totalRecords := 0
  for _, count := range serviceDataCount {
    totalRecords += count
  }
  color.Cyan("Total pricing records: %d across %d services\n", totalRecords, len(availableServices))
  fmt.Printf("\n")

  return nil
}


// FetchServicePricing fetches pricing from OTC Price API for any service
func FetchServicePricing(region string, serviceName string, filters map[string]string) ([]ServicePricing, error) {
	var results []ServicePricing

	pricingURL := fmt.Sprintf("https://calculator.otc-service.com/en/open-telekom-price-api/?serviceName=%s&region=%s&limitMax=1000",
		serviceName, region)

	req, _ := http.NewRequest("GET", pricingURL, nil)
	req.Header.Set("Content-Type", "application/json")

	httpClient := &http.Client{Timeout: 15 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	// First, try to parse as map structure
	var apiResponseMap struct {
		Response struct {
			Result map[string][]struct {
				OpiFlavour         string `json:"opiFlavour"`
				VCPU               string `json:"vCpu"`
				RAM                string `json:"ram"`
				PriceAmount        string `json:"priceAmount"`
				OSUnit             string `json:"osUnit"`
				Region             string `json:"region"`
				ProductIdParameter string `json:"productIdParameter"`
				Description        string `json:"description"`
				ResourceType       string `json:"resourceType"`
				BillingMode        string `json:"billingMode"`
				SpecCode           string `json:"specCode"`
				Unit               string `json:"unit"`
				ProductName        string `json:"productName"`
			} `json:"result"`
		} `json:"response"`
	}

	// Try to unmarshal as map first
	errMap := json.Unmarshal(body, &apiResponseMap)

	// If map parsing fails, try array structure
	if errMap != nil {
		var apiResponseArray struct {
			Response struct {
				Result []struct {
					OpiFlavour         string `json:"opiFlavour"`
					VCPU               string `json:"vCpu"`
					RAM                string `json:"ram"`
					PriceAmount        string `json:"priceAmount"`
					OSUnit             string `json:"osUnit"`
					Region             string `json:"region"`
					ProductIdParameter string `json:"productIdParameter"`
					Description        string `json:"description"`
					ResourceType       string `json:"resourceType"`
					BillingMode        string `json:"billingMode"`
					SpecCode           string `json:"specCode"`
					Unit               string `json:"unit"`
					ProductName        string `json:"productName"`
				} `json:"result"`
			} `json:"response"`
		}

		if err := json.Unmarshal(body, &apiResponseArray); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		// Process array structure
		results = processItems(apiResponseArray.Response.Result, region, serviceName, filters)
	} else {
		// Process map structure
		for _, items := range apiResponseMap.Response.Result {
			results = append(results, processItems(items, region, serviceName, filters)...)
		}
	}

	return results, nil
}

// processItems processes pricing items and returns ServicePricing slice
func processItems(items []struct {
	OpiFlavour         string `json:"opiFlavour"`
	VCPU               string `json:"vCpu"`
	RAM                string `json:"ram"`
	PriceAmount        string `json:"priceAmount"`
	OSUnit             string `json:"osUnit"`
	Region             string `json:"region"`
	ProductIdParameter string `json:"productIdParameter"`
	Description        string `json:"description"`
	ResourceType       string `json:"resourceType"`
	BillingMode        string `json:"billingMode"`
	SpecCode           string `json:"specCode"`
	Unit               string `json:"unit"`
	ProductName        string `json:"productName"`
}, region string, serviceName string, filters map[string]string) []ServicePricing {
	var results []ServicePricing
	seen := make(map[string]bool)

	for _, item := range items {
		if item.Region != region {
			continue
		}

		// Apply filters
		if !matchesFilters(item, filters) {
			continue
		}

		// Create unique key
		key := fmt.Sprintf("%s-%s-%s-%s", item.OpiFlavour, item.ProductIdParameter, item.SpecCode, item.PriceAmount)
		if seen[key] {
			continue
		}

		resourceName := item.OpiFlavour
		if resourceName == "" {
			resourceName = item.ProductName
		}
		if resourceName == "" {
			resourceName = item.ResourceType
		}
		if resourceName == "" {
			resourceName = item.SpecCode
		}
		if resourceName == "" {
			resourceName = item.ProductIdParameter
		}

		// Determine billing mode from unit field
		billingMode := item.BillingMode
		if billingMode == "" && item.Unit != "" {
			if strings.Contains(strings.ToLower(item.Unit), "month") {
				billingMode = "monthly"
			} else if strings.Contains(strings.ToLower(item.Unit), "hour") {
				billingMode = "hourly"
			}
		}

		results = append(results, ServicePricing{
			ServiceName:        serviceName,
			ResourceName:       resourceName,
			Region:             item.Region,
			PriceAmount:        item.PriceAmount,
			OSUnit:             item.OSUnit,
			ProductIdParameter: item.ProductIdParameter,
			Description:        item.Description,
			BillingMode:        billingMode,
			OpiFlavour:         item.OpiFlavour,
			VCPU:               item.VCPU,
			RAM:                item.RAM,
		})

		seen[key] = true
	}

	return results
}

// matchesFilters checks if an item matches the given filters
func matchesFilters(item interface{}, filters map[string]string) bool {
	if len(filters) == 0 {
		return true
	}

	// Convert struct to map for generic filtering
	data, _ := json.Marshal(item)
	var itemMap map[string]interface{}
	json.Unmarshal(data, &itemMap)

	for key, value := range filters {
		itemValue, exists := itemMap[key]
		if !exists {
			return false
		}

		itemValueStr := fmt.Sprintf("%v", itemValue)
		if !strings.Contains(strings.ToLower(itemValueStr), strings.ToLower(value)) {
			return false
		}
	}

	return true
}

// isComputeService checks if service is compute-related (VMs with vCPU/RAM specs)
func isComputeService(serviceName string) bool {
	computeServices := []string{"ecs", "ecsnoc", "memo", "uhio", "hps", "gpu", "deh", "dehl", "bms"}
	for _, s := range computeServices {
		if strings.EqualFold(s, serviceName) {
			return true
		}
	}
	return false
}
