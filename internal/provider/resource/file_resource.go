package resource

import (
	"context"
	"fmt"
	"github.com/askrella/askrella-ssh-provider/internal/provider/ssh"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"go.opentelemetry.io/otel"
	"os"
)

var (
	_ resource.Resource              = &FileResource{}
	_ resource.ResourceWithConfigure = &FileResource{}
)

var _ = resource.Resource(&FileResource{})

// FileResource defines the resource implementation.
type FileResource struct {
	pool *ssh.SSHPool
}

// FileResourceModel describes the resource data model.
type FileResourceModel struct {
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
	ID          types.String       `tfsdk:"id"`
}

// NewFileResource creates a new resource implementation.
func NewFileResource(pool *ssh.SSHPool) resource.Resource {
	return &FileResource{
		pool: pool,
	}
}

// Metadata returns the resource type name.
func (r *FileResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_file"
}

// Schema defines the schema for the resource.
func (r *FileResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a file on a remote server via SSH.",
		Attributes: map[string]schema.Attribute{
			"ssh": schema.SingleNestedAttribute{
				Description: "SSH connection configuration.",
				Required:    true,
				Attributes:  ssh.SSHBlockSchema(),
			},
			"path": schema.StringAttribute{
				Description: "The path where the file should be created on the remote server.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"content": schema.StringAttribute{
				Description: "The content of the file.",
				Required:    true,
			},
			"permissions": schema.StringAttribute{
				Description: "The file permissions in octal format (e.g., '0644').",
				Optional:    true,
			},
			"owner": schema.StringAttribute{
				Description: "The user owner of the file.",
				Optional:    true,
			},
			"group": schema.StringAttribute{
				Description: "The group owner of the file.",
				Optional:    true,
			},
			"immutable": schema.BoolAttribute{
				Description: "If true, the file cannot be modified/deleted/renamed.",
				Optional:    true,
			},
			"append_only": schema.BoolAttribute{
				Description: "If true, the file can only be opened in append mode for writing.",
				Optional:    true,
			},
			"no_dump": schema.BoolAttribute{
				Description: "If true, the file is not included in backups.",
				Optional:    true,
			},
			"synchronous": schema.BoolAttribute{
				Description: "If true, changes are written synchronously to disk.",
				Optional:    true,
			},
			"no_atime": schema.BoolAttribute{
				Description: "If true, access time is not updated.",
				Optional:    true,
			},
			"compressed": schema.BoolAttribute{
				Description: "If true, the file is compressed.",
				Optional:    true,
			},
			"no_cow": schema.BoolAttribute{
				Description: "If true, copy-on-write is disabled.",
				Optional:    true,
			},
			"undeletable": schema.BoolAttribute{
				Description: "If true, content is saved when deleted.",
				Optional:    true,
			},
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *FileResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	ctx, span := otel.Tracer("ssh-provider").Start(ctx, "FileResource.Create")
	defer span.End()

	var plan FileResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.getClient(ctx, plan.SSH)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating SSH client",
			fmt.Sprintf("Could not create SSH client: %s", err),
		)
		return
	}
	defer client.Close()

	permissions := parsePermissions(plan.Permissions.ValueString())

	err = client.CreateFile(ctx, plan.Path.ValueString(), plan.Content.ValueString(), os.FileMode(permissions))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating file",
			fmt.Sprintf("Could not create file: %s", err),
		)
		return
	}

	// Set ownership if specified
	if !plan.Owner.IsNull() || !plan.Group.IsNull() {
		err = client.SetFileOwnership(ctx, plan.Path.ValueString(), &ssh.FileOwnership{
			User:  plan.Owner.ValueString(),
			Group: plan.Group.ValueString(),
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Error setting file ownership",
				fmt.Sprintf("Could not set file ownership: %s", err),
			)
			return
		}
	}

	// Set attributes if any are specified
	if !plan.Immutable.IsNull() || !plan.AppendOnly.IsNull() || !plan.NoDump.IsNull() ||
		!plan.Synchronous.IsNull() || !plan.NoAtime.IsNull() || !plan.Compressed.IsNull() ||
		!plan.NoCoW.IsNull() || !plan.Undeletable.IsNull() {
		err = client.SetFileAttributes(ctx, plan.Path.ValueString(), &ssh.FileAttributes{
			Immutable:   plan.Immutable.ValueBool(),
			AppendOnly:  plan.AppendOnly.ValueBool(),
			NoDump:      plan.NoDump.ValueBool(),
			Synchronous: plan.Synchronous.ValueBool(),
			NoAtime:     plan.NoAtime.ValueBool(),
			Compressed:  plan.Compressed.ValueBool(),
			NoCoW:       plan.NoCoW.ValueBool(),
			Undeletable: plan.Undeletable.ValueBool(),
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Error setting file attributes",
				fmt.Sprintf("Could not set file attributes: %s", err),
			)
			return
		}
	}

	plan.ID = basetypes.NewStringValue(plan.Path.ValueString())

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Read refreshes the Terraform state with the latest data.
func (r *FileResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	ctx, span := otel.Tracer("ssh-provider").Start(ctx, "FileResource.Read")
	defer span.End()

	var state FileResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.getClient(ctx, state.SSH)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating SSH client",
			fmt.Sprintf("Could not create SSH client: %s", err),
		)
		return
	}
	defer client.Close()

	content, err := client.ReadFile(ctx, state.Path.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading file",
			fmt.Sprintf("Could not read file: %s", err),
		)
		return
	}
	state.Content = basetypes.NewStringValue(content)

	// Get file mode
	mode, err := client.GetFileMode(ctx, state.Path.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading file mode",
			fmt.Sprintf("Could not read file mode: %s", err),
		)
		return
	}
	state.Permissions = basetypes.NewStringValue(fmt.Sprintf("%04o", mode))

	// Get ownership if it was specified
	if !state.Owner.IsNull() || !state.Group.IsNull() {
		ownership, err := client.GetFileOwnership(ctx, state.Path.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error reading file ownership",
				fmt.Sprintf("Could not read file ownership: %s", err),
			)
			return
		}
		if !state.Owner.IsNull() {
			state.Owner = basetypes.NewStringValue(ownership.User)
		}
		if !state.Group.IsNull() {
			state.Group = basetypes.NewStringValue(ownership.Group)
		}
	}

	// Get attributes if any were specified
	if !state.Immutable.IsNull() || !state.AppendOnly.IsNull() || !state.NoDump.IsNull() ||
		!state.Synchronous.IsNull() || !state.NoAtime.IsNull() || !state.Compressed.IsNull() ||
		!state.NoCoW.IsNull() || !state.Undeletable.IsNull() {
		attrs, err := client.GetFileAttributes(ctx, state.Path.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error reading file attributes",
				fmt.Sprintf("Could not read file attributes: %s", err),
			)
			return
		}
		if !state.Immutable.IsNull() {
			state.Immutable = types.BoolValue(attrs.Immutable)
		}
		if !state.AppendOnly.IsNull() {
			state.AppendOnly = types.BoolValue(attrs.AppendOnly)
		}
		if !state.NoDump.IsNull() {
			state.NoDump = types.BoolValue(attrs.NoDump)
		}
		if !state.Synchronous.IsNull() {
			state.Synchronous = types.BoolValue(attrs.Synchronous)
		}
		if !state.NoAtime.IsNull() {
			state.NoAtime = types.BoolValue(attrs.NoAtime)
		}
		if !state.Compressed.IsNull() {
			state.Compressed = types.BoolValue(attrs.Compressed)
		}
		if !state.NoCoW.IsNull() {
			state.NoCoW = types.BoolValue(attrs.NoCoW)
		}
		if !state.Undeletable.IsNull() {
			state.Undeletable = types.BoolValue(attrs.Undeletable)
		}
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *FileResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	ctx, span := otel.Tracer("ssh-provider").Start(ctx, "FileResource.Update")
	defer span.End()

	var plan FileResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.getClient(ctx, plan.SSH)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating SSH client",
			fmt.Sprintf("Could not create SSH client: %s", err),
		)
		return
	}
	defer client.Close()

	permissions := parsePermissions(plan.Permissions.ValueString())

	err = client.CreateFile(ctx, plan.Path.ValueString(), plan.Content.ValueString(), os.FileMode(permissions))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating file",
			fmt.Sprintf("Could not update file: %s", err),
		)
		return
	}

	// Set ownership if specified
	if !plan.Owner.IsNull() || !plan.Group.IsNull() {
		err = client.SetFileOwnership(ctx, plan.Path.ValueString(), &ssh.FileOwnership{
			User:  plan.Owner.ValueString(),
			Group: plan.Group.ValueString(),
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Error setting file ownership",
				fmt.Sprintf("Could not set file ownership: %s", err),
			)
			return
		}
	}

	// Set attributes if any are specified
	if !plan.Immutable.IsNull() || !plan.AppendOnly.IsNull() || !plan.NoDump.IsNull() ||
		!plan.Synchronous.IsNull() || !plan.NoAtime.IsNull() || !plan.Compressed.IsNull() ||
		!plan.NoCoW.IsNull() || !plan.Undeletable.IsNull() {
		err = client.SetFileAttributes(ctx, plan.Path.ValueString(), &ssh.FileAttributes{
			Immutable:   plan.Immutable.ValueBool(),
			AppendOnly:  plan.AppendOnly.ValueBool(),
			NoDump:      plan.NoDump.ValueBool(),
			Synchronous: plan.Synchronous.ValueBool(),
			NoAtime:     plan.NoAtime.ValueBool(),
			Compressed:  plan.Compressed.ValueBool(),
			NoCoW:       plan.NoCoW.ValueBool(),
			Undeletable: plan.Undeletable.ValueBool(),
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Error setting file attributes",
				fmt.Sprintf("Could not set file attributes: %s", err),
			)
			return
		}
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *FileResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	ctx, span := otel.Tracer("ssh-provider").Start(ctx, "FileResource.Delete")
	defer span.End()

	var state FileResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.getClient(ctx, state.SSH)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating SSH client",
			fmt.Sprintf("Could not create SSH client: %s", err),
		)
		return
	}
	defer client.Close()

	err = client.DeleteFile(ctx, state.Path.ValueString())
	if err != nil {
		if os.IsNotExist(err) {
			// If the file is already gone, that's fine
			return
		}
		resp.Diagnostics.AddError(
			"Error deleting file",
			fmt.Sprintf("Could not delete file: %s", err),
		)
		return
	}
}

func (r *FileResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
}

func (r *FileResource) getClient(ctx context.Context, sshBlock *ssh.SSHBlockModel) (*ssh.SSHClient, error) {
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

	client, err := r.pool.GetClient(ctx, config)
	if err != nil {
		return nil, err
	}

	// Release the client when the context is done
	go func() {
		<-ctx.Done()
		r.pool.ReleaseClient(config)
	}()

	return client, nil
}
