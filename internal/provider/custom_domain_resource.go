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

func NewCustomDomainResource() resource.Resource {
	return &customDomainResource{}
}

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

	if data.OrgId.IsNull() || data.OrgId.IsUnknown() {
		data.OrgId = types.StringValue(orgID)
	}
	if data.Domain.IsNull() || data.Domain.IsUnknown() {
		data.Domain = types.StringValue(domainName)
	}
	if data.Id.IsNull() || data.Id.IsUnknown() {
		data.Id = data.Domain
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
		var httpErr *customdomains.ClientHttpError[struct{}]
		if errors.As(err, &httpErr) && httpErr.StatusCode == http.StatusNotFound {
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
		var httpErr *customdomains.ClientHttpError[struct{}]
		if errors.As(err, &httpErr) && httpErr.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to remove custom domain", err.Error())
	}
}

func CustomDomainResourceSchema(_ context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"org_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"domain": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"present": schema.BoolAttribute{
				Computed: true,
			},
			"id": schema.StringAttribute{
				Computed: true,
			},
			"status": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"auto_join": schema.BoolAttribute{
						Computed: true,
					},
					"pending": schema.SingleNestedAttribute{
						Computed: true,
						Attributes: map[string]schema.Attribute{
							"verification_token": schema.StringAttribute{
								Computed:  true,
								Sensitive: true,
							},
							"expires_at": schema.StringAttribute{
								Computed: true,
							},
						},
					},
					"verified": schema.SingleNestedAttribute{
						Computed: true,
						Attributes: map[string]schema.Attribute{
							"verified_at": schema.StringAttribute{
								Computed: true,
							},
						},
					},
				},
			},
		},
	}
}

type CustomDomainResourceModel struct {
	OrgId   types.String `tfsdk:"org_id"`
	Domain  types.String `tfsdk:"domain"`
	Present types.Bool   `tfsdk:"present"`
	Id      types.String `tfsdk:"id"`
	Status  types.Object `tfsdk:"status"`
}
