package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/rixlhq/rixl-go/sdk"
)

var (
	_ resource.Resource                = (*policyAttachmentResource)(nil)
	_ resource.ResourceWithImportState = (*policyAttachmentResource)(nil)
)

func NewPolicyAttachmentResource() resource.Resource {
	return &policyAttachmentResource{}
}

type policyAttachmentResource struct {
	client *sdk.Client
}

func (r *policyAttachmentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_policy_attachment"
}

func (r *policyAttachmentResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = PolicyAttachmentResourceSchema(ctx)
}

func (r *policyAttachmentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *policyAttachmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data PolicyAttachmentModel
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
	delete(body, "policy_id")
	delete(body, "created_at")

	res, err := r.client.AccessPolicies.AttachPolicy(ctx, data.OrgId.ValueString(), data.PolicyId.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to attach policy", err.Error())
		return
	}

	resp.Diagnostics.Append(mapResponseToModel(ctx, res, &data, PolicyAttachmentResourceSchema(ctx).Attributes)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *policyAttachmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data PolicyAttachmentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.client.AccessPolicies.ListPolicyAttachments(ctx, data.OrgId.ValueString(), data.PolicyId.ValueString(), nil)
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to list policy attachments", err.Error())
		return
	}

	var found any
	for _, att := range res.Attachments {
		if att.ID != nil && *att.ID == data.Id.ValueString() {
			found = att
			break
		}
	}
	if found == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(mapResponseToModel(ctx, found, &data, PolicyAttachmentResourceSchema(ctx).Attributes)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *policyAttachmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan PolicyAttachmentModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *policyAttachmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data PolicyAttachmentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.AccessPolicies.DetachPolicy(ctx, data.OrgId.ValueString(), data.Id.ValueString(), nil)
	if err != nil && !isNotFound(err) {
		resp.Diagnostics.AddError("Failed to detach policy", err.Error())
	}
}

func (r *policyAttachmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "/")
	if len(parts) == 3 {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("org_id"), parts[0])...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("policy_id"), parts[1])...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[2])...)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}
