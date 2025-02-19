package data

import (
	"context"
	"fmt"
	"github.com/askrella/askrella-ssh-provider/internal/provider/ssh"
	"os"
	"path/filepath"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.opentelemetry.io/otel"
)

var (
	_ datasource.DataSource              = &DirectoryDataSource{}
	_ datasource.DataSourceWithConfigure = &DirectoryDataSource{}
)

// DirectoryDataSource defines the data source implementation.
type DirectoryDataSource struct {
	pool *ssh.SSHPool
}

// DirectoryEntry represents a file or directory entry
type DirectoryEntry struct {
	Name        types.String `tfsdk:"name"`
	Path        types.String `tfsdk:"path"`
	Size        types.Int64  `tfsdk:"size"`
	IsDir       types.Bool   `tfsdk:"is_dir"`
	Permissions types.String `tfsdk:"permissions"`
	Owner       types.String `tfsdk:"owner"`
	Group       types.String `tfsdk:"group"`
	Immutable   types.Bool   `tfsdk:"immutable"`
	AppendOnly  types.Bool   `tfsdk:"append_only"`
	NoDump      types.Bool   `tfsdk:"no_dump"`
	Synchronous types.Bool   `tfsdk:"synchronous"`
	NoAtime     types.Bool   `tfsdk:"no_atime"`
	Compressed  types.Bool   `tfsdk:"compressed"`
	NoCoW       types.Bool   `tfsdk:"no_cow"`
	Undeletable types.Bool   `tfsdk:"undeletable"`
	ModTime     types.String `tfsdk:"mod_time"`
}

// DirectoryDataSourceModel describes the data source data model.
type DirectoryDataSourceModel struct {
	SSH         *ssh.SSHBlockModel `tfsdk:"ssh"`
	Path        types.String       `tfsdk:"path"`
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
	Entries     []DirectoryEntry   `tfsdk:"entries"`
	ID          types.String       `tfsdk:"id"`
}

// NewDirectoryDataSource creates a new data source implementation.
func NewDirectoryDataSource(pool *ssh.SSHPool) datasource.DataSource {
	return &DirectoryDataSource{
		pool: pool,
	}
}

// Metadata returns the data source type name.
func (d *DirectoryDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_directory_info"
}

// Schema defines the schema for the data source.
func (d *DirectoryDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads information about a directory on a remote server via SSH.",
		Attributes: map[string]schema.Attribute{
			"ssh": schema.SingleNestedAttribute{
				Description: "SSH connection configuration.",
				Required:    true,
				Attributes:  ssh.SSHBlockDataSourceSchema(),
			},
			"path": schema.StringAttribute{
				Description: "The path of the directory on the remote server.",
				Required:    true,
			},
			"permissions": schema.StringAttribute{
				Description: "The directory permissions in octal format (e.g., '0755').",
				Computed:    true,
			},
			"owner": schema.StringAttribute{
				Description: "The user owner of the directory.",
				Computed:    true,
			},
			"group": schema.StringAttribute{
				Description: "The group owner of the directory.",
				Computed:    true,
			},
			"immutable": schema.BoolAttribute{
				Description: "Whether the directory cannot be modified/deleted/renamed.",
				Computed:    true,
			},
			"append_only": schema.BoolAttribute{
				Description: "Whether the directory can only be opened in append mode for writing.",
				Computed:    true,
			},
			"no_dump": schema.BoolAttribute{
				Description: "Whether the directory is not included in backups.",
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
				Description: "Whether the directory is compressed.",
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
				Description: "Whether the directory exists.",
				Computed:    true,
			},
			"entries": schema.ListNestedAttribute{
				Description: "List of files and directories in this directory.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description: "The name of the file or directory.",
							Computed:    true,
						},
						"path": schema.StringAttribute{
							Description: "The full path of the file or directory.",
							Computed:    true,
						},
						"size": schema.Int64Attribute{
							Description: "The size of the file in bytes.",
							Computed:    true,
						},
						"is_dir": schema.BoolAttribute{
							Description: "Whether this entry is a directory.",
							Computed:    true,
						},
						"permissions": schema.StringAttribute{
							Description: "The permissions in octal format.",
							Computed:    true,
						},
						"owner": schema.StringAttribute{
							Description: "The user owner of the entry.",
							Computed:    true,
						},
						"group": schema.StringAttribute{
							Description: "The group owner of the entry.",
							Computed:    true,
						},
						"immutable": schema.BoolAttribute{
							Description: "Whether the entry cannot be modified/deleted/renamed.",
							Computed:    true,
						},
						"append_only": schema.BoolAttribute{
							Description: "Whether the entry can only be opened in append mode for writing.",
							Computed:    true,
						},
						"no_dump": schema.BoolAttribute{
							Description: "Whether the entry is not included in backups.",
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
							Description: "Whether the entry is compressed.",
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
						"mod_time": schema.StringAttribute{
							Description: "The last modification time in RFC3339 format.",
							Computed:    true,
						},
					},
				},
			},
			"id": schema.StringAttribute{
				Description: "The path of the directory.",
				Computed:    true,
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *DirectoryDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	ctx, span := otel.Tracer("ssh-provider").Start(ctx, "DirectoryDataSource.Read")
	defer span.End()

	var state DirectoryDataSourceModel
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

	// Check if directory exists
	dirInfo, err := client.SftpClient.Stat(state.Path.ValueString())
	if err != nil {
		if os.IsNotExist(err) {
			state.Exists = types.BoolValue(false)
			state.ID = types.StringValue(state.Path.ValueString())
			diags = resp.State.Set(ctx, &state)
			resp.Diagnostics.Append(diags...)
			return
		}
		resp.Diagnostics.AddError(
			"Error reading directory information",
			fmt.Sprintf("Could not read directory information: %s", err),
		)
		return
	}

	if !dirInfo.IsDir() {
		resp.Diagnostics.AddError(
			"Path is not a directory",
			fmt.Sprintf("The path %s exists but is not a directory", state.Path.ValueString()),
		)
		return
	}

	state.Exists = types.BoolValue(true)
	state.ID = types.StringValue(state.Path.ValueString())

	// Get directory permissions
	mode := dirInfo.Mode().Perm()
	state.Permissions = types.StringValue(fmt.Sprintf("%04o", mode))

	// Get directory ownership
	ownership, err := client.GetFileOwnership(ctx, state.Path.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading directory ownership",
			fmt.Sprintf("Could not read directory ownership: %s", err),
		)
		return
	}
	state.Owner = types.StringValue(ownership.User)
	state.Group = types.StringValue(ownership.Group)

	// Get directory attributes
	attrs, err := client.GetFileAttributes(ctx, state.Path.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading directory attributes",
			fmt.Sprintf("Could not read directory attributes: %s", err),
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

	// Read directory entries
	entries, err := client.SftpClient.ReadDir(state.Path.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading directory entries",
			fmt.Sprintf("Could not read directory entries: %s", err),
		)
		return
	}

	// Convert entries to model
	state.Entries = make([]DirectoryEntry, 0, len(entries))
	for _, entry := range entries {
		entryPath := filepath.Join(state.Path.ValueString(), entry.Name())
		ownership, err := client.GetFileOwnership(ctx, entryPath)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error reading entry ownership",
				fmt.Sprintf("Could not read ownership for %s: %s", entryPath, err),
			)
			return
		}

		attrs, err := client.GetFileAttributes(ctx, entryPath)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error reading entry attributes",
				fmt.Sprintf("Could not read attributes for %s: %s", entryPath, err),
			)
			return
		}

		state.Entries = append(state.Entries, DirectoryEntry{
			Name:        types.StringValue(entry.Name()),
			Path:        types.StringValue(entryPath),
			Size:        types.Int64Value(entry.Size()),
			IsDir:       types.BoolValue(entry.IsDir()),
			Permissions: types.StringValue(fmt.Sprintf("%04o", entry.Mode().Perm())),
			Owner:       types.StringValue(ownership.User),
			Group:       types.StringValue(ownership.Group),
			Immutable:   types.BoolValue(attrs.Immutable),
			AppendOnly:  types.BoolValue(attrs.AppendOnly),
			NoDump:      types.BoolValue(attrs.NoDump),
			Synchronous: types.BoolValue(attrs.Synchronous),
			NoAtime:     types.BoolValue(attrs.NoAtime),
			Compressed:  types.BoolValue(attrs.Compressed),
			NoCoW:       types.BoolValue(attrs.NoCoW),
			Undeletable: types.BoolValue(attrs.Undeletable),
			ModTime:     types.StringValue(entry.ModTime().Format(time.RFC3339)),
		})
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (d *DirectoryDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
}

func (d *DirectoryDataSource) getClient(ctx context.Context, sshBlock *ssh.SSHBlockModel) (*ssh.SSHClient, error) {
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
