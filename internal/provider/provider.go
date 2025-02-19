package provider

import (
	"context"

	"github.com/askrella/askrella-ssh-provider/internal/provider/data"
	resource2 "github.com/askrella/askrella-ssh-provider/internal/provider/resource"
	"github.com/askrella/askrella-ssh-provider/internal/provider/ssh"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/sirupsen/logrus"
)

// Ensure the implementation satisfies the expected interfaces
var (
	_ provider.Provider = &SSHProvider{}
)

// SSHProvider is the provider implementation.
type SSHProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
	pool    *ssh.SSHPool
}

// New creates a new provider instance
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &SSHProvider{
			version: version,
		}
	}
}

// Metadata returns the provider type name.
func (p *SSHProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "ssh"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data.
func (p *SSHProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{}
}

// Configure prepares a HashiCups API client for data sources and resources.
func (p *SSHProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// Initialize the SSH connection pool
	p.pool = ssh.NewSSHPool(ssh.PoolConfig{
		Logger: logrus.New(),
	})
}

// DataSources defines the data sources implemented in the provider.
func (p *SSHProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		func() datasource.DataSource {
			return data.NewFileDataSource(p.pool)
		},
		func() datasource.DataSource {
			return data.NewDirectoryDataSource(p.pool)
		},
	}
}

// Resources defines the resources implemented in the provider.
func (p *SSHProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		func() resource.Resource {
			return resource2.NewFileResource(p.pool)
		},
		func() resource.Resource {
			return resource2.NewDirectoryResource(p.pool)
		},
	}
}

// Close closes the provider's resources
func (p *SSHProvider) Close(ctx context.Context) error {
	if p.pool != nil {
		p.pool.Close()
	}
	return nil
}
