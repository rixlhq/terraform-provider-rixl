// Package provider implements the Rixl Terraform provider.
package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/rixlhq/terraform-provider-rixl/internal/providerdata"
	"github.com/rixlhq/terraform-provider-rixl/internal/rixlclient"
)

var _ provider.Provider = &RixlProvider{}

// RixlProvider implements the Rixl Terraform provider.
type RixlProvider struct {
	version string
}

// RixlProviderModel describes the provider configuration.
type RixlProviderModel struct {
	APIKey      types.String `tfsdk:"api_key"`
	BearerToken types.String `tfsdk:"bearer_token"`
	BaseURL     types.String `tfsdk:"base_url"`
}

// Metadata returns the provider type name and version.
func (p *RixlProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "rixl"
	resp.Version = p.version
}

// Schema returns the provider configuration schema.
func (p *RixlProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Terraform provider for managing Rixl platform resources.",
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				MarkdownDescription: "Rixl API key used for the `X-API-Key` header. Can be set via the `RIXL_API_KEY` environment variable.",
				Optional:            true,
				Sensitive:           true,
			},
			"bearer_token": schema.StringAttribute{
				MarkdownDescription: "Rixl bearer token used for the `Authorization` header. Can be set via the `RIXL_BEARER_TOKEN` environment variable.",
				Optional:            true,
				Sensitive:           true,
			},
			"base_url": schema.StringAttribute{
				MarkdownDescription: "Override the Rixl API base URL. Can be set via the `RIXL_BASE_URL` environment variable. Defaults to `https://api.rixl.com`.",
				Optional:            true,
			},
		},
	}
}

// Configure validates provider configuration and creates the API client.
func (p *RixlProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data RixlProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	data = applyEnvOverrides(data)

	if !isConfigured(data) {
		resp.Diagnostics.AddError(
			"Missing Credentials",
			"Either api_key or bearer_token must be configured.",
		)
		return
	}

	c, err := rixlclient.New(
		envStringValue(data.APIKey),
		envStringValue(data.BearerToken),
		envStringValue(data.BaseURL),
	)
	if err != nil {
		resp.Diagnostics.AddError("Client Configuration Error", err.Error())
		return
	}

	pd := &providerdata.Data{Client: c}
	resp.DataSourceData = pd
	resp.ResourceData = pd
}

func applyEnvOverrides(data RixlProviderModel) RixlProviderModel {
	data.APIKey = envOrString(data.APIKey, "RIXL_API_KEY")
	data.BearerToken = envOrString(data.BearerToken, "RIXL_BEARER_TOKEN")
	data.BaseURL = envOrString(data.BaseURL, "RIXL_BASE_URL")
	return data
}

func isConfigured(data RixlProviderModel) bool {
	return isSet(data.APIKey) || isSet(data.BearerToken)
}

func isSet(s types.String) bool {
	return !s.IsNull() && !s.IsUnknown() && s.ValueString() != ""
}

func envOrString(v types.String, env string) types.String {
	if !v.IsNull() && !v.IsUnknown() && v.ValueString() != "" {
		return v
	}
	if val := os.Getenv(env); val != "" {
		return types.StringValue(val)
	}
	return v
}

func envStringValue(v types.String) string {
	if v.IsNull() || v.IsUnknown() {
		return ""
	}
	return v.ValueString()
}

// Resources returns the resources supported by this provider.
func (p *RixlProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewFeedResource,
	}
}

// New returns a factory for the Rixl provider.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &RixlProvider{version: version}
	}
}
