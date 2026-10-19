package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rixlhq/rixl-go/sdk/models"
)

func buildWidgetInput(data DashboardWidgetResourceModel) models.AnalyticsV1WidgetInput {
	return models.AnalyticsV1WidgetInput{
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
