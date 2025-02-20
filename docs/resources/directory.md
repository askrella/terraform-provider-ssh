---
page_title: "ssh_directory Resource - SSH Provider"
subcategory: ""
description: |-
  Manages a directory on a remote server via SSH.
---

# ssh_directory (Resource)

Manages a directory on a remote server via SSH. This resource can create, update, and delete directories, as well as manage their permissions and attributes.

## Example Usage

```hcl
resource "ssh_directory" "example" {
  ssh = {
    host        = "example.com"
    port        = 22
    username    = "user"
    password    = "your-password"
    # private_key = file("~/.ssh/id_rsa")
  }

  path        = "/path/to/directory"
  permissions = "0755"
  owner       = "user"
  group       = "group"
}
```

## Argument Reference

The following arguments are supported:

* `ssh` - (Required) SSH connection configuration block. See [SSH Block Configuration](../index.md#ssh-block-configuration) for details.
* `path` - (Required) The path where the directory should be created on the remote server. **Note:** Changing this value forces a new resource to be created.
* `permissions` - (Optional) The directory permissions in octal format (e.g., '0755').
* `owner` - (Optional) The user owner of the directory.
* `group` - (Optional) The group owner of the directory.
* `immutable` - (Optional) If true, the directory cannot be modified/deleted/renamed.
* `append_only` - (Optional) If true, the directory can only be opened in append mode for writing.
* `no_dump` - (Optional) If true, the directory is not included in backups.
* `synchronous` - (Optional) If true, changes are written synchronously to disk.
* `no_atime` - (Optional) If true, access time is not updated.
* `compressed` - (Optional) If true, the directory is compressed.
* `no_cow` - (Optional) If true, copy-on-write is disabled.
* `undeletable` - (Optional) If true, content is saved when deleted.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The path of the directory.

## Import

Directories can be imported using their path. For example:

```shell
terraform import ssh_directory.example /path/to/directory
``` 