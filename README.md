# Askrella SSH Provider for Terraform

This Terraform provider enables you to manage files and directories on remote servers via SSH. It provides resources to create, read, update, and delete files and directories on remote systems using SSH authentication.

## Features

- Create, read, update, and delete files on remote servers
- Create and manage directories on remote servers
- Secure SSH authentication using various methods (password, private key)
- Support for file permissions and ownership
- Telemetry integration using OpenTelemetry

## Requirements

- Go 1.22 or higher
- Terraform 1.0 or higher
- MacOS, Linux, BSD or other *nix platforms where the file system is capable of understand file permissions (unlike Windows)

## Building The Provider

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
      source = "askrella/ssh"
    }
  }
}

provider "askrella-ssh" {
  # Provider configuration if needed
}

# Create a file on a remote server
resource "ssh_file" "example" {
  host     = "example.com"
  port     = 22
  username = "user"
  
  # Authentication (choose one)
  password    = "your-password"  # or
  private_key = file("~/.ssh/id_rsa")
  
  # File configuration
  path        = "/path/to/file"
  content     = "Hello, World!"
  permissions = "0644"
}

# Create a directory on a remote server
resource "ssh_directory" "example" {
  host     = "example.com"
  port     = 22
  username = "user"
  
  # Authentication (choose one)
  password    = "your-password"  # or
  private_key = file("~/.ssh/id_rsa")
  
  # Directory configuration
  path        = "/path/to/directory"
  permissions = "0755"
}
```

## Development

### Requirements

- [Go](https://golang.org/doc/install) 1.22 or higher
- [Terraform](https://www.terraform.io/downloads.html) 1.0 or higher

### Building

```bash
go build
```

### Testing

```bash
go test -v ./...
```

### Running with Local Provider

```bash
terraform {
  required_providers {
    ssh = {
      source = "askrella/ssh"
    }
  }
}
```

## Contributing

Contributions are welcome! Please read our contributing guidelines and submit pull requests.

## License

This project is licensed under the MIT License - see the LICENSE file for details. 