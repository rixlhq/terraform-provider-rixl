package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func OrganizationInvitationResourceSchema() schema.Schema {
	return schema.Schema{
		Description:         "Manages member invitations to a Rixl organization.",
		MarkdownDescription: "Manages member invitations to a Rixl organization.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"org_id": schema.StringAttribute{
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"user_id": schema.StringAttribute{
				Computed: true,
			},
			"username": schema.StringAttribute{
				Optional:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"role": schema.StringAttribute{
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"state": schema.StringAttribute{
				Computed: true,
			},
			"organization_username": schema.StringAttribute{
				Computed: true,
			},
			"organization_first_name": schema.StringAttribute{
				Computed: true,
			},
			"organization_last_name": schema.StringAttribute{
				Computed: true,
			},
			"created_at": schema.StringAttribute{
				Computed: true,
			},
			"decided_at": schema.StringAttribute{
				Computed: true,
			},
			"invitation_expires_at": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

type OrganizationInvitationResourceModel struct {
	Id                    types.String `tfsdk:"id"`
	OrgId                 types.String `tfsdk:"org_id"`
	UserId                types.String `tfsdk:"user_id"`
	Username              types.String `tfsdk:"username"`
	Role                  types.String `tfsdk:"role"`
	State                 types.String `tfsdk:"state"`
	OrganizationUsername  types.String `tfsdk:"organization_username"`
	OrganizationFirstName types.String `tfsdk:"organization_first_name"`
	OrganizationLastName  types.String `tfsdk:"organization_last_name"`
	CreatedAt             types.String `tfsdk:"created_at"`
	DecidedAt             types.String `tfsdk:"decided_at"`
	InvitationExpiresAt   types.String `tfsdk:"invitation_expires_at"`
}
