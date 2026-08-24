package provider

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rixlhq/rixl-go/sdk"
	"github.com/rixlhq/rixl-go/sdk/dashboards"
	"github.com/rixlhq/rixl-go/sdk/models"
)

var _ resource.Resource = (*dashboardResource)(nil)

func NewDashboardResource() resource.Resource {
	return &dashboardResource{}
}

type dashboardResource struct {
	client *sdk.Client
}

func (r *dashboardResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dashboard"
}

func (r *dashboardResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = DashboardResourceSchema(ctx)
}

func (r *dashboardResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *dashboardResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DashboardModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createBody := models.AnalyticsV1CreateDashboardRequest{
		Name:      data.Name.ValueString(),
		IsDefault: data.IsDefault.ValueBoolPointer(),
	}

	dashboard, err := r.client.Dashboards.CreateDashboard(ctx, createBody)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create dashboard", err.Error())
		return
	}

	resp.Diagnostics.Append(mapResponseToModel(ctx, dashboard, &data, DashboardResourceSchema(ctx).Attributes)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *dashboardResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DashboardModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := data.Id.ValueString()
	dashboard, err := r.client.Dashboards.GetDashboard(ctx, id)
	if err != nil {
		var httpErr *dashboards.ClientHttpError[struct{}]
		if errors.As(err, &httpErr) && httpErr.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read dashboard", err.Error())
		return
	}

	resp.Diagnostics.Append(mapResponseToModel(ctx, dashboard, &data, DashboardResourceSchema(ctx).Attributes)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.Id.IsNull() || data.Id.IsUnknown() {
		data.Id = types.StringValue(id)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *dashboardResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state, plan DashboardModel
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
	data := dataAny.(*DashboardModel)

	var visibility *string
	if !data.Visibility.IsNull() && !data.Visibility.IsUnknown() {
		v := data.Visibility.ValueString()
		if v != "" {
			visibility = &v
		}
	}

	rev, err := r.resolveRevision(ctx, data)
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve dashboard revision", err.Error())
		return
	}

	updateBody := models.AnalyticsV1UpdateDashboardRequest{
		ID:               data.Id.ValueString(),
		Name:             data.Name.ValueString(),
		Visibility:       visibility,
		ExpectedRevision: rev,
	}

	dashboard, err := r.client.Dashboards.UpdateDashboard(ctx, data.Id.ValueString(), updateBody)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update dashboard", err.Error())
		return
	}

	resp.Diagnostics.Append(mapResponseToModel(ctx, dashboard, data, DashboardResourceSchema(ctx).Attributes)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// The update response may omit the revision. Keep the value we used so
	// the next operation does not have to re-fetch.
	if data.Revision.IsNull() || data.Revision.IsUnknown() {
		data.Revision = types.Int64Value(int64(rev))
	}
	if data.Id.IsNull() || data.Id.IsUnknown() {
		data.Id = state.Id
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *dashboardResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DashboardModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	rev, err := r.resolveRevision(ctx, &data)
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve dashboard revision", err.Error())
		return
	}

	params := &dashboards.DeleteDashboardParams{
		ExpectedRevision: rev,
	}

	if _, err := r.client.Dashboards.DeleteDashboard(ctx, data.Id.ValueString(), params); err != nil {
		var httpErr *dashboards.ClientHttpError[struct{}]
		if errors.As(err, &httpErr) && httpErr.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to delete dashboard", err.Error())
	}
}

// resolveRevision returns the dashboard revision from state when it is known.
// It only falls back to a GET when the revision is absent (e.g. after import
// with a read that did not return it), avoiding an extra API call on every
// update/delete in the normal flow.
func (r *dashboardResource) resolveRevision(ctx context.Context, data *DashboardModel) (int32, error) {
	if !data.Revision.IsNull() && !data.Revision.IsUnknown() {
		return int32(data.Revision.ValueInt64()), nil
	}

	dashboard, err := r.client.Dashboards.GetDashboard(ctx, data.Id.ValueString())
	if err != nil {
		return 0, fmt.Errorf("failed to read dashboard for revision: %w", err)
	}
	if dashboard.Revision == nil {
		return 0, fmt.Errorf("dashboard %q has no revision", data.Id.ValueString())
	}

	data.Revision = types.Int64Value(int64(*dashboard.Revision))
	return *dashboard.Revision, nil
}

func DashboardResourceSchema(_ context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"org_id": schema.StringAttribute{
				Computed: true,
			},
			"owner_user_id": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"visibility": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.OneOf("private", "org"),
				},
			},
			"is_default": schema.BoolAttribute{
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.Bool{boolplanmodifier.RequiresReplace()},
			},
			"revision": schema.Int64Attribute{
				Computed: true,
			},
			"updated_by": schema.StringAttribute{
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

type DashboardModel struct {
	Id          types.String `tfsdk:"id"`
	OrgId       types.String `tfsdk:"org_id"`
	OwnerUserId types.String `tfsdk:"owner_user_id"`
	Name        types.String `tfsdk:"name"`
	Visibility  types.String `tfsdk:"visibility"`
	IsDefault   types.Bool   `tfsdk:"is_default"`
	Revision    types.Int64  `tfsdk:"revision"`
	UpdatedBy   types.String `tfsdk:"updated_by"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}
