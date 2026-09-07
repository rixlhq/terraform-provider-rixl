package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// paymentMethodModelV0 mirrors the v0 state where the payment provider was
// stored under the reserved root attribute "provider".
type paymentMethodModelV0 struct {
	Id              types.String `tfsdk:"id"`
	OrgId           types.String `tfsdk:"org_id"`
	PaymentMethodId types.String `tfsdk:"payment_method_id"`
	SetAsDefault    types.Bool   `tfsdk:"set_as_default"`
	Type            types.String `tfsdk:"type"`
	Provider        types.String `tfsdk:"provider"`
	Details         types.String `tfsdk:"details"`
	IsDefault       types.Bool   `tfsdk:"is_default"`
	Brand           types.String `tfsdk:"brand"`
	Last4           types.String `tfsdk:"last4"`
	ExpMonth        types.String `tfsdk:"exp_month"`
	ExpYear         types.String `tfsdk:"exp_year"`
	CreatedAt       types.String `tfsdk:"created_at"`
}

func paymentMethodSchemaV0(ctx context.Context) rschema.Schema {
	schema := PaymentMethodResourceSchema(ctx)
	attrs := make(map[string]rschema.Attribute, len(schema.Attributes)+1)
	for name, attr := range schema.Attributes {
		if name == "provider_name" {
			continue
		}
		attrs[name] = attr
	}
	attrs["provider"] = rschema.StringAttribute{
		Computed: true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	}
	schema.Version = 0
	schema.Attributes = attrs
	return schema
}

func upgradePaymentMethodStateV0toV1(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	var prior paymentMethodModelV0
	resp.Diagnostics.Append(req.State.Get(ctx, &prior)...)
	if resp.Diagnostics.HasError() {
		return
	}

	upgraded := PaymentMethodModel{
		Id:              prior.Id,
		OrgId:           prior.OrgId,
		PaymentMethodId: prior.PaymentMethodId,
		SetAsDefault:    prior.SetAsDefault,
		Type:            prior.Type,
		ProviderName:    prior.Provider,
		Details:         prior.Details,
		IsDefault:       prior.IsDefault,
		Brand:           prior.Brand,
		Last4:           prior.Last4,
		ExpMonth:        prior.ExpMonth,
		ExpYear:         prior.ExpYear,
		CreatedAt:       prior.CreatedAt,
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, upgraded)...)
}

// Ensure managedResource supports state upgrades for renamed attributes.
var _ resource.ResourceWithUpgradeState = (*managedResource)(nil)

func (r *managedResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	if r.descriptor.TypeName != "payment_method" {
		return nil
	}
	prior := paymentMethodSchemaV0(ctx)
	return map[int64]resource.StateUpgrader{
		0: {
			PriorSchema:   &prior,
			StateUpgrader: upgradePaymentMethodStateV0toV1,
		},
	}
}
