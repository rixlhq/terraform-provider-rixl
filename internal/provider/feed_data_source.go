package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rixlhq/rixl-go/sdk"
	"github.com/rixlhq/rixl-go/sdk/feeds"
	"github.com/rixlhq/rixl-go/sdk/models"
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

	mapFeedDataSourceToModel(feed, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func mapFeedDataSourceToModel(feed models.FeedsV1Feed, data *FeedDataSourceModel) {
	data.Id = types.StringPointerValue(feed.ID)
	data.ProjectId = types.StringPointerValue(feed.ProjectID)
	data.Name = types.StringPointerValue(feed.Name)
	data.Description = types.StringPointerValue(feed.Description)
	data.AllowImages = types.BoolPointerValue(feed.AllowImages)
	data.AllowVideos = types.BoolPointerValue(feed.AllowVideos)
	data.HasComments = types.BoolPointerValue(feed.HasComments)
	data.HasLikes = types.BoolPointerValue(feed.HasLikes)
	data.HasShares = types.BoolPointerValue(feed.HasShares)
	data.CreatedAt = timestampToString(feed.CreatedAt)
	data.UpdatedAt = timestampToString(feed.UpdatedAt)
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
	var data FeedsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := &feeds.ListFeedsParams{}
	if !data.Paginationlimit.IsNull() && !data.Paginationlimit.IsUnknown() {
		limit := int32(data.Paginationlimit.ValueInt64())
		params.PaginationLimit = &limit
	}
	if !data.Paginationoffset.IsNull() && !data.Paginationoffset.IsUnknown() {
		offset := int32(data.Paginationoffset.ValueInt64())
		params.PaginationOffset = &offset
	}

	list, err := d.client.Feeds.ListFeeds(ctx, data.ProjectId.ValueString(), params)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list feeds", err.Error())
		return
	}

	if list.Total != nil {
		data.Total = types.StringValue(strconv.FormatInt(*list.Total, 10))
	} else {
		data.Total = types.StringNull()
	}

	attrTypes := FeedsValue{}.AttributeTypes(ctx)
	elementType := FeedsType{ObjectType: types.ObjectType{AttrTypes: attrTypes}}
	elements := make([]attr.Value, 0, len(list.Feeds))
	for _, f := range list.Feeds {
		elements = append(elements, newFeedsValue(ctx, attrTypes, f))
	}
	data.Feeds, _ = types.ListValue(elementType, elements)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func newFeedsValue(_ context.Context, attrTypes map[string]attr.Type, f models.FeedsV1Feed) FeedsValue {
	attributes := map[string]attr.Value{
		"allow_images": types.BoolPointerValue(f.AllowImages),
		"allow_videos": types.BoolPointerValue(f.AllowVideos),
		"created_at":   timestampToString(f.CreatedAt),
		"description":  types.StringPointerValue(f.Description),
		"has_comments": types.BoolPointerValue(f.HasComments),
		"has_likes":    types.BoolPointerValue(f.HasLikes),
		"has_shares":   types.BoolPointerValue(f.HasShares),
		"id":           types.StringPointerValue(f.ID),
		"name":         types.StringPointerValue(f.Name),
		"project_id":   types.StringPointerValue(f.ProjectID),
		"updated_at":   timestampToString(f.UpdatedAt),
	}
	return NewFeedsValueMust(attrTypes, attributes)
}
