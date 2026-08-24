package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/rixlhq/rixl-go/sdk/models"
)

func TestPostsDataSourceMapMedia(t *testing.T) {
	ctx := context.Background()
	id := "post-1"
	feedID := "feed-1"
	imageID := "img-1"

	post := models.PostsV1Post{
		ID:                &id,
		FeedID:            &feedID,
		Type:              (*models.CommonV1MediaType)(new("MEDIA_TYPE_IMAGE")),
		PostsV1PostAllOf1: &models.PostsV1PostAllOf1{},
	}
	_ = post.PostsV1PostAllOf1.FromPostsV1PostAllOf1OneOf0(models.PostsV1PostAllOf1OneOf0{
		Image: models.ImagesV1Image{
			ID: &imageID,
		},
	})

	res := models.PostsV1ListPostsResponse{
		Posts: []models.PostsV1Post{post},
	}

	var data PostsDataSourceModel
	diags := mapResponseToModel(ctx, res, &data, PostsDataSourceSchema(ctx).Attributes)
	if diags.HasError() {
		t.Fatalf("mapResponseToModel failed: %v", diags)
	}

	if data.Posts.IsNull() || data.Posts.IsUnknown() {
		t.Fatalf("posts list is null/unknown")
	}

	elems := data.Posts.Elements()
	if len(elems) != 1 {
		t.Fatalf("expected 1 post, got %d", len(elems))
	}

	obj, ok := elems[0].(basetypes.ObjectValue)
	if !ok {
		t.Fatalf("expected object value, got %T", elems[0])
	}

	if obj.Attributes()["image"].IsNull() {
		t.Fatalf("image nested object should not be null")
	}

	t.Logf("mapped post: %v", obj)
}
