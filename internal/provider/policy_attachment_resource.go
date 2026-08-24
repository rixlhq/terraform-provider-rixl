package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func PolicyAttachmentResourceSchema(_ context.Context) rschema.Schema {
	return rschema.Schema{
		Attributes: map[string]rschema.Attribute{
			"id": rschema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"org_id": rschema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"policy_id": rschema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"identity_type": rschema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(
						"POLICY_IDENTITY_TYPE_UNSPECIFIED",
						"POLICY_IDENTITY_TYPE_USER",
						"POLICY_IDENTITY_TYPE_API_KEY",
						"POLICY_IDENTITY_TYPE_CLIENTAUTH_CREDENTIAL",
					),
				},
			},
			"identity_id": rschema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"created_at": rschema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

type PolicyAttachmentModel struct {
	Id           types.String `tfsdk:"id"`
	OrgId        types.String `tfsdk:"org_id"`
	PolicyId     types.String `tfsdk:"policy_id"`
	IdentityType types.String `tfsdk:"identity_type"`
	IdentityId   types.String `tfsdk:"identity_id"`
	CreatedAt    types.String `tfsdk:"created_at"`
}
