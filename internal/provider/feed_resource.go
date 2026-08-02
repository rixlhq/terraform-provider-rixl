package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/rixlhq/terraform-provider-rixl/internal/providerdata"
	"github.com/rixlhq/terraform-provider-rixl/internal/rixlclient"
	"github.com/rixlhq/terraform-provider-rixl/internal/rixlcommon"
)

var (
	_ resource.Resource                = &feedResource{}
	_ resource.ResourceWithImportState = &feedResource{}
)

// NewFeedResource returns a resource that manages Rixl feeds.
func NewFeedResource() resource.Resource {
	return &feedResource{}
}

type feedResource struct {
	client *rixlclient.Client
}

// Metadata returns the full resource type name.
func (r *feedResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_feed"
}

// Schema returns the feed resource schema.
func (r *feedResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	s := FeedResourceSchema(ctx)

	if attr, ok := s.Attributes["project_id"].(schema.StringAttribute); ok {
		attr.Optional = false
		attr.Computed = false
		attr.Required = true
		s.Attributes["project_id"] = attr
	}
	if attr, ok := s.Attributes["name"].(schema.StringAttribute); ok {
		attr.Optional = false
		attr.Computed = false
		attr.Required = true
		s.Attributes["name"] = attr
	}

	resp.Schema = s
}

// Configure prepares the API client for the resource.
func (r *feedResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	pd, ok := req.ProviderData.(*providerdata.Data)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Configure Type", fmt.Sprintf("Expected *providerdata.Data, got %T", req.ProviderData))
		return
	}
	r.client = pd.Client
}

// Create creates a new feed.
func (r *feedResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing Client", "Configure the provider to use this resource.")
		return
	}

	var plan FeedModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.ProjectId.IsNull() || plan.ProjectId.IsUnknown() || plan.ProjectId.ValueString() == "" {
		resp.Diagnostics.AddError("Missing project_id", "project_id is required to create a feed.")
		return
	}
	if plan.Name.IsNull() || plan.Name.IsUnknown() || plan.Name.ValueString() == "" {
		resp.Diagnostics.AddError("Missing name", "name is required to create a feed.")
		return
	}

	path := fmt.Sprintf("/feeds/v1/projects/%s/feeds", plan.ProjectId.ValueString())
	body, err := json.Marshal(feedRequestBody(plan, false))
	if err != nil {
		resp.Diagnostics.AddError("Request Encoding Error", err.Error())
		return
	}

	respBody, err := r.client.Post(ctx, path, body)
	if err != nil {
		resp.Diagnostics.AddError("Rixl API Error", err.Error())
		return
	}

	stateVal, err := r.responseToState(ctx, respBody, req.Plan.Raw)
	if err != nil {
		resp.Diagnostics.AddError("Response Conversion Error", err.Error())
		return
	}
	resp.State.Raw = stateVal
}

// Read refreshes the feed state from the API.
func (r *feedResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing Client", "Configure the provider to use this resource.")
		return
	}

	path, err := r.buildPath(req.State.Raw, "/feeds/v1/projects/{project_id}/feeds/{feed_id}")
	if err != nil {
		resp.Diagnostics.AddError("Path Error", err.Error())
		return
	}

	respBody, err := r.client.Get(ctx, path, nil)
	if err != nil {
		if rixlclient.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Rixl API Error", err.Error())
		return
	}

	stateVal, err := r.responseToState(ctx, respBody, req.State.Raw)
	if err != nil {
		resp.Diagnostics.AddError("Response Conversion Error", err.Error())
		return
	}
	resp.State.Raw = stateVal
}

// Update modifies an existing feed.
func (r *feedResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing Client", "Configure the provider to use this resource.")
		return
	}

	var plan FeedModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.ProjectId.IsNull() || plan.ProjectId.IsUnknown() || plan.ProjectId.ValueString() == "" {
		resp.Diagnostics.AddError("Missing project_id", "project_id is required to update a feed.")
		return
	}
	if plan.Id.IsNull() || plan.Id.IsUnknown() || plan.Id.ValueString() == "" {
		resp.Diagnostics.AddError("Missing id", "feed id is required to update a feed.")
		return
	}

	path := fmt.Sprintf("/feeds/v1/projects/%s/feeds/%s", plan.ProjectId.ValueString(), plan.Id.ValueString())
	body, err := json.Marshal(feedRequestBody(plan, true))
	if err != nil {
		resp.Diagnostics.AddError("Request Encoding Error", err.Error())
		return
	}

	respBody, err := r.client.Put(ctx, path, body)
	if err != nil {
		resp.Diagnostics.AddError("Rixl API Error", err.Error())
		return
	}

	stateVal, err := r.responseToState(ctx, respBody, req.Plan.Raw)
	if err != nil {
		resp.Diagnostics.AddError("Response Conversion Error", err.Error())
		return
	}
	resp.State.Raw = stateVal
}

// Delete removes a feed.
func (r *feedResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing Client", "Configure the provider to use this resource.")
		return
	}

	path, err := r.buildPath(req.State.Raw, "/feeds/v1/projects/{project_id}/feeds/{feed_id}")
	if err != nil {
		resp.Diagnostics.AddError("Path Error", err.Error())
		return
	}

	if _, err := r.client.Delete(ctx, path); err != nil && !rixlclient.IsNotFound(err) {
		resp.Diagnostics.AddError("Rixl API Error", err.Error())
		return
	}

	resp.State.RemoveResource(ctx)
}

// ImportState imports a feed using a project_id/feed_id format.
func (r *feedResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "/") // project_id/feed_id
	if len(parts) != 2 {
		resp.Diagnostics.AddError("Invalid Import ID", "Expected project_id/feed_id.")
		return
	}

	tfType := feedResourceSchemaType(ctx)
	objType := tfType.(tftypes.Object)
	vals := make(map[string]tftypes.Value, len(objType.AttributeTypes))
	for attr, attrType := range objType.AttributeTypes {
		switch attr {
		case "project_id":
			vals[attr] = tftypes.NewValue(attrType, parts[0])
		case "id":
			vals[attr] = tftypes.NewValue(attrType, parts[1])
		default:
			vals[attr] = tftypes.NewValue(attrType, nil)
		}
	}
	importVal := tftypes.NewValue(tfType, vals)

	path := fmt.Sprintf("/feeds/v1/projects/%s/feeds/%s", parts[0], parts[1])
	respBody, err := r.client.Get(ctx, path, nil)
	if err != nil {
		if rixlclient.IsNotFound(err) {
			resp.Diagnostics.AddError("Import Not Found", err.Error())
			return
		}
		resp.Diagnostics.AddError("Rixl API Error", err.Error())
		return
	}

	stateVal, err := r.responseToState(ctx, respBody, importVal)
	if err != nil {
		resp.Diagnostics.AddError("Response Conversion Error", err.Error())
		return
	}
	resp.State.Raw = stateVal
}

func (r *feedResource) buildPath(value tftypes.Value, template string) (string, error) {
	return buildPath(value, template, map[string]string{"feed_id": "id"})
}

func (r *feedResource) responseToState(ctx context.Context, body []byte, base tftypes.Value) (tftypes.Value, error) {
	tfVal, err := responseToState(ctx, feedResourceSchemaType(ctx), body, "")
	if err != nil {
		return tftypes.Value{}, err
	}
	return rixlcommon.OverlayKnown(base, tfVal), nil
}

func feedResourceSchemaType(ctx context.Context) tftypes.Type {
	return FeedResourceSchema(ctx).Type().TerraformType(ctx)
}

func feedRequestBody(plan FeedModel, isUpdate bool) map[string]any {
	body := make(map[string]any)

	body["project_id"] = plan.ProjectId.ValueString()
	body["name"] = plan.Name.ValueString()

	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		body["description"] = plan.Description.ValueString()
	}

	addBool := func(key string, v types.Bool) {
		if !v.IsNull() && !v.IsUnknown() {
			body[key] = v.ValueBool()
		}
	}
	addBool("allow_images", plan.AllowImages)
	addBool("allow_videos", plan.AllowVideos)
	addBool("has_likes", plan.HasLikes)
	addBool("has_shares", plan.HasShares)
	addBool("has_comments", plan.HasComments)

	if isUpdate {
		body["feed_id"] = plan.Id.ValueString()
	}

	return body
}
