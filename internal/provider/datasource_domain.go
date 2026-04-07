package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/xenioxcloud/terraform-provider-openprovider/internal/openprovider"
)

var _ datasource.DataSource = &DomainDataSource{}

// DomainDataSource reads domain information from Openprovider.
type DomainDataSource struct {
	client *openprovider.Client
}

type DomainDataSourceModel struct {
	Domain          types.String `tfsdk:"domain"`
	DomainID        types.Int64  `tfsdk:"domain_id"`
	Status          types.String `tfsdk:"status"`
	ExpirationDate  types.String `tfsdk:"expiration_date"`
	IsDNSSECEnabled types.Bool   `tfsdk:"is_dnssec_enabled"`
}

func NewDomainDataSource() datasource.DataSource {
	return &DomainDataSource{}
}

func (d *DomainDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain"
}

func (d *DomainDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads domain information from Openprovider.",
		Attributes: map[string]schema.Attribute{
			"domain": schema.StringAttribute{
				Description: "Full domain name (e.g., example.nl).",
				Required:    true,
			},
			"domain_id": schema.Int64Attribute{
				Description: "Openprovider domain ID.",
				Computed:    true,
			},
			"status": schema.StringAttribute{
				Description: "Domain status (e.g., ACT).",
				Computed:    true,
			},
			"expiration_date": schema.StringAttribute{
				Description: "Domain expiration date.",
				Computed:    true,
			},
			"is_dnssec_enabled": schema.BoolAttribute{
				Description: "Whether DNSSEC is enabled.",
				Computed:    true,
			},
		},
	}
}

func (d *DomainDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*openprovider.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data type", fmt.Sprintf("Expected *openprovider.Client, got %T", req.ProviderData))
		return
	}
	d.client = client
}

func (d *DomainDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DomainDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain, err := d.client.FindDomainByName(data.Domain.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to find domain", err.Error())
		return
	}

	data.DomainID = types.Int64Value(int64(domain.ID))
	data.Status = types.StringValue(domain.Status)
	data.ExpirationDate = types.StringValue(domain.ExpirationDate)
	data.IsDNSSECEnabled = types.BoolValue(domain.IsDNSSECEnabled)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
