---
page_title: "ssh_file_info Data Source - SSH Provider"
subcategory: ""
description: |-
  Reads information about a file on a remote server via SSH.
---

# ssh_file_info (Data Source)

Reads information about a file on a remote server via SSH. This data source can read file content, permissions, ownership, and various file attributes.

## Example Usage

```hcl
data "ssh_file_info" "example" {
  ssh = {
    host        = "example.com"
    port        = 22
    username    = "user"
    password    = "your-password"
    # private_key = file("~/.ssh/id_rsa")
  }

  path = "/path/to/file.txt"
}

output "file_content" {
  value = data.ssh_file_info.example.content
}

output "file_permissions" {
  value = data.ssh_file_info.example.permissions
}
```

## Argument Reference

The following arguments are supported:

* `ssh` - (Required) SSH connection configuration block. See [SSH Block Configuration](../index.md#ssh-block-configuration) for details.
* `path` - (Required) The path of the file to read on the remote server.

## Attribute Reference

The following attributes are exported:

* `content` - The content of the file.
* `permissions` - The file permissions in octal format (e.g., '0644').
* `owner` - The user owner of the file.
* `group` - The group owner of the file.
* `immutable` - Whether the file cannot be modified/deleted/renamed.
* `append_only` - Whether the file can only be opened in append mode for writing.
* `no_dump` - Whether the file is not included in backups.
* `synchronous` - Whether changes are written synchronously to disk.
* `no_atime` - Whether access time is not updated.
* `compressed` - Whether the file is compressed.
* `no_cow` - Whether copy-on-write is disabled.
* `undeletable` - Whether content is saved when deleted.
* `exists` - Whether the file exists. 