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
	"github.com/rixlhq/rixl-go/sdk/images"
	"github.com/rixlhq/rixl-go/sdk/models"
)

var _ resource.Resource = (*imageResource)(nil)

func NewImageResource() resource.Resource {
	return &imageResource{}
}

type imageResource struct {
	client *sdk.Client
}

func (r *imageResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_image"
}

func (r *imageResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = ImageResourceSchema(ctx)
}

func (r *imageResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *imageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ImageResourceModel
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

	uploadReq := models.ImagesV1CreateImageUploadRequest{
		ProjectID: &projectID,
		Name:      &name,
	}

	upload, err := r.client.Images.CreateImageUpload(ctx, projectID, uploadReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create image upload", err.Error())
		return
	}

	if upload.ImageID == nil || upload.UploadURL == nil {
		resp.Diagnostics.AddError("Image upload response missing id or url", "")
		return
	}

	if err := r.client.UploadFile(ctx, *upload.UploadURL, filePath); err != nil {
		resp.Diagnostics.AddError("Failed to upload image file", err.Error())
		return
	}

	img, err := r.waitForImage(ctx, *upload.ImageID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read image after upload", err.Error())
		return
	}

	resp.Diagnostics.Append(mapResponseToModel(ctx, img, &data, ImageResourceSchema(ctx).Attributes)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.ProjectId.IsNull() || data.ProjectId.IsUnknown() {
		data.ProjectId = types.StringValue(projectID)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *imageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ImageResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.client.Images.GetImage(ctx, data.Id.ValueString())
	if err != nil {
		var httpErr *images.ClientHttpError[struct{}]
		if errors.As(err, &httpErr) && httpErr.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read image", err.Error())
		return
	}

	resp.Diagnostics.Append(mapResponseToModel(ctx, res.Image, &data, ImageResourceSchema(ctx).Attributes)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *imageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state, plan ImageResourceModel
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
	data := dataAny.(*ImageResourceModel)

	if !plan.Visibility.Equal(state.Visibility) && !plan.Visibility.IsNull() && !plan.Visibility.IsUnknown() {
		vis := models.CommonV1Visibility(plan.Visibility.ValueString())
		projectID := data.ProjectId.ValueString()
		imageID := data.Id.ValueString()
		body := models.ImagesV1UpdateImageVisibilityRequest{
			ProjectID:  &projectID,
			ImageID:    &imageID,
			Visibility: &vis,
		}

		_, err := r.client.Images.UpdateImageVisibility(ctx, projectID, imageID, body)
		if err != nil {
			var httpErr *images.ClientHttpError[struct{}]
			if errors.As(err, &httpErr) && httpErr.StatusCode == http.StatusNotFound {
				resp.State.RemoveResource(ctx)
				return
			}
			resp.Diagnostics.AddError("Failed to update image visibility", err.Error())
			return
		}
	}

	res, err := r.client.Images.GetImage(ctx, data.Id.ValueString())
	if err != nil {
		var httpErr *images.ClientHttpError[struct{}]
		if errors.As(err, &httpErr) && httpErr.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read image after update", err.Error())
		return
	}

	resp.Diagnostics.Append(mapResponseToModel(ctx, res.Image, data, ImageResourceSchema(ctx).Attributes)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.ProjectId.IsNull() || data.ProjectId.IsUnknown() {
		data.ProjectId = state.ProjectId
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *imageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ImageResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := r.client.Images.DeleteImage(ctx, data.ProjectId.ValueString(), data.Id.ValueString()); err != nil {
		var httpErr *images.ClientHttpError[struct{}]
		if errors.As(err, &httpErr) && httpErr.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to delete image", err.Error())
	}
}

func (r *imageResource) waitForImage(ctx context.Context, imageID string) (*models.ImagesV1Image, error) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	timeout := time.NewTimer(30 * time.Second)
	defer timeout.Stop()

	var lastErr error
	for {
		res, err := r.client.Images.GetImage(ctx, imageID)
		if err == nil && res.Image != nil {
			return res.Image, nil
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
				return nil, fmt.Errorf("image %s not ready after 30s: %w", imageID, lastErr)
			}
			return nil, fmt.Errorf("image %s not ready after 30s", imageID)
		case <-ticker.C:
		}
	}
}

func ImageResourceSchema(_ context.Context) schema.Schema {
	attrs := imageResourceAttributes()
	attrs["file_path"] = schema.StringAttribute{
		Required:      true,
		PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
	}
	attrs["name"] = schema.StringAttribute{
		Optional: true,
		Computed: true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	}
	attrs["project_id"] = schema.StringAttribute{
		Required:      true,
		PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
	}
	return schema.Schema{Attributes: attrs}
}

type ImageResourceModel struct {
	ImageDataSourceModel

	ProjectId types.String `tfsdk:"project_id"`
	Name      types.String `tfsdk:"name"`
	FilePath  types.String `tfsdk:"file_path"`
}
