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
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/rixlhq/rixl-go/sdk"
	"github.com/rixlhq/rixl-go/sdk/images"
)

var _ datasource.DataSource = (*imageDataSource)(nil)
var _ datasource.DataSource = (*imagesDataSource)(nil)

func NewImageDataSource() datasource.DataSource {
	return &imageDataSource{}
}

func NewImagesDataSource() datasource.DataSource {
	return &imagesDataSource{}
}

type imageDataSource struct {
	client *sdk.Client
}

type imagesDataSource struct {
	client *sdk.Client
}

func (d *imageDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_image"
}

func (d *imageDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = ImageDataSourceSchema(ctx)
}

func (d *imageDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *imageDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ImageDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := d.client.Images.GetImage(ctx, data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read image", err.Error())
		return
	}

	obj, diags := imageV1ToObject(res.Image)
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

func (d *imagesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_images"
}

func (d *imagesDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = ImagesDataSourceSchema(ctx)
}

func (d *imagesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *imagesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ImagesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := &images.ListImagesParams{}
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

	res, err := d.client.Images.ListImages(ctx, data.ProjectId.ValueString(), params)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list images", err.Error())
		return
	}

	elemType := types.ObjectType{AttrTypes: imageV1AttrTypes()}
	items := make([]attr.Value, 0, len(res.Images))
	for i := range res.Images {
		obj, diags := imageV1ToObject(&res.Images[i])
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		items = append(items, obj)
	}

	imagesList, diags := types.ListValue(elemType, items)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	data.Images = imagesList
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

func ImageDataSourceSchema(_ context.Context) dschema.Schema {
	return dschema.Schema{
		Attributes: map[string]dschema.Attribute{
			"id": dschema.StringAttribute{
				Required: true,
			},
			"attached_to_video": dschema.BoolAttribute{
				Computed: true,
			},
			"file": dschema.SingleNestedAttribute{
				Computed:   true,
				Attributes: fileDataSourceAttributes(),
			},
			"height": dschema.Int64Attribute{
				Computed: true,
			},
			"thumbhash": dschema.StringAttribute{
				Computed: true,
			},
			"visibility": dschema.StringAttribute{
				Computed: true,
			},
			"width": dschema.Int64Attribute{
				Computed: true,
			},
		},
	}
}

func ImagesDataSourceSchema(_ context.Context) dschema.Schema {
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
			"images": dschema.ListNestedAttribute{
				Computed: true,
				NestedObject: dschema.NestedAttributeObject{
					Attributes: imageDataSourceAttributes(),
				},
			},
		},
	}
}

type ImageDataSourceModel struct {
	AttachedToVideo types.Bool   `tfsdk:"attached_to_video"`
	File            types.Object `tfsdk:"file"`
	Height          types.Int64  `tfsdk:"height"`
	Id              types.String `tfsdk:"id"`
	Thumbhash       types.String `tfsdk:"thumbhash"`
	Visibility      types.String `tfsdk:"visibility"`
	Width           types.Int64  `tfsdk:"width"`
}

type ImagesDataSourceModel struct {
	Images           types.List   `tfsdk:"images"`
	Limit            types.Int64  `tfsdk:"limit"`
	Offset           types.Int64  `tfsdk:"offset"`
	Paginationlimit  types.Int64  `tfsdk:"paginationlimit"`
	Paginationoffset types.Int64  `tfsdk:"paginationoffset"`
	ProjectId        types.String `tfsdk:"project_id"`
	SortDirection    types.String `tfsdk:"sort_direction"`
	SortField        types.String `tfsdk:"sort_field"`
	Total            types.String `tfsdk:"total"`
}
