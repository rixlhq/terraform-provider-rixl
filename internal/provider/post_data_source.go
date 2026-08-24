package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewPostDataSource() datasource.DataSource {
	return newManagedDataSource(DataSourceDescriptor{
		TypeName:          "post",
		SchemaFn:          PostDataSourceSchema,
		Model:             &PostDataSourceModel{},
		ClientField:       "Posts",
		ReadMethod:        "GetPost",
		PathParams:        []string{"project_id", "id"},
		ReadResponseField: "post",
	})
}

func PostDataSourceSchema(_ context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				Required: true,
			},
			"id": schema.StringAttribute{
				Required: true,
			},
			"feed_id": schema.StringAttribute{
				Computed: true,
			},
			"org_id": schema.StringAttribute{
				Computed: true,
			},
			"creator_id": schema.StringAttribute{
				Computed: true,
			},
			"type": schema.StringAttribute{
				Computed: true,
			},
			"description": schema.StringAttribute{
				Computed: true,
			},
			"created_at": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

type PostDataSourceModel struct {
	ProjectId   types.String `tfsdk:"project_id"`
	Id          types.String `tfsdk:"id"`
	FeedId      types.String `tfsdk:"feed_id"`
	OrgId       types.String `tfsdk:"org_id"`
	CreatorId   types.String `tfsdk:"creator_id"`
	Type        types.String `tfsdk:"type"`
	Description types.String `tfsdk:"description"`
	CreatedAt   types.String `tfsdk:"created_at"`
}
