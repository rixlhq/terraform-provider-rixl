package provider

import (
	"context"
	"testing"
)

func TestTypedResourceSchemas(t *testing.T) {
	ctx := context.Background()

	t.Run("ApiKeyResourceSchema", func(t *testing.T) {
		s := ApiKeyResourceSchema(ctx)
		if _, ok := s.Attributes["id"]; !ok {
			t.Fatalf("expected id attribute")
		}
		if _, ok := s.Attributes["name"]; !ok {
			t.Fatalf("expected name attribute")
		}
		if _, ok := s.Attributes["org_id"]; !ok {
			t.Fatalf("expected org_id attribute")
		}
		if _, ok := s.Attributes["secret"]; !ok {
			t.Fatalf("expected secret attribute")
		}
	})

	t.Run("ClientCredentialResourceSchema", func(t *testing.T) {
		s := ClientCredentialResourceSchema(ctx)
		if _, ok := s.Attributes["id"]; !ok {
			t.Fatalf("expected id attribute")
		}
		if _, ok := s.Attributes["client_secret"]; !ok {
			t.Fatalf("expected client_secret attribute")
		}
		if _, ok := s.Attributes["name"]; !ok {
			t.Fatalf("expected name attribute")
		}
	})

	t.Run("ProjectCustomDomainResourceSchema", func(t *testing.T) {
		s := ProjectCustomDomainResourceSchema(ctx)
		if _, ok := s.Attributes["custom_domain"]; !ok {
			t.Fatalf("expected custom_domain attribute")
		}
		if _, ok := s.Attributes["project_id"]; !ok {
			t.Fatalf("expected project_id attribute")
		}
	})

	t.Run("SubscriptionResourceSchema", func(t *testing.T) {
		s := SubscriptionResourceSchema(ctx)
		if _, ok := s.Attributes["stripe_price_id"]; !ok {
			t.Fatalf("expected stripe_price_id attribute")
		}
		if _, ok := s.Attributes["status"]; !ok {
			t.Fatalf("expected status attribute")
		}
	})
}
