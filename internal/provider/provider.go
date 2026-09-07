package provider

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rixlhq/rixl-go/sdk"
)

const (
	defaultBaseURL = "https://api.rixl.com"
	userAgent      = "terraform-provider-rixl"
	providerAPIKey = "api_key"
)

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
		Description:         "The Rixl provider allows Terraform to manage Rixl platform resources.",
		MarkdownDescription: "The Rixl provider allows Terraform to manage Rixl platform resources.",
		Attributes: map[string]schema.Attribute{
			providerAPIKey: schema.StringAttribute{
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

	baseURLOpt, err := baseURLRewriter(base)
	if err != nil {
		resp.Diagnostics.AddError("Invalid base_url", err.Error())
		return
	}

	opts := []sdk.Option{
		sdk.WithHTTPClient(newHTTPClient()),
		baseURLOpt,
	}
	if bearer != "" {
		// sdk.WithBearer replaces the entire editor list, which would drop the
		// base URL rewriter. Use a custom editor instead so both auth and base
		// URL rewriting are applied.
		apiKey = ""
		opts = append(opts, authEditor(bearer))
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
	return append([]func() resource.Resource{
		NewAccessPolicyResource,
		NewAudioTrackResource,
		NewBillingAddressResource,
		NewBlogSubscriptionResource,
		NewCustomDomainResource,
		NewDashboardResource,
		NewDashboardWidgetResource,
		NewDomainAutoJoinResource,
		NewFeedResource,
		NewImageResource,
		NewOrganizationInvitationResource,
		NewOrganizationMemberResource,
		NewOrganizationResource,
		NewPostResource,
		NewProjectResource,
		NewSubtitleResource,
		NewVideoResource,
	}, genericResourceConstructors()...)
}

func (p *rixlProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewApiKeysDataSource,
		NewAudioTracksDataSource,
		NewBandwidthUsageDataSource,
		NewBandwidthUsageHistoryDataSource,
		NewBillingAddressDataSource,
		NewBlogSubscriptionDataSource,
		NewChaptersDataSource,
		NewCheckMembershipDataSource,
		NewClientCredentialsDataSource,
		NewDashboardDataSource,
		NewDashboardDatasetsDataSource,
		NewDashboardStatsDataSource,
		NewDashboardsDataSource,
		NewDomainDataSource,
		NewDomainAutoJoinDataSource,
		NewFeedDataSource,
		NewFeedStatsDataSource,
		NewFeedsDataSource,
		NewImageDataSource,
		NewImageStatsDataSource,
		NewImagesDataSource,
		NewInternalMembershipInfoDataSource,
		NewInvoicesDataSource,
		NewLanguagesDataSource,
		NewMembershipApplicationsDataSource,
		NewMembershipsDataSource,
		NewOrganizationMembersDataSource,
		NewPasskeysDataSource,
		NewPaymentMethodFromPaymentIntentDataSource,
		NewPaymentMethodFromSetupIntentDataSource,
		NewPaymentMethodsDataSource,
		NewPermissionRegistriesDataSource,
		NewPlanDataSource,
		NewPlansDataSource,
		NewPoliciesDataSource,
		NewPolicyDataSource,
		NewPolicyAttachmentsDataSource,
		NewPostDataSource,
		NewPostStatsDataSource,
		NewPostsDataSource,
		NewProjectDataSource,
		NewProjectsDataSource,
		NewProvidersDataSource,
		NewRealtimeStatsDataSource,
		NewStorageUsageDataSource,
		NewStorageUsageHistoryDataSource,
		NewSubscriptionDataSource,
		NewSubscriptionHistoryDataSource,
		NewSubtitlesDataSource,
		NewTopFeedsDataSource,
		NewTopImagesDataSource,
		NewTopPostsDataSource,
		NewTopVideosDataSource,
		NewUserDataSource,
		NewUserInfoDataSource,
		NewUserPoliciesDataSource,
		NewVideoDataSource,
		NewVideoHeatmapDataSource,
		NewVideoHotSegmentsDataSource,
		NewVideoStatsDataSource,
		NewVideosDataSource,
	}
}

func valueOrEnv(v types.String, env string) string {
	if !v.IsNull() && !v.IsUnknown() {
		return v.ValueString()
	}
	return os.Getenv(env)
}

func newHTTPClient() *http.Client {
	var baseTransport http.RoundTripper
	if dt, ok := http.DefaultTransport.(*http.Transport); ok {
		cloned := dt.Clone()
		cloned.DialContext = (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext
		baseTransport = cloned
	} else {
		baseTransport = http.DefaultTransport
	}

	return &http.Client{
		Timeout:   30 * time.Second,
		Transport: &userAgentTransport{inner: baseTransport, ua: userAgent},
	}
}

type userAgentTransport struct {
	inner http.RoundTripper
	ua    string
}

func (t *userAgentTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	req.Header.Set("User-Agent", t.ua)
	return t.inner.RoundTrip(req)
}

func baseURLRewriter(base string) (sdk.Option, error) {
	baseURL, err := url.Parse(base)
	if err != nil {
		return nil, fmt.Errorf("parse base_url: %w", err)
	}
	if baseURL.Scheme == "" || baseURL.Host == "" {
		return nil, errors.New("base_url must include scheme and host")
	}

	return sdk.WithRequestEditor(func(_ context.Context, req *http.Request) error {
		req.URL.Scheme = baseURL.Scheme
		req.URL.Host = baseURL.Host
		if baseURL.Path != "" {
			prefix, _ := strings.CutSuffix(baseURL.Path, "/")
			req.URL.Path = prefix + req.URL.Path
			if req.URL.RawPath != "" {
				req.URL.RawPath = prefix + req.URL.RawPath
			}
		}
		if req.Host != "" {
			req.Host = baseURL.Host
		}
		return nil
	}), nil
}

func authEditor(token string) sdk.Option {
	return sdk.WithRequestEditor(func(_ context.Context, req *http.Request) error {
		req.Header.Set("Authorization", "Bearer "+token)
		return nil
	})
}
