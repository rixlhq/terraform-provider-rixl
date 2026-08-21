package provider

import (
	"context"
	"testing"
	"time"

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
