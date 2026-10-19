package provider

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rixlhq/rixl-go/sdk"
	"github.com/rixlhq/rixl-go/sdk/models"
	"github.com/rixlhq/rixl-go/sdk/subscriptions"
)

var (
	_ resource.Resource                = &subscriptionResource{}
	_ resource.ResourceWithConfigure   = &subscriptionResource{}
	_ resource.ResourceWithImportState = &subscriptionResource{}
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
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", "Expected *sdk.Client")
		return
	}
	r.client = client
}

func (r *subscriptionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan SubscriptionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := models.BillingV1CreateSubscriptionRequest{
		StripePriceID: plan.StripePriceId.ValueString(),
	}
	if !plan.OrgId.IsNull() && !plan.OrgId.IsUnknown() {
		org := plan.OrgId.ValueString()
		body.OrgID = &org
	}
	if !plan.PaymentMethodId.IsNull() && !plan.PaymentMethodId.IsUnknown() {
		pm := plan.PaymentMethodId.ValueString()
		body.PaymentMethodID = &pm
	}

	res, err := r.client.Subscriptions.CreateSubscription(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create subscription", err.Error())
		return
	}

	var data SubscriptionModel
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
	data.StripePriceId = plan.StripePriceId
	data.OrgId = plan.OrgId

	r.readSubscriptionIntoState(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

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
	var state SubscriptionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.StripePriceId.Equal(state.StripePriceId) {
		body := models.BillingV1UpgradeSubscriptionRequest{
			StripePriceID: plan.StripePriceId.ValueString(),
		}
		if !plan.OrgId.IsNull() && !plan.OrgId.IsUnknown() {
			org := plan.OrgId.ValueString()
			body.OrgID = &org
		}
		_, err := r.client.Payments.UpgradeSubscription(ctx, body)
		if err != nil {
			resp.Diagnostics.AddError("Failed to upgrade subscription", err.Error())
			return
		}
	}

	r.readSubscriptionIntoState(ctx, &plan, &resp.Diagnostics)
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
	parts := strings.Split(req.ID, "/")
	if len(parts) == 2 {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("org_id"), parts[0])...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}
