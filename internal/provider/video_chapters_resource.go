package provider

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rixlhq/rixl-go/sdk"
	"github.com/rixlhq/rixl-go/sdk/chapters"
	"github.com/rixlhq/rixl-go/sdk/models"
)

// rixl_video_chapters is a collection resource: one Terraform resource manages
// the full chapter set of a single video.
//
// Collection (not per-item) was chosen because the Chapters API exposes no
// per-chapter create endpoint (there is no POST) and chapters are value
// objects keyed by start_time_sec with no server-assigned IDs, so per-item
// resources would have no stable identity and no create path. The only write
// entrypoint, UpdateVideoChapters, operates on the collection path and returns
// the full videos.v1.VideoChapters set, which maps naturally onto
// collection-state refresh.
//
// SDK quirk (verified against rixl-go v0.7.0 sdk/chapters/chapters.gen.go):
// UpdateVideoChapters is a DELETE request on the collection path
// (/media/v1/projects/{project_id}/videos/{video_id}/chapters) with a nil
// body — NOT a DELETE-with-body. The chapter identity travels as query
// parameters (chapters.title, chapters.start_time_sec), so one call carries at
// most one chapter and the resource issues one call per chapter.
//
// Delete semantics:
//   - Resource Delete is delete-all: the API offers no delete-all endpoint for
//     chapters, so Delete iterates over the chapters in state and removes each
//     one via DeleteVideoChapter (DELETE .../chapters/{start_time_sec}).
//   - Resource Update diffs desired vs prior state: chapters present in state
//     but absent from plan are removed via DeleteVideoChapter, then every
//     planned chapter is upserted via UpdateVideoChapters.
//   - An empty chapters list is valid and converges to zero remote chapters
//     while keeping the (empty) resource in state.
//   - 404s encountered while deleting are treated as already-converged, not
//     errors; a 404 on Read removes the resource from state.

var _ resource.Resource = (*videoChaptersResource)(nil)

// NewVideoChaptersResource returns the rixl_video_chapters collection resource.
func NewVideoChaptersResource() resource.Resource {
	return &videoChaptersResource{}
}

type videoChaptersResource struct {
	client *sdk.Client
}

func (r *videoChaptersResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_video_chapters"
}

func (r *videoChaptersResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = VideoChaptersResourceSchema()
}

func (r *videoChaptersResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*sdk.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data type", fmt.Sprintf("expected *sdk.Client, got %T", req.ProviderData))
		return
	}
	r.client = client
}

func (r *videoChaptersResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data VideoChaptersResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectId.ValueString()
	videoID := data.VideoId.ValueString()

	planned, d := expandVideoChapters(ctx, data.Chapters)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	for i := range planned {
		if _, err := r.upsertVideoChapter(ctx, projectID, videoID, planned[i]); err != nil {
			resp.Diagnostics.AddError("Failed to create video chapters", err.Error())
			return
		}
	}

	got, err := r.client.Chapters.GetVideoChapters(ctx, projectID, videoID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read video chapters after create", err.Error())
		return
	}
	resp.Diagnostics.Append(mapResponseToModel(ctx, got, &data, VideoChaptersResourceSchema().Attributes)...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.Id = types.StringValue(videoChaptersID(projectID, videoID))
	if data.ProjectId.IsNull() || data.ProjectId.IsUnknown() {
		data.ProjectId = types.StringValue(projectID)
	}
	if data.VideoId.IsNull() || data.VideoId.IsUnknown() {
		data.VideoId = types.StringValue(videoID)
	}
	resp.Diagnostics.Append(normalizeVideoChapters(ctx, &data)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *videoChaptersResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data VideoChaptersResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectId.ValueString()
	videoID := data.VideoId.ValueString()
	stateID := data.Id.ValueString()

	got, err := r.client.Chapters.GetVideoChapters(ctx, projectID, videoID)
	if err != nil {
		if isVideoChaptersNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read video chapters", err.Error())
		return
	}

	resp.Diagnostics.Append(mapResponseToModel(ctx, got, &data, VideoChaptersResourceSchema().Attributes)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.ProjectId.IsNull() || data.ProjectId.IsUnknown() {
		data.ProjectId = types.StringValue(projectID)
	}
	if data.VideoId.IsNull() || data.VideoId.IsUnknown() {
		data.VideoId = types.StringValue(videoID)
	}
	if data.Id.IsNull() || data.Id.IsUnknown() {
		data.Id = types.StringValue(stateID)
	}
	resp.Diagnostics.Append(normalizeVideoChapters(ctx, &data)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *videoChaptersResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state, plan VideoChaptersResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataAny, diags := mergeStateAndPlan(ctx, &state, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data := dataAny.(*VideoChaptersResourceModel)

	projectID := data.ProjectId.ValueString()
	videoID := data.VideoId.ValueString()

	planned, d := expandVideoChapters(ctx, plan.Chapters)
	resp.Diagnostics.Append(d...)
	current, d := expandVideoChapters(ctx, state.Chapters)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	for _, start := range staleChapterStarts(current, planned) {
		if err := r.deleteVideoChapter(ctx, projectID, videoID, start); err != nil {
			if isVideoChaptersNotFound(err) {
				continue
			}
			resp.Diagnostics.AddError("Failed to remove stale video chapter", err.Error())
			return
		}
	}
	for i := range planned {
		if _, err := r.upsertVideoChapter(ctx, projectID, videoID, planned[i]); err != nil {
			resp.Diagnostics.AddError("Failed to update video chapters", err.Error())
			return
		}
	}

	got, err := r.client.Chapters.GetVideoChapters(ctx, projectID, videoID)
	if err != nil {
		if isVideoChaptersNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read video chapters after update", err.Error())
		return
	}
	resp.Diagnostics.Append(mapResponseToModel(ctx, got, data, VideoChaptersResourceSchema().Attributes)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.ProjectId.IsNull() || data.ProjectId.IsUnknown() {
		data.ProjectId = state.ProjectId
	}
	if data.VideoId.IsNull() || data.VideoId.IsUnknown() {
		data.VideoId = state.VideoId
	}
	if data.Id.IsNull() || data.Id.IsUnknown() {
		data.Id = state.Id
	}
	resp.Diagnostics.Append(normalizeVideoChapters(ctx, data)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *videoChaptersResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data VideoChaptersResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	current, d := expandVideoChapters(ctx, data.Chapters)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	for _, c := range current {
		if c.StartTimeSec == nil {
			continue
		}
		if err := r.deleteVideoChapter(ctx, data.ProjectId.ValueString(), data.VideoId.ValueString(), *c.StartTimeSec); err != nil {
			if isVideoChaptersNotFound(err) {
				continue
			}
			resp.Diagnostics.AddError("Failed to delete video chapters", err.Error())
			return
		}
	}
}

func (r *videoChaptersResource) upsertVideoChapter(ctx context.Context, projectID, videoID string, c models.VideosV1Chapter) (models.VideosV1VideoChapters, error) {
	params := &chapters.UpdateVideoChaptersParams{
		ChaptersTitle:        c.Title,
		ChaptersStartTimeSec: c.StartTimeSec,
	}
	return r.client.Chapters.UpdateVideoChapters(ctx, projectID, videoID, params)
}

func (r *videoChaptersResource) deleteVideoChapter(ctx context.Context, projectID, videoID string, startTimeSec float64) error {
	// DeleteVideoChapter addresses a chapter by integer seconds in the path
	// while chapter positions are doubles; truncate toward zero to satisfy the
	// API's int64 path parameter.
	_, err := r.client.Chapters.DeleteVideoChapter(ctx, projectID, videoID, int64(startTimeSec))
	return err
}

func isVideoChaptersNotFound(err error) bool {
	var httpErr *chapters.ClientHttpError[struct{}]
	return errors.As(err, &httpErr) && httpErr.StatusCode == http.StatusNotFound
}
