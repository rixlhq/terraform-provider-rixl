package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/rixlhq/rixl-go/sdk"
	"github.com/rixlhq/rixl-go/sdk/feeds"
)

var _ datasource.DataSource = (*feedDataSource)(nil)

func NewFeedDataSource() datasource.DataSource {
	return &feedDataSource{}
}

type feedDataSource struct {
	client *sdk.Client
}

func (d *feedDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_feed"
}

func (d *feedDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = FeedDataSourceSchema(ctx)
}

func (d *feedDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *feedDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data FeedDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	feed, err := d.client.Feeds.GetFeed(ctx, data.ProjectId.ValueString(), data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read feed", err.Error())
		return
	}

	resp.Diagnostics.Append(mapResponseToModel(ctx, feed, &data, FeedDataSourceSchema(ctx).Attributes)...)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

var _ datasource.DataSource = (*feedsDataSource)(nil)

func NewFeedsDataSource() datasource.DataSource {
	return &feedsDataSource{}
}

type feedsDataSource struct {
	client *sdk.Client
}

func (d *feedsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_feeds"
}

func (d *feedsDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = FeedsDataSourceSchema(ctx)
}

func (d *feedsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *feedsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data FeedsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := &feeds.ListFeedsParams{}
	if !data.Paginationlimit.IsNull() && !data.Paginationlimit.IsUnknown() {
		i := int32(data.Paginationlimit.ValueInt64())
		params.PaginationLimit = &i
	}
	if !data.Paginationoffset.IsNull() && !data.Paginationoffset.IsUnknown() {
		i := int32(data.Paginationoffset.ValueInt64())
		params.PaginationOffset = &i
	}

	list, err := d.client.Feeds.ListFeeds(ctx, data.ProjectId.ValueString(), params)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list feeds", err.Error())
		return
	}

	resp.Diagnostics.Append(mapResponseToModel(ctx, list, &data, FeedsDataSourceSchema(ctx).Attributes)...)

	m, err := responseToMap(list)
	if err != nil {
		resp.Diagnostics.AddError("Failed to convert list response", err.Error())
		return
	}
	resp.Diagnostics.Append(mapResponsePagination(ctx, m, &data)...)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
