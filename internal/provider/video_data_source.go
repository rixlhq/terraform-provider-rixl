package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/rixlhq/rixl-go/sdk"
)

var _ datasource.DataSource = (*videoDataSource)(nil)

func NewVideoDataSource() datasource.DataSource {
	return &videoDataSource{}
}

type videoDataSource struct {
	client *sdk.Client
}

func (d *videoDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_video"
}

func (d *videoDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = VideoDataSourceSchema(ctx)
}

func (d *videoDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *videoDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data VideoDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := d.client.Videos.GetVideo(ctx, data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read video", err.Error())
		return
	}

	obj, diags := videoV1ToObject(res.Video)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(obj.As(ctx, &data, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func videoDataSourceAttributes() map[string]dschema.Attribute {
	return map[string]dschema.Attribute{
		"id": dschema.StringAttribute{
			Computed: true,
		},
		"bitrate": dschema.Int64Attribute{
			Computed: true,
		},
		"codec": dschema.StringAttribute{
			Computed: true,
		},
		"duration": dschema.StringAttribute{
			Computed: true,
		},
		"file": dschema.SingleNestedAttribute{
			Computed:   true,
			Attributes: fileDataSourceAttributes(),
		},
		"framerate": dschema.StringAttribute{
			Computed: true,
		},
		"hdr": dschema.BoolAttribute{
			Computed: true,
		},
		"height": dschema.Int64Attribute{
			Computed: true,
		},
		"poster": dschema.SingleNestedAttribute{
			Computed:   true,
			Attributes: imageDataSourceAttributes(),
		},
		"visibility": dschema.StringAttribute{
			Computed: true,
		},
		"width": dschema.Int64Attribute{
			Computed: true,
		},
	}
}

func VideoDataSourceSchema(_ context.Context) dschema.Schema {
	attrs := videoDataSourceAttributes()
	attrs["id"] = dschema.StringAttribute{Required: true}
	return dschema.Schema{
		Attributes: attrs,
	}
}

type VideoDataSourceModel struct {
	Bitrate    types.Int64  `tfsdk:"bitrate"`
	Codec      types.String `tfsdk:"codec"`
	Duration   types.String `tfsdk:"duration"`
	File       types.Object `tfsdk:"file"`
	Framerate  types.String `tfsdk:"framerate"`
	Hdr        types.Bool   `tfsdk:"hdr"`
	Height     types.Int64  `tfsdk:"height"`
	Id         types.String `tfsdk:"id"`
	Poster     types.Object `tfsdk:"poster"`
	Visibility types.String `tfsdk:"visibility"`
	Width      types.Int64  `tfsdk:"width"`
}
