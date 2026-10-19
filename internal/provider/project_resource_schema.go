package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func ProjectResourceSchema(_ context.Context) schema.Schema {
	return schema.Schema{
		Description:         "Manages a Rixl project, providing an isolated container for videos, images, and content feeds.",
		MarkdownDescription: "Manages a Rixl project, providing an isolated container for videos, images, and content feeds.",
		Attributes: map[string]schema.Attribute{
			"org_id": schema.StringAttribute{
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"regions": schema.ListAttribute{
				ElementType:   types.StringType,
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.List{listplanmodifier.RequiresReplace()},
			},
			"video_quality": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"custom_domain": schema.StringAttribute{
				Computed: true,
			},
			"id": schema.StringAttribute{
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

type ProjectModel struct {
	CreatedAt    types.String `tfsdk:"created_at"`
	CustomDomain types.String `tfsdk:"custom_domain"`
	Id           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	OrgId        types.String `tfsdk:"org_id"`
	Regions      types.List   `tfsdk:"regions"`
	UpdatedAt    types.String `tfsdk:"updated_at"`
	VideoQuality types.String `tfsdk:"video_quality"`
}
