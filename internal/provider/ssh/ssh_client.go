package ssh

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pkg/sftp"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"golang.org/x/crypto/ssh"
)

// SSHClient represents a client for SSH operations
type SSHClient struct {
	sshClient  *ssh.Client
	SftpClient *sftp.Client
	logger     *logrus.Logger
}

// SSHConfig holds the configuration for SSH connections
type SSHConfig struct {
	Host       string
	Port       int
	Username   string
	Password   string
	PrivateKey string
}

// FileOwnership holds the user and group ownership of a file or directory
type FileOwnership struct {
	User  string
	Group string
}

// FileAttributes represents the attributes of a file or directory
type FileAttributes struct {
	Immutable   bool // 'i' attribute - cannot be modified/deleted/renamed
	AppendOnly  bool // 'a' attribute - can only be opened in append mode for writing
	NoDump      bool // 'd' attribute - not dumped in backups
	Synchronous bool // 'S' attribute - changes are written synchronously to disk
	NoAtime     bool // 'A' attribute - no atime updates
	Compressed  bool // 'c' attribute - compressed
	NoCoW       bool // 'C' attribute - no copy-on-write
	Undeletable bool // 'u' attribute - content saved when deleted
}

// NewSSHClient creates a new SSH client with the given configuration
func NewSSHClient(ctx context.Context, config SSHConfig) (*SSHClient, error) {
	ctx, span := otel.Tracer("ssh-provider").Start(ctx, "NewSSHClient")
	defer span.End()

	logger := logrus.New()

	var authMethods []ssh.AuthMethod

	if config.Password != "" {
		authMethods = append(authMethods, ssh.Password(config.Password))
	}

	if config.PrivateKey != "" {
		signer, err := ssh.ParsePrivateKey([]byte(config.PrivateKey))
		if err != nil {
			logger.WithContext(ctx).WithError(err).Error("Failed to parse private key")
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}

	if len(authMethods) == 0 {
		return nil, fmt.Errorf("no authentication method provided")
	}

	sshConfig := &ssh.ClientConfig{
		User:            config.Username,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // TODO: Allow configuring host key verification
	}

	host := config.Host
	isIpv6 := net.ParseIP(config.Host).To16() != nil
	if isIpv6 {
		host = fmt.Sprintf("[%s]", config.Host)
	}

	host += ":" + strconv.Itoa(config.Port)

	client, err := ssh.Dial("tcp", host, sshConfig)
	if err != nil {
		logger.WithContext(ctx).WithError(err).Error("Failed to connect to SSH server")
		return nil, fmt.Errorf("failed to connect to SSH server: %w", err)
	}

	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		logger.WithContext(ctx).WithError(err).Error("Failed to create SFTP client")
		client.Close()
		return nil, fmt.Errorf("failed to create SFTP client: %w", err)
	}

	return &SSHClient{
		sshClient:  client,
		SftpClient: sftpClient,
		logger:     logger,
	}, nil
}

// Close closes the SSH and SFTP connections
func (c *SSHClient) Close() error {
	if c.SftpClient != nil {
		if err := c.SftpClient.Close(); err != nil {
			return fmt.Errorf("error closing SFTP client: %w", err)
		}
	}
	if c.sshClient != nil {
		if err := c.sshClient.Close(); err != nil {
			return fmt.Errorf("error closing SSH client: %w", err)
		}
	}
	return nil
}

// CreateFile creates a file with the given content and permissions
func (c *SSHClient) CreateFile(ctx context.Context, path string, content string, permissions os.FileMode) error {
	ctx, span := otel.Tracer("ssh-provider").Start(ctx, "CreateFile")
	defer span.End()

	// Ensure parent directory exists
	parentDir := filepath.Dir(path)
	if exists, _ := c.Exists(ctx, parentDir); !exists {
		if err := c.CreateDirectory(ctx, parentDir, 0755); err != nil {
			return fmt.Errorf("failed to create parent directory: %w", err)
		}
	}

	file, err := c.SftpClient.Create(path)
	if err != nil {
		c.logger.WithContext(ctx).WithError(err).Error("Failed to create file")
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	if _, err := file.Write([]byte(content)); err != nil {
		c.logger.WithContext(ctx).WithError(err).Error("Failed to write file content")
		return fmt.Errorf("failed to write file content: %w", err)
	}

	if err := c.SftpClient.Chmod(path, permissions); err != nil {
		c.logger.WithContext(ctx).WithError(err).Error("Failed to set file permissions")
		return fmt.Errorf("failed to set file permissions: %w", err)
	}

	return nil
}

// ReadFile reads the content of a file
func (c *SSHClient) ReadFile(ctx context.Context, path string) (string, error) {
	ctx, span := otel.Tracer("ssh-provider").Start(ctx, "ReadFile")
	defer span.End()

	file, err := c.SftpClient.Open(path)
	if err != nil {
		c.logger.WithContext(ctx).WithError(err).Error("Failed to open file")
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		c.logger.WithContext(ctx).WithError(err).Error("Failed to read file content")
		return "", fmt.Errorf("failed to read file content: %w", err)
	}

	return string(content), nil
}

// DeleteFile deletes a file
func (c *SSHClient) DeleteFile(ctx context.Context, path string) error {
	ctx, span := otel.Tracer("ssh-provider").Start(ctx, "DeleteFile")
	defer span.End()

	if err := c.SftpClient.Remove(path); err != nil {
		c.logger.WithContext(ctx).WithError(err).Error("Failed to delete file")
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// CreateDirectory creates a directory with the given permissions
func (c *SSHClient) CreateDirectory(ctx context.Context, path string, permissions os.FileMode) error {
	ctx, span := otel.Tracer("ssh-provider").Start(ctx, "CreateDirectory")
	defer span.End()

	if exists, _ := c.Exists(ctx, path); exists {
		return fmt.Errorf("directory %s already exists", path)
	}

	if err := c.SftpClient.MkdirAll(path); err != nil {
		c.logger.WithContext(ctx).WithError(err).Error("Failed to create directory")
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := c.SftpClient.Chmod(path, permissions); err != nil {
		c.logger.WithContext(ctx).WithError(err).Error("Failed to set directory permissions")
		return fmt.Errorf("failed to set directory permissions: %w", err)
	}

	return nil
}

// DeleteDirectory deletes a directory
func (c *SSHClient) DeleteDirectory(ctx context.Context, path string) error {
	ctx, span := otel.Tracer("ssh-provider").Start(ctx, "DeleteDirectory")
	defer span.End()

	if err := c.SftpClient.RemoveAll(path); err != nil {
		c.logger.WithContext(ctx).WithError(err).Error("Failed to delete directory")
		return fmt.Errorf("failed to delete directory: %w", err)
	}

	return nil
}

// Exists checks if a directory or file exists
func (c *SSHClient) Exists(ctx context.Context, path string) (bool, error) {
	ctx, span := otel.Tracer("ssh-provider").Start(ctx, "Exists")
	defer span.End()

	_, err := c.SftpClient.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		c.logger.WithContext(ctx).WithError(err).Error("Failed to check existence")
		return false, fmt.Errorf("failed to check existence: %w", err)
	}

	return true, nil
}

// GetFileMode gets the permissions of a file or directory
func (c *SSHClient) GetFileMode(ctx context.Context, path string) (os.FileMode, error) {
	ctx, span := otel.Tracer("ssh-provider").Start(ctx, "GetFileMode")
	defer span.End()

	info, err := c.SftpClient.Stat(path)
	if err != nil {
		c.logger.WithContext(ctx).WithError(err).Error("Failed to get file mode")
		return 0, fmt.Errorf("failed to get file mode: %w", err)
	}

	return info.Mode().Perm(), nil
}

// GetFileOwnership gets the user and group ownership of a file or directory
func (c *SSHClient) GetFileOwnership(ctx context.Context, path string) (*FileOwnership, error) {
	ctx, span := otel.Tracer("ssh-provider").Start(ctx, "GetFileOwnership")
	defer span.End()

	// Run ls -ln to get numeric user/group IDs
	session, err := c.sshClient.NewSession()
	if err != nil {
		c.logger.WithContext(ctx).WithError(err).Error("Failed to create SSH session")
		return nil, fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer session.Close()

	output, err := session.Output(fmt.Sprintf("ls -ldn %q", path))
	if err != nil {
		c.logger.WithContext(ctx).WithError(err).Error("Failed to get file ownership")
		return nil, fmt.Errorf("failed to get file ownership: %w", err)
	}

	// Parse ls output (format: "-rw-r--r-- 1 1000 1000 0 Feb 19 13:23 /path/to/file")
	fields := strings.Fields(string(output))
	if len(fields) < 4 {
		c.logger.WithContext(ctx).WithError(err).Error("Invalid ls output format")
		return nil, fmt.Errorf("invalid ls output format: %s", string(output))
	}
	uid := fields[2]
	gid := fields[3]

	// Get user name from uid
	session, err = c.sshClient.NewSession()
	if err != nil {
		c.logger.WithContext(ctx).WithError(err).Error("Failed to create SSH session")
		return nil, fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer session.Close()

	userName, err := session.Output(fmt.Sprintf("getent passwd %s | cut -d: -f1", uid))
	if err != nil {
		c.logger.WithContext(ctx).WithError(err).Error("Failed to get username")
		return nil, fmt.Errorf("failed to get username: %w", err)
	}

	// Get group name from gid
	session, err = c.sshClient.NewSession()
	if err != nil {
		c.logger.WithContext(ctx).WithError(err).Error("Failed to create SSH session")
		return nil, fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer session.Close()

	groupName, err := session.Output(fmt.Sprintf("getent group %s | cut -d: -f1", gid))
	if err != nil {
		c.logger.WithContext(ctx).WithError(err).Error("Failed to get group name")
		return nil, fmt.Errorf("failed to get group name: %w", err)
	}

	return &FileOwnership{
		User:  strings.TrimSpace(string(userName)),
		Group: strings.TrimSpace(string(groupName)),
	}, nil
}

// SetFileOwnership sets the user and group ownership of a file or directory
func (c *SSHClient) SetFileOwnership(ctx context.Context, path string, ownership *FileOwnership) error {
	ctx, span := otel.Tracer("ssh-provider").Start(ctx, "SetFileOwnership")
	defer span.End()

	if ownership == nil {
		return nil
	}

	// Skip if both user and group are empty
	if ownership.User == "" && ownership.Group == "" {
		return nil
	}

	session, err := c.sshClient.NewSession()
	if err != nil {
		c.logger.WithContext(ctx).WithError(err).Error("Failed to create SSH session")
		return fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer session.Close()

	// Build chown command
	var cmd string
	switch {
	case ownership.User != "" && ownership.Group != "":
		cmd = fmt.Sprintf("chown %s:%s %q", ownership.User, ownership.Group, path)
	case ownership.User != "":
		// Get current group if only user is specified
		currentOwnership, err := c.GetFileOwnership(ctx, path)
		if err != nil {
			return fmt.Errorf("failed to get current ownership: %w", err)
		}
		cmd = fmt.Sprintf("chown %s:%s %q", ownership.User, currentOwnership.Group, path)
	case ownership.Group != "":
		// Get current user if only group is specified
		currentOwnership, err := c.GetFileOwnership(ctx, path)
		if err != nil {
			return fmt.Errorf("failed to get current ownership: %w", err)
		}
		cmd = fmt.Sprintf("chown %s:%s %q", currentOwnership.User, ownership.Group, path)
	default:
		return nil
	}

	err = session.Run(cmd)
	if err != nil {
		c.logger.WithContext(ctx).WithError(err).Error("Failed to set file ownership")
		return fmt.Errorf("failed to set file ownership: %w", err)
	}

	return nil
}

// GetFileAttributes gets the attributes of a file or directory
func (c *SSHClient) GetFileAttributes(ctx context.Context, path string) (*FileAttributes, error) {
	ctx, span := otel.Tracer("ssh-provider").Start(ctx, "GetFileAttributes")
	defer span.End()

	session, err := c.sshClient.NewSession()
	if err != nil {
		c.logger.WithContext(ctx).WithError(err).Error("Failed to create SSH session")
		return nil, fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer session.Close()

	output, err := session.Output(fmt.Sprintf("lsattr -d %q", path))
	if err != nil {
		c.logger.WithContext(ctx).WithError(err).Error("Failed to get file attributes")
		return nil, fmt.Errorf("failed to get file attributes: %w", err)
	}

	// Parse lsattr output (format: "----i-A------- /path/to/file")
	attrs := &FileAttributes{}
	if len(output) >= 16 {
		attrString := string(output[:16])
		attrs.Immutable = strings.Contains(attrString, "i")
		attrs.AppendOnly = strings.Contains(attrString, "a")
		attrs.NoDump = strings.Contains(attrString, "d")
		attrs.Synchronous = strings.Contains(attrString, "S")
		attrs.NoAtime = strings.Contains(attrString, "A")
		attrs.Compressed = strings.Contains(attrString, "c")
		attrs.NoCoW = strings.Contains(attrString, "C")
		attrs.Undeletable = strings.Contains(attrString, "u")
	}

	return attrs, nil
}

// SetFileAttributes sets the attributes of a file or directory
func (c *SSHClient) SetFileAttributes(ctx context.Context, path string, attrs *FileAttributes) error {
	ctx, span := otel.Tracer("ssh-provider").Start(ctx, "SetFileAttributes")
	defer span.End()

	if attrs == nil {
		return nil
	}

	// Build attribute string
	var addAttrs, removeAttrs []string

	// Map attributes to their flags
	type attrFlag struct {
		flag string
		set  *bool
	}

	attrFlags := []attrFlag{
		{flag: "i", set: &attrs.Immutable},
		{flag: "a", set: &attrs.AppendOnly},
		{flag: "d", set: &attrs.NoDump},
		{flag: "S", set: &attrs.Synchronous},
		{flag: "A", set: &attrs.NoAtime},
		{flag: "c", set: &attrs.Compressed},
		{flag: "C", set: &attrs.NoCoW},
		{flag: "u", set: &attrs.Undeletable},
	}

	// Get current attributes to determine what needs to change
	currentAttrs, err := c.GetFileAttributes(ctx, path)
	if err != nil {
		return err
	}

	currentAttrMap := map[string]bool{
		"i": currentAttrs.Immutable,
		"a": currentAttrs.AppendOnly,
		"d": currentAttrs.NoDump,
		"S": currentAttrs.Synchronous,
		"A": currentAttrs.NoAtime,
		"c": currentAttrs.Compressed,
		"C": currentAttrs.NoCoW,
		"u": currentAttrs.Undeletable,
	}

	// Determine which attributes need to be added or removed
	for _, attr := range attrFlags {
		if *attr.set && !currentAttrMap[attr.flag] {
			addAttrs = append(addAttrs, attr.flag)
		} else if !*attr.set && currentAttrMap[attr.flag] {
			removeAttrs = append(removeAttrs, attr.flag)
		}
	}

	// Apply changes if needed
	if len(addAttrs) > 0 {
		session, err := c.sshClient.NewSession()
		if err != nil {
			c.logger.WithContext(ctx).WithError(err).Error("Failed to create SSH session")
			return fmt.Errorf("failed to create SSH session: %w", err)
		}
		defer session.Close()

		cmd := fmt.Sprintf("chattr +%s %q", strings.Join(addAttrs, ""), path)
		if err := session.Run(cmd); err != nil {
			c.logger.WithContext(ctx).WithError(err).Error("Failed to add file attributes")
			return fmt.Errorf("failed to add file attributes: %w", err)
		}
	}

	if len(removeAttrs) > 0 {
		session, err := c.sshClient.NewSession()
		if err != nil {
			c.logger.WithContext(ctx).WithError(err).Error("Failed to create SSH session")
			return fmt.Errorf("failed to create SSH session: %w", err)
		}
		defer session.Close()

		cmd := fmt.Sprintf("chattr -%s %q", strings.Join(removeAttrs, ""), path)
		if err := session.Run(cmd); err != nil {
			c.logger.WithContext(ctx).WithError(err).Error("Failed to remove file attributes")
			return fmt.Errorf("failed to remove file attributes: %w", err)
		}
	}

	return nil
}
