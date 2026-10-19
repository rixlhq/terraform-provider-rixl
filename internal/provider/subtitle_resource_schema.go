package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func SubtitleResourceSchema() schema.Schema {
	return schema.Schema{
		Description:         "Uploads and manages subtitles (VTT or SRT) for a Rixl video.",
		MarkdownDescription: "Uploads and manages subtitles (VTT or SRT) for a Rixl video.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"project_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"video_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"language_code": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"label": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"file_path": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"format": schema.StringAttribute{
				Computed: true,
			},
			"vtt_path": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

type SubtitleResourceModel struct {
	Id           types.String `tfsdk:"id"`
	ProjectId    types.String `tfsdk:"project_id"`
	VideoId      types.String `tfsdk:"video_id"`
	LanguageCode types.String `tfsdk:"language_code"`
	Label        types.String `tfsdk:"label"`
	FilePath     types.String `tfsdk:"file_path"`
	Format       types.String `tfsdk:"format"`
	VttPath      types.String `tfsdk:"vtt_path"`
}
