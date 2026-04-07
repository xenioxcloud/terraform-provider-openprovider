package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/xenioxcloud/terraform-provider-openprovider/internal/openprovider"
)

var (
	_ resource.Resource                = &DomainSettingsResource{}
	_ resource.ResourceWithImportState = &DomainSettingsResource{}
)

type DomainSettingsResource struct {
	client *openprovider.Client
}

type DomainSettingsResourceModel struct {
	Domain    types.String `tfsdk:"domain"`
	DomainID  types.Int64  `tfsdk:"domain_id"`
	IsLocked  types.Bool   `tfsdk:"is_locked"`
	Autorenew types.String `tfsdk:"autorenew"`
}

func NewDomainSettingsResource() resource.Resource {
	return &DomainSettingsResource{}
}

func (r *DomainSettingsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain_settings"
}

func (r *DomainSettingsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages domain lock and auto-renewal settings at Openprovider.",
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

func (r *DomainSettingsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DomainSettingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DomainSettingsResourceModel
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

	var locked *bool
	if !data.IsLocked.IsNull() && !data.IsLocked.IsUnknown() {
		v := data.IsLocked.ValueBool()
		locked = &v
	}

	if err := r.client.UpdateDomainSettings(domain.ID, locked, data.Autorenew.ValueString()); err != nil {
		resp.Diagnostics.AddError("Failed to update domain settings", err.Error())
		return
	}

	// Read back to get computed values
	updated, err := r.client.GetDomain(domain.ID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read domain after update", err.Error())
		return
	}
	data.IsLocked = types.BoolValue(updated.IsLocked)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DomainSettingsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DomainSettingsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain, err := r.client.GetDomain(int(data.DomainID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Failed to read domain", err.Error())
		return
	}

	data.IsLocked = types.BoolValue(domain.IsLocked)
	data.Autorenew = types.StringValue(domain.Autorenew)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DomainSettingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data DomainSettingsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state DomainSettingsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.DomainID = state.DomainID

	var locked *bool
	if !data.IsLocked.IsNull() && !data.IsLocked.IsUnknown() {
		v := data.IsLocked.ValueBool()
		locked = &v
	}

	if err := r.client.UpdateDomainSettings(int(data.DomainID.ValueInt64()), locked, data.Autorenew.ValueString()); err != nil {
		resp.Diagnostics.AddError("Failed to update domain settings", err.Error())
		return
	}

	updated, err := r.client.GetDomain(int(data.DomainID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Failed to read domain after update", err.Error())
		return
	}
	data.IsLocked = types.BoolValue(updated.IsLocked)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DomainSettingsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Removing from state only — settings remain as-is on the domain.
}

func (r *DomainSettingsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	domain, err := r.client.FindDomainByName(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to find domain", err.Error())
		return
	}

	var data DomainSettingsResourceModel
	data.Domain = types.StringValue(req.ID)
	data.DomainID = types.Int64Value(int64(domain.ID))
	data.IsLocked = types.BoolValue(domain.IsLocked)
	data.Autorenew = types.StringValue(domain.Autorenew)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
