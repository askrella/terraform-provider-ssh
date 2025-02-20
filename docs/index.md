---
page_title: "SSH Provider"
subcategory: ""
description: |-
  The SSH provider enables management of files and directories on remote servers via SSH.
---

# SSH Provider

The SSH provider enables you to manage files and directories on remote servers via SSH. It provides resources to create, read, update, and delete files and directories on remote systems using SSH authentication.

## Example Usage

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

## Authentication

The provider supports two methods of authentication:
- Password authentication
- Private key authentication

## Provider Configuration

The provider itself requires no configuration. All SSH connection details are specified in the individual resources and data sources.

### SSH Block Configuration

The `ssh` block is required in all resources and data sources and accepts the following arguments:

* `host` - (Required) The hostname or IP address of the remote server.
* `port` - (Optional) The SSH port of the remote server. Defaults to 22.
* `username` - (Required) The username to use for SSH authentication.
* `password` - (Optional) The password to use for SSH authentication.
* `private_key` - (Optional) The private key to use for SSH authentication.

-> **Note:** Either `password` or `private_key` must be specified.
