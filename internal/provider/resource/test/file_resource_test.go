package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/askrella/askrella-ssh-provider/internal/provider/ssh"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/require"
)

func TestAccFileResource(t *testing.T) {
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

	testFilePath := "/home/testuser/test.txt"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccFileResourceConfig("test.txt", "Hello, World!", "0644", "testuser", "testuser"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ssh_file.test", "path", testFilePath),
					resource.TestCheckResourceAttr("ssh_file.test", "content", "Hello, World!"),
					resource.TestCheckResourceAttr("ssh_file.test", "permissions", "0644"),
					resource.TestCheckResourceAttr("ssh_file.test", "owner", "testuser"),
					resource.TestCheckResourceAttr("ssh_file.test", "group", "testuser"),
					resource.TestCheckResourceAttr("ssh_file.test", "ssh.host", "localhost"),
					resource.TestCheckResourceAttr("ssh_file.test", "ssh.port", "2222"),
					resource.TestCheckResourceAttr("ssh_file.test", "ssh.username", "testuser"),
					func(s *terraform.State) error {
						// Verify file exists and has correct content
						content, err := client.ReadFile(context.Background(), testFilePath)
						if err != nil {
							return fmt.Errorf("failed to read file: %v", err)
						}
						if content != "Hello, World!" {
							return fmt.Errorf("unexpected content: got %s, want Hello, World!", content)
						}

						// Verify ownership
						ownership, err := client.GetFileOwnership(context.Background(), testFilePath)
						if err != nil {
							return fmt.Errorf("failed to get file ownership: %v", err)
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
				Config: testAccFileResourceConfig("test.txt", "Updated content", "0644", "testuser", "testuser"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ssh_file.test", "path", testFilePath),
					resource.TestCheckResourceAttr("ssh_file.test", "content", "Updated content"),
					resource.TestCheckResourceAttr("ssh_file.test", "permissions", "0644"),
					resource.TestCheckResourceAttr("ssh_file.test", "owner", "testuser"),
					resource.TestCheckResourceAttr("ssh_file.test", "group", "testuser"),
					resource.TestCheckResourceAttr("ssh_file.test", "ssh.host", "localhost"),
					resource.TestCheckResourceAttr("ssh_file.test", "ssh.port", "2222"),
					resource.TestCheckResourceAttr("ssh_file.test", "ssh.username", "testuser"),
					func(s *terraform.State) error {
						// Verify file has updated content
						content, err := client.ReadFile(context.Background(), testFilePath)
						if err != nil {
							return fmt.Errorf("failed to read file: %v", err)
						}
						if content != "Updated content" {
							return fmt.Errorf("unexpected content: got %s, want Updated content", content)
						}

						// Verify ownership
						ownership, err := client.GetFileOwnership(context.Background(), testFilePath)
						if err != nil {
							return fmt.Errorf("failed to get file ownership: %v", err)
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

func testAccFileResourceConfig(name string, content string, permissions string, owner string, group string) string {
	return fmt.Sprintf(`
resource "ssh_file" "test" {
  ssh = {
    host        = "localhost"
    port        = 2222
    username    = "testuser"
    password    = "testpass"
  }
  path        = "/home/testuser/%s"
  content     = %q
  permissions = "%s"
  owner       = "%s"
  group       = "%s"
}
`, name, content, permissions, owner, group)
}
