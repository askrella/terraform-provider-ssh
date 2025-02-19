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
    host        = "localhost"
    port        = 2222
    username    = "testuser"
    password    = "testpass"  # or use private_key
  }
}

# Create a directory
resource "ssh_directory" "example_dir" {
  ssh = local.ssh_config
  path        = "/home/testuser/your/directory"
  permissions = "0755"
}

# Create a file in the directory
resource "ssh_file" "example_file" {
  ssh = local.ssh_config
  path        = "${ssh_directory.example_dir.path}/example.txt"
  content     = "Hello, World! Second take!"
  permissions = "0644"

  depends_on = [ssh_directory.example_dir]
} 