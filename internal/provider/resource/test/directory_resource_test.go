package test

import (
	"context"
	"crypto/rand"
	"fmt"
	"os"
	"testing"

	"github.com/askrella/askrella-ssh-provider/internal/provider/ssh"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/require"
)

func TestAccDirectoryResource(t *testing.T) {
	t.Parallel()

	// Setup SSH client for verification
	sshConfig := ssh.SSHConfig{
		Host:     "localhost",
		Port:     2222,
		Username: "testuser",
		Password: "testpass",
	}

	client, err := ssh.NewSSHClient(context.Background(), sshConfig)
	require.NoError(t, err)
	defer client.Close()

	dirName := "testdir_" + rand.Text()
	testDirPath := "/home/testuser/" + dirName

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccDirectoryResourceConfig(dirName, "0755", "testuser", "testuser"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ssh_directory.test", "path", testDirPath),
					resource.TestCheckResourceAttr("ssh_directory.test", "permissions", "0755"),
					resource.TestCheckResourceAttr("ssh_directory.test", "owner", "testuser"),
					resource.TestCheckResourceAttr("ssh_directory.test", "group", "testuser"),
					resource.TestCheckResourceAttr("ssh_directory.test", "ssh.host", "localhost"),
					resource.TestCheckResourceAttr("ssh_directory.test", "ssh.port", "2222"),
					resource.TestCheckResourceAttr("ssh_directory.test", "ssh.username", "testuser"),
					func(s *terraform.State) error {
						// Verify directory exists
						exists, err := client.DirectoryExists(context.Background(), testDirPath)
						if err != nil {
							return fmt.Errorf("failed to check directory: %v", err)
						}
						if !exists {
							return fmt.Errorf("directory does not exist: %s", testDirPath)
						}

						// Verify permissions
						mode, err := client.GetFileMode(context.Background(), testDirPath)
						if err != nil {
							return fmt.Errorf("failed to get directory permissions: %v", err)
						}
						if mode != os.FileMode(0755) {
							return fmt.Errorf("unexpected permissions for creation: got %o, want 0755", mode)
						}

						// Verify ownership
						ownership, err := client.GetFileOwnership(context.Background(), testDirPath)
						if err != nil {
							return fmt.Errorf("failed to get directory ownership: %v", err)
						}
						if ownership.User != "testuser" {
							return fmt.Errorf("unexpected owner: got %s, want testuser", ownership.User)
						}
						if ownership.Group != "testuser" {
							return fmt.Errorf("unexpected group: got %s, want testuser", ownership.Group)
						}

						return nil
					},
				),
			},
			// Update testing
			{
				Config: testAccDirectoryResourceConfig(dirName, "0775", "testuser", "testuser"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ssh_directory.test", "path", testDirPath),
					resource.TestCheckResourceAttr("ssh_directory.test", "permissions", "0775"),
					resource.TestCheckResourceAttr("ssh_directory.test", "owner", "testuser"),
					resource.TestCheckResourceAttr("ssh_directory.test", "group", "testuser"),
					resource.TestCheckResourceAttr("ssh_directory.test", "ssh.host", "localhost"),
					resource.TestCheckResourceAttr("ssh_directory.test", "ssh.port", "2222"),
					resource.TestCheckResourceAttr("ssh_directory.test", "ssh.username", "testuser"),
					func(s *terraform.State) error {
						// Verify directory exists
						exists, err := client.DirectoryExists(context.Background(), testDirPath)
						if err != nil {
							return fmt.Errorf("failed to check directory: %v", err)
						}
						if !exists {
							return fmt.Errorf("directory does not exist: %s", testDirPath)
						}

						// Verify updated permissions
						mode, err := client.GetFileMode(context.Background(), testDirPath)
						if err != nil {
							return fmt.Errorf("failed to get directory permissions: %v", err)
						}
						if mode != os.FileMode(0775) {
							return fmt.Errorf("unexpected permissions for updating: got %o, want 0775", mode)
						}

						// Verify ownership
						ownership, err := client.GetFileOwnership(context.Background(), testDirPath)
						if err != nil {
							return fmt.Errorf("failed to get directory ownership: %v", err)
						}
						if ownership.User != "testuser" {
							return fmt.Errorf("unexpected owner: got %s, want testuser", ownership.User)
						}
						if ownership.Group != "testuser" {
							return fmt.Errorf("unexpected group: got %s, want testuser", ownership.Group)
						}

						return nil
					},
				),
			},
		},
	})
}

func testAccDirectoryResourceConfig(name string, permissions string, owner string, group string) string {
	return fmt.Sprintf(`
resource "ssh_directory" "test" {
  ssh = {
    host        = "localhost"
    port        = 2222
    username    = "testuser"
    password    = "testpass"
  }
  path        = "/home/testuser/%s"
  permissions = "%s"
  owner       = "%s"
  group       = "%s"
}
`, name, permissions, owner, group)
}
