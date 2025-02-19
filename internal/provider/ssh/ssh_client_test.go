package ssh

import (
	"context"
	"crypto/rand"
	. "github.com/onsi/gomega"
	"os"
	"path"
	"testing"
)

func TestDirectoryOperations(t *testing.T) {
	RegisterTestingT(t)

	client, err := NewSSHClient(context.Background(), SSHConfig{
		Host:     "localhost",
		Port:     2222,
		Username: "testuser",
		Password: "testpass",
	})
	Expect(err).ToNot(HaveOccurred())

	basePath := "/home/testuser/ssh_test_" + rand.Text()

	directoryPath := path.Join(basePath, "dir/simple")

	t.Log("Check directory does not exist before we've done anything")
	exists, err := client.DirectoryExists(context.Background(), directoryPath)
	Expect(err).ToNot(HaveOccurred())
	Expect(exists).To(BeFalse())

	t.Log("Create directory")
	err = client.CreateDirectory(context.Background(), directoryPath, os.FileMode(0777))
	Expect(err).ToNot(HaveOccurred())

	t.Log("Check directory exists after creation")
	exists, err = client.DirectoryExists(context.Background(), directoryPath)
	Expect(err).ToNot(HaveOccurred())
	Expect(exists).To(BeTrue())

	t.Log("Creating directory twice should fail")
	err = client.CreateDirectory(context.Background(), directoryPath, os.FileMode(0777))
	Expect(err).To(HaveOccurred())

	t.Log("Delete directory")
	err = client.DeleteDirectory(context.Background(), directoryPath)
	Expect(err).ToNot(HaveOccurred())

	t.Log("Check directory exists after deletion")
	exists, err = client.DirectoryExists(context.Background(), directoryPath)
	Expect(err).ToNot(HaveOccurred())
	Expect(exists).To(BeFalse())
}
