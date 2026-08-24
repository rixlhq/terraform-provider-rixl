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
	"github.com/rixlhq/rixl-go/sdk/accesspolicies"
)

var _ resource.Resource = (*accessPolicyResource)(nil)

func NewAccessPolicyResource() resource.Resource {
	return &accessPolicyResource{}
}

type accessPolicyResource struct {
	client *sdk.Client
}

func (r *accessPolicyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_access_policy"
}

func (r *accessPolicyResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = AccessPolicyResourceSchema(ctx)
}

func (r *accessPolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *accessPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data AccessPolicyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgID := data.OrgId.ValueString()

	body, d := modelToMap(ctx, &data)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	delete(body, "id")
	delete(body, "org_id")
	delete(body, "created_by")
	delete(body, "created_at")
	delete(body, "updated_at")

	policy, err := r.client.AccessPolicies.CreatePolicy(ctx, orgID, body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create access policy", err.Error())
		return
	}

	resp.Diagnostics.Append(mapResponseToModel(ctx, policy, &data, AccessPolicyResourceSchema(ctx).Attributes)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.OrgId.IsNull() || data.OrgId.IsUnknown() {
		data.OrgId = types.StringValue(orgID)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *accessPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data AccessPolicyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgID := data.OrgId.ValueString()
	policyID := data.Id.ValueString()

	policy, err := r.client.AccessPolicies.GetPolicy(ctx, orgID, policyID, nil)
	if err != nil {
		var httpErr *accesspolicies.ClientHttpError[struct{}]
		if errors.As(err, &httpErr) && httpErr.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read access policy", err.Error())
		return
	}

	resp.Diagnostics.Append(mapResponseToModel(ctx, policy, &data, AccessPolicyResourceSchema(ctx).Attributes)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.OrgId.IsNull() || data.OrgId.IsUnknown() {
		data.OrgId = types.StringValue(orgID)
	}
	if data.Id.IsNull() || data.Id.IsUnknown() {
		data.Id = types.StringValue(policyID)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *accessPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state, plan AccessPolicyModel
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
	data := dataAny.(*AccessPolicyModel)

	body, d := modelToMap(ctx, data)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	delete(body, "id")
	delete(body, "org_id")
	delete(body, "created_by")
	delete(body, "created_at")
	delete(body, "updated_at")

	policy, err := r.client.AccessPolicies.UpdatePolicy(ctx, data.OrgId.ValueString(), data.Id.ValueString(), body)
	if err != nil {
		var httpErr *accesspolicies.ClientHttpError[struct{}]
		if errors.As(err, &httpErr) && httpErr.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to update access policy", err.Error())
		return
	}

	resp.Diagnostics.Append(mapResponseToModel(ctx, policy, data, AccessPolicyResourceSchema(ctx).Attributes)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.OrgId.IsNull() || data.OrgId.IsUnknown() {
		data.OrgId = state.OrgId
	}
	if data.Id.IsNull() || data.Id.IsUnknown() {
		data.Id = state.Id
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *accessPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data AccessPolicyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := r.client.AccessPolicies.DeletePolicy(ctx, data.OrgId.ValueString(), data.Id.ValueString(), nil); err != nil {
		var httpErr *accesspolicies.ClientHttpError[struct{}]
		if errors.As(err, &httpErr) && httpErr.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to delete access policy", err.Error())
	}
}

func AccessPolicyResourceSchema(_ context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"org_id": schema.StringAttribute{
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"description": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"permissions": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
			},
			"id": schema.StringAttribute{
				Computed: true,
			},
			"created_by": schema.StringAttribute{
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

type AccessPolicyModel struct {
	Id          types.String `tfsdk:"id"`
	OrgId       types.String `tfsdk:"org_id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Permissions types.List   `tfsdk:"permissions"`
	CreatedBy   types.String `tfsdk:"created_by"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}
