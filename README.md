# OTC CLI

A command-line interface tool for Open Telekom Cloud (OTC) that provides seamless authentication and resource management with federated Single Sign-On (SSO) support.

## Features

- 🔐 **Federated Authentication** - Login via OIDC/Keycloak SSO or IAM credentials
- ⚡ **Temporary Credentials** - Generate 24-hour AK/SK credentials
- 🔄 **Token Caching** - Automatic token management and refresh
- 📦 **Resource Management** - List and manage ECS, VPC, volumes, databases, and more
- ☸️ **CCE Integration** - Get kubeconfig for Kubernetes clusters
- 💰 **Pricing API** - Real-time pricing for 90+ OTC services with OS filtering
- 📊 **Multiple Outputs** - Table, JSON, and CSV export formats
- 🌐 **Multi-Project Support** - Work with multiple OTC projects
- 🖥️ **Remote Console** - Browser-based console access to ECS instances
- 🔍 **Advanced Filtering** - Filter by status, AZ, tags, OS type, and more

## Installation

### From Source

```bash
git clone https://github.com/abdo-farag/otc-cli.git
cd otc-cli
go build -o otc-cli cmd/otc-cli/main.go
sudo mv otc-cli /usr/local/bin/
```

### Using Go Install

```bash
go install github.com/abdo-farag/otc-cli@latest
```

## Quick Start

### Option 1: Login with Federated SSO (OIDC/Keycloak)

1. **Configure environment variables:**

```bash
# OIDC/Keycloak Configuration
export IDP_URL="https://your-keycloak.com/realms/YourRealm"
export IDP_CLIENT_ID="otc-client"
export IDP_CLIENT_SECRET="your-client-secret"  # Optional, for confidential clients
export IDP_PROVIDER_NAME="YourSSO"

# OTC Configuration
export OS_DOMAIN_NAME="OTC00000000001000001234"
export OS_REGION="eu-de"
```

2. **Login:**

```bash
otc-cli login
```

This will open your browser for SSO authentication. After successful login, credentials are saved to `otc-credentials.sh`.

3. **Load credentials:**

```bash
source otc-credentials.sh
```

### Option 2: Login with IAM Credentials

1. **Configure environment variables:**

```bash
export OS_USERNAME="your-username"
export OS_PASSWORD="your-password"
export OS_DOMAIN_NAME="OTC00000000001000001234"
export OS_REGION="eu-de"
```

2. **Login:**

```bash
otc-cli login --iam
```

3. **Load credentials:**

```bash
source otc-credentials.sh
```

## Basic Usage

### List Resources

```bash
# List all projects
otc-cli list projects

# List servers/instances
otc-cli list servers

# List servers in specific project
otc-cli list servers -p "Production"

# List VPCs
otc-cli list vpcs

# List subnets
otc-cli list subnets

# List volumes
otc-cli list volumes

# List CCE clusters
otc-cli list cce
```

### Get Kubeconfig

```bash
# Get kubeconfig for CCE cluster
otc-cli get kubeconfig -c cluster-name

# Save to specific file
otc-cli get kubeconfig -c cluster-id -o ~/.kube/otc-config
```

### Logout

Clear cached credentials:

```bash
otc-cli logout
```

## Configuration

### Keycloak OIDC Client Setup

If using federated SSO, configure your Keycloak client:

1. **Create OIDC Client** in Keycloak Admin Console
2. **Client Settings:**
   - Client Protocol: `openid-connect`
   - Access Type: `public` or `confidential`
   - Valid Redirect URIs: `http://localhost:9197/oidc/auth`
   - Web Origins: `http://localhost:9197`
3. **Required Scopes:** `openid`, `email`, `profile`, `roles`, `groups`, `offline_access`

### Environment Variables

| Variable | Description | Required | Default |
|----------|-------------|----------|---------|
| `IDP_URL` | Keycloak/OIDC provider URL | For SSO | - |
| `IDP_CLIENT_ID` | OIDC client ID | For SSO | - |
| `IDP_CLIENT_SECRET` | OIDC client secret | No | - |
| `IDP_PROVIDER_NAME` | Identity provider name in OTC | For SSO | - |
| `OS_USERNAME` | IAM username | For IAM | - |
| `OS_PASSWORD` | IAM password | For IAM | - |
| `OS_DOMAIN_NAME` | OTC domain name | Yes | - |
| `OS_REGION` | OTC region | No | `eu-de` |

## Documentation

For advanced usage, detailed examples, and troubleshooting, see the [complete documentation](otc-cli.md).

Topics covered in the advanced docs:
- Working with multiple projects
- JSON output and scripting
- Integration with AWS CLI, Terraform, and boto3
- Detailed troubleshooting guide
- Authentication flow diagrams
- CI/CD integration

## Quick Troubleshooting

### "Could not find OIDC configuration" error

- Verify `IDP_PROVIDER_NAME` matches the Identity Provider name in OTC IAM
- Check that federated identity is properly configured in OTC Console

### "Project not found" error

List available projects first:
```bash
otc-cli list projects
```

### Browser doesn't open

Use no-browser mode:
```bash
otc-cli login --no-browser
```

For more detailed troubleshooting, see [otc-cli.md](otc-cli.md).

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

MIT License - see [LICENSE](LICENSE) file for details

## Links

- **Documentation:** [otc-cli.md](otc-cli.md)
- **Issues:** [GitHub Issues](https://github.com/abdo-farag/otc-cli/issues)
- **Repository:** [github.com/abdo-farag/otc-cli](https://github.com/abdo-farag/otc-cli)

## Related Documentation

- **Keycloak Documentation:** [https://www.keycloak.org/documentation](https://www.keycloak.org/documentation)
  - [OIDC Client Configuration](https://www.keycloak.org/docs/latest/server_admin/#_oidc_clients)
  - [Identity Brokering](https://www.keycloak.org/docs/latest/server_admin/#_identity_broker)
- **Open Telekom Cloud Documentation:** [https://docs.otc.t-systems.com/](https://docs.otc.t-systems.com/)
  - [Identity and Access Management (IAM)](https://docs.otc.t-systems.com/identity-access-management/umn/service_overview/what_is_iam.html)
  - [IAM Federated Identity Authentication](https://docs.otc.t-systems.com/identity-access-management/umn/user_guide/federated_identity_authentication/index.html)
  - [API Reference](https://docs.otc.t-systems.com/api/api-ref.html)
- **Gopher Telekom Cloud SDK:** [https://github.com/opentelekomcloud/gophertelekomcloud](https://github.com/opentelekomcloud/gophertelekomcloud)
  - Go SDK for Open Telekom Cloud (this project uses it internally)
  
---

**Note:** This is an unofficial tool and is not supported by T-Systems or Deutsche Telekom.