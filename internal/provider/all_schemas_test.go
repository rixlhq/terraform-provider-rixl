package provider_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/rixlhq/terraform-provider-rixl/internal/provider"
)

func TestAllResourcesHaveValidSchemas(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	p := provider.New("1.7.0")()

	resources := p.Resources(ctx)
	if len(resources) == 0 {
		t.Fatalf("expected non-empty resources list")
	}

	for _, factory := range resources {
		res := factory()
		var metaResp resource.MetadataResponse
		res.Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "rixl"}, &metaResp)
		if metaResp.TypeName == "" {
			t.Fatalf("resource %T returned empty TypeName", res)
		}

		t.Run(metaResp.TypeName, func(t *testing.T) {
			t.Parallel()
			var schemaResp resource.SchemaResponse
			res.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
			if schemaResp.Diagnostics.HasError() {
				t.Fatalf("resource %s schema error: %v", metaResp.TypeName, schemaResp.Diagnostics)
			}
			if len(schemaResp.Schema.Attributes) == 0 {
				t.Fatalf("resource %s declared no attributes", metaResp.TypeName)
			}
			if schemaResp.Schema.MarkdownDescription == "" && schemaResp.Schema.Description == "" {
				t.Errorf("resource %s has no description or markdown description", metaResp.TypeName)
			}
		})
	}
}

func TestAllDataSourcesHaveValidSchemas(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	p := provider.New("1.7.0")()

	dataSources := p.DataSources(ctx)
	if len(dataSources) == 0 {
		t.Fatalf("expected non-empty data sources list")
	}

	for _, factory := range dataSources {
		ds := factory()
		var metaResp datasource.MetadataResponse
		ds.Metadata(ctx, datasource.MetadataRequest{ProviderTypeName: "rixl"}, &metaResp)
		if metaResp.TypeName == "" {
			t.Fatalf("datasource %T returned empty TypeName", ds)
		}

		t.Run(metaResp.TypeName, func(t *testing.T) {
			t.Parallel()
			var schemaResp datasource.SchemaResponse
			ds.Schema(ctx, datasource.SchemaRequest{}, &schemaResp)
			if schemaResp.Diagnostics.HasError() {
				t.Fatalf("datasource %s schema error: %v", metaResp.TypeName, schemaResp.Diagnostics)
			}
			if len(schemaResp.Schema.Attributes) == 0 {
				t.Fatalf("datasource %s declared no attributes", metaResp.TypeName)
			}
			if schemaResp.Schema.MarkdownDescription == "" && schemaResp.Schema.Description == "" {
				t.Errorf("datasource %s has no description or markdown description", metaResp.TypeName)
			}
		})
	}
}
