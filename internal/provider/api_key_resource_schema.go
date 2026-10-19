package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func ApiKeyResourceSchema(_ context.Context) schema.Schema {
	return schema.Schema{
		Description:         "Manages a Rixl API key for authenticating API requests and CI/CD automation.",
		MarkdownDescription: "Manages a Rixl API key for authenticating API requests and CI/CD automation.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"org_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"project_id": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_name": schema.StringAttribute{
				Computed: true,
			},
			"secret": schema.StringAttribute{
				Computed:  true,
				Sensitive: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "RFC 3339 timestamp, e.g. \"2024-12-25T12:00:00Z\".",
			},
			"expiring_at": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "RFC 3339 timestamp, e.g. \"2024-12-25T12:00:00Z\".",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"last_used": schema.StringAttribute{
				Computed:    true,
				Description: "RFC 3339 timestamp, e.g. \"2024-12-25T12:00:00Z\".",
			},
			"permissions": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
			"policy_ids": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
		},
	}
}

type ApiKeyModel struct {
	CreatedAt   types.String `tfsdk:"created_at"`
	ExpiringAt  types.String `tfsdk:"expiring_at"`
	Id          types.String `tfsdk:"id"`
	LastUsed    types.String `tfsdk:"last_used"`
	Name        types.String `tfsdk:"name"`
	OrgId       types.String `tfsdk:"org_id"`
	Permissions types.List   `tfsdk:"permissions"`
	PolicyIds   types.List   `tfsdk:"policy_ids"`
	ProjectId   types.String `tfsdk:"project_id"`
	ProjectName types.String `tfsdk:"project_name"`
	Secret      types.String `tfsdk:"secret"`
}
