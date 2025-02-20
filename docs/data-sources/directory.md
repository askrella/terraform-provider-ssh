---
page_title: "ssh_directory_info Data Source - SSH Provider"
subcategory: ""
description: |-
  Reads information about a directory on a remote server via SSH.
---

# ssh_directory_info (Data Source)

Reads information about a directory on a remote server via SSH. This data source can read directory permissions, ownership, attributes, and list its contents.

## Example Usage

```hcl
data "ssh_directory_info" "example" {
  ssh = {
    host        = "example.com"
    port        = 22
    username    = "user"
    password    = "your-password"
    # private_key = file("~/.ssh/id_rsa")
  }

  path = "/path/to/directory"
}

output "directory_permissions" {
  value = data.ssh_directory_info.example.permissions
}

output "directory_entries" {
  value = data.ssh_directory_info.example.entries
}
```

## Argument Reference

The following arguments are supported:

* `ssh` - (Required) SSH connection configuration block. See [SSH Block Configuration](../index.md#ssh-block-configuration) for details.
* `path` - (Required) The path of the directory to read on the remote server.

## Attribute Reference

The following attributes are exported:

* `permissions` - The directory permissions in octal format (e.g., '0755').
* `owner` - The user owner of the directory.
* `group` - The group owner of the directory.
* `immutable` - Whether the directory cannot be modified/deleted/renamed.
* `append_only` - Whether the directory can only be opened in append mode for writing.
* `no_dump` - Whether the directory is not included in backups.
* `synchronous` - Whether changes are written synchronously to disk.
* `no_atime` - Whether access time is not updated.
* `compressed` - Whether the directory is compressed.
* `no_cow` - Whether copy-on-write is disabled.
* `undeletable` - Whether content is saved when deleted.
* `exists` - Whether the directory exists.
* `entries` - A list of files and directories in this directory. Each entry contains:
  * `name` - The name of the file or directory.
  * `path` - The full path of the file or directory.
  * `size` - The size of the file in bytes.
  * `is_dir` - Whether this entry is a directory.
  * `permissions` - The permissions in octal format.
  * `owner` - The user owner of the entry.
  * `group` - The group owner of the entry.
  * `immutable` - Whether the entry cannot be modified/deleted/renamed.
  * `append_only` - Whether the entry can only be opened in append mode for writing.
  * `no_dump` - Whether the entry is not included in backups.
  * `synchronous` - Whether changes are written synchronously to disk.
  * `no_atime` - Whether access time is not updated.
  * `compressed` - Whether the entry is compressed.
  * `no_cow` - Whether copy-on-write is disabled.
  * `undeletable` - Whether content is saved when deleted.
  * `mod_time` - The last modification time in RFC3339 format. 