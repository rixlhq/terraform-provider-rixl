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
	"github.com/rixlhq/rixl-go/sdk/models"
	"github.com/rixlhq/rixl-go/sdk/subtitles"
)

var _ resource.Resource = (*subtitleResource)(nil)

func NewSubtitleResource() resource.Resource {
	return &subtitleResource{}
}

type subtitleResource struct {
	client *sdk.Client
}

func (r *subtitleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subtitle"
}

func (r *subtitleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = SubtitleResourceSchema()
}

func (r *subtitleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *subtitleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SubtitleResourceModel
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

	upload, err := r.client.Subtitles.CreateSubtitleUpload(ctx, projectID, videoID, body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create subtitle upload", err.Error())
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
		resp.Diagnostics.AddError("Subtitle upload response missing track id or upload url", "")
		return
	}

	if err := r.client.UploadFile(ctx, uploadURL, data.FilePath.ValueString()); err != nil {
		resp.Diagnostics.AddError("Failed to upload subtitle file", err.Error())
		return
	}

	subtitle, err := r.readSubtitleByID(ctx, videoID, trackID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read subtitle after upload", err.Error())
		return
	}

	resp.Diagnostics.Append(mapResponseToModel(ctx, subtitle, &data, SubtitleResourceSchema().Attributes)...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.ProjectId = types.StringValue(projectID)
	data.VideoId = types.StringValue(videoID)
	data.LanguageCode = types.StringValue(languageCode)
	data.Label = types.StringValue(label)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *subtitleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SubtitleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	subtitle, err := r.readSubtitleByID(ctx, data.VideoId.ValueString(), data.Id.ValueString())
	if err != nil {
		var httpErr *subtitles.ClientHttpError[struct{}]
		if errors.As(err, &httpErr) && httpErr.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read subtitle", err.Error())
		return
	}

	resp.Diagnostics.Append(mapResponseToModel(ctx, subtitle, &data, SubtitleResourceSchema().Attributes)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *subtitleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state, plan SubtitleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	subtitle, err := r.readSubtitleByID(ctx, state.VideoId.ValueString(), state.Id.ValueString())
	if err != nil {
		var httpErr *subtitles.ClientHttpError[struct{}]
		if errors.As(err, &httpErr) && httpErr.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read subtitle", err.Error())
		return
	}

	resp.Diagnostics.Append(mapResponseToModel(ctx, subtitle, &plan, SubtitleResourceSchema().Attributes)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *subtitleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SubtitleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := r.client.Subtitles.DeleteSubtitle(ctx, data.ProjectId.ValueString(), data.VideoId.ValueString(), data.Id.ValueString()); err != nil {
		var httpErr *subtitles.ClientHttpError[struct{}]
		if errors.As(err, &httpErr) && httpErr.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to delete subtitle", err.Error())
	}
}

func (r *subtitleResource) readSubtitleByID(ctx context.Context, videoID, subtitleID string) (*models.VideosV1Subtitle, error) {
	res, err := r.client.Subtitles.ListSubtitles(ctx, videoID)
	if err != nil {
		return nil, err
	}
	for i := range res.Subtitles {
		if res.Subtitles[i].ID != nil && *res.Subtitles[i].ID == subtitleID {
			return &res.Subtitles[i], nil
		}
	}
	return nil, fmt.Errorf("subtitle %s not found for video %s", subtitleID, videoID)
}

func SubtitleResourceSchema() schema.Schema {
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
			"format": schema.StringAttribute{
				Computed: true,
			},
			"vtt_path": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

type SubtitleResourceModel struct {
	Id           types.String `tfsdk:"id"`
	ProjectId    types.String `tfsdk:"project_id"`
	VideoId      types.String `tfsdk:"video_id"`
	LanguageCode types.String `tfsdk:"language_code"`
	Label        types.String `tfsdk:"label"`
	FilePath     types.String `tfsdk:"file_path"`
	Format       types.String `tfsdk:"format"`
	VttPath      types.String `tfsdk:"vtt_path"`
}
