package provider

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rixlhq/rixl-go/sdk"
	"github.com/rixlhq/rixl-go/sdk/models"
	"github.com/rixlhq/rixl-go/sdk/videos"
)

var _ resource.Resource = (*videoResource)(nil)

func NewVideoResource() resource.Resource {
	return &videoResource{}
}

type videoResource struct {
	client *sdk.Client
}

func (r *videoResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_video"
}

func (r *videoResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = VideoResourceSchema(ctx)
}

func (r *videoResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *videoResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data VideoResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectId.ValueString()
	filePath := data.FilePath.ValueString()

	name := data.Name.ValueString()
	if name == "" {
		name = filepath.Base(filePath)
		data.Name = types.StringValue(name)
	}

	uploadReq := models.VideosV1CreateVideoUploadRequest{
		ProjectID: &projectID,
		Name:      &name,
	}

	upload, err := r.client.Videos.CreateVideoUpload(ctx, projectID, uploadReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create video upload", err.Error())
		return
	}

	if upload.VideoID == nil || upload.VideoUploadURL == nil {
		resp.Diagnostics.AddError("Video upload response missing id or url", "")
		return
	}

	if err := r.client.UploadFile(ctx, *upload.VideoUploadURL, filePath); err != nil {
		resp.Diagnostics.AddError("Failed to upload video file", err.Error())
		return
	}

	if upload.PosterUploadURL != nil && !data.PosterFilePath.IsNull() && !data.PosterFilePath.IsUnknown() && data.PosterFilePath.ValueString() != "" {
		if err := r.client.UploadFile(ctx, *upload.PosterUploadURL, data.PosterFilePath.ValueString()); err != nil {
			resp.Diagnostics.AddError("Failed to upload poster file", err.Error())
			return
		}
	}

	video, err := r.waitForVideo(ctx, *upload.VideoID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read video after upload", err.Error())
		return
	}

	resp.Diagnostics.Append(mapResponseToModel(ctx, video, &data, VideoResourceSchema(ctx).Attributes)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.ProjectId.IsNull() || data.ProjectId.IsUnknown() {
		data.ProjectId = types.StringValue(projectID)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *videoResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data VideoResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.client.Videos.GetVideo(ctx, data.Id.ValueString())
	if err != nil {
		var httpErr *videos.ClientHttpError[struct{}]
		if errors.As(err, &httpErr) && httpErr.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read video", err.Error())
		return
	}

	resp.Diagnostics.Append(mapResponseToModel(ctx, res.Video, &data, VideoResourceSchema(ctx).Attributes)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *videoResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state, plan VideoResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Visibility.Equal(state.Visibility) && !plan.Visibility.IsNull() && !plan.Visibility.IsUnknown() {
		vis := models.CommonV1Visibility(plan.Visibility.ValueString())
		projectID := state.ProjectId.ValueString()
		videoID := state.Id.ValueString()
		body := models.VideosV1UpdateVideoVisibilityRequest{
			ProjectID:  &projectID,
			VideoID:    &videoID,
			Visibility: &vis,
		}

		_, err := r.client.Videos.UpdateVideoVisibility(ctx, projectID, videoID, body)
		if err != nil {
			var httpErr *videos.ClientHttpError[struct{}]
			if errors.As(err, &httpErr) && httpErr.StatusCode == http.StatusNotFound {
				resp.State.RemoveResource(ctx)
				return
			}
			resp.Diagnostics.AddError("Failed to update video visibility", err.Error())
			return
		}
	}

	res, err := r.client.Videos.GetVideo(ctx, state.Id.ValueString())
	if err != nil {
		var httpErr *videos.ClientHttpError[struct{}]
		if errors.As(err, &httpErr) && httpErr.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read video after update", err.Error())
		return
	}

	resp.Diagnostics.Append(mapResponseToModel(ctx, res.Video, &plan, VideoResourceSchema(ctx).Attributes)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.ProjectId.IsNull() || plan.ProjectId.IsUnknown() {
		plan.ProjectId = state.ProjectId
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *videoResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data VideoResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := r.client.Videos.DeleteVideo(ctx, data.ProjectId.ValueString(), data.Id.ValueString()); err != nil {
		var httpErr *videos.ClientHttpError[struct{}]
		if errors.As(err, &httpErr) && httpErr.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to delete video", err.Error())
	}
}

func (r *videoResource) waitForVideo(ctx context.Context, videoID string) (*models.VideosV1Video, error) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	timeout := time.NewTimer(60 * time.Second)
	defer timeout.Stop()

	var lastErr error
	for {
		res, err := r.client.Videos.GetVideo(ctx, videoID)
		if err == nil && res.Video != nil {
			return res.Video, nil
		}
		lastErr = err

		if !isTransientWaitError(err) {
			return nil, err
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-timeout.C:
			if lastErr != nil {
				return nil, fmt.Errorf("video %s not ready after 60s: %w", videoID, lastErr)
			}
			return nil, fmt.Errorf("video %s not ready after 60s", videoID)
		case <-ticker.C:
		}
	}
}

func VideoResourceSchema(_ context.Context) schema.Schema {
	attrs := videoResourceAttributes()
	attrs["file_path"] = schema.StringAttribute{
		Required:      true,
		PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
	}
	attrs["poster_file_path"] = schema.StringAttribute{
		Optional:      true,
		PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
	}
	attrs["name"] = schema.StringAttribute{
		Optional: true,
		Computed: true,
	}
	attrs["project_id"] = schema.StringAttribute{
		Required:      true,
		PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
	}
	return schema.Schema{Attributes: attrs}
}

func videoResourceAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"bitrate":   schema.Int64Attribute{Computed: true},
		"codec":     schema.StringAttribute{Computed: true},
		"duration":  schema.StringAttribute{Computed: true},
		"file":      schema.SingleNestedAttribute{Computed: true, Attributes: fileResourceAttributes()},
		"framerate": schema.StringAttribute{Computed: true},
		"hdr":       schema.BoolAttribute{Computed: true},
		"height":    schema.Int64Attribute{Computed: true},
		"id":        schema.StringAttribute{Computed: true},
		"poster":    schema.SingleNestedAttribute{Computed: true, Attributes: imageResourceAttributes()},
		"visibility": schema.StringAttribute{
			Optional: true,
			Computed: true,
		},
		"width": schema.Int64Attribute{Computed: true},
	}
}

type VideoResourceModel struct {
	VideoDataSourceModel
	ProjectId      types.String `tfsdk:"project_id"`
	Name           types.String `tfsdk:"name"`
	FilePath       types.String `tfsdk:"file_path"`
	PosterFilePath types.String `tfsdk:"poster_file_path"`
}
