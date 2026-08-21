package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rixlhq/rixl-go/sdk"
	"github.com/rixlhq/rixl-go/sdk/models"
)

var _ resource.Resource = (*feedResource)(nil)

func NewFeedResource() resource.Resource {
	return &feedResource{}
}

type feedResource struct {
	client *sdk.Client
}

func (r *feedResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_feed"
}

func (r *feedResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = FeedResourceSchema(ctx)
}

func (r *feedResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*sdk.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data type", fmt.Sprintf("expected *sdk.Client, got %T", req.ProviderData))
		return
	}
	r.client = client
}

func (r *feedResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data FeedModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectId.ValueString()
	body := feedModelToRequestBody(data)

	feed, err := r.client.Feeds.CreateFeed(ctx, projectID, body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create feed", err.Error())
		return
	}

	mapFeedToModel(ctx, feed, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *feedResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data FeedModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	feed, err := r.client.Feeds.GetFeed(ctx, data.ProjectId.ValueString(), data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read feed", err.Error())
		return
	}

	mapFeedToModel(ctx, feed, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *feedResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data FeedModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := feedModelToRequestBody(data)
	feed, err := r.client.Feeds.UpdateFeed(ctx, data.ProjectId.ValueString(), data.Id.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update feed", err.Error())
		return
	}

	mapFeedToModel(ctx, feed, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *feedResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data FeedModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := r.client.Feeds.DeleteFeed(ctx, data.ProjectId.ValueString(), data.Id.ValueString()); err != nil {
		resp.Diagnostics.AddError("Failed to delete feed", err.Error())
	}
}

func feedModelToRequestBody(data FeedModel) map[string]any {
	body := map[string]any{}
	if !data.Name.IsNull() && !data.Name.IsUnknown() {
		body["name"] = data.Name.ValueString()
	}
	if !data.Description.IsNull() && !data.Description.IsUnknown() {
		body["description"] = data.Description.ValueString()
	}
	if !data.AllowImages.IsNull() && !data.AllowImages.IsUnknown() {
		body["allow_images"] = data.AllowImages.ValueBool()
	}
	if !data.AllowVideos.IsNull() && !data.AllowVideos.IsUnknown() {
		body["allow_videos"] = data.AllowVideos.ValueBool()
	}
	if !data.HasComments.IsNull() && !data.HasComments.IsUnknown() {
		body["has_comments"] = data.HasComments.ValueBool()
	}
	if !data.HasLikes.IsNull() && !data.HasLikes.IsUnknown() {
		body["has_likes"] = data.HasLikes.ValueBool()
	}
	if !data.HasShares.IsNull() && !data.HasShares.IsUnknown() {
		body["has_shares"] = data.HasShares.ValueBool()
	}
	return body
}

func mapFeedToModel(_ context.Context, feed models.FeedsV1Feed, data *FeedModel) {
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

func timestampToString(v any) types.String {
	switch t := v.(type) {
	case time.Time:
		if t.IsZero() {
			return types.StringNull()
		}
		return types.StringValue(t.Format(time.RFC3339Nano))
	case *time.Time:
		if t == nil || t.IsZero() {
			return types.StringNull()
		}
		return types.StringValue(t.Format(time.RFC3339Nano))
	case string:
		if t == "" {
			return types.StringNull()
		}
		return types.StringValue(t)
	case *string:
		if t == nil || *t == "" {
			return types.StringNull()
		}
		return types.StringValue(*t)
	default:
		return types.StringNull()
	}
}
