package provider

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rixlhq/rixl-go/sdk"
	"github.com/rixlhq/rixl-go/sdk/dashboards"
	"github.com/rixlhq/rixl-go/sdk/models"
)

var _ resource.Resource = (*dashboardWidgetResource)(nil)

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
		var httpErr *dashboards.ClientHttpError[struct{}]
		if errors.As(err, &httpErr) && httpErr.StatusCode == http.StatusNotFound {
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
		var httpErr *dashboards.ClientHttpError[struct{}]
		if errors.As(err, &httpErr) && httpErr.StatusCode == http.StatusNotFound {
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
		var httpErr *dashboards.ClientHttpError[struct{}]
		if errors.As(err, &httpErr) && httpErr.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to fetch dashboard revision for delete", err.Error())
		return
	}

	params := &dashboards.DeleteWidgetParams{ExpectedRevision: revision}
	_, err = r.client.Dashboards.DeleteWidget(ctx, widgetID, params)
	if err != nil {
		var httpErr *dashboards.ClientHttpError[struct{}]
		if errors.As(err, &httpErr) && httpErr.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to delete dashboard widget", err.Error())
		return
	}
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

func buildWidgetInput(data DashboardWidgetResourceModel) models.AnalyticsV1WidgetInput {
	input := models.AnalyticsV1WidgetInput{
		Title:     data.Title.ValueString(),
		ChartType: data.ChartType.ValueString(),
		Dataset:   data.Dataset.ValueString(),
		Metric:    data.Metric.ValueString(),
		Interval:  tfStringPtr(data.Interval),
		Limit:     tfInt32Ptr(data.Limit),
		PosX:      tfInt32Ptr(data.PosX),
		PosY:      tfInt32Ptr(data.PosY),
		Width:     tfInt32Ptr(data.Width),
		Height:    tfInt32Ptr(data.Height),
		SortOrder: tfInt32Ptr(data.SortOrder),
		GroupBy:   stringListValue(data.GroupBy),
		Filters:   buildChartFilters(data.Filters),
	}
	return input
}

func buildChartFilters(list types.List) []models.AnalyticsV1ChartFilter {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}
	filters := make([]models.AnalyticsV1ChartFilter, 0, len(list.Elements()))
	for _, e := range list.Elements() {
		obj, ok := e.(types.Object)
		if !ok {
			continue
		}
		attrs := obj.Attributes()
		f := models.AnalyticsV1ChartFilter{}
		if v, ok := attrs["field"].(types.String); ok && !v.IsNull() {
			f.Field = v.ValueString()
		}
		if v, ok := attrs["operator"].(types.String); ok && !v.IsNull() {
			op := v.ValueString()
			f.Operator = &op
		}
		if v, ok := attrs["values"].(types.List); ok && !v.IsNull() {
			f.Values = stringListValue(v)
		}
		filters = append(filters, f)
	}
	return filters
}

func stringListValue(list types.List) []string {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}
	elements := make([]string, 0, len(list.Elements()))
	for _, e := range list.Elements() {
		if s, ok := e.(types.String); ok {
			elements = append(elements, s.ValueString())
		}
	}
	return elements
}
