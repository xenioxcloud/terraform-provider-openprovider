package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/xenioxcloud/terraform-provider-openprovider/internal/openprovider"
)

var _ provider.Provider = &OpenproviderProvider{}

// OpenproviderProvider implements the Openprovider Terraform provider.
type OpenproviderProvider struct {
	version string
}

type OpenproviderProviderModel struct {
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
	Sandbox  types.Bool   `tfsdk:"sandbox"`
}

// New returns a new provider factory function.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &OpenproviderProvider{
			version: version,
		}
	}
}

func (p *OpenproviderProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "openprovider"
	resp.Version = p.version
}

func (p *OpenproviderProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Terraform provider for managing domains at Openprovider.",
		Attributes: map[string]schema.Attribute{
			"username": schema.StringAttribute{
				Description: "Openprovider API username. Can also be set via OPENPROVIDER_USERNAME env var.",
				Optional:    true,
			},
			"password": schema.StringAttribute{
				Description: "Openprovider API password. Can also be set via OPENPROVIDER_PASSWORD env var.",
				Optional:    true,
				Sensitive:   true,
			},
			"sandbox": schema.BoolAttribute{
				Description: "Use the Openprovider sandbox environment. Can also be set via OPENPROVIDER_SANDBOX env var.",
				Optional:    true,
			},
		},
	}
}

func (p *OpenproviderProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config OpenproviderProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	username := os.Getenv("OPENPROVIDER_USERNAME")
	password := os.Getenv("OPENPROVIDER_PASSWORD")

	if !config.Username.IsNull() {
		username = config.Username.ValueString()
	}
	if !config.Password.IsNull() {
		password = config.Password.ValueString()
	}

	sandbox := os.Getenv("OPENPROVIDER_SANDBOX") == "true"
	if !config.Sandbox.IsNull() {
		sandbox = config.Sandbox.ValueBool()
	}

	if username == "" {
		resp.Diagnostics.AddError("Missing username", "Set username in provider config or OPENPROVIDER_USERNAME env var")
		return
	}
	if password == "" {
		resp.Diagnostics.AddError("Missing password", "Set password in provider config or OPENPROVIDER_PASSWORD env var")
		return
	}

	client := openprovider.NewClient(username, password, sandbox)
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *OpenproviderProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewDomainNameserversResource,
		NewDomainDNSSECResource,
	}
}

func (p *OpenproviderProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewDomainDataSource,
	}
}
