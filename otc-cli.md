# Otc-Cli Manual

A command-line tool for managing Open Telekom Cloud (OTC) resources.
Supports authentication, resource management, and automation.

```text
otc-cli [command] [global flags] [command flags]
```

### Global Flags

```text
      --json             Output raw JSON response (alias)
  -p, --project string   Project ID or name
      --raw              Output raw JSON response
```

### Commands

* [otc-cli completion](#otc-cli-completion)
* [otc-cli completion bash](#otc-cli-completion-bash)
* [otc-cli completion fish](#otc-cli-completion-fish)
* [otc-cli completion help](#otc-cli-completion-help)
* [otc-cli completion powershell](#otc-cli-completion-powershell)
* [otc-cli completion zsh](#otc-cli-completion-zsh)
* [otc-cli docs](#otc-cli-docs)
* [otc-cli help](#otc-cli-help)
* [otc-cli list](#otc-cli-list)
* [otc-cli list flavor](#otc-cli-list-flavor)
* [otc-cli list help](#otc-cli-list-help)
* [otc-cli list image](#otc-cli-list-image)
* [otc-cli list projects](#otc-cli-list-projects)
* [otc-cli login](#otc-cli-login)
* [otc-cli logout](#otc-cli-logout)
* [otc-cli version](#otc-cli-version)

# Commands

## `otc-cli completion`

Generate the autocompletion script for otc-cli for the specified shell.
See each sub-command's help for details on how to use the generated script.


```text
otc-cli completion [flags]
```

### Command Flags

```text
  -h, --help   help for completion
```

## `otc-cli completion bash`

Generate the autocompletion script for the bash shell.

This script depends on the 'bash-completion' package.
If it is not installed already, you can install it via your OS's package manager.

To load completions in your current shell session:

	source <(otc-cli completion bash)

To load completions for every new session, execute once:

#### Linux:

	otc-cli completion bash > /etc/bash_completion.d/otc-cli

#### macOS:

	otc-cli completion bash > $(brew --prefix)/etc/bash_completion.d/otc-cli

You will need to start a new shell for this setup to take effect.


```text
otc-cli completion bash
```

### Command Flags

```text
  -h, --help              help for bash
      --no-descriptions   disable completion descriptions
```

## `otc-cli completion fish`

Generate the autocompletion script for the fish shell.

To load completions in your current shell session:

	otc-cli completion fish | source

To load completions for every new session, execute once:

	otc-cli completion fish > ~/.config/fish/completions/otc-cli.fish

You will need to start a new shell for this setup to take effect.


```text
otc-cli completion fish [flags]
```

### Command Flags

```text
  -h, --help              help for fish
      --no-descriptions   disable completion descriptions
```

## `otc-cli completion help`

Help provides help for any command in the application.
Simply type completion help [path to command] for full details.

```text
otc-cli completion help [command] [flags]
```

### Command Flags

```text
  -h, --help   help for help
```

## `otc-cli completion powershell`

Generate the autocompletion script for powershell.

To load completions in your current shell session:

	otc-cli completion powershell | Out-String | Invoke-Expression

To load completions for every new session, add the output of the above command
to your powershell profile.


```text
otc-cli completion powershell [flags]
```

### Command Flags

```text
  -h, --help              help for powershell
      --no-descriptions   disable completion descriptions
```

## `otc-cli completion zsh`

Generate the autocompletion script for the zsh shell.

If shell completion is not already enabled in your environment you will need
to enable it.  You can execute the following once:

	echo "autoload -U compinit; compinit" >> ~/.zshrc

To load completions in your current shell session:

	source <(otc-cli completion zsh)

To load completions for every new session, execute once:

#### Linux:

	otc-cli completion zsh > "${fpath[1]}/_otc-cli"

#### macOS:

	otc-cli completion zsh > $(brew --prefix)/share/zsh/site-functions/_otc-cli

You will need to start a new shell for this setup to take effect.


```text
otc-cli completion zsh [flags]
```

### Command Flags

```text
  -h, --help              help for zsh
      --no-descriptions   disable completion descriptions
```

## `otc-cli docs`

Generate comprehensive markdown documentation for all CLI commands.

```text
otc-cli docs [flags]
```

### Command Flags

```text
  -h, --help            help for docs
  -o, --output string   Output file path (default "otc-cli.md")
```

## `otc-cli help`

Help provides help for any command in the application.
Simply type otc-cli help [path to command] for full details.

```text
otc-cli help [command] [flags]
```

### Command Flags

```text
  -h, --help   help for help
```

## `otc-cli list`

List OTC resources such as servers, VPCs, volumes, and more.

```text
otc-cli list [flags]
```

### Command Flags

```text
  -h, --help   help for list
```

## `otc-cli list flavor`

List server flavors with pricing

```text
otc-cli list flavor [flags]
```

### Command Flags

```text
  -h, --help        help for flavor
  -o, --os string   OS type for pricing (openlinux, redhat, oracle, windows) (default "openlinux")
```

## `otc-cli list help`

Help provides help for any command in the application.
Simply type list help [path to command] for full details.

```text
otc-cli list help [command] [flags]
```

### Command Flags

```text
  -h, --help   help for help
```

## `otc-cli list image`

List system and custom images

```text
otc-cli list image [flags]
```

### Command Flags

```text
  -h, --help                help for image
      --name string         Filter by image name (partial match)
      --platform string     Filter by platform (Ubuntu, CentOS, Windows, etc.)
      --status string       Filter by status (active, queued, etc.)
      --visibility string   Filter by visibility (private, public, shared)
```

## `otc-cli list projects`

List OTC projects

```text
otc-cli list projects [flags]
```

### Command Flags

```text
  -h, --help   help for projects
```

## `otc-cli login`

Authenticate with OTC using either OIDC federation or IAM direct authentication.

OIDC Authentication (default):
  Uses OAuth2/OIDC flow with your identity provider.
  
IAM Authentication (--iam flag):
  Uses direct username/password authentication with OTC IAM.

```text
otc-cli login [flags]
```

### Command Flags

```text
      --auth-url string                OTC IAM endpoint
      --code-challenge-method string   PKCE method (S256 or plain) (default "S256")
      --domain-name string             OTC domain name
  -h, --help                           help for login
      --iam                            Use IAM direct authentication
      --idp-client-id string           IDP client ID
      --idp-provider string            IDP provider name
      --idp-url string                 IDP URL
      --no-browser                     Don't open browser automatically
      --output string                  Output file
      --password string                IAM password
      --port int                       Callback port (default 9197)
      --region string                  Region
      --scope string                   OIDC scopes (default "openid email profile roles groups organization")
      --username string                IAM username
```

## `otc-cli logout`

Remove the cached authentication token from local storage.

```text
otc-cli logout [flags]
```

### Command Flags

```text
  -h, --help   help for logout
```

## `otc-cli version`

Print the version number

```text
otc-cli version [flags]
```

### Command Flags

```text
  -h, --help   help for version
```
