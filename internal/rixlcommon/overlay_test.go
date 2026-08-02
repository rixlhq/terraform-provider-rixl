package rixlcommon

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestOverlayKnownUsesBaseWhenResponseNull(t *testing.T) {
	base := tftypes.NewValue(tftypes.String, "hello")
	resp := tftypes.NewValue(tftypes.String, nil)

	got := OverlayKnown(base, resp)
	if !got.Equal(base) {
		t.Fatalf("OverlayKnown(%v, %v) = %v, want %v", base, resp, got, base)
	}
}

func TestOverlayKnownUsesResponseWhenKnown(t *testing.T) {
	base := tftypes.NewValue(tftypes.String, "hello")
	resp := tftypes.NewValue(tftypes.String, "world")

	got := OverlayKnown(base, resp)
	if !got.Equal(resp) {
		t.Fatalf("OverlayKnown(%v, %v) = %v, want %v", base, resp, got, resp)
	}
}

func TestOverlayKnownMergesObjectAttributes(t *testing.T) {
	objType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"project_id": tftypes.String,
		"name":       tftypes.String,
	}}

	base := tftypes.NewValue(objType, map[string]tftypes.Value{
		"project_id": tftypes.NewValue(tftypes.String, "p1"),
		"name":       tftypes.NewValue(tftypes.String, nil),
	})

	resp := tftypes.NewValue(objType, map[string]tftypes.Value{
		"project_id": tftypes.NewValue(tftypes.String, nil),
		"name":       tftypes.NewValue(tftypes.String, "feed-a"),
	})

	want := tftypes.NewValue(objType, map[string]tftypes.Value{
		"project_id": tftypes.NewValue(tftypes.String, "p1"),
		"name":       tftypes.NewValue(tftypes.String, "feed-a"),
	})

	got := OverlayKnown(base, resp)
	if !got.Equal(want) {
		t.Fatalf("OverlayKnown(%v, %v) = %v, want %v", base, resp, got, want)
	}
}

func TestOverlayKnownReturnsResponseForList(t *testing.T) {
	listType := tftypes.List{ElementType: tftypes.String}
	base := tftypes.NewValue(listType, []tftypes.Value{
		tftypes.NewValue(tftypes.String, "a"),
	})
	resp := tftypes.NewValue(listType, []tftypes.Value{
		tftypes.NewValue(tftypes.String, "b"),
	})

	got := OverlayKnown(base, resp)
	if !got.Equal(resp) {
		t.Fatalf("OverlayKnown(%v, %v) = %v, want %v", base, resp, got, resp)
	}
}
