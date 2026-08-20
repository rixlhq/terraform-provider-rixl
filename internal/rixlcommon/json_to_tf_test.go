package rixlcommon_test

import (
	"encoding/json"
	"math/big"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/rixlhq/terraform-provider-rixl/internal/rixlcommon"
)

func TestJSONToTfValue(t *testing.T) {
	t.Parallel()

	typ := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"name":  tftypes.String,
			"count": tftypes.Number,
			"ok":    tftypes.Bool,
		},
	}

	input := map[string]any{
		"name":  "example",
		"count": json.Number("42"),
		"ok":    true,
	}

	val, err := rixlcommon.JSONToTfValue(t.Context(), typ, input)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	var m map[string]tftypes.Value
	if err := val.As(&m); err != nil {
		t.Fatalf("failed to convert value: %s", err)
	}

	var name string
	if err := m["name"].As(&name); err != nil || name != "example" {
		t.Fatalf("unexpected name: %v / %s", name, err)
	}

	var count big.Float
	if err := m["count"].As(&count); err != nil {
		t.Fatalf("failed to get count: %s", err)
	}
	if count.Text('f', 0) != "42" {
		t.Fatalf("unexpected count: %s", count.Text('f', 0))
	}

	var ok bool
	if err := m["ok"].As(&ok); err != nil || !ok {
		t.Fatalf("unexpected ok: %v / %s", ok, err)
	}
}

func TestJSONToTfValueNested(t *testing.T) {
	t.Parallel()

	typ := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"items": tftypes.List{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
				"id": tftypes.String,
			}}},
		},
	}

	input := map[string]any{
		"items": []any{
			map[string]any{"id": "a"},
			map[string]any{"id": "b"},
		},
	}

	val, err := rixlcommon.JSONToTfValue(t.Context(), typ, input)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	var m map[string]tftypes.Value
	if err := val.As(&m); err != nil {
		t.Fatalf("failed to convert value: %s", err)
	}

	var items []tftypes.Value
	if err := m["items"].As(&items); err != nil {
		t.Fatalf("failed to get items: %s", err)
	}

	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}

	for i, item := range items {
		var itemMap map[string]tftypes.Value
		if err := item.As(&itemMap); err != nil {
			t.Fatalf("failed to get item %d: %s", i, err)
		}
		var id string
		if err := itemMap["id"].As(&id); err != nil {
			t.Fatalf("failed to get item %d id: %s", i, err)
		}
		expected := []string{"a", "b"}
		if id != expected[i] {
			t.Fatalf("expected id %q, got %q", expected[i], id)
		}
	}
}
