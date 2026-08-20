package provider

import (
	"context"
	"fmt"
	"net/url"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/rixlhq/terraform-provider-rixl/internal/providerdata"
	"github.com/rixlhq/terraform-provider-rixl/internal/rixlclient"
	"github.com/rixlhq/terraform-provider-rixl/internal/rixlcommon"
)

type rixlDataSourceSpec struct {
	typeName     string
	readPath     string
	paramAliases map[string]string
	responseRoot string
	queryParams  map[string]string
	schemaFn     func(context.Context) schema.Schema
}

var _ datasource.DataSource = &rixlDataSource{}

type rixlDataSource struct {
	client *rixlclient.Client
	schema schema.Schema
	spec   rixlDataSourceSpec
}

// Metadata returns the full data source type name.
func (d *rixlDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.spec.typeName
}

// Schema returns the data source schema.
func (d *rixlDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = d.schema
}

// Configure prepares the API client for the data source.
func (d *rixlDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	pd, ok := req.ProviderData.(*providerdata.Data)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Configure Type", fmt.Sprintf("Expected *providerdata.Data, got %T", req.ProviderData))
		return
	}
	d.client = pd.Client
}

// Read queries the Rixl API and sets the data source state.
func (d *rixlDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Missing Client", "Configure the provider to use this data source.")
		return
	}

	path, err := buildPath(req.Config.Raw, d.spec.readPath, d.spec.paramAliases)
	if err != nil {
		resp.Diagnostics.AddError("Path Error", err.Error())
		return
	}

	query, err := d.queryValues(ctx, req.Config.Raw)
	if err != nil {
		resp.Diagnostics.AddError("Query Parameter Error", err.Error())
		return
	}

	respBody, err := d.client.Get(ctx, path, query)
	if err != nil {
		resp.Diagnostics.AddError("Rixl API Error", err.Error())
		return
	}

	tfType := d.schema.Type().TerraformType(ctx)
	tfVal, err := responseToState(ctx, tfType, respBody, d.spec.responseRoot)
	if err != nil {
		resp.Diagnostics.AddError("Response Conversion Error", err.Error())
		return
	}

	resp.State.Raw = rixlcommon.OverlayKnown(req.Config.Raw, tfVal)
}

func newRixlDataSource(spec rixlDataSourceSpec) datasource.DataSource {
	return &rixlDataSource{
		schema: spec.schemaFn(context.Background()),
		spec:   spec,
	}
}

func (d *rixlDataSource) queryValues(_ context.Context, config tftypes.Value) (url.Values, error) {
	if len(d.spec.queryParams) == 0 {
		return nil, nil
	}

	obj, err := asObject(config)
	if err != nil {
		return nil, err
	}

	q := make(url.Values)
	for attr, paramName := range d.spec.queryParams {
		v, ok := obj[attr]
		if !ok {
			continue
		}
		s, err := rixlcommon.ValueAsString(v)
		if err != nil || s == "" {
			continue
		}
		q.Set(paramName, s)
	}
	return q, nil
}
