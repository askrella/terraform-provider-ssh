package ssh

import (
	"context"
	"crypto/rand"
	"os"
	"path"
	"testing"

	. "github.com/onsi/gomega"
)

var sshConfig = SSHConfig{
	Host:     "localhost",
	Port:     2222,
	Username: "testuser",
	Password: "testpass",
}

func TestFilePermissions(t *testing.T) {
	RegisterTestingT(t)

	client, err := NewSSHClient(context.Background(), sshConfig)
	Expect(err).ToNot(HaveOccurred())
	ctx := context.Background()
	basePath := "/home/testuser/ssh_test_" + rand.Text()

	testCases := []struct {
		name        string
		filePath    string
		content     string
		permissions os.FileMode
	}{
		{
			name:        "Test File Permissions 0777",
			filePath:    basePath + "_1",
			content:     "Hello World",
			permissions: 0777,
		},
		{
			name:        "Test File Permissions 0644",
			filePath:    basePath + "_2",
			content:     "Hello World",
			permissions: 0644,
		},
		{
			name:        "Test File Permissions 0600",
			filePath:    basePath + "_3",
			content:     "Hello World",
			permissions: 0600,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			RegisterTestingT(t)

			Expect(client.CreateFile(ctx, tc.filePath, tc.content, tc.permissions)).Should(Succeed())
			Expect(client.GetFileMode(ctx, tc.filePath)).To(BeEquivalentTo(tc.permissions))
		})
	}
}

func TestDirectoryOperations(t *testing.T) {
	RegisterTestingT(t)

	client, err := NewSSHClient(context.Background(), sshConfig)
	Expect(err).ToNot(HaveOccurred())

	basePath := "/home/testuser/ssh_test_" + rand.Text()

	directoryPath := path.Join(basePath, "dir/simple")

	t.Log("Check directory does not exist before we've done anything")
	exists, err := client.Exists(context.Background(), directoryPath)
	Expect(err).ToNot(HaveOccurred())
	Expect(exists).To(BeFalse())

	t.Log("Create directory")
	err = client.CreateDirectory(context.Background(), directoryPath, os.FileMode(0777))
	Expect(err).ToNot(HaveOccurred())

	t.Log("Check directory exists after creation")
	exists, err = client.Exists(context.Background(), directoryPath)
	Expect(err).ToNot(HaveOccurred())
	Expect(exists).To(BeTrue())

	t.Log("Creating directory twice should fail")
	err = client.CreateDirectory(context.Background(), directoryPath, os.FileMode(0777))
	Expect(err).To(HaveOccurred())

	t.Log("Delete directory")
	err = client.DeleteDirectory(context.Background(), directoryPath)
	Expect(err).ToNot(HaveOccurred())

	t.Log("Check directory exists after deletion")
	exists, err = client.Exists(context.Background(), directoryPath)
	Expect(err).ToNot(HaveOccurred())
	Expect(exists).To(BeFalse())
}
