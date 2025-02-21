package test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/askrella/askrella-ssh-provider/internal/provider/ssh"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
)

func TestAccFileDataSource(t *testing.T) {
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
	testContent := "Hello, World!"

	// Create test file
	err = client.CreateFile(context.Background(), testFilePath, testContent, 0644)
	require.NoError(t, err)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccFileDataSourceConfig(testFilePath),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.ssh_file_info.test", "path", testFilePath),
					resource.TestCheckResourceAttr("data.ssh_file_info.test", "content", testContent),
					resource.TestCheckResourceAttr("data.ssh_file_info.test", "permissions", "0644"),
					resource.TestCheckResourceAttr("data.ssh_file_info.test", "exists", "true"),
					resource.TestCheckResourceAttr("data.ssh_file_info.test", "ssh.host", "localhost"),
					resource.TestCheckResourceAttr("data.ssh_file_info.test", "ssh.port", "2222"),
					resource.TestCheckResourceAttr("data.ssh_file_info.test", "ssh.username", "testuser"),
				),
			},
			// Test non-existent file
			{
				Config: testAccFileDataSourceConfig("/home/testuser/nonexistent.txt"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.ssh_file_info.test", "exists", "false"),
				),
			},
		},
	})

	err = client.DeleteFile(context.Background(), testFilePath)
	if err != nil && !os.IsNotExist(err) {
		// Only log if it's not a "file not exist" error
		t.Logf("Failed to cleanup test file: %v", err)
	}
}

func testAccFileDataSourceConfig(path string) string {
	return fmt.Sprintf(`
data "ssh_file_info" "test" {
  ssh = {
    host        = "localhost"
    port        = 2222
    username    = "testuser"
    password    = "testpass"
  }
  path = %q
}
`, path)
}
