package cli

import (
	"fmt"
	"os"
	"strconv"

	"github.com/abdo-farag/otc-cli/internal/commands"
	"github.com/abdo-farag/otc-cli/internal/config"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	// Login flags
	idpURL              string
	idpClientID         string
	domainName          string
	authURL             string
	idpProviderName     string
	region              string
	redirectPort        int
	outputFile          string
	noBrowser           bool
	codeChallengeMethod string
	scope               string
	iamMode             bool
	username            string
	password            string
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate and generate credentials",
	Long: `Authenticate with OTC using either OIDC federation or IAM direct authentication.

OIDC Authentication (default):
  Uses OAuth2/OIDC flow with your identity provider.
  
IAM Authentication (--iam flag):
  Uses direct username/password authentication with OTC IAM.`,
	Example: `  # OIDC authentication
  otc-cli login --idp-url https://idp.example.com --idp-client-id myclient

  # IAM authentication
  otc-cli login --iam --username myuser

  # IAM with environment variables
  export OS_USERNAME=myuser
  export OS_PASSWORD=mypassword
  otc-cli login --iam`,
	RunE: runLogin,
}

func init() {
  // OIDC flags - use empty defaults
  loginCmd.Flags().StringVar(&idpURL, "idp-url", "", "IDP URL")
  loginCmd.Flags().StringVar(&idpClientID, "idp-client-id", "", "IDP client ID")
  loginCmd.Flags().StringVar(&domainName, "domain-name", "", "OTC domain name")
  loginCmd.Flags().StringVar(&authURL, "auth-url", "", "OTC IAM endpoint")
  loginCmd.Flags().StringVar(&idpProviderName, "idp-provider", "", "IDP provider name")
  loginCmd.Flags().StringVar(&region, "region", "", "Region")
  loginCmd.Flags().IntVar(&redirectPort, "port", 9197, "Callback port")
  loginCmd.Flags().StringVar(&outputFile, "output", "", "Output file")
  loginCmd.Flags().BoolVar(&noBrowser, "no-browser", false, "Don't open browser automatically")
  loginCmd.Flags().StringVar(&codeChallengeMethod, "code-challenge-method", "S256", "PKCE method (S256 or plain)")
  loginCmd.Flags().StringVar(&scope, "scope", "openid email profile roles groups organization", "OIDC scopes")

  // IAM flags - use empty defaults
  loginCmd.Flags().BoolVar(&iamMode, "iam", false, "Use IAM direct authentication")
  loginCmd.Flags().StringVar(&username, "username", "", "IAM username")
  loginCmd.Flags().StringVar(&password, "password", "", "IAM password")
}


func runLogin(cmd *cobra.Command, args []string) error {
	cfg := buildConfig()

	// Handle IAM authentication
	if iamMode {
		return handleIAMLogin(cfg)
	}

	// Handle OIDC authentication
	return handleOIDCLogin(cfg)
}

func buildConfig() *config.Config {
  cfg := config.New()
  
  // Priority: flag > env var > config default
  if idpURL != "" {
    cfg.IdpURL = idpURL
  } else if env := os.Getenv("IDP_URL"); env != "" {
    cfg.IdpURL = env
  }
  
  if idpClientID != "" {
    cfg.IdpClientID = idpClientID
  } else if env := os.Getenv("IDP_CLIENT_ID"); env != "" {
    cfg.IdpClientID = env
  }
  
  if domainName != "" {
    cfg.DomainName = domainName
  } else if env := os.Getenv("OS_DOMAIN_NAME"); env != "" {
    cfg.DomainName = env
  }
  
  if authURL != "" {
    cfg.AUTHURL = authURL
  } else if env := os.Getenv("OS_AUTH_URL"); env != "" {
    cfg.AUTHURL = env
  }
  
  if idpProviderName != "" {
    cfg.IDPProviderName = idpProviderName
  } else if env := os.Getenv("IDP_PROVIDER_NAME"); env != "" {
    cfg.IDPProviderName = env
  }
  
  if region != "" {
    cfg.Region = region
  } else if env := os.Getenv("OS_REGION_NAME"); env != "" {
    cfg.Region = env
  }
  
  if redirectPort == 9197 { // Check if still default
    if portEnv := os.Getenv("REDIRECT_PORT"); portEnv != "" {
      if port, err := strconv.Atoi(portEnv); err == nil {
        cfg.RedirectPort = port
      }
    } else {
      cfg.RedirectPort = redirectPort
    }
  } else {
    cfg.RedirectPort = redirectPort
  }
  
  if outputFile != "" {
    cfg.OutputFile = outputFile
  }
  
  if codeChallengeMethod != "" {
    cfg.CodeChallengeMethod = codeChallengeMethod
  }
  
  if scope != "" {
    cfg.Scope = scope
  }
  
  cfg.NoBrowser = noBrowser

  return cfg
}

func handleIAMLogin(cfg *config.Config) error {
  // Priority: flag > env var > prompt
  user := username
  if user == "" {
    user = os.Getenv("OS_USERNAME")
  }
  if user == "" {
    user = commands.PromptUsername()
  }
  
  pass := password
  if pass == "" {
    pass = os.Getenv("OS_PASSWORD")
  }
  if pass == "" {
    pass = commands.PromptPassword()
  }

  if user == "" || pass == "" {
    return fmt.Errorf("username and password are required")
  }

  if err := commands.LoginIAM(cfg, user, pass); err != nil {
    return err
  }

  color.Green("✓ Successfully authenticated with IAM")
  return nil
}

func handleOIDCLogin(cfg *config.Config) error {
	if err := validateOIDCConfig(cfg); err != nil {
		return err
	}

	if err := cfg.ValidateCodeChallengeMethod(); err != nil {
		return err
	}

	if err := commands.Login(cfg); err != nil {
		return err
	}

	color.Green("✓ Successfully authenticated with OIDC")
	return nil
}

func validateOIDCConfig(cfg *config.Config) error {
	missing := []string{}
	if cfg.DomainName == "" {
		missing = append(missing, "OS_DOMAIN_NAME / --domain-name")
	}
	if cfg.IdpURL == "" {
		missing = append(missing, "IDP_URL / --idp-url")
	}
	if cfg.IdpClientID == "" {
		missing = append(missing, "IDP_CLIENT_ID / --idp-client-id")
	}
	if cfg.IDPProviderName == "" {
		missing = append(missing, "IDP_PROVIDER_NAME / --idp-provider")
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required configuration: %v", missing)
	}
	return nil
}