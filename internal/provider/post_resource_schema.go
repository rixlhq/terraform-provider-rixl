package provider

import (
	"context"

	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func PostResourceSchema(_ context.Context) rschema.Schema {
	return rschema.Schema{
		Description:         "Publishes and manages a media post within a Rixl content feed.",
		MarkdownDescription: "Publishes and manages a media post within a Rixl content feed.",
		Attributes: map[string]rschema.Attribute{
			"project_id": rschema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"feed_id": rschema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"id": rschema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"org_id": rschema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"creator_id": rschema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": rschema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": rschema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"image_id": rschema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"video_id": rschema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"created_at": rschema.StringAttribute{
				Computed: true,
			},
		},
	}
}

type PostModel struct {
	ProjectId   types.String `tfsdk:"project_id"`
	FeedId      types.String `tfsdk:"feed_id"`
	Id          types.String `tfsdk:"id"`
	OrgId       types.String `tfsdk:"org_id"`
	CreatorId   types.String `tfsdk:"creator_id"`
	Type        types.String `tfsdk:"type"`
	Description types.String `tfsdk:"description"`
	ImageId     types.String `tfsdk:"image_id"`
	VideoId     types.String `tfsdk:"video_id"`
	CreatedAt   types.String `tfsdk:"created_at"`
}
