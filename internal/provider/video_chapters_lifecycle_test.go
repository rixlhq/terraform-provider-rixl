package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	validatorPkg "github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rixlhq/rixl-go/sdk"
)

func TestVideoChaptersValidatorRejectsUnsorted(t *testing.T) {
	ctx := context.Background()
	validator := videoChaptersListValidator{}

	items := []VideoChapterItemModel{
		{Title: types.StringValue("Outro"), StartTimeSec: types.Float64Value(60.0)},
		{Title: types.StringValue("Intro"), StartTimeSec: types.Float64Value(0.0)},
	}
	listVal, diags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: videoChapterAttrTypes()}, items)
	if diags.HasError() {
		t.Fatalf("failed to create test list: %v", diags)
	}

	req := validatorPkg.ListRequest{
		Path:        path.Root("chapters"),
		ConfigValue: listVal,
	}
	var resp validatorPkg.ListResponse
	validator.ValidateList(ctx, req, &resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected validator error for unsorted chapters, got none")
	}
	found := false
	for _, diag := range resp.Diagnostics.Errors() {
		if strings.Contains(diag.Summary(), "Unsorted") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected Unsorted error, got: %v", resp.Diagnostics)
	}
}

func TestVideoChaptersValidatorRejectsDuplicates(t *testing.T) {
	ctx := context.Background()
	validator := videoChaptersListValidator{}

	items := []VideoChapterItemModel{
		{Title: types.StringValue("Part 1"), StartTimeSec: types.Float64Value(10.0)},
		{Title: types.StringValue("Part 2"), StartTimeSec: types.Float64Value(10.0)},
	}
	listVal, diags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: videoChapterAttrTypes()}, items)
	if diags.HasError() {
		t.Fatalf("failed to create test list: %v", diags)
	}

	req := validatorPkg.ListRequest{
		Path:        path.Root("chapters"),
		ConfigValue: listVal,
	}
	var resp validatorPkg.ListResponse
	validator.ValidateList(ctx, req, &resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected validator error for duplicate start_time_sec, got none")
	}
	found := false
	for _, diag := range resp.Diagnostics.Errors() {
		if strings.Contains(diag.Summary(), "Duplicate") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected Duplicate error, got: %v", resp.Diagnostics)
	}
}

func TestVideoChaptersValidatorAcceptsValidChapters(t *testing.T) {
	ctx := context.Background()
	validator := videoChaptersListValidator{}

	items := []VideoChapterItemModel{
		{Title: types.StringValue("Intro"), StartTimeSec: types.Float64Value(0.0)},
		{Title: types.StringValue("Main"), StartTimeSec: types.Float64Value(12.5)},
		{Title: types.StringValue("Outro"), StartTimeSec: types.Float64Value(60.0)},
	}
	listVal, diags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: videoChapterAttrTypes()}, items)
	if diags.HasError() {
		t.Fatalf("failed to create test list: %v", diags)
	}

	req := validatorPkg.ListRequest{
		Path:        path.Root("chapters"),
		ConfigValue: listVal,
	}
	var resp validatorPkg.ListResponse
	validator.ValidateList(ctx, req, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("expected no errors for valid sorted chapters, got: %v", resp.Diagnostics)
	}
}

func TestVideoChaptersDeletePreservesFractionalPosition(t *testing.T) {
	var receivedMethod, receivedPath, receivedQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedMethod = r.Method
		receivedPath = r.URL.Path
		receivedQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	opt, err := baseURLRewriter(srv.URL)
	if err != nil {
		t.Fatalf("baseURLRewriter: %v", err)
	}
	client, err := sdk.New("test-key", sdk.WithHTTPClient(newHTTPClient()), opt)
	if err != nil {
		t.Fatalf("sdk.New: %v", err)
	}

	r := &videoChaptersResource{client: client}
	ctx := context.Background()

	err = r.deleteVideoChapter(ctx, "p1", "v1", 12.5)
	if err != nil {
		t.Fatalf("deleteVideoChapter: %v", err)
	}

	if receivedMethod != http.MethodDelete {
		t.Errorf("expected DELETE method, got %q", receivedMethod)
	}
	expectedPath := "/media/v1/projects/p1/videos/v1/chapters"
	if receivedPath != expectedPath {
		t.Errorf("expected path %q, got %q", expectedPath, receivedPath)
	}
	if !strings.Contains(receivedQuery, "chapters.start_time_sec=12.5") {
		t.Errorf("expected query to preserve fractional start_time_sec=12.5, got %q", receivedQuery)
	}
}

func TestVideoChaptersCreateCompensatesOnUpsertFailure(t *testing.T) {
	var deletedStarts []string
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete && r.URL.Query().Get("chapters.title") == "" {
			deletedStarts = append(deletedStarts, r.URL.Query().Get("chapters.start_time_sec"))
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
			return
		}
		callCount++
		// First upsert succeeds, second upsert fails with 500
		if callCount == 1 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
		} else {
			http.Error(w, "server error", http.StatusInternalServerError)
		}
	}))
	defer srv.Close()

	opt, err := baseURLRewriter(srv.URL)
	if err != nil {
		t.Fatalf("baseURLRewriter: %v", err)
	}
	client, err := sdk.New("test-key", sdk.WithHTTPClient(newHTTPClient()), opt)
	if err != nil {
		t.Fatalf("sdk.New: %v", err)
	}

	r := &videoChaptersResource{client: client}
	ctx := context.Background()

	items := []VideoChapterItemModel{
		{Title: types.StringValue("Intro"), StartTimeSec: types.Float64Value(0.0)},
		{Title: types.StringValue("Main"), StartTimeSec: types.Float64Value(12.5)},
	}
	chaptersList, _ := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: videoChapterAttrTypes()}, items)
	model := VideoChaptersResourceModel{
		ProjectId: types.StringValue("p1"),
		VideoId:   types.StringValue("v1"),
		Chapters:  chaptersList,
	}

	sch := VideoChaptersResourceSchema()
	plan := tfsdk.Plan{Schema: sch}
	diags := plan.Set(ctx, &model)
	if diags.HasError() {
		t.Fatalf("plan.Set: %v", diags)
	}

	state := tfsdk.State{Schema: sch}
	emptyModel := VideoChaptersResourceModel{
		Chapters: types.ListNull(types.ObjectType{AttrTypes: videoChapterAttrTypes()}),
	}
	diags = state.Set(ctx, &emptyModel)
	if diags.HasError() {
		t.Fatalf("state.Set: %v", diags)
	}

	req := resource.CreateRequest{
		Plan: plan,
	}
	resp := &resource.CreateResponse{
		State: state,
	}

	r.Create(ctx, req, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected Create to fail with error")
	}
	// The first chapter (0.0) should have been compensated (deleted)
	if len(deletedStarts) != 1 || deletedStarts[0] != "0" {
		t.Fatalf("expected compensation delete for chapter 0, got %v", deletedStarts)
	}
}
