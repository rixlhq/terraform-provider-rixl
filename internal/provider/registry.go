//nolint:revive // registry file containing many generated constructors and descriptors
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

func NewClientCredentialsDataSource() datasource.DataSource {
	return newManagedDataSource(client_credentialsDataSourceDescriptor)
}

func NewDashboardDataSource() datasource.DataSource {
	return newManagedDataSource(dashboardDataSourceDescriptor)
}

func NewDashboardsDataSource() datasource.DataSource {
	return newManagedDataSource(dashboardsDataSourceDescriptor)
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

func NewPasskeysDataSource() datasource.DataSource {
	return newManagedDataSource(passkeysDataSourceDescriptor)
}

func NewPaymentMethodsDataSource() datasource.DataSource {
	return newManagedDataSource(payment_methodsDataSourceDescriptor)
}

func NewPlanDataSource() datasource.DataSource {
	return newManagedDataSource(planDataSourceDescriptor)
}

func NewPlansDataSource() datasource.DataSource {
	return newManagedDataSource(plansDataSourceDescriptor)
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

func NewUserDataSource() datasource.DataSource {
	return newManagedDataSource(userDataSourceDescriptor)
}

func NewUserInfoDataSource() datasource.DataSource {
	return newManagedDataSource(user_infoDataSourceDescriptor)
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

var dashboardsDataSourceDescriptor = DataSourceDescriptor{
	TypeName:    "dashboards",
	SchemaFn:    DashboardsDataSourceSchema,
	Model:       (*DashboardsDataSourceModel)(nil),
	ClientField: "Dashboards",
	ReadMethod:  "ListDashboards",
	PathParams:  []string{},
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

var passkeysDataSourceDescriptor = DataSourceDescriptor{
	TypeName:    "passkeys",
	SchemaFn:    PasskeysDataSourceSchema,
	Model:       (*PasskeysDataSourceModel)(nil),
	ClientField: "Passkeys",
	ReadMethod:  "ListPasskeys",
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
