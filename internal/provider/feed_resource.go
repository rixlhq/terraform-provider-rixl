package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rixlhq/rixl-go/sdk"
)

var (
	_ resource.Resource                = (*feedResource)(nil)
	_ resource.ResourceWithImportState = (*feedResource)(nil)
)

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

	body, d := modelToMap(ctx, &data)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	delete(body, "id")
	delete(body, "project_id")

	feed, err := r.client.Feeds.CreateFeed(ctx, projectID, body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create feed", err.Error())
		return
	}

	resp.Diagnostics.Append(mapResponseToModel(ctx, feed, &data, FeedResourceSchema(ctx).Attributes)...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.ProjectId = types.StringValue(projectID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *feedResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data FeedModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectId.ValueString()
	feedID := data.Id.ValueString()

	feed, err := r.client.Feeds.GetFeed(ctx, projectID, feedID)
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read feed", err.Error())
		return
	}

	resp.Diagnostics.Append(mapResponseToModel(ctx, feed, &data, FeedResourceSchema(ctx).Attributes)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.ProjectId.IsNull() || data.ProjectId.IsUnknown() {
		data.ProjectId = types.StringValue(projectID)
	}
	if data.Id.IsNull() || data.Id.IsUnknown() {
		data.Id = types.StringValue(feedID)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *feedResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state, plan FeedModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataAny, diags := mergeStateAndPlan(ctx, &state, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data := dataAny.(*FeedModel)

	body, d := modelToMap(ctx, data)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	delete(body, "id")
	delete(body, "created_at")
	delete(body, "updated_at")
	// The API body identifies the resource by feed_id rather than id.
	tbodyFeedID := data.Id.ValueString()
	body["feed_id"] = tbodyFeedID
	body["project_id"] = data.ProjectId.ValueString()

	feed, err := r.client.Feeds.UpdateFeed(ctx, data.ProjectId.ValueString(), data.Id.ValueString(), body)
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to update feed", err.Error())
		return
	}

	resp.Diagnostics.Append(mapResponseToModel(ctx, feed, data, FeedResourceSchema(ctx).Attributes)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.ProjectId.IsNull() || data.ProjectId.IsUnknown() {
		data.ProjectId = state.ProjectId
	}
	if data.Id.IsNull() || data.Id.IsUnknown() {
		data.Id = state.Id
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *feedResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data FeedModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := r.client.Feeds.DeleteFeed(ctx, data.ProjectId.ValueString(), data.Id.ValueString()); err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to delete feed", err.Error())
	}
}

func (r *feedResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "/")
	if len(parts) == 2 {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_id"), parts[0])...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
		return
	}
	resp.Diagnostics.AddError(
		"Unexpected Import Identifier",
		fmt.Sprintf("Expected import identifier with format: project_id/id. Got: %q", req.ID),
	)
}
