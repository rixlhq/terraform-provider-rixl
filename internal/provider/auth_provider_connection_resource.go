package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/rixlhq/rixl-go/sdk"
	"github.com/rixlhq/rixl-go/sdk/models"
	"github.com/rixlhq/rixl-go/sdk/socialproviders"
)

var (
	_ resource.Resource                = (*authProviderConnectionResource)(nil)
	_ resource.ResourceWithImportState = (*authProviderConnectionResource)(nil)
)

func NewAuthProviderConnectionResource() resource.Resource {
	return &authProviderConnectionResource{}
}

type authProviderConnectionResource struct {
	client *sdk.Client
}

func (r *authProviderConnectionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_auth_provider_connection"
}

func (r *authProviderConnectionResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = AuthProviderConnectionResourceSchema(ctx)
}

func (r *authProviderConnectionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *authProviderConnectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data AuthProviderConnectionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	providerType := models.AuthV1ExternalAccountProvider(data.Id.ValueString())
	tokenVal := data.Token.ValueString()
	body := models.AuthV1ConnectProviderRequest{
		Provider: &providerType,
		Token:    &tokenVal,
	}
	if !data.CountryCode.IsNull() && !data.CountryCode.IsUnknown() {
		cc := data.CountryCode.ValueString()
		body.CountryCode = &cc
	}
	if !data.Origin.IsNull() && !data.Origin.IsUnknown() {
		orig := data.Origin.ValueString()
		body.Origin = &orig
	}

	_, err := r.client.SocialProviders.ConnectProvider(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to connect auth provider", err.Error())
		return
	}

	res, err := r.client.SocialProviders.ListProviders(ctx, nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read connected auth provider", err.Error())
		return
	}

	var found any
	for _, p := range res.Providers {
		if p.Provider != nil && string(*p.Provider) == data.Id.ValueString() {
			found = p
			break
		}
	}
	if found != nil {
		savedToken := data.Token
		resp.Diagnostics.Append(mapResponseToModel(ctx, found, &data, AuthProviderConnectionResourceSchema(ctx).Attributes)...)
		data.Token = savedToken
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *authProviderConnectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data AuthProviderConnectionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.client.SocialProviders.ListProviders(ctx, nil)
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to list auth providers", err.Error())
		return
	}

	var found any
	for _, p := range res.Providers {
		if p.Provider != nil && string(*p.Provider) == data.Id.ValueString() {
			found = p
			break
		}
	}
	if found == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	savedToken := data.Token
	resp.Diagnostics.Append(mapResponseToModel(ctx, found, &data, AuthProviderConnectionResourceSchema(ctx).Attributes)...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Token = savedToken

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *authProviderConnectionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan AuthProviderConnectionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *authProviderConnectionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data AuthProviderConnectionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.SocialProviders.DisconnectProvider(ctx, socialproviders.AuthV1ExternalAccountProvider(data.Id.ValueString()), nil)
	if err != nil && !isNotFound(err) {
		resp.Diagnostics.AddError("Failed to disconnect auth provider", err.Error())
	}
}

func (r *authProviderConnectionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}
