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
	"github.com/rixlhq/rixl-go/sdk/customdomains"
)

var _ resource.Resource = (*domainAutoJoinResource)(nil)

func NewDomainAutoJoinResource() resource.Resource {
	return &domainAutoJoinResource{}
}

type domainAutoJoinResource struct {
	client *sdk.Client
}

func (r *domainAutoJoinResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain_auto_join"
}

func (r *domainAutoJoinResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = DomainAutoJoinResourceSchema()
}

func (r *domainAutoJoinResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *domainAutoJoinResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DomainAutoJoinResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgID := data.OrgId.ValueString()
	body := r.buildRequest(data)

	setting, err := r.client.CustomDomains.SetDomainAutoJoin(ctx, orgID, body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to set domain auto-join", err.Error())
		return
	}

	resp.Diagnostics.Append(mapResponseToModel(ctx, setting, &data, DomainAutoJoinResourceSchema().Attributes)...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.OrgId = types.StringValue(orgID)
	if data.Id.IsNull() || data.Id.IsUnknown() {
		data.Id = types.StringValue(orgID)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *domainAutoJoinResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DomainAutoJoinResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgID := data.OrgId.ValueString()

	setting, err := r.client.CustomDomains.GetDomainAutoJoin(ctx, orgID, nil)
	if err != nil {
		var httpErr *customdomains.ClientHttpError[struct{}]
		if errors.As(err, &httpErr) && httpErr.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read domain auto-join", err.Error())
		return
	}

	if setting.Present != nil && !*setting.Present {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(mapResponseToModel(ctx, setting, &data, DomainAutoJoinResourceSchema().Attributes)...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.OrgId = types.StringValue(orgID)
	if data.Id.IsNull() || data.Id.IsUnknown() {
		data.Id = types.StringValue(orgID)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *domainAutoJoinResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan DomainAutoJoinResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgID := plan.OrgId.ValueString()
	body := r.buildRequest(plan)

	setting, err := r.client.CustomDomains.SetDomainAutoJoin(ctx, orgID, body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update domain auto-join", err.Error())
		return
	}

	resp.Diagnostics.Append(mapResponseToModel(ctx, setting, &plan, DomainAutoJoinResourceSchema().Attributes)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.OrgId = types.StringValue(orgID)
	if plan.Id.IsNull() || plan.Id.IsUnknown() {
		plan.Id = types.StringValue(orgID)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *domainAutoJoinResource) Delete(ctx context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	// The Rixl API does not expose a delete-auto-join endpoint.
	// Disable auto-join and remove from state.
	resp.Diagnostics.AddWarning(
		"Domain auto-join not deleted",
		"The Rixl API does not support deleting domain auto-join settings. The resource has been removed from Terraform state, but the setting may still exist in the platform.",
	)
	resp.State.RemoveResource(ctx)
}

func (r *domainAutoJoinResource) buildRequest(data DomainAutoJoinResourceModel) map[string]any {
	orgID := data.OrgId.ValueString()
	enabled := data.Enabled.ValueBool()
	return map[string]any{
		"enabled": enabled,
		"user": map[string]any{
			"org_id": orgID,
		},
	}
}

func DomainAutoJoinResourceSchema() schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"org_id": schema.StringAttribute{
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"enabled": schema.BoolAttribute{
				Required: true,
			},
			"present": schema.BoolAttribute{
				Computed: true,
			},
		},
	}
}

type DomainAutoJoinResourceModel struct {
	Id      types.String `tfsdk:"id"`
	OrgId   types.String `tfsdk:"org_id"`
	Enabled types.Bool   `tfsdk:"enabled"`
	Present types.Bool   `tfsdk:"present"`
}
