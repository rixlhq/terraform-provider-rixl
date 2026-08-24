package provider

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rixlhq/rixl-go/sdk"
	"github.com/rixlhq/rixl-go/sdk/feeds"
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

	body, d := modelToMap(ctx, &data)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	delete(body, "id")
	delete(body, "project_id")
	delete(body, "created_at")
	delete(body, "updated_at")

	feed, err := r.client.Feeds.CreateFeed(ctx, projectID, body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create feed", err.Error())
		return
	}

	resp.Diagnostics.Append(mapResponseToModel(ctx, feed, &data, FeedResourceSchema(ctx).Attributes)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.ProjectId.IsNull() || data.ProjectId.IsUnknown() {
		data.ProjectId = types.StringValue(projectID)
	}

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
		var httpErr *feeds.ClientHttpError[struct{}]
		if errors.As(err, &httpErr) && httpErr.StatusCode == http.StatusNotFound {
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
	body["feed_id"] = data.Id.ValueString()
	body["project_id"] = data.ProjectId.ValueString()

	feed, err := r.client.Feeds.UpdateFeed(ctx, data.ProjectId.ValueString(), data.Id.ValueString(), body)
	if err != nil {
		var httpErr *feeds.ClientHttpError[struct{}]
		if errors.As(err, &httpErr) && httpErr.StatusCode == http.StatusNotFound {
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
		var httpErr *feeds.ClientHttpError[struct{}]
		if errors.As(err, &httpErr) && httpErr.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to delete feed", err.Error())
	}
}

func FeedResourceSchema(_ context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"allow_images": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"allow_videos": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"created_at": schema.StringAttribute{
				Computed: true,
			},
			"description": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"has_comments": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"has_likes": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"has_shares": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"id": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"project_id": schema.StringAttribute{
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"updated_at": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

type FeedModel struct {
	AllowImages types.Bool   `tfsdk:"allow_images"`
	AllowVideos types.Bool   `tfsdk:"allow_videos"`
	CreatedAt   types.String `tfsdk:"created_at"`
	Description types.String `tfsdk:"description"`
	HasComments types.Bool   `tfsdk:"has_comments"`
	HasLikes    types.Bool   `tfsdk:"has_likes"`
	HasShares   types.Bool   `tfsdk:"has_shares"`
	Id          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	ProjectId   types.String `tfsdk:"project_id"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}
