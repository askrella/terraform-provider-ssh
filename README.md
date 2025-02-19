# Askrella SSH Provider for Terraform

This Terraform provider enables you to manage files and directories on remote servers via SSH. It provides resources to create, read, update, and delete files and directories on remote systems using SSH authentication.

## Features

- Create, read, update, and delete files on remote servers
- Create and manage directories on remote servers
- Secure SSH authentication using various methods (password, private key)
- Support for file permissions and ownership
- Telemetry integration using OpenTelemetry
- Robust error handling and logging
- Comprehensive test coverage

## Requirements

- Go 1.22 or higher
- Terraform 1.0 or higher
- MacOS, Linux, BSD or other *nix platforms where the file system is capable of understanding file permissions (unlike Windows)

## Installation

### From Terraform Registry

```hcl
terraform {
  required_providers {
    ssh = {
      source  = "askrella/ssh"
      version = "0.1.0"
    }
  }
}
```

### Building From Source

1. Clone the repository
```bash
git clone https://github.com/askrella/askrella-ssh-provider.git
```

2. Enter the repository directory
```bash
cd askrella-ssh-provider
```

3. Build the provider
```bash
go build -o terraform-provider-askrella-ssh
```

## Using the Provider

```hcl
terraform {
  required_providers {
    ssh = {
      source  = "askrella/ssh"
      version = "0.1.0"
    }
  }
}

provider "ssh" {}

locals {
  ssh_config = {
    host        = "example.com"
    port        = 22
    username    = "user"
    password    = "your-password"  # or use private_key
    # private_key = file("~/.ssh/id_rsa")
  }
}

# Create a directory on a remote server
resource "ssh_directory" "example_dir" {
  ssh         = local.ssh_config
  path        = "/path/to/directory"
  permissions = "0755"
}

# Create a file in the directory
resource "ssh_file" "example_file" {
  ssh         = local.ssh_config
  path        = "${ssh_directory.example_dir.path}/example.txt"
  content     = "Hello, World!"
  permissions = "0644"

  depends_on = [ssh_directory.example_dir]
}
```

## Development

### Requirements

- [Go](https://golang.org/doc/install) 1.22 or higher
- [Terraform](https://www.terraform.io/downloads.html) 1.0 or higher
- [golangci-lint](https://golangci-lint.run/usage/install/) for code linting

### Building

```bash
go build
```

### Testing

The project uses Go's testing framework with Gomega matchers. Tests are designed to run in parallel and are isolated from each other.

```bash
# Run all tests
go test -v ./...

# Run tests with race detection
go test -race -v ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Telemetry

This provider integrates with OpenTelemetry for distributed tracing and monitoring. Traces are automatically created for provider operations and can be exported to your preferred observability backend.

### Code Quality

The project maintains high code quality standards through:

- Comprehensive test coverage
- Static code analysis using golangci-lint
- Mutation testing
- Continuous Integration via GitHub Actions
- Automated dependency updates via Dependabot

### Local Development

To use a locally built version of the provider:

1. Build the provider as described above
2. Configure your Terraform project to use the local provider:

```hcl
terraform {
  required_providers {
    ssh = {
      source = "askrella/ssh"
    }
  }
}
```

## Contributing

Contributions are welcome! Please follow these steps:

1. Fork the repository
2. Create a feature branch
3. Commit your changes with descriptive commit messages
4. Add tests for any new functionality
5. Create a pull request

Please ensure your code:
- Passes all tests
- Includes appropriate documentation
- Follows the project's code style
- Includes relevant test cases

## License

This project is licensed under the MIT License - see the LICENSE file for details. 