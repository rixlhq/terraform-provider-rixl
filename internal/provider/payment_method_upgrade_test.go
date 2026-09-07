package provider

import (
	"context"
	"testing"
)

func TestPaymentMethodSchemaVersion(t *testing.T) {
	schema := PaymentMethodResourceSchema(context.Background())
	if schema.Version != 1 {
		t.Fatalf("expected schema version 1, got %d", schema.Version)
	}
	if _, ok := schema.Attributes["provider_name"]; !ok {
		t.Fatalf("expected provider_name attribute in current schema")
	}
	if _, ok := schema.Attributes["provider"]; ok {
		t.Fatalf("current schema must not contain reserved provider attribute")
	}
}

func TestPaymentMethodPriorSchema(t *testing.T) {
	ctx := context.Background()
	prior := paymentMethodSchemaV0(ctx)
	if prior.Version != 0 {
		t.Fatalf("expected prior schema version 0, got %d", prior.Version)
	}
	if _, ok := prior.Attributes["provider"]; !ok {
		t.Fatalf("expected provider attribute in prior schema")
	}
	if _, ok := prior.Attributes["provider_name"]; ok {
		t.Fatalf("prior schema must not contain provider_name attribute")
	}
}

func TestManagedResourceUpgradeStateWiring(t *testing.T) {
	ctx := context.Background()

	payment := &managedResource{descriptor: ResourceDescriptor{TypeName: "payment_method"}}
	upgrades := payment.UpgradeState(ctx)
	upgrader, ok := upgrades[0]
	if !ok {
		t.Fatalf("expected state upgrader for version 0")
	}
	if upgrader.PriorSchema == nil {
		t.Fatalf("expected prior schema for payment_method upgrade")
	}
	if upgrader.StateUpgrader == nil {
		t.Fatalf("expected state upgrader func for payment_method upgrade")
	}
	if _, ok := upgrader.PriorSchema.Attributes["provider"]; !ok {
		t.Fatalf("prior schema must contain provider attribute")
	}

	other := &managedResource{descriptor: ResourceDescriptor{TypeName: "subscription"}}
	if upgrades := other.UpgradeState(ctx); upgrades != nil {
		t.Fatalf("expected no state upgraders for unrelated resource, got %v", upgrades)
	}
}
