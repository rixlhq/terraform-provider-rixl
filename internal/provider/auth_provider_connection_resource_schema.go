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

func AuthProviderConnectionResourceSchema(_ context.Context) rschema.Schema {
	return rschema.Schema{
		Attributes: map[string]rschema.Attribute{
			"id": rschema.StringAttribute{
				Required:    true,
				Description: "Connected OAuth provider. This is the resource identity and maps to the API provider field.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(
						"EXTERNAL_ACCOUNT_PROVIDER_UNSPECIFIED",
						"EXTERNAL_ACCOUNT_PROVIDER_GOOGLE",
						"EXTERNAL_ACCOUNT_PROVIDER_APPLE",
						"EXTERNAL_ACCOUNT_PROVIDER_MICROSOFT",
						"EXTERNAL_ACCOUNT_PROVIDER_FACEBOOK",
						"EXTERNAL_ACCOUNT_PROVIDER_TELEGRAM",
					),
				},
			},
			"token": rschema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "OAuth token used to connect the provider. Write-only: it is never returned by the API.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"country_code": rschema.StringAttribute{
				Optional:    true,
				Description: "Country code sent when connecting the provider.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"origin": rschema.StringAttribute{
				Optional:    true,
				Description: "Origin sent when connecting the provider.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"username": rschema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"first_name": rschema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"last_name": rschema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"email_address": rschema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"image_url": rschema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

type AuthProviderConnectionModel struct {
	Id           types.String `tfsdk:"id"`
	Token        types.String `tfsdk:"token"`
	CountryCode  types.String `tfsdk:"country_code"`
	Origin       types.String `tfsdk:"origin"`
	Username     types.String `tfsdk:"username"`
	FirstName    types.String `tfsdk:"first_name"`
	LastName     types.String `tfsdk:"last_name"`
	EmailAddress types.String `tfsdk:"email_address"`
	ImageUrl     types.String `tfsdk:"image_url"`
}
