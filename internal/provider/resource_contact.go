package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/xenioxcloud/terraform-provider-openprovider/internal/openprovider"
)

var (
	_ resource.Resource                = &ContactResource{}
	_ resource.ResourceWithImportState = &ContactResource{}
)

type ContactResource struct {
	client *openprovider.Client
}

type ContactResourceModel struct {
	Handle          types.String `tfsdk:"handle"`
	CompanyName     types.String `tfsdk:"company_name"`
	FirstName       types.String `tfsdk:"first_name"`
	LastName        types.String `tfsdk:"last_name"`
	Email           types.String `tfsdk:"email"`
	PhoneCountry    types.String `tfsdk:"phone_country_code"`
	PhoneArea       types.String `tfsdk:"phone_area_code"`
	PhoneSubscriber types.String `tfsdk:"phone_subscriber_number"`
	Street          types.String `tfsdk:"street"`
	Number          types.String `tfsdk:"number"`
	Zipcode         types.String `tfsdk:"zipcode"`
	City            types.String `tfsdk:"city"`
	State           types.String `tfsdk:"state"`
	Country         types.String `tfsdk:"country"`
}

func NewContactResource() resource.Resource {
	return &ContactResource{}
}

func (r *ContactResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_contact"
}

func (r *ContactResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a contact (customer handle) at Openprovider.",
		Attributes: map[string]schema.Attribute{
			"handle": schema.StringAttribute{
				Description: "Openprovider contact handle (e.g., XX000001-NL). Computed on create.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"company_name": schema.StringAttribute{
				Description: "Company name.",
				Optional:    true,
			},
			"first_name": schema.StringAttribute{
				Description: "First name.",
				Required:    true,
			},
			"last_name": schema.StringAttribute{
				Description: "Last name.",
				Required:    true,
			},
			"email": schema.StringAttribute{
				Description: "Email address.",
				Required:    true,
			},
			"phone_country_code": schema.StringAttribute{
				Description: "Phone country code (e.g., +31).",
				Required:    true,
			},
			"phone_area_code": schema.StringAttribute{
				Description: "Phone area code (e.g., 70).",
				Required:    true,
			},
			"phone_subscriber_number": schema.StringAttribute{
				Description: "Phone subscriber number.",
				Required:    true,
			},
			"street": schema.StringAttribute{
				Description: "Street name.",
				Required:    true,
			},
			"number": schema.StringAttribute{
				Description: "House/building number.",
				Required:    true,
			},
			"zipcode": schema.StringAttribute{
				Description: "Postal/zip code.",
				Required:    true,
			},
			"city": schema.StringAttribute{
				Description: "City.",
				Required:    true,
			},
			"state": schema.StringAttribute{
				Description: "State/province (required for some countries).",
				Optional:    true,
			},
			"country": schema.StringAttribute{
				Description: "Two-letter country code (e.g., NL).",
				Required:    true,
			},
		},
	}
}

func (r *ContactResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ContactResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ContactResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	contact := modelToContact(&data)
	handle, err := r.client.CreateContact(contact)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create contact", err.Error())
		return
	}

	data.Handle = types.StringValue(handle)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ContactResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ContactResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	contact, err := r.client.GetContact(data.Handle.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read contact", err.Error())
		return
	}

	contactToModel(contact, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ContactResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ContactResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state ContactResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Handle = state.Handle

	contact := modelToContact(&data)
	if err := r.client.UpdateContact(data.Handle.ValueString(), contact); err != nil {
		resp.Diagnostics.AddError("Failed to update contact", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ContactResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ContactResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteContact(data.Handle.ValueString()); err != nil {
		resp.Diagnostics.AddError("Failed to delete contact", err.Error())
		return
	}
}

func (r *ContactResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by handle, e.g.: terraform import openprovider_contact.example XX000001-NL
	contact, err := r.client.GetContact(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to find contact", err.Error())
		return
	}

	var data ContactResourceModel
	contactToModel(contact, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func modelToContact(data *ContactResourceModel) *openprovider.Contact {
	return &openprovider.Contact{
		CompanyName: data.CompanyName.ValueString(),
		Name: openprovider.ContactName{
			FirstName: data.FirstName.ValueString(),
			LastName:  data.LastName.ValueString(),
		},
		Phone: openprovider.ContactPhone{
			CountryCode:      data.PhoneCountry.ValueString(),
			AreaCode:         data.PhoneArea.ValueString(),
			SubscriberNumber: data.PhoneSubscriber.ValueString(),
		},
		Address: openprovider.ContactAddress{
			Street:  data.Street.ValueString(),
			Number:  data.Number.ValueString(),
			Zipcode: data.Zipcode.ValueString(),
			City:    data.City.ValueString(),
			State:   data.State.ValueString(),
			Country: data.Country.ValueString(),
		},
		Email: data.Email.ValueString(),
	}
}

func contactToModel(contact *openprovider.Contact, data *ContactResourceModel) {
	data.Handle = types.StringValue(contact.Handle)
	data.CompanyName = types.StringValue(contact.CompanyName)
	data.FirstName = types.StringValue(contact.Name.FirstName)
	data.LastName = types.StringValue(contact.Name.LastName)
	data.Email = types.StringValue(contact.Email)
	data.PhoneCountry = types.StringValue(contact.Phone.CountryCode)
	data.PhoneArea = types.StringValue(contact.Phone.AreaCode)
	data.PhoneSubscriber = types.StringValue(contact.Phone.SubscriberNumber)
	data.Street = types.StringValue(contact.Address.Street)
	data.Number = types.StringValue(contact.Address.Number)
	data.Zipcode = types.StringValue(contact.Address.Zipcode)
	data.City = types.StringValue(contact.Address.City)
	if contact.Address.State != "" {
		data.State = types.StringValue(contact.Address.State)
	} else {
		data.State = types.StringNull()
	}
	data.Country = types.StringValue(contact.Address.Country)
}
