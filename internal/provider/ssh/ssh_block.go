package ssh

import (
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// SSHBlockModel represents the shared SSH configuration block
type SSHBlockModel struct {
	Host       types.String `tfsdk:"host"`
	Port       types.Int64  `tfsdk:"port"`
	Username   types.String `tfsdk:"username"`
	Password   types.String `tfsdk:"password"`
	PrivateKey types.String `tfsdk:"private_key"`
}

// SSHBlockSchema returns the schema for the SSH block
func SSHBlockSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"host": schema.StringAttribute{
			Description: "The hostname or IP address of the remote server.",
			Required:    true,
		},
		"port": schema.Int64Attribute{
			Description: "The SSH port of the remote server.",
			Optional:    true,
		},
		"username": schema.StringAttribute{
			Description: "The username to use for SSH authentication.",
			Required:    true,
		},
		"password": schema.StringAttribute{
			Description: "The password to use for SSH authentication.",
			Optional:    true,
			Sensitive:   true,
		},
		"private_key": schema.StringAttribute{
			Description: "The private key to use for SSH authentication.",
			Optional:    true,
			Sensitive:   true,
		},
	}
}

// SSHBlockDataSourceSchema returns the schema for the SSH block in data sources
func SSHBlockDataSourceSchema() map[string]dschema.Attribute {
	return map[string]dschema.Attribute{
		"host": dschema.StringAttribute{
			Description: "The hostname or IP address of the remote server.",
			Required:    true,
		},
		"port": dschema.Int64Attribute{
			Description: "The SSH port of the remote server.",
			Optional:    true,
		},
		"username": dschema.StringAttribute{
			Description: "The username to use for SSH authentication.",
			Required:    true,
		},
		"password": dschema.StringAttribute{
			Description: "The password to use for SSH authentication.",
			Optional:    true,
			Sensitive:   true,
		},
		"private_key": dschema.StringAttribute{
			Description: "The private key to use for SSH authentication.",
			Optional:    true,
			Sensitive:   true,
		},
	}
}
