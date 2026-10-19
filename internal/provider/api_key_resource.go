package provider

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rixlhq/rixl-go/sdk"
	"github.com/rixlhq/rixl-go/sdk/models"
)

var (
	_ resource.Resource                = &apiKeyResource{}
	_ resource.ResourceWithConfigure   = &apiKeyResource{}
	_ resource.ResourceWithImportState = &apiKeyResource{}
)

func NewApiKeyResource() resource.Resource {
	return &apiKeyResource{}
}

type apiKeyResource struct {
	client *sdk.Client
}

func (r *apiKeyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_key"
}

func (r *apiKeyResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = ApiKeyResourceSchema(ctx)
}

func (r *apiKeyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*sdk.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", "Expected *sdk.Client")
		return
	}
	r.client = client
}

func (r *apiKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ApiKeyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	body := models.ApikeysV1CreateAPIKeyRequest{
		Name: &name,
	}
	if !plan.ProjectId.IsNull() && !plan.ProjectId.IsUnknown() {
		pid := plan.ProjectId.ValueString()
		body.ProjectID = &pid
	}

	res, err := r.client.APIKeys.CreateApiKey(ctx, plan.OrgId.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create API key", err.Error())
		return
	}

	var data ApiKeyModel
	resp.Diagnostics.Append(mapResponseToModel(ctx, res.APIKey, &data, ApiKeyResourceSchema(ctx).Attributes)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if res.APIKey != nil && res.APIKey.Secret != nil {
		data.Secret = types.StringValue(*res.APIKey.Secret)
	}
	data.OrgId = plan.OrgId

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *apiKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ApiKeyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.client.APIKeys.ListApiKeys(ctx, state.OrgId.ValueString(), nil)
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to list API keys", err.Error())
		return
	}

	keyID := state.Id.ValueString()
	var found any
	for _, key := range res.APIKeys {
		if key.ID != nil && *key.ID == keyID {
			found = key
			break
		}
	}
	if found == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	savedSecret := state.Secret
	var data ApiKeyModel
	resp.Diagnostics.Append(mapResponseToModel(ctx, found, &data, ApiKeyResourceSchema(ctx).Attributes)...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Secret = savedSecret

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *apiKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ApiKeyModel
	var state ApiKeyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Name.Equal(state.Name) {
		body := map[string]any{
			"name": plan.Name.ValueString(),
		}
		_, err := r.client.APIKeys.UpdateApiKey(ctx, state.OrgId.ValueString(), state.Id.ValueString(), body)
		if err != nil {
			resp.Diagnostics.AddError("Failed to update API key", err.Error())
			return
		}
	}

	savedSecret := state.Secret
	if !savedSecret.IsNull() && !savedSecret.IsUnknown() {
		plan.Secret = savedSecret
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *apiKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ApiKeyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.APIKeys.DeleteApiKey(ctx, data.OrgId.ValueString(), data.Id.ValueString())
	if err != nil && !isNotFound(err) {
		resp.Diagnostics.AddError("Failed to delete API key", err.Error())
	}
}

func (r *apiKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "/")
	if len(parts) == 2 {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("org_id"), parts[0])...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}
