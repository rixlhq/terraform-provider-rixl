package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func ImageResourceSchema(_ context.Context) schema.Schema {
	attrs := imageResourceAttributes()
	attrs["file_path"] = schema.StringAttribute{
		Required:      true,
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
	return schema.Schema{
		Description:         "Uploads and manages image assets in Rixl with automatic dimensions calculation and thumbhash generation.",
		MarkdownDescription: "Uploads and manages image assets in Rixl with automatic dimensions calculation and thumbhash generation.",
		Attributes:          attrs,
	}
}

type ImageResourceModel struct {
	ImageDataSourceModel

	ProjectId types.String `tfsdk:"project_id"`
	Name      types.String `tfsdk:"name"`
	FilePath  types.String `tfsdk:"file_path"`
}
