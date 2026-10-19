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

func NewCustomDomainResource() resource.Resource {
	return &customDomainResource{}
}

var (
	_ resource.Resource                = (*customDomainResource)(nil)
	_ resource.ResourceWithImportState = (*customDomainResource)(nil)
)

type customDomainResource struct {
	client *sdk.Client
}

func (r *customDomainResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_custom_domain"
}

func (r *customDomainResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = CustomDomainResourceSchema(ctx)
}

func (r *customDomainResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *customDomainResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data CustomDomainResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgID := data.OrgId.ValueString()
	domainName := data.Domain.ValueString()

	body := map[string]any{
		"domain": domainName,
	}

	domain, err := r.client.CustomDomains.CreateDomainVerification(ctx, orgID, body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create custom domain", err.Error())
		return
	}

	resp.Diagnostics.Append(mapResponseToModel(ctx, domain, &data, CustomDomainResourceSchema(ctx).Attributes)...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.OrgId = types.StringValue(orgID)
	data.Domain = types.StringValue(domainName)
	if data.Id.IsNull() || data.Id.IsUnknown() {
		data.Id = types.StringValue(domainName)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *customDomainResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data CustomDomainResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain, err := r.client.CustomDomains.GetDomainStatus(ctx, data.OrgId.ValueString(), nil)
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read custom domain", err.Error())
		return
	}

	resp.Diagnostics.Append(mapResponseToModel(ctx, domain, &data, CustomDomainResourceSchema(ctx).Attributes)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.Id.IsNull() || data.Id.IsUnknown() {
		data.Id = data.Domain
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *customDomainResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state, plan CustomDomainResourceModel
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
	data := dataAny.(*CustomDomainResourceModel)

	// The API does not support an in-place domain update. Recreate to change the domain.
	resp.Diagnostics.AddWarning(
		"Custom domain update not supported",
		"The Rixl API does not support updating an existing custom domain in place. Changing 'domain' or 'org_id' forces a new resource.",
	)

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *customDomainResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data CustomDomainResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := r.client.CustomDomains.RemoveDomain(ctx, data.OrgId.ValueString(), nil); err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to remove custom domain", err.Error())
	}
}

func (r *customDomainResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "/")
	if len(parts) == 2 {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("org_id"), parts[0])...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("domain"), parts[1])...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("org_id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}
