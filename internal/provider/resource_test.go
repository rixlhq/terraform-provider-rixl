package provider

import (
	"context"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestMethodPathParams(t *testing.T) {
	r := &managedResource{
		descriptor: ResourceDescriptor{
			PathParams:       []string{"a", "b"},
			CreatePathParams: []string{"c"},
			UpdatePathParams: []string{"d", "e"},
		},
	}

	if got := r.methodPathParams("create"); len(got) != 1 || got[0] != "c" {
		t.Fatalf("create path params mismatch: got %v", got)
	}
	if got := r.methodPathParams("read"); len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Fatalf("read path params mismatch: got %v", got)
	}
	if got := r.methodPathParams("update"); len(got) != 2 || got[0] != "d" || got[1] != "e" {
		t.Fatalf("update path params mismatch: got %v", got)
	}
	if got := r.methodPathParams("delete"); len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Fatalf("delete path params mismatch: got %v", got)
	}
}

func TestBuildBodyRemovesPathParamsAndComputed(t *testing.T) {
	ctx := context.Background()
	r := &managedResource{
		descriptor: ResourceDescriptor{
			TypeName:           "test",
			CreateKeepPathKeys: []string{},
			ComputedBodyKeys:   []string{"id", "created_at"},
		},
	}

	model := &PostModel{
		ProjectId:   types.StringValue("p"),
		FeedId:      types.StringValue("f"),
		Id:          types.StringValue("id"),
		Description: types.StringValue("desc"),
		CreatedAt:   types.StringValue("now"),
	}

	body, diags := r.buildBody(ctx, model, []string{"project_id", "feed_id"}, nil)
	if diags.HasError() {
		t.Fatalf("diags: %v", diags)
	}

	if _, ok := body["project_id"]; ok {
		t.Fatalf("project_id should be removed from body")
	}
	if _, ok := body["feed_id"]; ok {
		t.Fatalf("feed_id should be removed from body")
	}
	if _, ok := body["id"]; ok {
		t.Fatalf("id should be removed as computed")
	}
	if _, ok := body["created_at"]; ok {
		t.Fatalf("created_at should be removed as computed")
	}
	if body["description"] != "desc" {
		t.Fatalf("description should be preserved")
	}
}

func TestBuildBodyKeepsPathKeys(t *testing.T) {
	ctx := context.Background()
	r := &managedResource{
		descriptor: ResourceDescriptor{
			TypeName:         "test",
			ComputedBodyKeys: []string{"id"},
		},
	}

	model := &PostModel{
		ProjectId: types.StringValue("p"),
		FeedId:    types.StringValue("f"),
	}

	body, diags := r.buildBody(ctx, model, []string{"project_id", "feed_id"}, []string{"project_id"})
	if diags.HasError() {
		t.Fatalf("diags: %v", diags)
	}

	if body["project_id"] != "p" {
		t.Fatalf("project_id should be kept")
	}
	if _, ok := body["feed_id"]; ok {
		t.Fatalf("feed_id should be removed")
	}
}

func TestApplyBodyRenames(t *testing.T) {
	r := &managedResource{
		descriptor: ResourceDescriptor{
			BodyRenames: map[string]string{"image_id": "media_id"},
		},
	}

	body := r.applyBodyRenames(map[string]any{"image_id": "img"})
	if body["media_id"] != "img" {
		t.Fatalf("rename failed: got %v", body)
	}
	if _, ok := body["image_id"]; ok {
		t.Fatalf("old key should be removed")
	}
}

func TestResponseToReadMapUnwraps(t *testing.T) {
	ctx := context.Background()
	r := &managedResource{
		descriptor: ResourceDescriptor{
			ReadResponseField: "post",
		},
	}

	m, diags := r.responseToReadMap(ctx, reflect.ValueOf(map[string]any{
		"post": map[string]any{"id": "post-1", "description": "hello"},
	}), &PostModel{})
	if diags.HasError() {
		t.Fatalf("diags: %v", diags)
	}

	if m["id"] != "post-1" {
		t.Fatalf("id mismatch: got %v", m["id"])
	}
	if m["description"] != "hello" {
		t.Fatalf("description mismatch")
	}
}

func TestResponseToResourceMapUnwrapsAndFlattens(t *testing.T) {
	r := &managedResource{
		descriptor: ResourceDescriptor{
			CreateResponseField: "credential",
		},
	}

	m, diags := r.responseToResourceMap(reflect.ValueOf(map[string]any{
		"credential": map[string]any{"id": "c", "client_id": "client"},
	}))
	if diags.HasError() {
		t.Fatalf("diags: %v", diags)
	}

	if m["id"] != "c" {
		t.Fatalf("id mismatch: got %v", m["id"])
	}
	if m["client_id"] != "client" {
		t.Fatalf("client_id mismatch")
	}
}

func TestResponseToResourceMapFlattens(t *testing.T) {
	r := &managedResource{
		descriptor: ResourceDescriptor{
			CreateFlattenField: "credential",
		},
	}

	m, diags := r.responseToResourceMap(reflect.ValueOf(map[string]any{
		"client_secret": "s",
		"credential":    map[string]any{"id": "c", "client_id": "client"},
	}))
	if diags.HasError() {
		t.Fatalf("diags: %v", diags)
	}

	if m["id"] != "c" {
		t.Fatalf("id mismatch: got %v", m["id"])
	}
	if m["client_id"] != "client" {
		t.Fatalf("client_id mismatch")
	}
	if m["client_secret"] != "s" {
		t.Fatalf("client_secret should be flattened")
	}
}

func TestSelectFromList(t *testing.T) {
	ctx := context.Background()
	r := &managedResource{
		descriptor: ResourceDescriptor{
			ReadListField:   "posts",
			ReadListIDField: "id",
		},
	}

	list := map[string]any{
		"posts": []any{
			map[string]any{"id": "a", "description": "first"},
			map[string]any{"id": "b", "description": "second"},
		},
	}

	model := &PostModel{Id: types.StringValue("b")}
	m, diags := r.selectFromList(ctx, list, model)
	if diags.HasError() {
		t.Fatalf("diags: %v", diags)
	}

	if m["id"] != "b" {
		t.Fatalf("id mismatch: got %v", m["id"])
	}
	if m["description"] != "second" {
		t.Fatalf("description mismatch: got %v", m["description"])
	}
}
