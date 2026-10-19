package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func DomainAutoJoinResourceSchema() schema.Schema {
	return schema.Schema{
		Description:         "Configures whether users with matching corporate email domains can automatically join a Rixl organization.",
		MarkdownDescription: "Configures whether users with matching corporate email domains can automatically join a Rixl organization.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"org_id": schema.StringAttribute{
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"enabled": schema.BoolAttribute{
				Required: true,
			},
			"present": schema.BoolAttribute{
				Computed: true,
			},
		},
	}
}

type DomainAutoJoinResourceModel struct {
	Id      types.String `tfsdk:"id"`
	OrgId   types.String `tfsdk:"org_id"`
	Enabled types.Bool   `tfsdk:"enabled"`
	Present types.Bool   `tfsdk:"present"`
}
