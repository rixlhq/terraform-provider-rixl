package provider

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rixlhq/rixl-go/sdk"
	"github.com/rixlhq/rixl-go/sdk/memberships"
)

var _ resource.Resource = (*organizationResource)(nil)

func NewOrganizationResource() resource.Resource {
	return &organizationResource{}
}

type organizationResource struct {
	client *sdk.Client
}

func (r *organizationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization"
}

func (r *organizationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = OrganizationResourceSchema()
}

func (r *organizationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *organizationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Organizations are not created via this API. This resource adopts an
	// existing organization into Terraform management by setting the desired
	// full_name and username.
	var data OrganizationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgID := data.OrgId.ValueString()

	if !data.FullName.IsNull() && !data.FullName.IsUnknown() {
		body := map[string]any{
			"user":      map[string]any{"org_id": orgID},
			"full_name": data.FullName.ValueString(),
		}
		if _, err := r.client.Memberships.UpdateOrgName(ctx, orgID, body); err != nil {
			resp.Diagnostics.AddError("Failed to set organization name", err.Error())
			return
		}
	}

	if !data.Username.IsNull() && !data.Username.IsUnknown() {
		body := map[string]any{
			"user":     map[string]any{"org_id": orgID},
			"username": data.Username.ValueString(),
		}
		if _, err := r.client.Memberships.UpdateOrgUsername(ctx, orgID, body); err != nil {
			resp.Diagnostics.AddError("Failed to set organization username", err.Error())
			return
		}
	}

	r.readBack(ctx, orgID, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *organizationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data OrganizationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgID := data.OrgId.ValueString()
	r.readBack(ctx, orgID, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *organizationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state, plan OrganizationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgID := plan.OrgId.ValueString()

	if state.FullName.ValueString() != plan.FullName.ValueString() && !plan.FullName.IsNull() && !plan.FullName.IsUnknown() {
		body := map[string]any{
			"user":      map[string]any{"org_id": orgID},
			"full_name": plan.FullName.ValueString(),
		}
		if _, err := r.client.Memberships.UpdateOrgName(ctx, orgID, body); err != nil {
			resp.Diagnostics.AddError("Failed to update organization name", err.Error())
			return
		}
	}

	if state.Username.ValueString() != plan.Username.ValueString() && !plan.Username.IsNull() && !plan.Username.IsUnknown() {
		body := map[string]any{
			"user":     map[string]any{"org_id": orgID},
			"username": plan.Username.ValueString(),
		}
		if _, err := r.client.Memberships.UpdateOrgUsername(ctx, orgID, body); err != nil {
			resp.Diagnostics.AddError("Failed to update organization username", err.Error())
			return
		}
	}

	r.readBack(ctx, orgID, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *organizationResource) Delete(ctx context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	// The Rixl API does not support deleting organizations.
	resp.Diagnostics.AddWarning(
		"Organization not deleted",
		"The Rixl API does not support deleting organizations. The resource has been removed from Terraform state, but the organization still exists in the platform.",
	)
	resp.State.RemoveResource(ctx)
}

// readBack fetches the current organization info via ListMemberships and
// populates the model with the latest first_name, last_name, and username.
func (r *organizationResource) readBack(ctx context.Context, orgID string, data *OrganizationResourceModel, diags *diag.Diagnostics) {
	listResp, err := r.client.Memberships.ListMemberships(ctx, &memberships.ListMembershipsParams{
		Limit: new(int32(100)),
	}, nil)
	if err != nil {
		var httpErr *memberships.ClientHttpError[struct{}]
		if errors.As(err, &httpErr) && httpErr.StatusCode == http.StatusNotFound {
			diags.AddWarning("Organization not found", fmt.Sprintf("Organization %s not found in memberships", orgID))
			return
		}
		diags.AddError("Failed to list memberships", err.Error())
		return
	}

	for i := range listResp.Memberships {
		m := listResp.Memberships[i]
		if m.OrgID != nil && *m.OrgID == orgID {
			data.OrgId = types.StringValue(orgID)
			if m.OrganizationFirstName != nil {
				data.FirstName = types.StringValue(*m.OrganizationFirstName)
			}
			if m.OrganizationLastName != nil {
				data.LastName = types.StringValue(*m.OrganizationLastName)
			}
			if m.OrganizationUsername != nil {
				data.Username = types.StringValue(*m.OrganizationUsername)
			}
			if data.Id.IsNull() || data.Id.IsUnknown() {
				data.Id = types.StringValue(orgID)
			}
			return
		}
	}

	diags.AddWarning("Organization not found", fmt.Sprintf("Organization %s not found in current user's memberships", orgID))
}
