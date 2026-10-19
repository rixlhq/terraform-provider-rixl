package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func CustomDomainResourceSchema(_ context.Context) schema.Schema {
	return schema.Schema{
		Description:         "Manages and verifies an organization custom domain in Rixl.",
		MarkdownDescription: "Manages and verifies an organization custom domain in Rixl.",
		Attributes: map[string]schema.Attribute{
			"org_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"domain": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"present": schema.BoolAttribute{
				Computed: true,
			},
			"id": schema.StringAttribute{
				Computed: true,
			},
			"status": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"auto_join": schema.BoolAttribute{
						Computed: true,
					},
					"pending": schema.SingleNestedAttribute{
						Computed: true,
						Attributes: map[string]schema.Attribute{
							"verification_token": schema.StringAttribute{
								Computed:  true,
								Sensitive: true,
							},
							"expires_at": schema.StringAttribute{
								Computed: true,
							},
						},
					},
					"verified": schema.SingleNestedAttribute{
						Computed: true,
						Attributes: map[string]schema.Attribute{
							"verified_at": schema.StringAttribute{
								Computed: true,
							},
						},
					},
				},
			},
		},
	}
}

type CustomDomainResourceModel struct {
	OrgId   types.String `tfsdk:"org_id"`
	Domain  types.String `tfsdk:"domain"`
	Present types.Bool   `tfsdk:"present"`
	Id      types.String `tfsdk:"id"`
	Status  types.Object `tfsdk:"status"`
}
