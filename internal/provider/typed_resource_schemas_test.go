package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func TestTypedResourceSchemas(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		get      func() schema.Schema
		required []string
	}{
		{"AccessPolicy", func() schema.Schema { return AccessPolicyResourceSchema(ctx) }, []string{"id", "permissions", "org_id"}},
		{"ApiKey", func() schema.Schema { return ApiKeyResourceSchema(ctx) }, []string{"id", "name", "org_id", "secret"}},
		{"AudioTrack", AudioTrackResourceSchema, []string{"id", "language_code", "file_path"}},
		{"AuthProviderConnection", func() schema.Schema { return AuthProviderConnectionResourceSchema(ctx) }, []string{"id", "token"}},
		{"BillingAddress", func() schema.Schema { return BillingAddressResourceSchema(ctx) }, []string{"id", "org_id"}},
		{"BlogSubscription", BlogSubscriptionResourceSchema, []string{"user_id", "subscribed"}},
		{"ClientCredential", func() schema.Schema { return ClientCredentialResourceSchema(ctx) }, []string{"id", "client_secret", "name"}},
		{"CustomDomain", func() schema.Schema { return CustomDomainResourceSchema(ctx) }, []string{"id", "domain", "org_id"}},
		{"Dashboard", func() schema.Schema { return DashboardResourceSchema(ctx) }, []string{"id", "name"}},
		{"DashboardWidget", DashboardWidgetResourceSchema, []string{"id", "dashboard_id", "title"}},
		{"DomainAutoJoin", DomainAutoJoinResourceSchema, []string{"org_id", "enabled"}},
		{"Feed", func() schema.Schema { return FeedResourceSchema(ctx) }, []string{"id", "name", "project_id"}},
		{"Image", func() schema.Schema { return ImageResourceSchema(ctx) }, []string{"id", "file_path", "project_id"}},
		{"OrganizationInvitation", OrganizationInvitationResourceSchema, []string{"id", "org_id", "role"}},
		{"OrganizationMember", OrganizationMemberResourceSchema, []string{"id", "role"}},
		{"Organization", OrganizationResourceSchema, []string{"id", "org_id"}},
		{"PaymentMethod", func() schema.Schema { return PaymentMethodResourceSchema(ctx) }, []string{"id", "payment_method_id"}},
		{"PolicyAttachment", func() schema.Schema { return PolicyAttachmentResourceSchema(ctx) }, []string{"id", "policy_id", "identity_id"}},
		{"Post", func() schema.Schema { return PostResourceSchema(ctx) }, []string{"id", "project_id", "feed_id"}},
		{"ProjectCustomDomain", func() schema.Schema { return ProjectCustomDomainResourceSchema(ctx) }, []string{"custom_domain", "project_id"}},
		{"Project", func() schema.Schema { return ProjectResourceSchema(ctx) }, []string{"id", "name", "org_id"}},
		{"Subscription", func() schema.Schema { return SubscriptionResourceSchema(ctx) }, []string{"stripe_price_id", "status"}},
		{"Subtitle", SubtitleResourceSchema, []string{"id", "language_code", "file_path"}},
		{"Video", func() schema.Schema { return VideoResourceSchema(ctx) }, []string{"id", "file_path", "project_id"}},
		{"VideoChapters", VideoChaptersResourceSchema, []string{"id", "chapters", "video_id"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.get()
			for _, attr := range tt.required {
				if _, ok := s.Attributes[attr]; !ok {
					t.Fatalf("expected attribute %q in %s schema", attr, tt.name)
				}
			}
		})
	}
}
