package provider

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rixlhq/rixl-go/sdk"
	"github.com/rixlhq/rixl-go/sdk/blog"
)

var _ resource.Resource = (*blogSubscriptionResource)(nil)

func NewBlogSubscriptionResource() resource.Resource {
	return &blogSubscriptionResource{}
}

type blogSubscriptionResource struct {
	client *sdk.Client
}

func (r *blogSubscriptionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_blog_subscription"
}

func (r *blogSubscriptionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = BlogSubscriptionResourceSchema()
}

func (r *blogSubscriptionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*sdk.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data type", fmt.Sprintf("expected *sdk.Client, got %T", req.ProviderData))
		return
	}
	r.client = client
}

func (r *blogSubscriptionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data BlogSubscriptionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.Subscribed.ValueBool() {
		params := &blog.SubscribeBlogParams{}
		if !data.UserId.IsNull() && !data.UserId.IsUnknown() {
			uid := data.UserId.ValueString()
			params.UserId = &uid
		}
		if _, err := r.client.Blog.SubscribeBlog(ctx, params, nil); err != nil {
			resp.Diagnostics.AddError("Failed to subscribe to blog", err.Error())
			return
		}
	}

	r.readBack(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *blogSubscriptionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data BlogSubscriptionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.readBack(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *blogSubscriptionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state, plan BlogSubscriptionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.Subscribed.ValueBool() != plan.Subscribed.ValueBool() {
		if plan.Subscribed.ValueBool() {
			params := &blog.SubscribeBlogParams{}
			if !plan.UserId.IsNull() && !plan.UserId.IsUnknown() {
				uid := plan.UserId.ValueString()
				params.UserId = &uid
			}
			if _, err := r.client.Blog.SubscribeBlog(ctx, params, nil); err != nil {
				resp.Diagnostics.AddError("Failed to subscribe to blog", err.Error())
				return
			}
		} else {
			params := &blog.UnsubscribeBlogParams{}
			if !plan.UserId.IsNull() && !plan.UserId.IsUnknown() {
				uid := plan.UserId.ValueString()
				params.UserId = &uid
			}
			if _, err := r.client.Blog.UnsubscribeBlog(ctx, params, nil); err != nil {
				resp.Diagnostics.AddError("Failed to unsubscribe from blog", err.Error())
				return
			}
		}
	}

	r.readBack(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *blogSubscriptionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data BlogSubscriptionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.Subscribed.ValueBool() {
		params := &blog.UnsubscribeBlogParams{}
		if !data.UserId.IsNull() && !data.UserId.IsUnknown() {
			uid := data.UserId.ValueString()
			params.UserId = &uid
		}
		if _, err := r.client.Blog.UnsubscribeBlog(ctx, params, nil); err != nil {
			var httpErr *blog.ClientHttpError[struct{}]
			if errors.As(err, &httpErr) && httpErr.StatusCode == http.StatusNotFound {
				resp.State.RemoveResource(ctx)
				return
			}
			resp.Diagnostics.AddError("Failed to unsubscribe from blog", err.Error())
			return
		}
	}
}

func (r *blogSubscriptionResource) readBack(ctx context.Context, data *BlogSubscriptionResourceModel, diags *diag.Diagnostics) {
	params := &blog.GetBlogSubscriptionParams{}
	if !data.UserId.IsNull() && !data.UserId.IsUnknown() {
		uid := data.UserId.ValueString()
		params.UserId = &uid
	}

	sub, err := r.client.Blog.GetBlogSubscription(ctx, params, nil)
	if err != nil {
		var httpErr *blog.ClientHttpError[struct{}]
		if errors.As(err, &httpErr) && httpErr.StatusCode == http.StatusNotFound {
			diags.AddWarning("Blog subscription not found", "The blog subscription was not found.")
			return
		}
		diags.AddError("Failed to read blog subscription", err.Error())
		return
	}

	if sub.Subscribed != nil {
		data.Subscribed = types.BoolValue(*sub.Subscribed)
	}
	if sub.SubscribedAt != nil {
		ts, err := responseToMap(sub.SubscribedAt)
		if err == nil {
			if secs, ok := ts["seconds"]; ok {
				data.SubscribedAt = types.StringValue(fmt.Sprintf("%v", secs))
			}
		}
	}

	if data.Id.IsNull() || data.Id.IsUnknown() {
		id := "blog-subscription"
		if !data.UserId.IsNull() && !data.UserId.IsUnknown() {
			id = data.UserId.ValueString()
		}
		data.Id = types.StringValue(id)
	}
}

func BlogSubscriptionResourceSchema() schema.Schema {
	return schema.Schema{
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
