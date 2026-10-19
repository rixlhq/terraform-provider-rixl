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
	"github.com/rixlhq/rixl-go/sdk/payments"
)

var (
	_ resource.Resource                = (*paymentMethodResource)(nil)
	_ resource.ResourceWithImportState = (*paymentMethodResource)(nil)
)

func NewPaymentMethodResource() resource.Resource {
	return &paymentMethodResource{}
}

type paymentMethodResource struct {
	client *sdk.Client
}

func (r *paymentMethodResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_payment_method"
}

func (r *paymentMethodResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = PaymentMethodResourceSchema(ctx)
}

func (r *paymentMethodResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *paymentMethodResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data PaymentMethodModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	pmID := data.PaymentMethodId.ValueString()
	body := models.BillingV1UpsertPaymentMethodRequest{
		PaymentMethodID: &pmID,
	}
	if !data.OrgId.IsNull() && !data.OrgId.IsUnknown() {
		org := data.OrgId.ValueString()
		body.OrgID = &org
	}
	if !data.SetAsDefault.IsNull() && !data.SetAsDefault.IsUnknown() {
		def := data.SetAsDefault.ValueBool()
		body.SetAsDefault = &def
	}

	res, err := r.client.Payments.UpsertPaymentMethod(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to upsert payment method", err.Error())
		return
	}

	data.Id = data.PaymentMethodId
	resp.Diagnostics.Append(mapResponseToModel(ctx, res, &data, PaymentMethodResourceSchema(ctx).Attributes)...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Id = data.PaymentMethodId

	r.populateDetails(ctx, &data, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *paymentMethodResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data PaymentMethodModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var params payments.ListPaymentMethodsParams
	if !data.OrgId.IsNull() && !data.OrgId.IsUnknown() {
		org := data.OrgId.ValueString()
		params.OrgId = &org
	}

	res, err := r.client.Payments.ListPaymentMethods(ctx, &params)
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to list payment methods", err.Error())
		return
	}

	var found any
	targetID := data.PaymentMethodId.ValueString()
	for _, pm := range res.PaymentMethods {
		if pm.ID != nil && *pm.ID == targetID {
			found = pm
			break
		}
	}
	if found == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(mapResponseToModel(ctx, found, &data, PaymentMethodResourceSchema(ctx).Attributes)...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Id = data.PaymentMethodId

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *paymentMethodResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data PaymentMethodModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	pmID := data.PaymentMethodId.ValueString()
	body := models.BillingV1UpsertPaymentMethodRequest{
		PaymentMethodID: &pmID,
	}
	if !data.OrgId.IsNull() && !data.OrgId.IsUnknown() {
		org := data.OrgId.ValueString()
		body.OrgID = &org
	}
	if !data.SetAsDefault.IsNull() && !data.SetAsDefault.IsUnknown() {
		def := data.SetAsDefault.ValueBool()
		body.SetAsDefault = &def
	}

	res, err := r.client.Payments.UpsertPaymentMethod(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update payment method", err.Error())
		return
	}

	data.Id = data.PaymentMethodId
	resp.Diagnostics.Append(mapResponseToModel(ctx, res, &data, PaymentMethodResourceSchema(ctx).Attributes)...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Id = data.PaymentMethodId

	r.populateDetails(ctx, &data, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *paymentMethodResource) populateDetails(ctx context.Context, data *PaymentMethodModel, diags *diag.Diagnostics) {
	var params payments.ListPaymentMethodsParams
	if !data.OrgId.IsNull() && !data.OrgId.IsUnknown() {
		org := data.OrgId.ValueString()
		params.OrgId = &org
	}

	res, err := r.client.Payments.ListPaymentMethods(ctx, &params)
	if err != nil {
		return
	}

	targetID := data.PaymentMethodId.ValueString()
	for _, pm := range res.PaymentMethods {
		if pm.ID != nil && *pm.ID == targetID {
			diags.Append(mapResponseToModel(ctx, pm, data, PaymentMethodResourceSchema(ctx).Attributes)...)
			if pm.Provider != nil {
				data.ProviderName = types.StringValue(*pm.Provider)
			}
			data.Id = data.PaymentMethodId
			break
		}
	}
}

func (r *paymentMethodResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data PaymentMethodModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var params payments.DeletePaymentMethodParams
	if !data.OrgId.IsNull() && !data.OrgId.IsUnknown() {
		org := data.OrgId.ValueString()
		params.OrgId = &org
	}

	_, err := r.client.Payments.DeletePaymentMethod(ctx, data.PaymentMethodId.ValueString(), &params)
	if err != nil && !isNotFound(err) {
		resp.Diagnostics.AddError("Failed to delete payment method", err.Error())
	}
}

func (r *paymentMethodResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("payment_method_id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}
