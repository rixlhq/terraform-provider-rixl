package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rixlhq/rixl-go/sdk"
	"github.com/rixlhq/rixl-go/sdk/posts"
)

func NewPostsDataSource() datasource.DataSource {
	return &postsDataSource{}
}

type postsDataSource struct {
	client *sdk.Client
}

func (d *postsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_posts"
}

func (d *postsDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = PostsDataSourceSchema(ctx)
}

func (d *postsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *postsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data PostsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := &posts.ListPostsParams{}
	if !data.CreatorId.IsNull() && !data.CreatorId.IsUnknown() {
		s := data.CreatorId.ValueString()
		params.CreatorId = &s
	}
	if !data.Paginationlimit.IsNull() && !data.Paginationlimit.IsUnknown() {
		i := int32(data.Paginationlimit.ValueInt64())
		params.PaginationLimit = &i
	}
	if !data.Paginationoffset.IsNull() && !data.Paginationoffset.IsUnknown() {
		i := int32(data.Paginationoffset.ValueInt64())
		params.PaginationOffset = &i
	}

	res, err := d.client.Posts.ListPosts(ctx, data.ProjectId.ValueString(), data.FeedId.ValueString(), params)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list posts", err.Error())
		return
	}

	resp.Diagnostics.Append(mapResponseToModel(ctx, res, &data, PostsDataSourceSchema(ctx).Attributes)...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.Limit = ptrInt32(res.Limit)
	data.Offset = ptrInt32(res.Offset)
	if res.Total != nil {
		data.Total = types.StringValue(strconv.FormatInt(*res.Total, 10))
	} else {
		data.Total = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func postDataSourceAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed: true,
		},
		"feed_id": schema.StringAttribute{
			Computed: true,
		},
		"org_id": schema.StringAttribute{
			Computed: true,
		},
		"creator_id": schema.StringAttribute{
			Computed: true,
		},
		"type": schema.StringAttribute{
			Computed: true,
		},
		"description": schema.StringAttribute{
			Computed: true,
		},
		"created_at": schema.StringAttribute{
			Computed: true,
		},
		"image": schema.SingleNestedAttribute{
			Computed:   true,
			Attributes: imageDataSourceAttributes(),
		},
		"video": schema.SingleNestedAttribute{
			Computed:   true,
			Attributes: videoDataSourceAttributes(),
		},
	}
}

func PostsDataSourceSchema(_ context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				Required: true,
			},
			"feed_id": schema.StringAttribute{
				Required: true,
			},
			"creator_id": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"paginationlimit": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"paginationoffset": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"limit": schema.Int64Attribute{
				Computed: true,
			},
			"offset": schema.Int64Attribute{
				Computed: true,
			},
			"total": schema.StringAttribute{
				Computed: true,
			},
			"posts": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: postDataSourceAttributes(),
				},
			},
		},
	}
}

type PostsDataSourceModel struct {
	ProjectId        types.String `tfsdk:"project_id"`
	FeedId           types.String `tfsdk:"feed_id"`
	CreatorId        types.String `tfsdk:"creator_id"`
	Paginationlimit  types.Int64  `tfsdk:"paginationlimit"`
	Paginationoffset types.Int64  `tfsdk:"paginationoffset"`
	Limit            types.Int64  `tfsdk:"limit"`
	Offset           types.Int64  `tfsdk:"offset"`
	Total            types.String `tfsdk:"total"`
	Posts            types.List   `tfsdk:"posts"`
}
