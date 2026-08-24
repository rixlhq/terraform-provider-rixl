//nolint:goconst // tfsdk attribute names are repeated throughout the provider
package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewDomainDataSource() datasource.DataSource {
	return newManagedDataSource(DataSourceDescriptor{
		TypeName:    "domain",
		SchemaFn:    DomainDataSourceSchema,
		Model:       &DomainDataSourceModel{},
		ClientField: "CustomDomains",
		ReadMethod:  "GetDomainStatus",
		PathParams:  []string{"org_id"},
	})
}

func DomainDataSourceSchema(_ context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"org_id": schema.StringAttribute{
				Required: true,
			},
			"present": schema.BoolAttribute{
				Computed: true,
			},
			"id": schema.StringAttribute{
				Computed: true,
			},
			"domain": schema.StringAttribute{
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
								Computed: true,
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

type DomainDataSourceModel struct {
	OrgId   types.String `tfsdk:"org_id"`
	Present types.Bool   `tfsdk:"present"`
	Id      types.String `tfsdk:"id"`
	Domain  types.String `tfsdk:"domain"`
	Status  types.Object `tfsdk:"status"`
}
