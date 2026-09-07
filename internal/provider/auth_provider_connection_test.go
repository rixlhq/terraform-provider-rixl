package provider

import (
	"context"
	"strings"
	"testing"

	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func TestAuthProviderConnectionSchemaAttributes(t *testing.T) {
	schema := AuthProviderConnectionResourceSchema(context.Background())
	for _, name := range []string{
		"id", "token", "country_code", "origin",
		"username", "first_name", "last_name", "email_address", "image_url",
	} {
		if _, ok := schema.Attributes[name]; !ok {
			t.Fatalf("expected %s attribute in auth_provider_connection schema", name)
		}
	}

	idAttr, ok := schema.Attributes["id"].(rschema.StringAttribute)
	if !ok {
		t.Fatalf("expected id to be a string attribute")
	}
	if !idAttr.Required {
		t.Fatalf("expected id to be required: it carries the provider identity")
	}
	if !hasRequiresReplace(idAttr) {
		t.Fatalf("expected id to require replacement on change")
	}

	tokenAttr, ok := schema.Attributes["token"].(rschema.StringAttribute)
	if !ok {
		t.Fatalf("expected token to be a string attribute")
	}
	if !tokenAttr.Required || !tokenAttr.Sensitive {
		t.Fatalf("expected token to be required and sensitive")
	}
	if !hasRequiresReplace(tokenAttr) {
		t.Fatalf("expected token to require replacement on change")
	}
}

func hasRequiresReplace(attr rschema.StringAttribute) bool {
	ctx := context.Background()
	for _, m := range attr.PlanModifiers {
		if desc := m.Description(ctx); strings.Contains(desc, "destroy and recreate") {
			return true
		}
	}
	return false
}

func TestAuthProviderConnectionUpgradeWiringNil(t *testing.T) {
	ctx := context.Background()

	r := &managedResource{descriptor: ResourceDescriptor{TypeName: "auth_provider_connection"}}
	if upgrades := r.UpgradeState(ctx); upgrades != nil {
		t.Fatalf("expected no state upgraders for auth_provider_connection, got %v", upgrades)
	}
}
