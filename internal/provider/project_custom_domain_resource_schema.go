package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func ProjectCustomDomainResourceSchema(_ context.Context) schema.Schema {
	return schema.Schema{
		Description:         "Associates a verified custom domain with a Rixl project.",
		MarkdownDescription: "Associates a verified custom domain with a Rixl project.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"org_id": schema.StringAttribute{
				Required:    true,
				Description: "The org_id path parameter.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"project_id": schema.StringAttribute{
				Required:    true,
				Description: "The project_id path parameter.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"custom_domain": schema.StringAttribute{
				Required: true,
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "RFC 3339 timestamp, e.g. \"2024-12-25T12:00:00Z\".",
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "RFC 3339 timestamp, e.g. \"2024-12-25T12:00:00Z\".",
			},
		},
	}
}

type ProjectCustomDomainModel struct {
	Id           types.String `tfsdk:"id"`
	OrgId        types.String `tfsdk:"org_id"`
	ProjectId    types.String `tfsdk:"project_id"`
	CustomDomain types.String `tfsdk:"custom_domain"`
	CreatedAt    types.String `tfsdk:"created_at"`
	UpdatedAt    types.String `tfsdk:"updated_at"`
}
