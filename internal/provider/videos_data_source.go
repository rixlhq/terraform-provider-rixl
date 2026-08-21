package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rixlhq/rixl-go/sdk"
	"github.com/rixlhq/rixl-go/sdk/videos"
)

var _ datasource.DataSource = (*videosDataSource)(nil)

func NewVideosDataSource() datasource.DataSource {
	return &videosDataSource{}
}

type videosDataSource struct {
	client *sdk.Client
}

func (d *videosDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_videos"
}

func (d *videosDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = VideosDataSourceSchema(ctx)
}

func (d *videosDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*sdk.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data type", fmt.Sprintf("expected *sdk.Client, got %T", req.ProviderData))
		return
	}
	d.client = client
}

func (d *videosDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data VideosDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := &videos.ListVideosParams{}
	if !data.Paginationlimit.IsNull() && !data.Paginationlimit.IsUnknown() {
		i := int32(data.Paginationlimit.ValueInt64())
		params.PaginationLimit = &i
	}
	if !data.Paginationoffset.IsNull() && !data.Paginationoffset.IsUnknown() {
		i := int32(data.Paginationoffset.ValueInt64())
		params.PaginationOffset = &i
	}
	if !data.SortField.IsNull() && !data.SortField.IsUnknown() {
		s := data.SortField.ValueString()
		params.SortField = &s
	}
	if !data.SortDirection.IsNull() && !data.SortDirection.IsUnknown() {
		s := data.SortDirection.ValueString()
		params.SortDirection = &s
	}

	res, err := d.client.Videos.ListVideos(ctx, data.ProjectId.ValueString(), params)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list videos", err.Error())
		return
	}

	elemType := types.ObjectType{AttrTypes: videoV1AttrTypes()}
	items := make([]attr.Value, 0, len(res.Videos))
	for i := range res.Videos {
		obj, diags := videoV1ToObject(&res.Videos[i])
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		items = append(items, obj)
	}

	videosList, diags := types.ListValue(elemType, items)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	data.Videos = videosList
	data.Limit = ptrInt32(res.Limit)
	data.Offset = ptrInt32(res.Offset)
	data.SortField = ptrString(res.SortField)
	data.SortDirection = ptrString(res.SortDirection)
	if res.Total != nil {
		data.Total = types.StringValue(strconv.FormatInt(*res.Total, 10))
	} else {
		data.Total = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func VideosDataSourceSchema(_ context.Context) dschema.Schema {
	return dschema.Schema{
		Attributes: map[string]dschema.Attribute{
			"project_id": dschema.StringAttribute{
				Required: true,
			},
			"paginationlimit": dschema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				Description:         "Maximum number of items to return.",
				MarkdownDescription: "Maximum number of items to return.",
				Validators: []validator.Int64{
					int64validator.Between(1, 100),
				},
			},
			"paginationoffset": dschema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				Description:         "Number of items to skip before collecting the result set.",
				MarkdownDescription: "Number of items to skip before collecting the result set.",
			},
			"sort_direction": dschema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"sort_field": dschema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"limit": dschema.Int64Attribute{
				Computed:            true,
				Description:         "Maximum number of items returned.",
				MarkdownDescription: "Maximum number of items returned.",
			},
			"offset": dschema.Int64Attribute{
				Computed:            true,
				Description:         "Number of items skipped before this page.",
				MarkdownDescription: "Number of items skipped before this page.",
			},
			"total": dschema.StringAttribute{
				Computed: true,
			},
			"videos": dschema.ListNestedAttribute{
				Computed: true,
				NestedObject: dschema.NestedAttributeObject{
					Attributes: map[string]dschema.Attribute{
						"bitrate":  dschema.Int64Attribute{Computed: true},
						"codec":    dschema.StringAttribute{Computed: true},
						"duration": dschema.StringAttribute{Computed: true},
						"file": dschema.SingleNestedAttribute{
							Computed:   true,
							Attributes: fileDataSourceAttributes(),
						},
						"framerate": dschema.StringAttribute{Computed: true},
						"hdr":       dschema.BoolAttribute{Computed: true},
						"height":    dschema.Int64Attribute{Computed: true},
						"id":        dschema.StringAttribute{Computed: true},
						"poster": dschema.SingleNestedAttribute{
							Computed:   true,
							Attributes: imageDataSourceAttributes(),
						},
						"visibility": dschema.StringAttribute{Computed: true},
						"width":      dschema.Int64Attribute{Computed: true},
					},
				},
			},
		},
	}
}

type VideosDataSourceModel struct {
	Limit            types.Int64  `tfsdk:"limit"`
	Offset           types.Int64  `tfsdk:"offset"`
	Paginationlimit  types.Int64  `tfsdk:"paginationlimit"`
	Paginationoffset types.Int64  `tfsdk:"paginationoffset"`
	ProjectId        types.String `tfsdk:"project_id"`
	SortDirection    types.String `tfsdk:"sort_direction"`
	SortField        types.String `tfsdk:"sort_field"`
	Total            types.String `tfsdk:"total"`
	Videos           types.List   `tfsdk:"videos"`
}
