package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rixlhq/rixl-go/sdk/models"
)

func TestMapResponseToModelEmbedded(t *testing.T) {
	img := &models.ImagesV1Image{
		ID:         new("img-1"),
		Visibility: (*models.CommonV1Visibility)(new("VISIBILITY_PUBLIC")),
	}

	var data ImageResourceModel
	data.ProjectId = types.StringValue("proj-1")
	data.Name = types.StringValue("name")
	data.FilePath = types.StringValue("/tmp/foo.jpg")

	diags := mapResponseToModel(context.Background(), img, &data, ImageResourceSchema(context.Background()).Attributes)
	if diags.HasError() {
		t.Fatalf("mapResponseToModel failed: %v", diags)
	}

	if data.Id.IsNull() || data.Id.ValueString() != "img-1" {
		t.Fatalf("expected id img-1, got %v", data.Id)
	}
	if data.ProjectId.ValueString() != "proj-1" {
		t.Fatalf("expected project_id to be preserved, got %v", data.ProjectId)
	}
}
