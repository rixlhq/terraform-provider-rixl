package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func BlogSubscriptionResourceSchema() schema.Schema {
	return schema.Schema{
		Description:         "Manages a user's subscription preference to Rixl blog and product updates.",
		MarkdownDescription: "Manages a user's subscription preference to Rixl blog and product updates.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"user_id": schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"subscribed": schema.BoolAttribute{
				Required: true,
			},
			"subscribed_at": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

type BlogSubscriptionResourceModel struct {
	Id           types.String `tfsdk:"id"`
	UserId       types.String `tfsdk:"user_id"`
	Subscribed   types.Bool   `tfsdk:"subscribed"`
	SubscribedAt types.String `tfsdk:"subscribed_at"`
}
