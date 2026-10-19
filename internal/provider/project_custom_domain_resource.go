package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rixlhq/rixl-go/sdk"
)

var (
	_ resource.Resource                = (*projectCustomDomainResource)(nil)
	_ resource.ResourceWithImportState = (*projectCustomDomainResource)(nil)
)

func NewProjectCustomDomainResource() resource.Resource {
	return &projectCustomDomainResource{}
}

type projectCustomDomainResource struct {
	client *sdk.Client
}

func (r *projectCustomDomainResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_custom_domain"
}

func (r *projectCustomDomainResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = ProjectCustomDomainResourceSchema(ctx)
}

func (r *projectCustomDomainResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *projectCustomDomainResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ProjectCustomDomainModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, diags := modelToMap(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	delete(body, "id")
	delete(body, "org_id")
	delete(body, "project_id")
	delete(body, "created_at")
	delete(body, "updated_at")

	proj, err := r.client.Projects.SetCustomDomain(ctx, data.OrgId.ValueString(), data.ProjectId.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to set project custom domain", err.Error())
		return
	}

	data.Id = types.StringValue(fmt.Sprintf("%s/%s", data.OrgId.ValueString(), data.ProjectId.ValueString()))
	resp.Diagnostics.Append(mapResponseToModel(ctx, proj, &data, ProjectCustomDomainResourceSchema(ctx).Attributes)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *projectCustomDomainResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ProjectCustomDomainModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	proj, err := r.client.Projects.GetProject(ctx, data.OrgId.ValueString(), data.ProjectId.ValueString())
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to get project for custom domain", err.Error())
		return
	}

	if proj.CustomDomain == nil || *proj.CustomDomain == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(mapResponseToModel(ctx, proj, &data, ProjectCustomDomainResourceSchema(ctx).Attributes)...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Id = types.StringValue(fmt.Sprintf("%s/%s", data.OrgId.ValueString(), data.ProjectId.ValueString()))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *projectCustomDomainResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ProjectCustomDomainModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, diags := modelToMap(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	delete(body, "id")
	delete(body, "org_id")
	delete(body, "project_id")
	delete(body, "created_at")
	delete(body, "updated_at")

	proj, err := r.client.Projects.SetCustomDomain(ctx, data.OrgId.ValueString(), data.ProjectId.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update project custom domain", err.Error())
		return
	}

	data.Id = types.StringValue(fmt.Sprintf("%s/%s", data.OrgId.ValueString(), data.ProjectId.ValueString()))
	resp.Diagnostics.Append(mapResponseToModel(ctx, proj, &data, ProjectCustomDomainResourceSchema(ctx).Attributes)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *projectCustomDomainResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ProjectCustomDomainModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.Projects.RemoveCustomDomain(ctx, data.OrgId.ValueString(), data.ProjectId.ValueString())
	if err != nil && !isNotFound(err) {
		resp.Diagnostics.AddError("Failed to remove project custom domain", err.Error())
	}
}

func (r *projectCustomDomainResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "/")
	if len(parts) == 2 {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("org_id"), parts[0])...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_id"), parts[1])...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}
