package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func BillingAddressResourceSchema(_ context.Context) schema.Schema {
	return schema.Schema{
		Description:         "Manages the organization billing and invoicing address in Rixl.",
		MarkdownDescription: "Manages the organization billing and invoicing address in Rixl.",
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
