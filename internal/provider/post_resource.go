package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/rixlhq/rixl-go/sdk"
)

var (
	_ resource.Resource                = (*postResource)(nil)
	_ resource.ResourceWithImportState = (*postResource)(nil)
)

func NewPostResource() resource.Resource {
	return &postResource{}
}

type postResource struct {
	client *sdk.Client
}

func (r *postResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_post"
}

func (r *postResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = PostResourceSchema(ctx)
}

func (r *postResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *postResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data PostModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, diags := modelToMap(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	delete(body, "project_id")
	delete(body, "feed_id")
	delete(body, "id")
	delete(body, "created_at")

	res, err := r.client.Posts.CreatePost(ctx, data.ProjectId.ValueString(), data.FeedId.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create post", err.Error())
		return
	}

	resp.Diagnostics.Append(mapResponseToModel(ctx, res.Post, &data, PostResourceSchema(ctx).Attributes)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *postResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data PostModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.client.Posts.GetPost(ctx, data.ProjectId.ValueString(), data.Id.ValueString(), nil)
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read post", err.Error())
		return
	}

	resp.Diagnostics.Append(mapResponseToModel(ctx, res.Post, &data, PostResourceSchema(ctx).Attributes)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *postResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Post attributes require replacement; in-place update is not supported.
	var plan PostModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *postResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data PostModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.Posts.DeletePost(ctx, data.ProjectId.ValueString(), data.Id.ValueString(), nil)
	if err != nil && !isNotFound(err) {
		resp.Diagnostics.AddError("Failed to delete post", err.Error())
	}
}

func (r *postResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "/")
	if len(parts) == 2 {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_id"), parts[0])...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}
