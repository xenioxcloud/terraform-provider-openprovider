package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/xenioxcloud/terraform-provider-openprovider/internal/openprovider"
)

var _ resource.Resource = &DomainDNSSECResource{}

// DomainDNSSECResource manages DNSSEC for an Openprovider domain.
type DomainDNSSECResource struct {
	client *openprovider.Client
}

type DomainDNSSECResourceModel struct {
	Domain   types.String `tfsdk:"domain"`
	DomainID types.Int64  `tfsdk:"domain_id"`
	Enabled  types.Bool   `tfsdk:"enabled"`
}

func NewDomainDNSSECResource() resource.Resource {
	return &DomainDNSSECResource{}
}

func (r *DomainDNSSECResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain_dnssec"
}

func (r *DomainDNSSECResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages DNSSEC for a domain registered at Openprovider.",
		Attributes: map[string]schema.Attribute{
			"domain": schema.StringAttribute{
				Description: "Full domain name (e.g., example.nl).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"domain_id": schema.Int64Attribute{
				Description: "Openprovider domain ID (resolved automatically).",
				Computed:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether DNSSEC is enabled.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
		},
	}
}

func (r *DomainDNSSECResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*openprovider.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data type", fmt.Sprintf("Expected *openprovider.Client, got %T", req.ProviderData))
		return
	}
	r.client = client
}

func (r *DomainDNSSECResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DomainDNSSECResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain, err := r.client.FindDomainByName(data.Domain.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to find domain", err.Error())
		return
	}
	data.DomainID = types.Int64Value(int64(domain.ID))

	if err := r.client.UpdateDomainDNSSEC(domain.ID, data.Enabled.ValueBool()); err != nil {
		resp.Diagnostics.AddError("Failed to update DNSSEC", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DomainDNSSECResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DomainDNSSECResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain, err := r.client.GetDomain(int(data.DomainID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Failed to read domain", err.Error())
		return
	}

	data.Enabled = types.BoolValue(domain.IsDNSSECEnabled)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DomainDNSSECResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data DomainDNSSECResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state DomainDNSSECResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.DomainID = state.DomainID

	if err := r.client.UpdateDomainDNSSEC(int(data.DomainID.ValueInt64()), data.Enabled.ValueBool()); err != nil {
		resp.Diagnostics.AddError("Failed to update DNSSEC", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DomainDNSSECResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DomainDNSSECResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Disable DNSSEC on delete
	if err := r.client.UpdateDomainDNSSEC(int(data.DomainID.ValueInt64()), false); err != nil {
		resp.Diagnostics.AddError("Failed to disable DNSSEC", err.Error())
		return
	}
}
