package provider

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rixlhq/rixl-go/sdk"
)

const defaultBaseURL = "https://api.rixl.com"

var _ provider.Provider = (*rixlProvider)(nil)

type rixlProvider struct{}

type providerConfig struct {
	APIKey      types.String `tfsdk:"api_key"`
	BearerToken types.String `tfsdk:"bearer_token"`
	BaseURL     types.String `tfsdk:"base_url"`
}

func New() provider.Provider {
	return &rixlProvider{}
}

func (p *rixlProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "rixl"
}

func (p *rixlProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Rixl API key. May also be set via the RIXL_API_KEY environment variable.",
			},
			"bearer_token": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Rixl bearer token. May also be set via the RIXL_BEARER_TOKEN environment variable.",
			},
			"base_url": schema.StringAttribute{
				Optional:    true,
				Description: "Rixl API base URL. Defaults to https://api.rixl.com. May also be set via the RIXL_BASE_URL environment variable.",
			},
		},
	}
}

func (p *rixlProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var cfg providerConfig
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiKey := valueOrEnv(cfg.APIKey, "RIXL_API_KEY")
	bearer := valueOrEnv(cfg.BearerToken, "RIXL_BEARER_TOKEN")
	base := valueOrEnv(cfg.BaseURL, "RIXL_BASE_URL")
	if base == "" {
		base = defaultBaseURL
	}

	if apiKey == "" && bearer == "" {
		resp.Diagnostics.AddError(
			"Missing Authentication",
			"Either api_key or bearer_token must be configured, or the RIXL_API_KEY / RIXL_BEARER_TOKEN environment variables must be set.",
		)
		return
	}

	opts := []sdk.Option{}
	if bearer != "" {
		opts = append(opts, sdk.WithBearer(bearer))
	}

	if base != defaultBaseURL {
		httpClient, err := newHTTPClient(base)
		if err != nil {
			resp.Diagnostics.AddError("Invalid base_url", err.Error())
			return
		}
		opts = append(opts, sdk.WithHTTPClient(httpClient))
	}

	client, err := sdk.New(apiKey, opts...)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create Rixl client", err.Error())
		return
	}

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *rixlProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewFeedResource,
		NewProjectResource,
	}
}

func (p *rixlProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewFeedDataSource,
		NewFeedsDataSource,
		NewImageDataSource,
		NewImagesDataSource,
		NewVideoDataSource,
		NewVideosDataSource,
		NewProjectDataSource,
		NewProjectsDataSource,
		NewApiKeysDataSource,
	}
}

func valueOrEnv(v types.String, env string) string {
	if !v.IsNull() && !v.IsUnknown() {
		return v.ValueString()
	}
	return os.Getenv(env)
}

func newHTTPClient(base string) (*http.Client, error) {
	baseURL, err := url.Parse(base)
	if err != nil {
		return nil, fmt.Errorf("parse base_url: %w", err)
	}
	if baseURL.Scheme == "" || baseURL.Host == "" {
		return nil, errors.New("base_url must include scheme and host")
	}
	return &http.Client{
		Transport: &baseURLTransport{base: baseURL, inner: http.DefaultTransport},
	}, nil
}

type baseURLTransport struct {
	base  *url.URL
	inner http.RoundTripper
}

func (t *baseURLTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	newURL := *req.URL
	newURL.Scheme = t.base.Scheme
	newURL.Host = t.base.Host
	if t.base.Path != "" {
		newURL.Path = strings.TrimSuffix(t.base.Path, "/") + newURL.Path
	}
	newURL.RawPath = ""
	req = req.Clone(req.Context())
	req.URL = &newURL
	if req.Host != "" {
		req.Host = t.base.Host
	}
	return t.inner.RoundTrip(req)
}
