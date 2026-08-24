package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func DashboardWidgetResourceSchema() schema.Schema {
	filterAttrs := map[string]schema.Attribute{
		"field":    schema.StringAttribute{Optional: true},
		"operator": schema.StringAttribute{Optional: true},
		"values": schema.ListAttribute{
			ElementType: types.StringType,
			Optional:    true,
		},
	}

	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"dashboard_id": schema.StringAttribute{
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"title": schema.StringAttribute{
				Required: true,
			},
			"chart_type": schema.StringAttribute{
				Required: true,
			},
			"dataset": schema.StringAttribute{
				Required: true,
			},
			"metric": schema.StringAttribute{
				Required: true,
			},
			"group_by": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
			},
			"filters": schema.ListNestedAttribute{
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: filterAttrs,
				},
			},
			"interval": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"limit": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"pos_x": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"pos_y": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"width": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"height": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"sort_order": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"created_at": schema.StringAttribute{
				Computed: true,
			},
			"updated_at": schema.StringAttribute{
				Computed: true,
			},
			"spec_version": schema.Int64Attribute{
				Computed: true,
			},
			"dashboard_revision": schema.Int64Attribute{
				Computed: true,
			},
		},
	}
}

type DashboardWidgetResourceModel struct {
	Id                types.String `tfsdk:"id"`
	DashboardId       types.String `tfsdk:"dashboard_id"`
	Title             types.String `tfsdk:"title"`
	ChartType         types.String `tfsdk:"chart_type"`
	Dataset           types.String `tfsdk:"dataset"`
	Metric            types.String `tfsdk:"metric"`
	GroupBy           types.List   `tfsdk:"group_by"`
	Filters           types.List   `tfsdk:"filters"`
	Interval          types.String `tfsdk:"interval"`
	Limit             types.Int64  `tfsdk:"limit"`
	PosX              types.Int64  `tfsdk:"pos_x"`
	PosY              types.Int64  `tfsdk:"pos_y"`
	Width             types.Int64  `tfsdk:"width"`
	Height            types.Int64  `tfsdk:"height"`
	SortOrder         types.Int64  `tfsdk:"sort_order"`
	CreatedAt         types.String `tfsdk:"created_at"`
	UpdatedAt         types.String `tfsdk:"updated_at"`
	SpecVersion       types.Int64  `tfsdk:"spec_version"`
	DashboardRevision types.Int64  `tfsdk:"dashboard_revision"`
}
