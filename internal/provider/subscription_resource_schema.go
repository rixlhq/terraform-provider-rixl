package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func SubscriptionResourceSchema(_ context.Context) schema.Schema {
	return schema.Schema{
		Description:         "Manages an organization billing subscription tier and Stripe payment plan.",
		MarkdownDescription: "Manages an organization billing subscription tier and Stripe payment plan.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"org_id": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"payment_method_id": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"stripe_price_id": schema.StringAttribute{
				Required: true,
			},
			"plan_id": schema.StringAttribute{
				Computed: true,
			},
			"plan_name": schema.StringAttribute{
				Computed: true,
			},
			"plan_type": schema.StringAttribute{
				Computed: true,
			},
			"status": schema.StringAttribute{
				Computed: true,
			},
			"price": schema.StringAttribute{
				Computed: true,
			},
			"currency": schema.StringAttribute{
				Computed: true,
			},
			"current_period_end": schema.StringAttribute{
				Computed:    true,
				Description: "RFC 3339 timestamp, e.g. \"2024-12-25T12:00:00Z\".",
			},
			"cancel_at_period_end": schema.BoolAttribute{
				Computed: true,
			},
			"expiring_soon": schema.BoolAttribute{
				Computed: true,
			},
			"trials_ending_soon": schema.BoolAttribute{
				Computed: true,
			},
			"stripe_customer_id": schema.StringAttribute{
				Computed: true,
			},
			"stripe_subscription_id": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

type SubscriptionModel struct {
	CancelAtPeriodEnd    types.Bool   `tfsdk:"cancel_at_period_end"`
	Currency             types.String `tfsdk:"currency"`
	CurrentPeriodEnd     types.String `tfsdk:"current_period_end"`
	ExpiringSoon         types.Bool   `tfsdk:"expiring_soon"`
	Id                   types.String `tfsdk:"id"`
	OrgId                types.String `tfsdk:"org_id"`
	PaymentMethodId      types.String `tfsdk:"payment_method_id"`
	PlanId               types.String `tfsdk:"plan_id"`
	PlanName             types.String `tfsdk:"plan_name"`
	PlanType             types.String `tfsdk:"plan_type"`
	Price                types.String `tfsdk:"price"`
	Status               types.String `tfsdk:"status"`
	StripeCustomerId     types.String `tfsdk:"stripe_customer_id"`
	StripePriceId        types.String `tfsdk:"stripe_price_id"`
	StripeSubscriptionId types.String `tfsdk:"stripe_subscription_id"`
	TrialsEndingSoon     types.Bool   `tfsdk:"trials_ending_soon"`
}
