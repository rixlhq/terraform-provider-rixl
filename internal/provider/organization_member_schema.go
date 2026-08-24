package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func OrganizationMemberResourceSchema() schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"org_id": schema.StringAttribute{
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"user_id": schema.StringAttribute{
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"role": schema.StringAttribute{
				Required: true,
			},
			"state": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"username": schema.StringAttribute{
				Computed: true,
			},
			"first_name": schema.StringAttribute{
				Computed: true,
			},
			"last_name": schema.StringAttribute{
				Computed: true,
			},
			"joined_at": schema.StringAttribute{
				Computed: true,
			},
			"invitation_expires_at": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

type OrganizationMemberResourceModel struct {
	Id                  types.String `tfsdk:"id"`
	OrgId               types.String `tfsdk:"org_id"`
	UserId              types.String `tfsdk:"user_id"`
	Role                types.String `tfsdk:"role"`
	State               types.String `tfsdk:"state"`
	Username            types.String `tfsdk:"username"`
	FirstName           types.String `tfsdk:"first_name"`
	LastName            types.String `tfsdk:"last_name"`
	JoinedAt            types.String `tfsdk:"joined_at"`
	InvitationExpiresAt types.String `tfsdk:"invitation_expires_at"`
}
