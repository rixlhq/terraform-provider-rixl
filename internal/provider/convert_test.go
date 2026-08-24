package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rixlhq/rixl-go/sdk/models"
)

func TestMapResponseToModel(t *testing.T) {
	ctx := context.Background()
	n := "test feed"
	desc := "desc"
	pid := "project-1"
	id := "feed-1"
	trueVal := true
	falseVal := false
	created := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)

	feed := models.FeedsV1Feed{
		ID:          &id,
		ProjectID:   &pid,
		Name:        &n,
		Description: &desc,
		AllowImages: &trueVal,
		AllowVideos: &falseVal,
		HasComments: &trueVal,
		HasLikes:    &trueVal,
		HasShares:   &trueVal,
		CreatedAt:   &created,
		UpdatedAt:   &created,
	}

	var model FeedModel
	diags := mapResponseToModel(ctx, feed, &model, FeedResourceSchema(ctx).Attributes)
	if diags.HasError() {
		t.Fatalf("diags: %v", diags)
	}

	if model.Id.ValueString() != id {
		t.Fatalf("id mismatch: got %q", model.Id.ValueString())
	}
	if model.ProjectId.ValueString() != pid {
		t.Fatalf("project_id mismatch")
	}
	if model.Name.ValueString() != n {
		t.Fatalf("name mismatch")
	}
	if model.AllowImages.ValueBool() != true {
		t.Fatalf("allow_images mismatch")
	}
	if model.CreatedAt.ValueString() != created.Format(time.RFC3339Nano) {
		t.Fatalf("created_at mismatch: got %q", model.CreatedAt.ValueString())
	}
}

func TestModelToMap(t *testing.T) {
	ctx := context.Background()
	model := FeedModel{
		Name:        types.StringValue("feed"),
		Description: types.StringValue("desc"),
		AllowImages: types.BoolValue(true),
		ProjectId:   types.StringValue("p"),
		Id:          types.StringValue("id"),
	}

	m, diags := modelToMap(ctx, &model)
	if diags.HasError() {
		t.Fatalf("diags: %v", diags)
	}
	if m["name"] != "feed" {
		t.Fatalf("name mismatch")
	}
	if m["description"] != "desc" {
		t.Fatalf("description mismatch")
	}
	if m["allow_images"] != true {
		t.Fatalf("allow_images mismatch")
	}
	if m["id"] != "id" {
		t.Fatalf("id mismatch")
	}
}

func TestNativeToString(t *testing.T) {
	cases := []struct {
		in   any
		want string
	}{
		{"hello", "hello"},
		{json.Number("12345678901234"), "12345678901234"},
		{int64(12345678901234), "12345678901234"},
		{float64(1000000), "1000000"},
		{float64(12345678901234), "12345678901234"},
		{true, "true"},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("%T_%v", c.in, c.in), func(t *testing.T) {
			got := nativeToString(c.in)
			if got != c.want {
				t.Fatalf("nativeToString(%#v) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}

func TestFromNativeNumberToString(t *testing.T) {
	ctx := context.Background()

	for _, c := range []struct {
		in   any
		want string
	}{
		{float64(1000000), "1000000"},
		{int64(1000000), "1000000"},
		{json.Number("1000000"), "1000000"},
	} {
		av, diags := fromNative(ctx, c.in, types.StringType)
		if diags.HasError() {
			t.Fatalf("fromNative diags: %v", diags)
		}
		sv, ok := av.(types.String)
		if !ok {
			t.Fatalf("expected types.String, got %T", av)
		}
		if sv.ValueString() != c.want {
			t.Fatalf("fromNative(%#v) = %q, want %q", c.in, sv.ValueString(), c.want)
		}
	}
}

type paginationTotalModel struct {
	Total types.String `tfsdk:"total"`
}

func TestMapResponsePaginationLargeTotal(t *testing.T) {
	ctx := context.Background()

	for _, c := range []struct {
		name string
		raw  any
		want string
	}{
		{"float64", float64(1000000), "1000000"},
		{"int64", int64(1000000), "1000000"},
		{"json.Number", json.Number("1000000"), "1000000"},
	} {
		t.Run(c.name, func(t *testing.T) {
			var m paginationTotalModel
			diags := mapResponsePagination(ctx, map[string]any{"total": c.raw}, &m)
			if diags.HasError() {
				t.Fatalf("mapResponsePagination diags: %v", diags)
			}
			if m.Total.ValueString() != c.want {
				t.Fatalf("total = %q, want %q", m.Total.ValueString(), c.want)
			}
		})
	}
}

type mapToModelTestModel struct {
	Name  types.String `tfsdk:"name"`
	Desc  types.String `tfsdk:"description"`
	Extra types.String `tfsdk:"extra"`
}

func TestMapToModelClearsNilAndPreservesMissing(t *testing.T) {
	ctx := context.Background()

	var m mapToModelTestModel
	m.Name = types.StringValue("old name")
	m.Desc = types.StringValue("old desc")
	m.Extra = types.StringValue("keep me")

	attrs := map[string]attr.Type{
		"name":        types.StringType,
		"description": types.StringType,
		"extra":       types.StringType,
	}

	preserve := map[string]bool{"extra": true}

	diags := mapToModel(ctx, map[string]any{
		"name":        "new name",
		"description": nil,
	}, &m, attrs, preserve)
	if diags.HasError() {
		t.Fatalf("mapToModel diags: %v", diags)
	}

	if m.Name.ValueString() != "new name" {
		t.Fatalf("name = %q, want %q", m.Name.ValueString(), "new name")
	}
	if !m.Desc.IsNull() {
		t.Fatalf("description should be null, got %q", m.Desc.ValueString())
	}
	if m.Extra.ValueString() != "keep me" {
		t.Fatalf("extra should be preserved, got %q", m.Extra.ValueString())
	}
}
