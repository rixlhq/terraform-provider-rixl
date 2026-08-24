package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func VideoResourceSchema(_ context.Context) schema.Schema {
	attrs := videoResourceAttributes()
	attrs["file_path"] = schema.StringAttribute{
		Required:      true,
		PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
	}
	attrs["poster_file_path"] = schema.StringAttribute{
		Optional:      true,
		PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
	}
	attrs["name"] = schema.StringAttribute{
		Optional: true,
		Computed: true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	}
	attrs["project_id"] = schema.StringAttribute{
		Required:      true,
		PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
	}
	return schema.Schema{Attributes: attrs}
}

func videoResourceAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"bitrate":   schema.Int64Attribute{Computed: true},
		"codec":     schema.StringAttribute{Computed: true},
		"duration":  schema.StringAttribute{Computed: true},
		"file":      schema.SingleNestedAttribute{Computed: true, Attributes: fileResourceAttributes()},
		"framerate": schema.StringAttribute{Computed: true},
		"hdr":       schema.BoolAttribute{Computed: true},
		"height":    schema.Int64Attribute{Computed: true},
		"id":        schema.StringAttribute{Computed: true},
		"poster":    schema.SingleNestedAttribute{Computed: true, Attributes: imageResourceAttributes()},
		"visibility": schema.StringAttribute{
			Optional: true,
			Computed: true,
		},
		"width": schema.Int64Attribute{Computed: true},
	}
}

type VideoResourceModel struct {
	VideoDataSourceModel

	ProjectId      types.String `tfsdk:"project_id"`
	Name           types.String `tfsdk:"name"`
	FilePath       types.String `tfsdk:"file_path"`
	PosterFilePath types.String `tfsdk:"poster_file_path"`
}
