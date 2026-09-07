package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/rixlhq/rixl-go/sdk/models"
)

func TestVideoChaptersResourceSchemaShape(t *testing.T) {
	sch := VideoChaptersResourceSchema()

	projectID, ok := sch.Attributes["project_id"].(schema.StringAttribute)
	if !ok {
		t.Fatalf("project_id should be a StringAttribute")
	}
	if !projectID.Required {
		t.Fatalf("project_id should be required")
	}
	if len(projectID.PlanModifiers) != 1 {
		t.Fatalf("project_id should require replacement on change")
	}

	videoID, ok := sch.Attributes["video_id"].(schema.StringAttribute)
	if !ok {
		t.Fatalf("video_id should be a StringAttribute")
	}
	if !videoID.Required {
		t.Fatalf("video_id should be required")
	}
	if len(videoID.PlanModifiers) != 1 {
		t.Fatalf("video_id should require replacement on change")
	}

	nested, ok := sch.Attributes["chapters"].(schema.ListNestedAttribute)
	if !ok {
		t.Fatalf("chapters should be a ListNestedAttribute")
	}
	if !nested.Required {
		t.Fatalf("chapters should be required: the resource manages the full set")
	}

	title, ok := nested.NestedObject.Attributes["title"].(schema.StringAttribute)
	if !ok {
		t.Fatalf("chapters.title should be a StringAttribute")
	}
	if !title.Required {
		t.Fatalf("chapters.title should be required")
	}

	start, ok := nested.NestedObject.Attributes["start_time_sec"].(schema.Float64Attribute)
	if !ok {
		t.Fatalf("chapters.start_time_sec should be a Float64Attribute")
	}
	if !start.Required {
		t.Fatalf("chapters.start_time_sec should be required")
	}

	id, ok := sch.Attributes["id"].(schema.StringAttribute)
	if !ok {
		t.Fatalf("id should be a StringAttribute")
	}
	if !id.Computed || id.Required || id.Optional {
		t.Fatalf("id should be computed-only")
	}

	// The chapters API returns no timestamps, so none are modelled.
	for _, key := range []string{"created_at", "updated_at"} {
		if _, ok := sch.Attributes[key]; ok {
			t.Fatalf("%s should not be modelled: the API returns no chapter timestamps", key)
		}
	}
}

func TestVideoChaptersIDScopesVideoUnderProject(t *testing.T) {
	if got := videoChaptersID("proj-1", "vid-2"); got != "proj-1/vid-2" {
		t.Fatalf("id mismatch: got %q", got)
	}
}

func TestVideoChaptersStaleStartsFindsRemovedChaptersForDeleteOne(t *testing.T) {
	current := []models.VideosV1Chapter{
		{Title: new("Intro"), StartTimeSec: new(0.0)},
		{Title: new("Main"), StartTimeSec: new(12.5)},
		{Title: new("Outro"), StartTimeSec: new(60.0)},
	}
	planned := []models.VideosV1Chapter{
		{Title: new("Intro"), StartTimeSec: new(0.0)},
		{Title: new("Main v2"), StartTimeSec: new(12.5)},
	}

	stale := staleChapterStarts(current, planned)
	if len(stale) != 1 || stale[0] != 60 {
		t.Fatalf("expected only start 60 to be deleted one-by-one, got %v", stale)
	}
}

func TestVideoChaptersEmptyPlanDeletesAllRemoteChapters(t *testing.T) {
	current := []models.VideosV1Chapter{
		{Title: new("Intro"), StartTimeSec: new(0.0)},
		{Title: new("Outro"), StartTimeSec: new(60.0)},
	}

	stale := staleChapterStarts(current, nil)
	if len(stale) != 2 {
		t.Fatalf("empty plan should delete every remote chapter, got %v", stale)
	}
}

func TestVideoChaptersDeleteAllSkipsChaptersWithoutStartTime(t *testing.T) {
	current := []models.VideosV1Chapter{{Title: new("Untitled")}}

	if stale := staleChapterStarts(current, nil); len(stale) != 0 {
		t.Fatalf("chapters without a start time cannot be addressed by delete-one, got %v", stale)
	}
}

func TestVideoChaptersFlattenSortsByStartTimeForStableState(t *testing.T) {
	ctx := context.Background()
	in := []models.VideosV1Chapter{
		{Title: new("Outro"), StartTimeSec: new(60.0)},
		{Title: new("Intro"), StartTimeSec: new(0.0)},
		{Title: new("Main"), StartTimeSec: new(12.5)},
	}

	flat, diags := flattenVideoChapters(ctx, in)
	if diags.HasError() {
		t.Fatalf("diags: %v", diags)
	}

	expanded, diags := expandVideoChapters(ctx, flat)
	if diags.HasError() {
		t.Fatalf("diags: %v", diags)
	}
	if len(expanded) != 3 {
		t.Fatalf("round trip should preserve all chapters, got %v", expanded)
	}
	for i, want := range []float64{0, 12.5, 60} {
		if expanded[i].StartTimeSec == nil || *expanded[i].StartTimeSec != want {
			t.Fatalf("position %d should be start %v, got %v", i, want, expanded[i].StartTimeSec)
		}
	}
}
