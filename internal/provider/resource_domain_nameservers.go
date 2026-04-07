package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/xenioxcloud/terraform-provider-openprovider/internal/openprovider"
)

var _ resource.Resource = &DomainNameserversResource{}

// DomainNameserversResource manages nameservers for an Openprovider domain.
type DomainNameserversResource struct {
	client *openprovider.Client
}

type DomainNameserversResourceModel struct {
	Domain      types.String          `tfsdk:"domain"`
	DomainID    types.Int64           `tfsdk:"domain_id"`
	Nameservers []NameserverModel     `tfsdk:"nameservers"`
}

type NameserverModel struct {
	Name types.String `tfsdk:"name"`
	IP4  types.String `tfsdk:"ip4"`
	IP6  types.String `tfsdk:"ip6"`
}

func NewDomainNameserversResource() resource.Resource {
	return &DomainNameserversResource{}
}

func (r *DomainNameserversResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain_nameservers"
}

func (r *DomainNameserversResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages nameservers for a domain registered at Openprovider.",
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
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"nameservers": schema.ListNestedAttribute{
				Description: "List of nameservers to set on the domain.",
				Required:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description: "Nameserver hostname.",
							Required:    true,
						},
						"ip4": schema.StringAttribute{
							Description: "IPv4 glue record (only needed for in-domain nameservers).",
							Optional:    true,
						},
						"ip6": schema.StringAttribute{
							Description: "IPv6 glue record (only needed for in-domain nameservers).",
							Optional:    true,
						},
					},
				},
			},
		},
	}
}

func (r *DomainNameserversResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DomainNameserversResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DomainNameserversResourceModel
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

	nameservers := toAPINameservers(data.Nameservers)
	if err := r.client.UpdateDomainNameservers(domain.ID, nameservers); err != nil {
		resp.Diagnostics.AddError("Failed to update nameservers", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DomainNameserversResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DomainNameserversResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain, err := r.client.GetDomain(int(data.DomainID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Failed to read domain", err.Error())
		return
	}

	data.Nameservers = fromAPINameservers(domain.Nameservers)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DomainNameserversResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data DomainNameserversResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state DomainNameserversResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.DomainID = state.DomainID

	nameservers := toAPINameservers(data.Nameservers)
	if err := r.client.UpdateDomainNameservers(int(data.DomainID.ValueInt64()), nameservers); err != nil {
		resp.Diagnostics.AddError("Failed to update nameservers", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DomainNameserversResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Nameservers can't be "deleted" — removing from state is sufficient.
	// The domain keeps whatever nameservers were last set.
}

func toAPINameservers(models []NameserverModel) []openprovider.Nameserver {
	ns := make([]openprovider.Nameserver, len(models))
	for i, m := range models {
		ns[i] = openprovider.Nameserver{
			Name: m.Name.ValueString(),
			IP4:  m.IP4.ValueString(),
			IP6:  m.IP6.ValueString(),
		}
	}
	return ns
}

func fromAPINameservers(apiNS []openprovider.Nameserver) []NameserverModel {
	models := make([]NameserverModel, len(apiNS))
	for i, ns := range apiNS {
		m := NameserverModel{
			Name: types.StringValue(ns.Name),
		}
		if ns.IP4 != "" {
			m.IP4 = types.StringValue(ns.IP4)
		} else {
			m.IP4 = types.StringNull()
		}
		if ns.IP6 != "" {
			m.IP6 = types.StringValue(ns.IP6)
		} else {
			m.IP6 = types.StringNull()
		}
		models[i] = m
	}
	return models
}
