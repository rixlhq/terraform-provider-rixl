package provider_test

import (
	"testing"

	fwprovider "github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	rixlprovider "github.com/rixlhq/terraform-provider-rixl/internal/provider"
)

func TestProviderMetadata(t *testing.T) {
	t.Parallel()

	p := rixlprovider.New("dev")()
	var resp fwprovider.MetadataResponse
	p.Metadata(t.Context(), fwprovider.MetadataRequest{}, &resp)

	if resp.TypeName != "rixl" {
		t.Fatalf("expected type name rixl, got %q", resp.TypeName)
	}
}

func TestProviderSchema(t *testing.T) {
	t.Parallel()

	p := rixlprovider.New("dev")()
	var resp fwprovider.SchemaResponse
	p.Schema(t.Context(), fwprovider.SchemaRequest{}, &resp)

	attrs := resp.Schema.Attributes
	required := []string{"api_key", "bearer_token", "base_url"}
	for _, attr := range required {
		if _, ok := attrs[attr]; !ok {
			t.Fatalf("expected schema to contain %q", attr)
		}
	}
}

func TestProviderConfigureMissingCredentials(t *testing.T) {
	t.Parallel()

	p := rixlprovider.New("dev")()

	var schemaResp fwprovider.SchemaResponse
	p.Schema(t.Context(), fwprovider.SchemaRequest{}, &schemaResp)

	typ := schemaResp.Schema.Type().TerraformType(t.Context())
	objType := typ.(tftypes.Object)
	vals := make(map[string]tftypes.Value, len(objType.AttributeTypes))
	for attr, attrType := range objType.AttributeTypes {
		vals[attr] = tftypes.NewValue(attrType, nil)
	}

	var configureResp fwprovider.ConfigureResponse
	p.Configure(t.Context(), fwprovider.ConfigureRequest{
		Config: tfsdk.Config{
			Raw:    tftypes.NewValue(typ, vals),
			Schema: schemaResp.Schema,
		},
	}, &configureResp)

	if !configureResp.Diagnostics.HasError() {
		t.Fatal("expected error for missing credentials")
	}
}
