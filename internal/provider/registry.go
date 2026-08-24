//nolint:revive // registry file is long by design
package provider

import "github.com/hashicorp/terraform-plugin-framework/datasource"

// Data sources.

func NewApiKeysDataSource() datasource.DataSource {
	return newManagedDataSource(api_keysDataSourceDescriptor)
}

func NewAudioTracksDataSource() datasource.DataSource {
	return newManagedDataSource(audio_tracksDataSourceDescriptor)
}

func NewBandwidthUsageDataSource() datasource.DataSource {
	return newManagedDataSource(bandwidth_usageDataSourceDescriptor)
}

func NewBandwidthUsageHistoryDataSource() datasource.DataSource {
	return newManagedDataSource(bandwidth_usage_historyDataSourceDescriptor)
}

func NewBillingAddressDataSource() datasource.DataSource {
	return newManagedDataSource(billing_addressDataSourceDescriptor)
}

func NewBlogSubscriptionDataSource() datasource.DataSource {
	return newManagedDataSource(blog_subscriptionDataSourceDescriptor)
}

func NewChaptersDataSource() datasource.DataSource {
	return newManagedDataSource(chaptersDataSourceDescriptor)
}

func NewCheckMembershipDataSource() datasource.DataSource {
	return newManagedDataSource(check_membershipDataSourceDescriptor)
}

func NewClientCredentialsDataSource() datasource.DataSource {
	return newManagedDataSource(client_credentialsDataSourceDescriptor)
}

func NewDashboardDataSource() datasource.DataSource {
	return newManagedDataSource(dashboardDataSourceDescriptor)
}

func NewDashboardDatasetsDataSource() datasource.DataSource {
	return newManagedDataSource(dashboard_datasetsDataSourceDescriptor)
}

func NewDashboardStatsDataSource() datasource.DataSource {
	return newManagedDataSource(dashboard_statsDataSourceDescriptor)
}

func NewDashboardsDataSource() datasource.DataSource {
	return newManagedDataSource(dashboardsDataSourceDescriptor)
}

func NewDomainAutoJoinDataSource() datasource.DataSource {
	return newManagedDataSource(domain_auto_joinDataSourceDescriptor)
}

func NewFeedStatsDataSource() datasource.DataSource {
	return newManagedDataSource(feed_statsDataSourceDescriptor)
}

func NewImageStatsDataSource() datasource.DataSource {
	return newManagedDataSource(image_statsDataSourceDescriptor)
}

func NewInternalMembershipInfoDataSource() datasource.DataSource {
	return newManagedDataSource(internal_membership_infoDataSourceDescriptor)
}

func NewInvoicesDataSource() datasource.DataSource {
	return newManagedDataSource(invoicesDataSourceDescriptor)
}

func NewLanguagesDataSource() datasource.DataSource {
	return newManagedDataSource(languagesDataSourceDescriptor)
}

func NewMembershipApplicationsDataSource() datasource.DataSource {
	return newManagedDataSource(membership_applicationsDataSourceDescriptor)
}

func NewMembershipsDataSource() datasource.DataSource {
	return newManagedDataSource(membershipsDataSourceDescriptor)
}

func NewOrganizationMembersDataSource() datasource.DataSource {
	return newManagedDataSource(organization_membersDataSourceDescriptor)
}

func NewPasskeysDataSource() datasource.DataSource {
	return newManagedDataSource(passkeysDataSourceDescriptor)
}

func NewPaymentMethodFromPaymentIntentDataSource() datasource.DataSource {
	return newManagedDataSource(payment_method_from_payment_intentDataSourceDescriptor)
}

func NewPaymentMethodFromSetupIntentDataSource() datasource.DataSource {
	return newManagedDataSource(payment_method_from_setup_intentDataSourceDescriptor)
}

func NewPaymentMethodsDataSource() datasource.DataSource {
	return newManagedDataSource(payment_methodsDataSourceDescriptor)
}

func NewPermissionRegistriesDataSource() datasource.DataSource {
	return newManagedDataSource(permission_registriesDataSourceDescriptor)
}

func NewPlanDataSource() datasource.DataSource {
	return newManagedDataSource(planDataSourceDescriptor)
}

func NewPlansDataSource() datasource.DataSource {
	return newManagedDataSource(plansDataSourceDescriptor)
}

func NewPoliciesDataSource() datasource.DataSource {
	return newManagedDataSource(policiesDataSourceDescriptor)
}

func NewPolicyDataSource() datasource.DataSource {
	return newManagedDataSource(policyDataSourceDescriptor)
}

func NewPolicyAttachmentsDataSource() datasource.DataSource {
	return newManagedDataSource(policy_attachmentsDataSourceDescriptor)
}

func NewPostStatsDataSource() datasource.DataSource {
	return newManagedDataSource(post_statsDataSourceDescriptor)
}

func NewProjectDataSource() datasource.DataSource {
	return newManagedDataSource(projectDataSourceDescriptor)
}

func NewProjectsDataSource() datasource.DataSource {
	return newManagedDataSource(projectsDataSourceDescriptor)
}

func NewProvidersDataSource() datasource.DataSource {
	return newManagedDataSource(providersDataSourceDescriptor)
}

func NewRealtimeStatsDataSource() datasource.DataSource {
	return newManagedDataSource(realtime_statsDataSourceDescriptor)
}

func NewStorageUsageDataSource() datasource.DataSource {
	return newManagedDataSource(storage_usageDataSourceDescriptor)
}

func NewStorageUsageHistoryDataSource() datasource.DataSource {
	return newManagedDataSource(storage_usage_historyDataSourceDescriptor)
}

func NewSubscriptionDataSource() datasource.DataSource {
	return newManagedDataSource(subscriptionDataSourceDescriptor)
}

func NewSubscriptionHistoryDataSource() datasource.DataSource {
	return newManagedDataSource(subscription_historyDataSourceDescriptor)
}

func NewSubtitlesDataSource() datasource.DataSource {
	return newManagedDataSource(subtitlesDataSourceDescriptor)
}

func NewTopFeedsDataSource() datasource.DataSource {
	return newManagedDataSource(top_feedsDataSourceDescriptor)
}

func NewTopImagesDataSource() datasource.DataSource {
	return newManagedDataSource(top_imagesDataSourceDescriptor)
}

func NewTopPostsDataSource() datasource.DataSource {
	return newManagedDataSource(top_postsDataSourceDescriptor)
}

func NewTopVideosDataSource() datasource.DataSource {
	return newManagedDataSource(top_videosDataSourceDescriptor)
}

func NewUserDataSource() datasource.DataSource {
	return newManagedDataSource(userDataSourceDescriptor)
}

func NewUserInfoDataSource() datasource.DataSource {
	return newManagedDataSource(user_infoDataSourceDescriptor)
}

func NewUserPoliciesDataSource() datasource.DataSource {
	return newManagedDataSource(user_policiesDataSourceDescriptor)
}

func NewVideoHeatmapDataSource() datasource.DataSource {
	return newManagedDataSource(video_heatmapDataSourceDescriptor)
}

func NewVideoHotSegmentsDataSource() datasource.DataSource {
	return newManagedDataSource(video_hot_segmentsDataSourceDescriptor)
}

func NewVideoStatsDataSource() datasource.DataSource {
	return newManagedDataSource(video_statsDataSourceDescriptor)
}

var api_keysDataSourceDescriptor = DataSourceDescriptor{
	TypeName:    "api_keys",
	SchemaFn:    ApiKeysDataSourceSchema,
	Model:       (*ApiKeysDataSourceModel)(nil),
	ClientField: "APIKeys",
	ReadMethod:  "ListApiKeys",
	PathParams:  []string{"org_id"},
}

var audio_tracksDataSourceDescriptor = DataSourceDescriptor{
	TypeName:    "audio_tracks",
	SchemaFn:    AudioTracksDataSourceSchema,
	Model:       (*AudioTracksDataSourceModel)(nil),
	ClientField: "AudioTracks",
	ReadMethod:  "ListAudioTracks",
	PathParams:  []string{"video_id"},
}

var bandwidth_usageDataSourceDescriptor = DataSourceDescriptor{
	TypeName:    "bandwidth_usage",
	SchemaFn:    BandwidthUsageDataSourceSchema,
	Model:       (*BandwidthUsageDataSourceModel)(nil),
	ClientField: "Usage",
	ReadMethod:  "GetBandwidthUsage",
	PathParams:  []string{},
}

var bandwidth_usage_historyDataSourceDescriptor = DataSourceDescriptor{
	TypeName:    "bandwidth_usage_history",
	SchemaFn:    BandwidthUsageHistoryDataSourceSchema,
	Model:       (*BandwidthUsageHistoryDataSourceModel)(nil),
	ClientField: "Usage",
	ReadMethod:  "GetBandwidthUsageHistory",
	PathParams:  []string{},
}

var billing_addressDataSourceDescriptor = DataSourceDescriptor{
	TypeName:    "billing_address",
	SchemaFn:    BillingAddressDataSourceSchema,
	Model:       (*BillingAddressDataSourceModel)(nil),
	ClientField: "Payments",
	ReadMethod:  "GetBillingAddress",
	PathParams:  []string{},
}

var blog_subscriptionDataSourceDescriptor = DataSourceDescriptor{
	TypeName:    "blog_subscription",
	SchemaFn:    BlogSubscriptionDataSourceSchema,
	Model:       (*BlogSubscriptionDataSourceModel)(nil),
	ClientField: "Blog",
	ReadMethod:  "GetBlogSubscription",
	PathParams:  []string{},
}

var chaptersDataSourceDescriptor = DataSourceDescriptor{
	TypeName:    "chapters",
	SchemaFn:    ChaptersDataSourceSchema,
	Model:       (*ChaptersDataSourceModel)(nil),
	ClientField: "Chapters",
	ReadMethod:  "GetVideoChapters",
	PathParams:  []string{"project_id", "video_id"},
}

var check_membershipDataSourceDescriptor = DataSourceDescriptor{
	TypeName:    "check_membership",
	SchemaFn:    CheckMembershipDataSourceSchema,
	Model:       (*CheckMembershipDataSourceModel)(nil),
	ClientField: "Memberships",
	ReadMethod:  "CheckMembership",
	PathParams:  []string{"org_id"},
}

var client_credentialsDataSourceDescriptor = DataSourceDescriptor{
	TypeName:    "client_credentials",
	SchemaFn:    ClientCredentialsDataSourceSchema,
	Model:       (*ClientCredentialsDataSourceModel)(nil),
	ClientField: "ClientCredentials",
	ReadMethod:  "ListClientCredentials",
	PathParams:  []string{},
}

var dashboardDataSourceDescriptor = DataSourceDescriptor{
	TypeName:    "dashboard",
	SchemaFn:    DashboardDataSourceSchema,
	Model:       (*DashboardDataSourceModel)(nil),
	ClientField: "Dashboards",
	ReadMethod:  "GetDashboard",
	PathParams:  []string{"id"},
}

var dashboard_datasetsDataSourceDescriptor = DataSourceDescriptor{
	TypeName:    "dashboard_datasets",
	SchemaFn:    DashboardDatasetsDataSourceSchema,
	Model:       (*DashboardDatasetsDataSourceModel)(nil),
	ClientField: "Dashboards",
	ReadMethod:  "ListDatasets",
	PathParams:  []string{},
}

var dashboard_statsDataSourceDescriptor = DataSourceDescriptor{
	TypeName:    "dashboard_stats",
	SchemaFn:    DashboardStatsDataSourceSchema,
	Model:       (*DashboardStatsDataSourceModel)(nil),
	ClientField: "Dashboards",
	ReadMethod:  "GetDashboardStats",
	PathParams:  []string{},
}

var dashboardsDataSourceDescriptor = DataSourceDescriptor{
	TypeName:    "dashboards",
	SchemaFn:    DashboardsDataSourceSchema,
	Model:       (*DashboardsDataSourceModel)(nil),
	ClientField: "Dashboards",
	ReadMethod:  "ListDashboards",
	PathParams:  []string{},
}

var domain_auto_joinDataSourceDescriptor = DataSourceDescriptor{
	TypeName:    "domain_auto_join",
	SchemaFn:    DomainAutoJoinDataSourceSchema,
	Model:       (*DomainAutoJoinDataSourceModel)(nil),
	ClientField: "CustomDomains",
	ReadMethod:  "GetDomainAutoJoin",
	PathParams:  []string{"org_id"},
}

var feed_statsDataSourceDescriptor = DataSourceDescriptor{
	TypeName:    "feed_stats",
	SchemaFn:    FeedStatsDataSourceSchema,
	Model:       (*FeedStatsDataSourceModel)(nil),
	ClientField: "FeedAnalytics",
	ReadMethod:  "GetFeedStats",
	PathParams:  []string{"feed_id"},
}

var image_statsDataSourceDescriptor = DataSourceDescriptor{
	TypeName:    "image_stats",
	SchemaFn:    ImageStatsDataSourceSchema,
	Model:       (*ImageStatsDataSourceModel)(nil),
	ClientField: "ImageAnalytics",
	ReadMethod:  "GetImageStats",
	PathParams:  []string{"image_id"},
}

var internal_membership_infoDataSourceDescriptor = DataSourceDescriptor{
	TypeName:    "internal_membership_info",
	SchemaFn:    InternalMembershipInfoDataSourceSchema,
	Model:       (*InternalMembershipInfoDataSourceModel)(nil),
	ClientField: "Memberships",
	ReadMethod:  "GetInternalMembershipInfo",
	PathParams:  []string{"org_id"},
}

var invoicesDataSourceDescriptor = DataSourceDescriptor{
	TypeName:    "invoices",
	SchemaFn:    InvoicesDataSourceSchema,
	Model:       (*InvoicesDataSourceModel)(nil),
	ClientField: "Invoices",
	ReadMethod:  "ListInvoices",
	PathParams:  []string{},
}

var languagesDataSourceDescriptor = DataSourceDescriptor{
	TypeName:    "languages",
	SchemaFn:    LanguagesDataSourceSchema,
	Model:       (*LanguagesDataSourceModel)(nil),
	ClientField: "Languages",
	ReadMethod:  "ListLanguages",
	PathParams:  []string{},
}

var membership_applicationsDataSourceDescriptor = DataSourceDescriptor{
	TypeName:    "membership_applications",
	SchemaFn:    MembershipApplicationsDataSourceSchema,
	Model:       (*MembershipApplicationsDataSourceModel)(nil),
	ClientField: "Memberships",
	ReadMethod:  "ListMembershipApplications",
	PathParams:  []string{},
}

var membershipsDataSourceDescriptor = DataSourceDescriptor{
	TypeName:    "memberships",
	SchemaFn:    MembershipsDataSourceSchema,
	Model:       (*MembershipsDataSourceModel)(nil),
	ClientField: "Memberships",
	ReadMethod:  "ListMemberships",
	PathParams:  []string{},
}

var organization_membersDataSourceDescriptor = DataSourceDescriptor{
	TypeName:    "organization_members",
	SchemaFn:    OrganizationMembersDataSourceSchema,
	Model:       (*OrganizationMembersDataSourceModel)(nil),
	ClientField: "Memberships",
	ReadMethod:  "ListOrganizationMembers",
	PathParams:  []string{"org_id"},
}

var passkeysDataSourceDescriptor = DataSourceDescriptor{
	TypeName:    "passkeys",
	SchemaFn:    PasskeysDataSourceSchema,
	Model:       (*PasskeysDataSourceModel)(nil),
	ClientField: "Passkeys",
	ReadMethod:  "ListPasskeys",
	PathParams:  []string{},
}

var payment_method_from_payment_intentDataSourceDescriptor = DataSourceDescriptor{
	TypeName:    "payment_method_from_payment_intent",
	SchemaFn:    PaymentMethodFromPaymentIntentDataSourceSchema,
	Model:       (*PaymentMethodFromPaymentIntentDataSourceModel)(nil),
	ClientField: "Payments",
	ReadMethod:  "GetPaymentMethodFromPaymentIntent",
	PathParams:  []string{},
}

var payment_method_from_setup_intentDataSourceDescriptor = DataSourceDescriptor{
	TypeName:    "payment_method_from_setup_intent",
	SchemaFn:    PaymentMethodFromSetupIntentDataSourceSchema,
	Model:       (*PaymentMethodFromSetupIntentDataSourceModel)(nil),
	ClientField: "Payments",
	ReadMethod:  "GetPaymentMethodFromSetupIntent",
	PathParams:  []string{},
}

var payment_methodsDataSourceDescriptor = DataSourceDescriptor{
	TypeName:    "payment_methods",
	SchemaFn:    PaymentMethodsDataSourceSchema,
	Model:       (*PaymentMethodsDataSourceModel)(nil),
	ClientField: "Payments",
	ReadMethod:  "ListPaymentMethods",
	PathParams:  []string{},
}

var permission_registriesDataSourceDescriptor = DataSourceDescriptor{
	TypeName:    "permission_registries",
	SchemaFn:    PermissionRegistriesDataSourceSchema,
	Model:       (*PermissionRegistriesDataSourceModel)(nil),
	ClientField: "AccessPolicies",
	ReadMethod:  "ListPermissionRegistry",
	PathParams:  []string{"org_id"},
}

var planDataSourceDescriptor = DataSourceDescriptor{
	TypeName:    "plan",
	SchemaFn:    PlanDataSourceSchema,
	Model:       (*PlanDataSourceModel)(nil),
	ClientField: "Plans",
	ReadMethod:  "GetPlan",
	PathParams:  []string{"id"},
}

var plansDataSourceDescriptor = DataSourceDescriptor{
	TypeName:    "plans",
	SchemaFn:    PlansDataSourceSchema,
	Model:       (*PlansDataSourceModel)(nil),
	ClientField: "Plans",
	ReadMethod:  "ListPlans",
	PathParams:  []string{},
}

var policiesDataSourceDescriptor = DataSourceDescriptor{
	TypeName:    "policies",
	SchemaFn:    PoliciesDataSourceSchema,
	Model:       (*PoliciesDataSourceModel)(nil),
	ClientField: "AccessPolicies",
	ReadMethod:  "ListPolicies",
	PathParams:  []string{"org_id"},
}

var policyDataSourceDescriptor = DataSourceDescriptor{
	TypeName:    "policy",
	SchemaFn:    PolicyDataSourceSchema,
	Model:       (*PolicyDataSourceModel)(nil),
	ClientField: "AccessPolicies",
	ReadMethod:  "GetPolicy",
	PathParams:  []string{"org_id", "id"},
}

var policy_attachmentsDataSourceDescriptor = DataSourceDescriptor{
	TypeName:    "policy_attachments",
	SchemaFn:    PolicyAttachmentsDataSourceSchema,
	Model:       (*PolicyAttachmentsDataSourceModel)(nil),
	ClientField: "AccessPolicies",
	ReadMethod:  "ListPolicyAttachments",
	PathParams:  []string{"org_id", "policy_id"},
}

var post_statsDataSourceDescriptor = DataSourceDescriptor{
	TypeName:    "post_stats",
	SchemaFn:    PostStatsDataSourceSchema,
	Model:       (*PostStatsDataSourceModel)(nil),
	ClientField: "PostAnalytics",
	ReadMethod:  "GetPostStats",
	PathParams:  []string{"post_id"},
}

var projectDataSourceDescriptor = DataSourceDescriptor{
	TypeName:    "project",
	SchemaFn:    ProjectDataSourceSchema,
	Model:       (*ProjectDataSourceModel)(nil),
	ClientField: "Projects",
	ReadMethod:  "GetProject",
	PathParams:  []string{"org_id", "id"},
}

var projectsDataSourceDescriptor = DataSourceDescriptor{
	TypeName:    "projects",
	SchemaFn:    ProjectsDataSourceSchema,
	Model:       (*ProjectsDataSourceModel)(nil),
	ClientField: "Projects",
	ReadMethod:  "ListProjects",
	PathParams:  []string{"org_id"},
}

var providersDataSourceDescriptor = DataSourceDescriptor{
	TypeName:    "providers",
	SchemaFn:    ProvidersDataSourceSchema,
	Model:       (*ProvidersDataSourceModel)(nil),
	ClientField: "SocialProviders",
	ReadMethod:  "ListProviders",
	PathParams:  []string{},
}

var realtime_statsDataSourceDescriptor = DataSourceDescriptor{
	TypeName:    "realtime_stats",
	SchemaFn:    RealtimeStatsDataSourceSchema,
	Model:       (*RealtimeStatsDataSourceModel)(nil),
	ClientField: "Realtime",
	ReadMethod:  "GetRealtimeStats",
	PathParams:  []string{},
}

var storage_usageDataSourceDescriptor = DataSourceDescriptor{
	TypeName:    "storage_usage",
	SchemaFn:    StorageUsageDataSourceSchema,
	Model:       (*StorageUsageDataSourceModel)(nil),
	ClientField: "Usage",
	ReadMethod:  "GetStorageUsage",
	PathParams:  []string{},
}

var storage_usage_historyDataSourceDescriptor = DataSourceDescriptor{
	TypeName:    "storage_usage_history",
	SchemaFn:    StorageUsageHistoryDataSourceSchema,
	Model:       (*StorageUsageHistoryDataSourceModel)(nil),
	ClientField: "Usage",
	ReadMethod:  "GetStorageUsageHistory",
	PathParams:  []string{},
}

var subscriptionDataSourceDescriptor = DataSourceDescriptor{
	TypeName:    "subscription",
	SchemaFn:    SubscriptionDataSourceSchema,
	Model:       (*SubscriptionDataSourceModel)(nil),
	ClientField: "Subscriptions",
	ReadMethod:  "GetSubscription",
	PathParams:  []string{},
}

var subscription_historyDataSourceDescriptor = DataSourceDescriptor{
	TypeName:    "subscription_history",
	SchemaFn:    SubscriptionHistoryDataSourceSchema,
	Model:       (*SubscriptionHistoryDataSourceModel)(nil),
	ClientField: "Subscriptions",
	ReadMethod:  "GetSubscriptionHistory",
	PathParams:  []string{},
}

var subtitlesDataSourceDescriptor = DataSourceDescriptor{
	TypeName:    "subtitles",
	SchemaFn:    SubtitlesDataSourceSchema,
	Model:       (*SubtitlesDataSourceModel)(nil),
	ClientField: "Subtitles",
	ReadMethod:  "ListSubtitles",
	PathParams:  []string{"video_id"},
}

var top_feedsDataSourceDescriptor = DataSourceDescriptor{
	TypeName:    "top_feeds",
	SchemaFn:    TopFeedsDataSourceSchema,
	Model:       (*TopFeedsDataSourceModel)(nil),
	ClientField: "FeedAnalytics",
	ReadMethod:  "GetTopFeeds",
	PathParams:  []string{},
}

var top_imagesDataSourceDescriptor = DataSourceDescriptor{
	TypeName:    "top_images",
	SchemaFn:    TopImagesDataSourceSchema,
	Model:       (*TopImagesDataSourceModel)(nil),
	ClientField: "ImageAnalytics",
	ReadMethod:  "GetTopImages",
	PathParams:  []string{},
}

var top_postsDataSourceDescriptor = DataSourceDescriptor{
	TypeName:    "top_posts",
	SchemaFn:    TopPostsDataSourceSchema,
	Model:       (*TopPostsDataSourceModel)(nil),
	ClientField: "PostAnalytics",
	ReadMethod:  "GetTopPosts",
	PathParams:  []string{},
}

var top_videosDataSourceDescriptor = DataSourceDescriptor{
	TypeName:    "top_videos",
	SchemaFn:    TopVideosDataSourceSchema,
	Model:       (*TopVideosDataSourceModel)(nil),
	ClientField: "VideoAnalytics",
	ReadMethod:  "GetTopVideos",
	PathParams:  []string{},
}

var userDataSourceDescriptor = DataSourceDescriptor{
	TypeName:    "user",
	SchemaFn:    UserDataSourceSchema,
	Model:       (*UserDataSourceModel)(nil),
	ClientField: "Users",
	ReadMethod:  "GetUser",
	PathParams:  []string{},
}

var user_infoDataSourceDescriptor = DataSourceDescriptor{
	TypeName:    "user_info",
	SchemaFn:    UserInfoDataSourceSchema,
	Model:       (*UserInfoDataSourceModel)(nil),
	ClientField: "Users",
	ReadMethod:  "GetUserInfo",
	PathParams:  []string{},
}

var user_policiesDataSourceDescriptor = DataSourceDescriptor{
	TypeName:    "user_policies",
	SchemaFn:    UserPoliciesDataSourceSchema,
	Model:       (*UserPoliciesDataSourceModel)(nil),
	ClientField: "AccessPolicies",
	ReadMethod:  "ListUserPolicies",
	PathParams:  []string{"org_id", "member_id"},
}

var video_heatmapDataSourceDescriptor = DataSourceDescriptor{
	TypeName:    "video_heatmap",
	SchemaFn:    VideoHeatmapDataSourceSchema,
	Model:       (*VideoHeatmapDataSourceModel)(nil),
	ClientField: "Heatmaps",
	ReadMethod:  "GetVideoHeatmap",
	PathParams:  []string{"video_id"},
}

var video_hot_segmentsDataSourceDescriptor = DataSourceDescriptor{
	TypeName:    "video_hot_segments",
	SchemaFn:    VideoHotSegmentsDataSourceSchema,
	Model:       (*VideoHotSegmentsDataSourceModel)(nil),
	ClientField: "Heatmaps",
	ReadMethod:  "GetHotSegments",
	PathParams:  []string{"video_id"},
}

var video_statsDataSourceDescriptor = DataSourceDescriptor{
	TypeName:    "video_stats",
	SchemaFn:    VideoStatsDataSourceSchema,
	Model:       (*VideoStatsDataSourceModel)(nil),
	ClientField: "VideoAnalytics",
	ReadMethod:  "GetVideoStats",
	PathParams:  []string{"video_id"},
}
