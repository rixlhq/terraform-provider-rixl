package provider_test

import (
	"testing"

	fwprovider "github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	rixlprovider "github.com/rixlhq/terraform-provider-rixl/internal/provider"
)

func newEmptyConfig(t *testing.T, schemaResp *fwprovider.SchemaResponse) tfsdk.Config {
	t.Helper()

	typ := schemaResp.Schema.Type().TerraformType(t.Context())
	objType := typ.(tftypes.Object)
	vals := make(map[string]tftypes.Value, len(objType.AttributeTypes))
	for attr, attrType := range objType.AttributeTypes {
		vals[attr] = tftypes.NewValue(attrType, nil)
	}

	return tfsdk.Config{
		Raw:    tftypes.NewValue(typ, vals),
		Schema: schemaResp.Schema,
	}
}

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

	var configureResp fwprovider.ConfigureResponse
	p.Configure(t.Context(), fwprovider.ConfigureRequest{
		Config: newEmptyConfig(t, &schemaResp),
	}, &configureResp)

	if !configureResp.Diagnostics.HasError() {
		t.Fatal("expected error for missing credentials")
	}
}

func TestProviderConfigureReadsEnv(t *testing.T) {
	t.Setenv("RIXL_API_KEY", "api-key")
	t.Setenv("RIXL_BASE_URL", "https://api.example.com")

	p := rixlprovider.New("dev")()

	var schemaResp fwprovider.SchemaResponse
	p.Schema(t.Context(), fwprovider.SchemaRequest{}, &schemaResp)

	var configureResp fwprovider.ConfigureResponse
	p.Configure(t.Context(), fwprovider.ConfigureRequest{
		Config: newEmptyConfig(t, &schemaResp),
	}, &configureResp)

	if configureResp.Diagnostics.HasError() {
		t.Fatalf("unexpected error: %s", configureResp.Diagnostics.Errors())
	}
}
