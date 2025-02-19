package data

import (
	"context"
	"fmt"
	"github.com/askrella/askrella-ssh-provider/internal/provider/ssh"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.opentelemetry.io/otel"
)

var (
	_ datasource.DataSource              = &FileDataSource{}
	_ datasource.DataSourceWithConfigure = &FileDataSource{}
)

// FileDataSource defines the data source implementation.
type FileDataSource struct {
	pool *ssh.SSHPool
}

// FileDataSourceModel describes the data source data model.
type FileDataSourceModel struct {
	SSH         *ssh.SSHBlockModel `tfsdk:"ssh"`
	Path        types.String       `tfsdk:"path"`
	Content     types.String       `tfsdk:"content"`
	Permissions types.String       `tfsdk:"permissions"`
	Owner       types.String       `tfsdk:"owner"`
	Group       types.String       `tfsdk:"group"`
	Immutable   types.Bool         `tfsdk:"immutable"`
	AppendOnly  types.Bool         `tfsdk:"append_only"`
	NoDump      types.Bool         `tfsdk:"no_dump"`
	Synchronous types.Bool         `tfsdk:"synchronous"`
	NoAtime     types.Bool         `tfsdk:"no_atime"`
	Compressed  types.Bool         `tfsdk:"compressed"`
	NoCoW       types.Bool         `tfsdk:"no_cow"`
	Undeletable types.Bool         `tfsdk:"undeletable"`
	Exists      types.Bool         `tfsdk:"exists"`
	ID          types.String       `tfsdk:"id"`
}

// NewFileDataSource creates a new data source implementation.
func NewFileDataSource(pool *ssh.SSHPool) datasource.DataSource {
	return &FileDataSource{
		pool: pool,
	}
}

// Metadata returns the data source type name.
func (d *FileDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_file_info"
}

// Schema defines the schema for the data source.
func (d *FileDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads information about a file on a remote server via SSH.",
		Attributes: map[string]schema.Attribute{
			"ssh": schema.SingleNestedAttribute{
				Description: "SSH connection configuration.",
				Required:    true,
				Attributes:  ssh.SSHBlockDataSourceSchema(),
			},
			"path": schema.StringAttribute{
				Description: "The path of the file on the remote server.",
				Required:    true,
			},
			"content": schema.StringAttribute{
				Description: "The content of the file.",
				Computed:    true,
			},
			"permissions": schema.StringAttribute{
				Description: "The file permissions in octal format (e.g., '0644').",
				Computed:    true,
			},
			"owner": schema.StringAttribute{
				Description: "The user owner of the file.",
				Computed:    true,
			},
			"group": schema.StringAttribute{
				Description: "The group owner of the file.",
				Computed:    true,
			},
			"immutable": schema.BoolAttribute{
				Description: "Whether the file cannot be modified/deleted/renamed.",
				Computed:    true,
			},
			"append_only": schema.BoolAttribute{
				Description: "Whether the file can only be opened in append mode for writing.",
				Computed:    true,
			},
			"no_dump": schema.BoolAttribute{
				Description: "Whether the file is not included in backups.",
				Computed:    true,
			},
			"synchronous": schema.BoolAttribute{
				Description: "Whether changes are written synchronously to disk.",
				Computed:    true,
			},
			"no_atime": schema.BoolAttribute{
				Description: "Whether access time is not updated.",
				Computed:    true,
			},
			"compressed": schema.BoolAttribute{
				Description: "Whether the file is compressed.",
				Computed:    true,
			},
			"no_cow": schema.BoolAttribute{
				Description: "Whether copy-on-write is disabled.",
				Computed:    true,
			},
			"undeletable": schema.BoolAttribute{
				Description: "Whether content is saved when deleted.",
				Computed:    true,
			},
			"exists": schema.BoolAttribute{
				Description: "Whether the file exists.",
				Computed:    true,
			},
			"id": schema.StringAttribute{
				Description: "The path of the file.",
				Computed:    true,
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *FileDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	ctx, span := otel.Tracer("ssh-provider").Start(ctx, "FileDataSource.Read")
	defer span.End()

	var state FileDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := d.getClient(ctx, state.SSH)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating SSH client",
			fmt.Sprintf("Could not create SSH client: %s", err),
		)
		return
	}
	defer client.Close()

	// Check if file exists
	fileInfo, err := client.SftpClient.Stat(state.Path.ValueString())
	if err != nil {
		if os.IsNotExist(err) {
			state.Exists = types.BoolValue(false)
			state.ID = types.StringValue(state.Path.ValueString())
			diags = resp.State.Set(ctx, &state)
			resp.Diagnostics.Append(diags...)
			return
		}
		resp.Diagnostics.AddError(
			"Error reading file information",
			fmt.Sprintf("Could not read file information: %s", err),
		)
		return
	}

	state.Exists = types.BoolValue(true)
	state.ID = types.StringValue(state.Path.ValueString())

	// Get file permissions
	mode := fileInfo.Mode().Perm()
	state.Permissions = types.StringValue(fmt.Sprintf("%04o", mode))

	// Get file ownership
	ownership, err := client.GetFileOwnership(ctx, state.Path.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading file ownership",
			fmt.Sprintf("Could not read file ownership: %s", err),
		)
		return
	}
	state.Owner = types.StringValue(ownership.User)
	state.Group = types.StringValue(ownership.Group)

	// Get file attributes
	attrs, err := client.GetFileAttributes(ctx, state.Path.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading file attributes",
			fmt.Sprintf("Could not read file attributes: %s", err),
		)
		return
	}
	state.Immutable = types.BoolValue(attrs.Immutable)
	state.AppendOnly = types.BoolValue(attrs.AppendOnly)
	state.NoDump = types.BoolValue(attrs.NoDump)
	state.Synchronous = types.BoolValue(attrs.Synchronous)
	state.NoAtime = types.BoolValue(attrs.NoAtime)
	state.Compressed = types.BoolValue(attrs.Compressed)
	state.NoCoW = types.BoolValue(attrs.NoCoW)
	state.Undeletable = types.BoolValue(attrs.Undeletable)

	// Read file content
	content, err := client.ReadFile(ctx, state.Path.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading file content",
			fmt.Sprintf("Could not read file content: %s", err),
		)
		return
	}
	state.Content = types.StringValue(content)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (d *FileDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
}

func (d *FileDataSource) getClient(ctx context.Context, sshBlock *ssh.SSHBlockModel) (*ssh.SSHClient, error) {
	port := int(sshBlock.Port.ValueInt64())
	if port == 0 {
		port = 22
	}

	config := ssh.SSHConfig{
		Host:       sshBlock.Host.ValueString(),
		Port:       port,
		Username:   sshBlock.Username.ValueString(),
		Password:   sshBlock.Password.ValueString(),
		PrivateKey: sshBlock.PrivateKey.ValueString(),
	}

	client, err := d.pool.GetClient(ctx, config)
	if err != nil {
		return nil, err
	}

	// Release the client when the context is done
	go func() {
		<-ctx.Done()
		d.pool.ReleaseClient(config)
	}()

	return client, nil
}
