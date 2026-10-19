package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rixlhq/rixl-go/sdk"
	"github.com/rixlhq/rixl-go/sdk/dashboards"
)

var (
	_ resource.Resource                = (*dashboardWidgetResource)(nil)
	_ resource.ResourceWithImportState = (*dashboardWidgetResource)(nil)
)

func NewDashboardWidgetResource() resource.Resource {
	return &dashboardWidgetResource{}
}

type dashboardWidgetResource struct {
	client *sdk.Client
}

func (r *dashboardWidgetResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dashboard_widget"
}

func (r *dashboardWidgetResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = DashboardWidgetResourceSchema()
}

func (r *dashboardWidgetResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *dashboardWidgetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DashboardWidgetResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dashboardID := data.DashboardId.ValueString()
	revision, err := r.fetchRevision(ctx, dashboardID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to fetch dashboard revision", err.Error())
		return
	}

	params := &dashboards.CreateWidgetParams{ExpectedRevision: revision}
	widget, err := r.client.Dashboards.CreateWidget(ctx, dashboardID, params, buildWidgetInput(data))
	if err != nil {
		resp.Diagnostics.AddError("Failed to create dashboard widget", err.Error())
		return
	}

	resp.Diagnostics.Append(mapResponseToModel(ctx, widget, &data, DashboardWidgetResourceSchema().Attributes)...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.DashboardId = types.StringValue(dashboardID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *dashboardWidgetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DashboardWidgetResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dashboardID := data.DashboardId.ValueString()
	widgetID := data.Id.ValueString()

	dashboard, err := r.client.Dashboards.GetDashboard(ctx, dashboardID)
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read dashboard", err.Error())
		return
	}

	for i := range dashboard.Widgets {
		widget := dashboard.Widgets[i]
		widgetMap, err := responseToMap(widget)
		if err != nil {
			continue
		}
		if id, ok := widgetMap["id"].(string); ok && id == widgetID {
			resp.Diagnostics.Append(mapResponseToModel(ctx, widget, &data, DashboardWidgetResourceSchema().Attributes)...)
			if resp.Diagnostics.HasError() {
				return
			}
			data.DashboardId = types.StringValue(dashboardID)
			resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
			return
		}
	}

	resp.State.RemoveResource(ctx)
}

func (r *dashboardWidgetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan DashboardWidgetResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	widgetID := plan.Id.ValueString()
	dashboardID := plan.DashboardId.ValueString()
	revision, err := r.fetchRevision(ctx, dashboardID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to fetch dashboard revision", err.Error())
		return
	}

	patchMap, err := responseToMap(buildWidgetInput(plan))
	if err != nil {
		resp.Diagnostics.AddError("Failed to build widget patch", err.Error())
		return
	}
	body := map[string]any{
		"expected_revision": revision,
		"patch":             patchMap,
	}

	widget, err := r.client.Dashboards.UpdateWidget(ctx, widgetID, body)
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to update dashboard widget", err.Error())
		return
	}

	resp.Diagnostics.Append(mapResponseToModel(ctx, widget, &plan, DashboardWidgetResourceSchema().Attributes)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.DashboardId = types.StringValue(dashboardID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *dashboardWidgetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DashboardWidgetResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	widgetID := data.Id.ValueString()
	dashboardID := data.DashboardId.ValueString()
	revision, err := r.fetchRevision(ctx, dashboardID)
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to fetch dashboard revision for delete", err.Error())
		return
	}

	params := &dashboards.DeleteWidgetParams{ExpectedRevision: revision}
	_, err = r.client.Dashboards.DeleteWidget(ctx, widgetID, params)
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to delete dashboard widget", err.Error())
		return
	}
}

func (r *dashboardWidgetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "/")
	if len(parts) == 2 {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("dashboard_id"), parts[0])...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
		return
	}
	resp.Diagnostics.AddError(
		"Unexpected Import Identifier",
		fmt.Sprintf("Expected import identifier with format: dashboard_id/widget_id. Got: %q", req.ID),
	)
}

func (r *dashboardWidgetResource) fetchRevision(ctx context.Context, dashboardID string) (int32, error) {
	dashboard, err := r.client.Dashboards.GetDashboard(ctx, dashboardID)
	if err != nil {
		return 0, err
	}
	if dashboard.Revision == nil {
		return 0, fmt.Errorf("dashboard %s has no revision", dashboardID)
	}
	return *dashboard.Revision, nil
}
