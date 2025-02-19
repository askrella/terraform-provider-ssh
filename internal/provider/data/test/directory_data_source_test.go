package test

import (
	"context"
	"crypto/rand"
	"fmt"
	"testing"

	"github.com/askrella/askrella-ssh-provider/internal/provider/ssh"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
)

func TestAccDirectoryDataSource(t *testing.T) {
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

	testDirPath := "/home/testuser/testdir_" + rand.Text()
	testFilePath := testDirPath + "/test.txt"
	testContent := "Hello, World!"

	// Create test directory and file
	err = client.CreateDirectory(context.Background(), testDirPath, 0755)
	require.NoError(t, err)
	err = client.CreateFile(context.Background(), testFilePath, testContent, 0644)
	require.NoError(t, err)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccDirectoryDataSourceConfig(testDirPath),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.ssh_directory_info.test", "path", testDirPath),
					resource.TestCheckResourceAttr("data.ssh_directory_info.test", "permissions", "0755"),
					resource.TestCheckResourceAttr("data.ssh_directory_info.test", "exists", "true"),
					resource.TestCheckResourceAttr("data.ssh_directory_info.test", "ssh.host", "localhost"),
					resource.TestCheckResourceAttr("data.ssh_directory_info.test", "ssh.port", "2222"),
					resource.TestCheckResourceAttr("data.ssh_directory_info.test", "ssh.username", "testuser"),
					// Check that we have one entry (test.txt)
					resource.TestCheckResourceAttr("data.ssh_directory_info.test", "entries.#", "1"),
					resource.TestCheckResourceAttr("data.ssh_directory_info.test", "entries.0.name", "test.txt"),
					resource.TestCheckResourceAttr("data.ssh_directory_info.test", "entries.0.path", testFilePath),
					resource.TestCheckResourceAttr("data.ssh_directory_info.test", "entries.0.permissions", "0644"),
					resource.TestCheckResourceAttr("data.ssh_directory_info.test", "entries.0.is_dir", "false"),
				),
			},
			// Test non-existent directory
			{
				Config: testAccDirectoryDataSourceConfig("/home/testuser/nonexistent"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.ssh_directory_info.test", "exists", "false"),
				),
			},
		},
	})
}

func testAccDirectoryDataSourceConfig(path string) string {
	return fmt.Sprintf(`
data "ssh_directory_info" "test" {
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
