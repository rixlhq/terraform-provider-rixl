package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rixlhq/rixl-go/sdk"
	"github.com/rixlhq/rixl-go/sdk/models"
)

var _ resource.Resource = (*projectResource)(nil)

func NewProjectResource() resource.Resource {
	return &projectResource{}
}

type projectResource struct {
	client *sdk.Client
}

func (r *projectResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

func (r *projectResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = ProjectResourceSchema(ctx)
}

func (r *projectResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *projectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ProjectModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := data.Name.ValueString()
	body := models.CreateProjectJSONRequest{
		Name: stringPtr(name),
	}

	if !data.Regions.IsNull() && !data.Regions.IsUnknown() {
		var regions []string
		resp.Diagnostics.Append(data.Regions.ElementsAs(ctx, &regions, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		body.Regions = regions
	}

	if !data.VideoQuality.IsNull() && !data.VideoQuality.IsUnknown() {
		q := models.CommonV1VideoQuality(data.VideoQuality.ValueString())
		body.VideoQuality = &q
	}

	project, err := r.client.Projects.CreateProject(ctx, data.OrgId.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create project", err.Error())
		return
	}

	resp.Diagnostics.Append(mapResponseToModel(ctx, project, &data, ProjectResourceSchema(ctx).Attributes)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *projectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ProjectModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	project, err := r.client.Projects.GetProject(ctx, data.OrgId.ValueString(), data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read project", err.Error())
		return
	}

	resp.Diagnostics.Append(mapResponseToModel(ctx, project, &data, ProjectResourceSchema(ctx).Attributes)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *projectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ProjectModel
	var state ProjectModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgID := plan.OrgId.ValueString()
	projectID := state.Id.ValueString()

	if !plan.Name.Equal(state.Name) {
		body := models.UpdateProjectNameJSONRequest{
			ProjectID: stringPtr(projectID),
			OrgID:     stringPtr(orgID),
			Name:      stringPtr(plan.Name.ValueString()),
		}
		if _, err := r.client.Projects.UpdateProjectName(ctx, orgID, projectID, body); err != nil {
			resp.Diagnostics.AddError("Failed to update project name", err.Error())
			return
		}
	}

	if !plan.VideoQuality.Equal(state.VideoQuality) && !plan.VideoQuality.IsUnknown() && !plan.VideoQuality.IsNull() {
		body := models.UpdateVideoQualityJSONRequest{
			ProjectID:    stringPtr(projectID),
			OrgID:        stringPtr(orgID),
			VideoQuality: models.CommonV1VideoQuality(plan.VideoQuality.ValueString()),
		}
		if _, err := r.client.Projects.UpdateVideoQuality(ctx, orgID, projectID, body); err != nil {
			resp.Diagnostics.AddError("Failed to update project video quality", err.Error())
			return
		}
	}

	project, err := r.client.Projects.GetProject(ctx, orgID, projectID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read project after update", err.Error())
		return
	}

	var data ProjectModel
	resp.Diagnostics.Append(mapResponseToModel(ctx, project, &data, ProjectResourceSchema(ctx).Attributes)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *projectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ProjectModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := r.client.Projects.DeleteProject(ctx, data.OrgId.ValueString(), data.Id.ValueString()); err != nil {
		resp.Diagnostics.AddError("Failed to delete project", err.Error())
		return
	}
}

func ProjectResourceSchema(_ context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"org_id": schema.StringAttribute{
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"regions": schema.ListAttribute{
				ElementType:   types.StringType,
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.List{listplanmodifier.RequiresReplace()},
			},
			"video_quality": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"custom_domain": schema.StringAttribute{
				Computed: true,
			},
			"id": schema.StringAttribute{
				Computed: true,
			},
			"created_at": schema.StringAttribute{
				Computed: true,
			},
			"updated_at": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

type ProjectModel struct {
	CreatedAt    types.String `tfsdk:"created_at"`
	CustomDomain types.String `tfsdk:"custom_domain"`
	Id           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	OrgId        types.String `tfsdk:"org_id"`
	Regions      types.List   `tfsdk:"regions"`
	UpdatedAt    types.String `tfsdk:"updated_at"`
	VideoQuality types.String `tfsdk:"video_quality"`
}
