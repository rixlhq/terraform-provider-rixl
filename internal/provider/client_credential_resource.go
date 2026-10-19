package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rixlhq/rixl-go/sdk"
	"github.com/rixlhq/rixl-go/sdk/models"
)

var (
	_ resource.Resource                = (*clientCredentialResource)(nil)
	_ resource.ResourceWithImportState = (*clientCredentialResource)(nil)
)

func NewClientCredentialResource() resource.Resource {
	return &clientCredentialResource{}
}

type clientCredentialResource struct {
	client *sdk.Client
}

func (r *clientCredentialResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_client_credential"
}

func (r *clientCredentialResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = ClientCredentialResourceSchema(ctx)
}

func (r *clientCredentialResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *clientCredentialResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ClientCredentialModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := models.ClientauthV1CreateClientCredentialRequest{
		Name: data.Name.ValueString(),
	}
	if !data.OrgId.IsNull() && !data.OrgId.IsUnknown() {
		org := data.OrgId.ValueString()
		body.OrgID = &org
	}
	if !data.Alg.IsNull() && !data.Alg.IsUnknown() {
		alg := data.Alg.ValueString()
		body.Alg = &alg
	}

	res, err := r.client.ClientCredentials.CreateClientCredential(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create client credential", err.Error())
		return
	}

	resp.Diagnostics.Append(mapResponseToModel(ctx, res.Credential, &data, ClientCredentialResourceSchema(ctx).Attributes)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if res.ClientSecret != nil {
		data.ClientSecret = types.StringValue(*res.ClientSecret)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *clientCredentialResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ClientCredentialModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.client.ClientCredentials.ListClientCredentials(ctx, nil)
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to list client credentials", err.Error())
		return
	}

	var found any
	for _, cred := range res.Credentials {
		if cred.ID != nil && *cred.ID == data.Id.ValueString() {
			found = cred
			break
		}
	}
	if found == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	savedSecret := data.ClientSecret
	resp.Diagnostics.Append(mapResponseToModel(ctx, found, &data, ClientCredentialResourceSchema(ctx).Attributes)...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.ClientSecret = savedSecret

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *clientCredentialResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ClientCredentialModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *clientCredentialResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ClientCredentialModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.ClientCredentials.RevokeClientCredential(ctx, data.Id.ValueString(), nil)
	if err != nil && !isNotFound(err) {
		resp.Diagnostics.AddError("Failed to revoke client credential", err.Error())
	}
}

func (r *clientCredentialResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}
