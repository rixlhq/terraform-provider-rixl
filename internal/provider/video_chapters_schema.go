package provider

import (
	"cmp"
	"context"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rixlhq/rixl-go/sdk/models"
)

func videoChaptersID(projectID, videoID string) string {
	return projectID + "/" + videoID
}

// staleChapterStarts returns the start times present in current but absent
// from planned, i.e. the chapters Update/Delete must remove one by one.
func staleChapterStarts(current, planned []models.VideosV1Chapter) []float64 {
	plannedKeys := make(map[float64]bool, len(planned))
	for _, c := range planned {
		if c.StartTimeSec != nil {
			plannedKeys[*c.StartTimeSec] = true
		}
	}
	var out []float64
	for _, c := range current {
		if c.StartTimeSec == nil || plannedKeys[*c.StartTimeSec] {
			continue
		}
		out = append(out, *c.StartTimeSec)
	}
	return out
}

func videoChapterAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"title":          types.StringType,
		"start_time_sec": types.Float64Type,
	}
}

func expandVideoChapters(ctx context.Context, list types.List) ([]models.VideosV1Chapter, diag.Diagnostics) {
	var diags diag.Diagnostics
	if list.IsNull() || list.IsUnknown() {
		return nil, diags
	}
	var items []VideoChapterItemModel
	diags.Append(list.ElementsAs(ctx, &items, false)...)
	if diags.HasError() {
		return nil, diags
	}
	out := make([]models.VideosV1Chapter, 0, len(items))
	for _, item := range items {
		var start *float64
		if !item.StartTimeSec.IsNull() && !item.StartTimeSec.IsUnknown() {
			start = new(item.StartTimeSec.ValueFloat64())
		}
		out = append(out, models.VideosV1Chapter{
			Title:        tfStringPtr(item.Title),
			StartTimeSec: start,
		})
	}
	return out, diags
}

func flattenVideoChapters(ctx context.Context, in []models.VideosV1Chapter) (types.List, diag.Diagnostics) {
	items := make([]VideoChapterItemModel, 0, len(in))
	for _, c := range in {
		var start types.Float64
		if c.StartTimeSec == nil {
			start = types.Float64Null()
		} else {
			start = types.Float64Value(*c.StartTimeSec)
		}
		items = append(items, VideoChapterItemModel{
			Title:        ptrString(c.Title),
			StartTimeSec: start,
		})
	}
	slices.SortFunc(items, func(a, b VideoChapterItemModel) int {
		return cmp.Compare(a.StartTimeSec.ValueFloat64(), b.StartTimeSec.ValueFloat64())
	})
	flat, diags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: videoChapterAttrTypes()}, items)
	return flat, diags
}

func normalizeVideoChapters(ctx context.Context, data *VideoChaptersResourceModel) diag.Diagnostics {
	expanded, diags := expandVideoChapters(ctx, data.Chapters)
	if diags.HasError() {
		return diags
	}
	flat, d := flattenVideoChapters(ctx, expanded)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}
	data.Chapters = flat
	return diags
}

// VideoChaptersResourceSchema defines the rixl_video_chapters collection
// schema. The API returns no timestamps for chapters, so the schema carries
// none as computed attributes.
func VideoChaptersResourceSchema() schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"project_id": schema.StringAttribute{
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"video_id": schema.StringAttribute{
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"chapters": schema.ListNestedAttribute{
				Required: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"title": schema.StringAttribute{
							Required: true,
						},
						"start_time_sec": schema.Float64Attribute{
							Required: true,
						},
					},
				},
			},
		},
	}
}

// VideoChaptersResourceModel is the Terraform state for rixl_video_chapters.
type VideoChaptersResourceModel struct {
	Id        types.String `tfsdk:"id"`
	ProjectId types.String `tfsdk:"project_id"`
	VideoId   types.String `tfsdk:"video_id"`
	Chapters  types.List   `tfsdk:"chapters"`
}

// VideoChapterItemModel is a single entry of the chapters list.
type VideoChapterItemModel struct {
	Title        types.String  `tfsdk:"title"`
	StartTimeSec types.Float64 `tfsdk:"start_time_sec"`
}
