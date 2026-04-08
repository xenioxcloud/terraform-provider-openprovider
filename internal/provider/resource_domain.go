package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/xenioxcloud/terraform-provider-openprovider/internal/openprovider"
)

var (
	_ resource.Resource                = &DomainResource{}
	_ resource.ResourceWithImportState = &DomainResource{}
)

type DomainResource struct {
	client *openprovider.Client
}

type DomainResourceModel struct {
	Domain      types.String      `tfsdk:"domain"`
	DomainID    types.Int64       `tfsdk:"domain_id"`
	Nameservers []NameserverModel `tfsdk:"nameservers"`
	DNSSEC      types.Bool        `tfsdk:"dnssec"`
	IsLocked    types.Bool        `tfsdk:"is_locked"`
	Autorenew   types.String      `tfsdk:"autorenew"`
}

func NewDomainResource() resource.Resource {
	return &DomainResource{}
}

func (r *DomainResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain"
}

func (r *DomainResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a domain at Openprovider: nameservers, DNSSEC, lock, and auto-renewal.",
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
			"dnssec": schema.BoolAttribute{
				Description: "Whether DNSSEC is enabled.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"is_locked": schema.BoolAttribute{
				Description: "Whether the domain transfer lock is enabled. Leave unset for TLDs that don't support locking.",
				Optional:    true,
				Computed:    true,
			},
			"autorenew": schema.StringAttribute{
				Description: "Auto-renewal mode: \"on\", \"off\", or \"default\".",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("default"),
			},
		},
	}
}

func (r *DomainResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DomainResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DomainResourceModel
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

	r.applyChanges(domain.ID, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	r.readDomain(domain.ID, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DomainResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DomainResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.readDomain(int(data.DomainID.ValueInt64()), &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DomainResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data DomainResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state DomainResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.DomainID = state.DomainID

	domainID := int(data.DomainID.ValueInt64())
	r.applyChanges(domainID, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	r.readDomain(domainID, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DomainResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Removing from state only — domain continues to exist at Openprovider.
}

func (r *DomainResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	domain, err := r.client.FindDomainByName(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to find domain", err.Error())
		return
	}

	var data DomainResourceModel
	data.Domain = types.StringValue(req.ID)
	data.DomainID = types.Int64Value(int64(domain.ID))
	data.Nameservers = fromAPINameservers(domain.Nameservers)
	data.DNSSEC = types.BoolValue(domain.IsDNSSECEnabled)
	data.IsLocked = types.BoolValue(domain.IsLocked)
	data.Autorenew = types.StringValue(domain.Autorenew)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DomainResource) applyChanges(domainID int, data *DomainResourceModel, diags *diag.Diagnostics) {
	nameservers := toAPINameservers(data.Nameservers)
	if err := r.client.UpdateDomainNameservers(domainID, nameservers); err != nil {
		diags.AddError("Failed to update nameservers", err.Error())
		return
	}

	if err := r.client.UpdateDomainDNSSEC(domainID, data.DNSSEC.ValueBool()); err != nil {
		diags.AddError("Failed to update DNSSEC", err.Error())
		return
	}

	var locked *bool
	if !data.IsLocked.IsNull() && !data.IsLocked.IsUnknown() {
		v := data.IsLocked.ValueBool()
		locked = &v
	}
	if err := r.client.UpdateDomainSettings(domainID, locked, data.Autorenew.ValueString()); err != nil {
		diags.AddError("Failed to update domain settings", err.Error())
		return
	}
}

func (r *DomainResource) readDomain(domainID int, data *DomainResourceModel, diags *diag.Diagnostics) {
	domain, err := r.client.GetDomain(domainID)
	if err != nil {
		diags.AddError("Failed to read domain", err.Error())
		return
	}

	data.Nameservers = fromAPINameservers(domain.Nameservers)
	data.DNSSEC = types.BoolValue(domain.IsDNSSECEnabled)
	data.IsLocked = types.BoolValue(domain.IsLocked)
	data.Autorenew = types.StringValue(domain.Autorenew)
}
