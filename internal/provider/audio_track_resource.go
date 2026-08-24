package provider

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rixlhq/rixl-go/sdk"
	"github.com/rixlhq/rixl-go/sdk/audiotracks"
	"github.com/rixlhq/rixl-go/sdk/models"
)

var _ resource.Resource = (*audioTrackResource)(nil)

func NewAudioTrackResource() resource.Resource {
	return &audioTrackResource{}
}

type audioTrackResource struct {
	client *sdk.Client
}

func (r *audioTrackResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_audio_track"
}

func (r *audioTrackResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = AudioTrackResourceSchema()
}

func (r *audioTrackResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *audioTrackResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data AudioTrackResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectId.ValueString()
	videoID := data.VideoId.ValueString()
	languageCode := data.LanguageCode.ValueString()
	label := data.Label.ValueString()

	body := map[string]any{
		"project_id": projectID,
		"video_id":   videoID,
		"items": []map[string]any{{
			"language_code": languageCode,
			"label":         label,
		}},
	}

	upload, err := r.client.AudioTracks.CreateAudioTrackUpload(ctx, projectID, videoID, body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create audio track upload", err.Error())
		return
	}

	var trackID, uploadURL string
	for _, t := range upload.Targets {
		if t.LanguageCode != nil && *t.LanguageCode == languageCode {
			if t.ID != nil {
				trackID = *t.ID
			}
			if t.UploadURL != nil {
				uploadURL = *t.UploadURL
			}
			break
		}
	}

	if trackID == "" || uploadURL == "" {
		resp.Diagnostics.AddError("Audio track upload response missing track id or upload url", "")
		return
	}

	if err := r.client.UploadFile(ctx, uploadURL, data.FilePath.ValueString()); err != nil {
		resp.Diagnostics.AddError("Failed to upload audio track file", err.Error())
		return
	}

	audioTrack, err := r.readAudioTrackByID(ctx, videoID, trackID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read audio track after upload", err.Error())
		return
	}

	resp.Diagnostics.Append(mapResponseToModel(ctx, audioTrack, &data, AudioTrackResourceSchema().Attributes)...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.ProjectId = types.StringValue(projectID)
	data.VideoId = types.StringValue(videoID)
	data.LanguageCode = types.StringValue(languageCode)
	data.Label = types.StringValue(label)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *audioTrackResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data AudioTrackResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	audioTrack, err := r.readAudioTrackByID(ctx, data.VideoId.ValueString(), data.Id.ValueString())
	if err != nil {
		var httpErr *audiotracks.ClientHttpError[struct{}]
		if errors.As(err, &httpErr) && httpErr.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read audio track", err.Error())
		return
	}

	resp.Diagnostics.Append(mapResponseToModel(ctx, audioTrack, &data, AudioTrackResourceSchema().Attributes)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *audioTrackResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// The API does not support updating an audio track in place.
	// All mutable inputs use RequiresReplace, so Terraform will replace the resource instead.
	var state, plan AudioTrackResourceModel
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
	data := dataAny.(*AudioTrackResourceModel)

	audioTrack, err := r.readAudioTrackByID(ctx, data.VideoId.ValueString(), data.Id.ValueString())
	if err != nil {
		var httpErr *audiotracks.ClientHttpError[struct{}]
		if errors.As(err, &httpErr) && httpErr.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read audio track", err.Error())
		return
	}

	resp.Diagnostics.Append(mapResponseToModel(ctx, audioTrack, data, AudioTrackResourceSchema().Attributes)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *audioTrackResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data AudioTrackResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := r.client.AudioTracks.DeleteAudioTrack(ctx, data.ProjectId.ValueString(), data.VideoId.ValueString(), data.Id.ValueString()); err != nil {
		var httpErr *audiotracks.ClientHttpError[struct{}]
		if errors.As(err, &httpErr) && httpErr.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to delete audio track", err.Error())
	}
}

func (r *audioTrackResource) readAudioTrackByID(ctx context.Context, videoID, trackID string) (*models.VideosV1AudioTrack, error) {
	res, err := r.client.AudioTracks.ListAudioTracks(ctx, videoID)
	if err != nil {
		return nil, err
	}
	for i := range res.AudioTracks {
		if res.AudioTracks[i].ID != nil && *res.AudioTracks[i].ID == trackID {
			return &res.AudioTracks[i], nil
		}
	}
	return nil, fmt.Errorf("audio track %s not found for video %s", trackID, videoID)
}

func AudioTrackResourceSchema() schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"project_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"video_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"language_code": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"label": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"file_path": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"codec": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

type AudioTrackResourceModel struct {
	Id           types.String `tfsdk:"id"`
	ProjectId    types.String `tfsdk:"project_id"`
	VideoId      types.String `tfsdk:"video_id"`
	LanguageCode types.String `tfsdk:"language_code"`
	Label        types.String `tfsdk:"label"`
	FilePath     types.String `tfsdk:"file_path"`
	Codec        types.String `tfsdk:"codec"`
}
