package provider

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rixlhq/rixl-go/sdk"
	"github.com/rixlhq/rixl-go/sdk/memberships"
)

var _ resource.Resource = (*organizationMemberResource)(nil)

func NewOrganizationMemberResource() resource.Resource {
	return &organizationMemberResource{}
}

type organizationMemberResource struct {
	client *sdk.Client
}

func (r *organizationMemberResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_member"
}

func (r *organizationMemberResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = OrganizationMemberResourceSchema()
}

func (r *organizationMemberResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *organizationMemberResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Members are not created via API; they join through invitations. This
	// resource adopts an existing member into Terraform management by setting
	// the desired role and state.
	var data OrganizationMemberResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgID := data.OrgId.ValueString()
	userID := data.UserId.ValueString()

	member, err := r.findMember(ctx, orgID, userID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to find organization member", err.Error())
		return
	}
	if member == nil {
		resp.Diagnostics.AddError("Organization member not found", fmt.Sprintf("User %s is not a member of organization %s", userID, orgID))
		return
	}

	// Set the desired role if it differs from the current one.
	if !data.Role.IsNull() && !data.Role.IsUnknown() {
		role := data.Role.ValueString()
		body := map[string]any{
			"user":    map[string]any{"org_id": orgID},
			"user_id": userID,
			"role":    role,
		}
		if _, err := r.client.Memberships.UpdateMemberRole(ctx, orgID, userID, body); err != nil {
			resp.Diagnostics.AddError("Failed to set member role", err.Error())
			return
		}
	}

	// Set the desired state (suspend/reactivate).
	if !data.State.IsNull() && !data.State.IsUnknown() {
		if err := r.applyState(ctx, orgID, userID, data.State.ValueString()); err != nil {
			resp.Diagnostics.AddError("Failed to set member state", err.Error())
			return
		}
	}

	// Re-read to get the final state.
	member, err = r.findMember(ctx, orgID, userID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read organization member after create", err.Error())
		return
	}
	if member == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(mapResponseToModel(ctx, member, &data, OrganizationMemberResourceSchema().Attributes)...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.OrgId = types.StringValue(orgID)
	data.UserId = types.StringValue(userID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *organizationMemberResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data OrganizationMemberResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgID := data.OrgId.ValueString()
	userID := data.UserId.ValueString()

	member, err := r.findMember(ctx, orgID, userID)
	if err != nil {
		var httpErr *memberships.ClientHttpError[struct{}]
		if errors.As(err, &httpErr) && httpErr.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read organization member", err.Error())
		return
	}
	if member == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(mapResponseToModel(ctx, member, &data, OrganizationMemberResourceSchema().Attributes)...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.OrgId = types.StringValue(orgID)
	data.UserId = types.StringValue(userID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *organizationMemberResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state, plan OrganizationMemberResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgID := plan.OrgId.ValueString()
	userID := plan.UserId.ValueString()

	// Update role if changed.
	if state.Role.ValueString() != plan.Role.ValueString() && !plan.Role.IsNull() && !plan.Role.IsUnknown() {
		body := map[string]any{
			"user":    map[string]any{"org_id": orgID},
			"user_id": userID,
			"role":    plan.Role.ValueString(),
		}
		if _, err := r.client.Memberships.UpdateMemberRole(ctx, orgID, userID, body); err != nil {
			resp.Diagnostics.AddError("Failed to update member role", err.Error())
			return
		}
	}

	// Update state if changed.
	if state.State.ValueString() != plan.State.ValueString() && !plan.State.IsNull() && !plan.State.IsUnknown() {
		if err := r.applyState(ctx, orgID, userID, plan.State.ValueString()); err != nil {
			resp.Diagnostics.AddError("Failed to update member state", err.Error())
			return
		}
	}

	// Re-read to get the final state.
	member, err := r.findMember(ctx, orgID, userID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read organization member after update", err.Error())
		return
	}
	if member == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(mapResponseToModel(ctx, member, &plan, OrganizationMemberResourceSchema().Attributes)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.OrgId = types.StringValue(orgID)
	plan.UserId = types.StringValue(userID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *organizationMemberResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data OrganizationMemberResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.Memberships.RemoveMember(ctx, data.OrgId.ValueString(), data.UserId.ValueString(), nil)
	if err != nil {
		var httpErr *memberships.ClientHttpError[struct{}]
		if errors.As(err, &httpErr) && httpErr.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to remove organization member", err.Error())
		return
	}
}
