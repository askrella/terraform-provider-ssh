package resource

import (
	"context"
	"fmt"
	"os"

	"github.com/askrella/askrella-ssh-provider/internal/provider/ssh"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"go.opentelemetry.io/otel"
)

var (
	_ resource.Resource              = &DirectoryResource{}
	_ resource.ResourceWithConfigure = &DirectoryResource{}
)

// DirectoryResource defines the resource implementation.
type DirectoryResource struct {
	pool *ssh.SSHPool
}

// DirectoryResourceModel describes the resource data model.
type DirectoryResourceModel struct {
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
	ID          types.String       `tfsdk:"id"`
}

// NewDirectoryResource creates a new resource implementation.
func NewDirectoryResource(pool *ssh.SSHPool) resource.Resource {
	return &DirectoryResource{
		pool: pool,
	}
}

// Metadata returns the resource type name.
func (r *DirectoryResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_directory"
}

// Schema defines the schema for the resource.
func (r *DirectoryResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a directory on a remote server via SSH.",
		Attributes: map[string]schema.Attribute{
			"ssh": schema.SingleNestedAttribute{
				Description: "SSH connection configuration.",
				Required:    true,
				Attributes:  ssh.SSHBlockSchema(),
			},
			"path": schema.StringAttribute{
				Description: "The path where the directory should be created on the remote server.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"permissions": schema.StringAttribute{
				Description: "The directory permissions in octal format (e.g., '0755').",
				Optional:    true,
			},
			"owner": schema.StringAttribute{
				Description: "The user owner of the directory.",
				Optional:    true,
			},
			"group": schema.StringAttribute{
				Description: "The group owner of the directory.",
				Optional:    true,
			},
			"immutable": schema.BoolAttribute{
				Description: "If true, the directory cannot be modified/deleted/renamed.",
				Optional:    true,
			},
			"append_only": schema.BoolAttribute{
				Description: "If true, the directory can only be opened in append mode for writing.",
				Optional:    true,
			},
			"no_dump": schema.BoolAttribute{
				Description: "If true, the directory is not included in backups.",
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
				Description: "If true, the directory is compressed.",
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
func (r *DirectoryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	ctx, span := otel.Tracer("ssh-provider").Start(ctx, "DirectoryResource.Create")
	defer span.End()

	var plan DirectoryResourceModel
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

	permissions := ssh.ParsePermissions(plan.Permissions.ValueString())

	if exists, _ := client.Exists(ctx, plan.Path.ValueString()); !exists {
		err = client.CreateDirectory(ctx, plan.Path.ValueString(), os.FileMode(permissions))
		if err != nil {
			resp.Diagnostics.AddError(
				"Error creating directory",
				fmt.Sprintf("Could not create directory: %s", err),
			)
			return
		}
	}

	// Set ownership if specified
	if !plan.Owner.IsNull() || !plan.Group.IsNull() {
		err = client.SetFileOwnership(ctx, plan.Path.ValueString(), &ssh.FileOwnership{
			User:  plan.Owner.ValueString(),
			Group: plan.Group.ValueString(),
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Error setting directory ownership",
				fmt.Sprintf("Could not set directory ownership: %s", err),
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
				"Error setting directory attributes",
				fmt.Sprintf("Could not set directory attributes: %s", err),
			)
			return
		}
	}

	plan.ID = basetypes.NewStringValue(plan.Path.ValueString())

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Read refreshes the Terraform state with the latest data.
func (r *DirectoryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	ctx, span := otel.Tracer("ssh-provider").Start(ctx, "DirectoryResource.Read")
	defer span.End()

	var state DirectoryResourceModel
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

	exists, err := client.Exists(ctx, state.Path.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error determining directory existence",
			fmt.Sprintf("Could not determine directory existence: %s", err),
		)
		return
	}
	if !exists {
		resp.State.RemoveResource(ctx)
		return
	}

	// Get directory mode
	mode, err := client.GetFileMode(ctx, state.Path.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading directory mode",
			fmt.Sprintf("Could not read directory mode: %s", err),
		)
		return
	}
	state.Permissions = basetypes.NewStringValue(fmt.Sprintf("%04o", mode))

	// Get ownership if it was specified
	if !state.Owner.IsNull() || !state.Group.IsNull() {
		ownership, err := client.GetFileOwnership(ctx, state.Path.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error reading directory ownership",
				fmt.Sprintf("Could not read directory ownership: %s", err),
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
				"Error reading directory attributes",
				fmt.Sprintf("Could not read directory attributes: %s", err),
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
func (r *DirectoryResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	ctx, span := otel.Tracer("ssh-provider").Start(ctx, "DirectoryResource.Update")
	defer span.End()

	var plan DirectoryResourceModel
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

	permissions := ssh.ParsePermissions(plan.Permissions.ValueString())
	wantedFileMode := os.FileMode(permissions)

	if exists, _ := client.Exists(ctx, plan.Path.ValueString()); !exists {
		err = client.CreateDirectory(ctx, plan.Path.ValueString(), wantedFileMode)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error updating directory",
				fmt.Sprintf("Could not update directory: %s", err),
			)
			return
		}
	}

	fileMode, err := client.GetFileMode(ctx, plan.Path.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error retrieving permissions",
			fmt.Sprintf("Could not retrieve permissions: %s", err),
		)
	}
	if fileMode != wantedFileMode {
		err := client.SetFileMode(ctx, plan.Path.ValueString(), wantedFileMode)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error updating permissions",
				fmt.Sprintf("Could not set permissions: %s", err),
			)
			return
		}
	}

	// Set ownership if specified
	if !plan.Owner.IsNull() || !plan.Group.IsNull() {
		err = client.SetFileOwnership(ctx, plan.Path.ValueString(), &ssh.FileOwnership{
			User:  plan.Owner.ValueString(),
			Group: plan.Group.ValueString(),
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Error setting directory ownership",
				fmt.Sprintf("Could not set directory ownership: %s", err),
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
				"Error setting directory attributes",
				fmt.Sprintf("Could not set directory attributes: %s", err),
			)
			return
		}
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *DirectoryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	ctx, span := otel.Tracer("ssh-provider").Start(ctx, "DirectoryResource.Delete")
	defer span.End()

	var state DirectoryResourceModel
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

	err = client.DeleteDirectory(ctx, state.Path.ValueString())
	if err != nil {
		if os.IsNotExist(err) {
			// If the directory is already gone, that's fine
			return
		}
		resp.Diagnostics.AddError(
			"Error deleting directory",
			fmt.Sprintf("Could not delete directory: %s", err),
		)
		return
	}
}

func (r *DirectoryResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
}

func (r *DirectoryResource) getClient(ctx context.Context, sshBlock *ssh.SSHBlockModel) (*ssh.SSHClient, error) {
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
