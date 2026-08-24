package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rixlhq/rixl-go/sdk"
	"github.com/rixlhq/rixl-go/sdk/models"
	"github.com/rixlhq/rixl-go/sdk/payments"
	rttypes "github.com/rixlhq/rixl-go/sdk/runtime/types"
)

var _ resource.Resource = (*billingAddressResource)(nil)

func NewBillingAddressResource() resource.Resource {
	return &billingAddressResource{}
}

type billingAddressResource struct {
	client *sdk.Client
}

func (r *billingAddressResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_billing_address"
}

func (r *billingAddressResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = BillingAddressResourceSchema(ctx)
}

func (r *billingAddressResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*sdk.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data type", fmt.Sprintf("expected *sdk.Client, got %T", req.ProviderData))
		return
	}
	r.client = client
}

func (r *billingAddressResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data BillingAddressResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := r.buildRequest(data)
	address, err := r.client.Payments.UpsertBillingAddress(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to upsert billing address", err.Error())
		return
	}

	resp.Diagnostics.Append(mapResponseToModel(ctx, address, &data, BillingAddressResourceSchema(ctx).Attributes)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.Id.IsNull() || data.Id.IsUnknown() {
		if data.OrgId.IsNull() || data.OrgId.IsUnknown() {
			data.Id = types.StringValue("default")
		} else {
			data.Id = data.OrgId
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *billingAddressResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data BillingAddressResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := &payments.GetBillingAddressParams{}
	if !data.OrgId.IsNull() && !data.OrgId.IsUnknown() {
		orgID := data.OrgId.ValueString()
		params.OrgId = &orgID
	}

	address, err := r.client.Payments.GetBillingAddress(ctx, params)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read billing address", err.Error())
		return
	}

	resp.Diagnostics.Append(mapResponseToModel(ctx, address, &data, BillingAddressResourceSchema(ctx).Attributes)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.Id.IsNull() || data.Id.IsUnknown() {
		if data.OrgId.IsNull() || data.OrgId.IsUnknown() {
			data.Id = types.StringValue("default")
		} else {
			data.Id = data.OrgId
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *billingAddressResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state, plan BillingAddressResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataAny, diags := mergeStateAndPlan(ctx, &state, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data := dataAny.(*BillingAddressResourceModel)

	body := r.buildRequest(*data)
	address, err := r.client.Payments.UpsertBillingAddress(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to upsert billing address", err.Error())
		return
	}

	resp.Diagnostics.Append(mapResponseToModel(ctx, address, data, BillingAddressResourceSchema(ctx).Attributes)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.Id.IsNull() || data.Id.IsUnknown() {
		if data.OrgId.IsNull() || data.OrgId.IsUnknown() {
			data.Id = types.StringValue("default")
		} else {
			data.Id = data.OrgId
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *billingAddressResource) Delete(ctx context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	// The Rixl API does not expose a delete-billing-address endpoint.
	// Remove the resource from state and warn the user that the remote
	// billing address was not modified.
	resp.Diagnostics.AddWarning(
		"Billing address not deleted",
		"The Rixl API does not support deleting billing addresses. The resource has been removed from Terraform state, but the billing address may still exist in the platform.",
	)
	resp.State.RemoveResource(ctx)
}

func (r *billingAddressResource) buildRequest(data BillingAddressResourceModel) models.BillingV1UpsertBillingAddressRequest {
	addr := models.BillingV1BillingAddress{
		Name:       data.Name.ValueString(),
		Line1:      data.Line1.ValueString(),
		City:       data.City.ValueString(),
		State:      data.State.ValueString(),
		PostalCode: data.PostalCode.ValueString(),
		Country:    data.Country.ValueString(),
	}
	if !data.Line2.IsNull() && !data.Line2.IsUnknown() {
		v := data.Line2.ValueString()
		addr.Line2 = &v
	}
	if !data.Phone.IsNull() && !data.Phone.IsUnknown() {
		v := data.Phone.ValueString()
		addr.Phone = &v
	}
	if !data.Email.IsNull() && !data.Email.IsUnknown() {
		v := rttypes.Email(data.Email.ValueString())
		addr.Email = &v
	}

	req := models.BillingV1UpsertBillingAddressRequest{Address: &addr}
	if !data.OrgId.IsNull() && !data.OrgId.IsUnknown() {
		v := data.OrgId.ValueString()
		req.OrgID = &v
	}
	return req
}

func BillingAddressResourceSchema(_ context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"city": schema.StringAttribute{
				Required: true,
			},
			"country": schema.StringAttribute{
				Required: true,
			},
			"email": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"id": schema.StringAttribute{
				Computed: true,
			},
			"line1": schema.StringAttribute{
				Required: true,
			},
			"line2": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"org_id": schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"phone": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"postal_code": schema.StringAttribute{
				Required: true,
			},
			"state": schema.StringAttribute{
				Required: true,
			},
		},
	}
}

type BillingAddressResourceModel struct {
	Id         types.String `tfsdk:"id"`
	OrgId      types.String `tfsdk:"org_id"`
	Name       types.String `tfsdk:"name"`
	Line1      types.String `tfsdk:"line1"`
	Line2      types.String `tfsdk:"line2"`
	City       types.String `tfsdk:"city"`
	State      types.String `tfsdk:"state"`
	PostalCode types.String `tfsdk:"postal_code"`
	Country    types.String `tfsdk:"country"`
	Phone      types.String `tfsdk:"phone"`
	Email      types.String `tfsdk:"email"`
}
