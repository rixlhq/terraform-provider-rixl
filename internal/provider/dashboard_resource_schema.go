package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func DashboardResourceSchema(_ context.Context) schema.Schema {
	return schema.Schema{
		Description:         "Manages an analytics and observability dashboard in Rixl.",
		MarkdownDescription: "Manages an analytics and observability dashboard in Rixl.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"org_id": schema.StringAttribute{
				Computed: true,
			},
			"owner_user_id": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"visibility": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.OneOf("private", "org"),
				},
			},
			"is_default": schema.BoolAttribute{
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.Bool{boolplanmodifier.RequiresReplace()},
			},
			"revision": schema.Int64Attribute{
				Computed: true,
			},
			"updated_by": schema.StringAttribute{
				Computed: true,
			},
			"created_at": schema.StringAttribute{
				Computed: true,
			},
			"updated_at": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

type DashboardModel struct {
	Id          types.String `tfsdk:"id"`
	OrgId       types.String `tfsdk:"org_id"`
	OwnerUserId types.String `tfsdk:"owner_user_id"`
	Name        types.String `tfsdk:"name"`
	Visibility  types.String `tfsdk:"visibility"`
	IsDefault   types.Bool   `tfsdk:"is_default"`
	Revision    types.Int64  `tfsdk:"revision"`
	UpdatedBy   types.String `tfsdk:"updated_by"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}
