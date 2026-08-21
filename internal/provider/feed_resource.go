package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/rixlhq/rixl-go/sdk"
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

	body, d := modelToMap(ctx, &data)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	delete(body, "id")
	delete(body, "project_id")

	feed, err := r.client.Feeds.CreateFeed(ctx, data.ProjectId.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create feed", err.Error())
		return
	}

	resp.Diagnostics.Append(mapResponseToModel(ctx, feed, &data, FeedResourceSchema(ctx).Attributes)...)
	if resp.Diagnostics.HasError() {
		return
	}

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

	resp.Diagnostics.Append(mapResponseToModel(ctx, feed, &data, FeedResourceSchema(ctx).Attributes)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *feedResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data FeedModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, d := modelToMap(ctx, &data)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	delete(body, "id")
	delete(body, "project_id")

	feed, err := r.client.Feeds.UpdateFeed(ctx, data.ProjectId.ValueString(), data.Id.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update feed", err.Error())
		return
	}

	resp.Diagnostics.Append(mapResponseToModel(ctx, feed, &data, FeedResourceSchema(ctx).Attributes)...)
	if resp.Diagnostics.HasError() {
		return
	}

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
