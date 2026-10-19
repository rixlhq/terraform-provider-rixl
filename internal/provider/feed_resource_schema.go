package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func FeedResourceSchema(_ context.Context) schema.Schema {
	return schema.Schema{
		Description:         "Manages a content feed within a Rixl project, defining allowed media types and engagement settings.",
		MarkdownDescription: "Manages a content feed within a Rixl project, defining allowed media types and engagement settings.",
		Attributes: map[string]schema.Attribute{
			"allow_images": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"allow_videos": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"created_at": schema.StringAttribute{
				Computed: true,
			},
			"description": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"has_comments": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"has_likes": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"has_shares": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"id": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"project_id": schema.StringAttribute{
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"updated_at": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

type FeedModel struct {
	AllowImages types.Bool   `tfsdk:"allow_images"`
	AllowVideos types.Bool   `tfsdk:"allow_videos"`
	CreatedAt   types.String `tfsdk:"created_at"`
	Description types.String `tfsdk:"description"`
	HasComments types.Bool   `tfsdk:"has_comments"`
	HasLikes    types.Bool   `tfsdk:"has_likes"`
	HasShares   types.Bool   `tfsdk:"has_shares"`
	Id          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	ProjectId   types.String `tfsdk:"project_id"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}
