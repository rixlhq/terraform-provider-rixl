package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rixlhq/rixl-go/sdk"
	"github.com/rixlhq/rixl-go/sdk/models"
	"github.com/rixlhq/rixl-go/sdk/subscriptions"
)

var (
	_ resource.Resource                = (*subscriptionResource)(nil)
	_ resource.ResourceWithImportState = (*subscriptionResource)(nil)
)

func NewSubscriptionResource() resource.Resource {
	return &subscriptionResource{}
}

type subscriptionResource struct {
	client *sdk.Client
}

func (r *subscriptionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subscription"
}

func (r *subscriptionResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = SubscriptionResourceSchema(ctx)
}

func (r *subscriptionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *subscriptionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SubscriptionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := models.BillingV1CreateSubscriptionRequest{
		StripePriceID: data.StripePriceId.ValueString(),
	}
	if !data.OrgId.IsNull() && !data.OrgId.IsUnknown() {
		org := data.OrgId.ValueString()
		body.OrgID = &org
	}
	if !data.PaymentMethodId.IsNull() && !data.PaymentMethodId.IsUnknown() {
		pm := data.PaymentMethodId.ValueString()
		body.PaymentMethodID = &pm
	}

	res, err := r.client.Subscriptions.CreateSubscription(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create subscription", err.Error())
		return
	}

	resp.Diagnostics.Append(mapResponseToModel(ctx, res, &data, SubscriptionResourceSchema(ctx).Attributes)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.Id.IsNull() || data.Id.IsUnknown() {
		if res.ID != nil {
			data.Id = types.StringValue(*res.ID)
		} else {
			data.Id = data.OrgId
		}
	}

	r.readSubscriptionIntoState(ctx, &data, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *subscriptionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SubscriptionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.readSubscriptionIntoState(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *subscriptionResource) readSubscriptionIntoState(ctx context.Context, data *SubscriptionModel, diags *diag.Diagnostics) {
	var params subscriptions.GetSubscriptionParams
	if !data.OrgId.IsNull() && !data.OrgId.IsUnknown() {
		org := data.OrgId.ValueString()
		params.OrgId = &org
	}

	sub, err := r.client.Subscriptions.GetSubscription(ctx, &params)
	if err != nil {
		if isNotFound(err) {
			return
		}
		diags.AddError("Failed to get subscription", err.Error())
		return
	}

	diags.Append(mapResponseToModel(ctx, sub, data, SubscriptionResourceSchema(ctx).Attributes)...)
}

func (r *subscriptionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan SubscriptionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *subscriptionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SubscriptionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var params subscriptions.CancelSubscriptionParams
	if !data.OrgId.IsNull() && !data.OrgId.IsUnknown() {
		org := data.OrgId.ValueString()
		params.OrgId = &org
	}

	_, err := r.client.Subscriptions.CancelSubscription(ctx, &params)
	if err != nil && !isNotFound(err) {
		resp.Diagnostics.AddError("Failed to cancel subscription", err.Error())
	}
}

func (r *subscriptionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("org_id"), req.ID)...)
}
