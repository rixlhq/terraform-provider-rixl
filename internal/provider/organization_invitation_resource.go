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
	"github.com/rixlhq/rixl-go/sdk/memberships"
)

var _ resource.Resource = (*organizationInvitationResource)(nil)

func NewOrganizationInvitationResource() resource.Resource {
	return &organizationInvitationResource{}
}

type organizationInvitationResource struct {
	client *sdk.Client
}

func (r *organizationInvitationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_invitation"
}

func (r *organizationInvitationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = OrganizationInvitationResourceSchema()
}

func (r *organizationInvitationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *organizationInvitationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data OrganizationInvitationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgID := data.OrgId.ValueString()
	body := r.buildCreateBody(data)

	application, err := r.client.Memberships.InviteMember(ctx, orgID, body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create organization invitation", err.Error())
		return
	}

	resp.Diagnostics.Append(mapResponseToModel(ctx, application, &data, OrganizationInvitationResourceSchema().Attributes)...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.OrgId = types.StringValue(orgID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *organizationInvitationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data OrganizationInvitationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgID := data.OrgId.ValueString()
	userID := data.UserId.ValueString()

	params := &memberships.ListMembershipApplicationsParams{
		Limit: new(int32(100)),
	}
	if userID != "" {
		params.UserUserId = new(userID)
	}

	listResp, err := r.client.Memberships.ListMembershipApplications(ctx, params, nil)
	if err != nil {
		var httpErr *memberships.ClientHttpError[struct{}]
		if errors.As(err, &httpErr) && httpErr.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to list membership applications", err.Error())
		return
	}

	for i := range listResp.Applications {
		app := listResp.Applications[i]
		appMap, err := responseToMap(app)
		if err != nil {
			continue
		}
		if id, ok := appMap["id"].(string); ok && id == data.Id.ValueString() {
			resp.Diagnostics.Append(mapResponseToModel(ctx, app, &data, OrganizationInvitationResourceSchema().Attributes)...)
			if resp.Diagnostics.HasError() {
				return
			}
			data.OrgId = types.StringValue(orgID)
			resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
			return
		}
		if uid, ok := appMap["user_id"].(string); ok && uid == userID && data.Id.IsNull() {
			resp.Diagnostics.Append(mapResponseToModel(ctx, app, &data, OrganizationInvitationResourceSchema().Attributes)...)
			if resp.Diagnostics.HasError() {
				return
			}
			data.OrgId = types.StringValue(orgID)
			resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
			return
		}
	}

	resp.State.RemoveResource(ctx)
}

func (r *organizationInvitationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Invitations cannot be updated; changing any field forces replacement.
	var state, plan OrganizationInvitationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Re-send the invitation with the new values by creating a new one.
	orgID := plan.OrgId.ValueString()
	body := r.buildCreateBody(plan)

	_, err := r.client.Memberships.CancelInvitation(ctx, state.OrgId.ValueString(), state.UserId.ValueString(), nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to cancel old invitation during replacement", err.Error())
		return
	}

	application, err := r.client.Memberships.InviteMember(ctx, orgID, body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create new organization invitation", err.Error())
		return
	}

	resp.Diagnostics.Append(mapResponseToModel(ctx, application, &plan, OrganizationInvitationResourceSchema().Attributes)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.OrgId = types.StringValue(orgID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *organizationInvitationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data OrganizationInvitationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.Memberships.CancelInvitation(ctx, data.OrgId.ValueString(), data.UserId.ValueString(), nil)
	if err != nil {
		var httpErr *memberships.ClientHttpError[struct{}]
		if errors.As(err, &httpErr) && httpErr.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to cancel organization invitation", err.Error())
		return
	}
}

func (r *organizationInvitationResource) buildCreateBody(data OrganizationInvitationResourceModel) map[string]any {
	body := map[string]any{
		"user": map[string]any{
			"org_id": data.OrgId.ValueString(),
		},
		"role": data.Role.ValueString(),
	}
	if !data.Username.IsNull() && !data.Username.IsUnknown() {
		body["username"] = data.Username.ValueString()
	}
	return body
}

func OrganizationInvitationResourceSchema() schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"org_id": schema.StringAttribute{
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"user_id": schema.StringAttribute{
				Computed: true,
			},
			"username": schema.StringAttribute{
				Optional:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"role": schema.StringAttribute{
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"state": schema.StringAttribute{
				Computed: true,
			},
			"organization_username": schema.StringAttribute{
				Computed: true,
			},
			"organization_first_name": schema.StringAttribute{
				Computed: true,
			},
			"organization_last_name": schema.StringAttribute{
				Computed: true,
			},
			"created_at": schema.StringAttribute{
				Computed: true,
			},
			"decided_at": schema.StringAttribute{
				Computed: true,
			},
			"invitation_expires_at": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

type OrganizationInvitationResourceModel struct {
	Id                    types.String `tfsdk:"id"`
	OrgId                 types.String `tfsdk:"org_id"`
	UserId                types.String `tfsdk:"user_id"`
	Username              types.String `tfsdk:"username"`
	Role                  types.String `tfsdk:"role"`
	State                 types.String `tfsdk:"state"`
	OrganizationUsername  types.String `tfsdk:"organization_username"`
	OrganizationFirstName types.String `tfsdk:"organization_first_name"`
	OrganizationLastName  types.String `tfsdk:"organization_last_name"`
	CreatedAt             types.String `tfsdk:"created_at"`
	DecidedAt             types.String `tfsdk:"decided_at"`
	InvitationExpiresAt   types.String `tfsdk:"invitation_expires_at"`
}
