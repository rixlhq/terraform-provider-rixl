package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func initImportState(ctx context.Context, sch schema.Schema) tfsdk.State {
	return tfsdk.State{
		Schema: sch,
		Raw:    tftypes.NewValue(sch.Type().TerraformType(ctx), nil),
	}
}

func TestImportState_Project(t *testing.T) {
	ctx := context.Background()
	r := NewProjectResource().(resource.ResourceWithImportState)
	sch := ProjectResourceSchema(ctx)

	t.Run("valid id", func(t *testing.T) {
		req := resource.ImportStateRequest{ID: "org-1/proj-1"}
		resp := &resource.ImportStateResponse{State: initImportState(ctx, sch)}
		r.ImportState(ctx, req, resp)
		if resp.Diagnostics.HasError() {
			t.Fatalf("unexpected error: %v", resp.Diagnostics)
		}

		var orgID, id types.String
		resp.State.GetAttribute(ctx, path.Root("org_id"), &orgID)
		resp.State.GetAttribute(ctx, path.Root("id"), &id)
		if orgID.ValueString() != "org-1" || id.ValueString() != "proj-1" {
			t.Errorf("got orgID=%s, id=%s; want org-1, proj-1", orgID.ValueString(), id.ValueString())
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		req := resource.ImportStateRequest{ID: "invalid-id-without-slash"}
		resp := &resource.ImportStateResponse{State: initImportState(ctx, sch)}
		r.ImportState(ctx, req, resp)
		if !resp.Diagnostics.HasError() {
			t.Fatal("expected error for invalid import id format")
		}
	})
}

func TestImportState_AccessPolicy(t *testing.T) {
	ctx := context.Background()
	r := NewAccessPolicyResource().(resource.ResourceWithImportState)
	sch := AccessPolicyResourceSchema(ctx)

	t.Run("valid id", func(t *testing.T) {
		req := resource.ImportStateRequest{ID: "org-1/policy-1"}
		resp := &resource.ImportStateResponse{State: initImportState(ctx, sch)}
		r.ImportState(ctx, req, resp)
		if resp.Diagnostics.HasError() {
			t.Fatalf("unexpected error: %v", resp.Diagnostics)
		}

		var orgID, id types.String
		resp.State.GetAttribute(ctx, path.Root("org_id"), &orgID)
		resp.State.GetAttribute(ctx, path.Root("id"), &id)
		if orgID.ValueString() != "org-1" || id.ValueString() != "policy-1" {
			t.Errorf("got orgID=%s, id=%s; want org-1, policy-1", orgID.ValueString(), id.ValueString())
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		req := resource.ImportStateRequest{ID: "no-slash"}
		resp := &resource.ImportStateResponse{State: initImportState(ctx, sch)}
		r.ImportState(ctx, req, resp)
		if !resp.Diagnostics.HasError() {
			t.Fatal("expected error for invalid import id format")
		}
	})
}

func TestImportState_Feed(t *testing.T) {
	ctx := context.Background()
	r := NewFeedResource().(resource.ResourceWithImportState)
	sch := FeedResourceSchema(ctx)

	t.Run("valid id", func(t *testing.T) {
		req := resource.ImportStateRequest{ID: "proj-1/feed-1"}
		resp := &resource.ImportStateResponse{State: initImportState(ctx, sch)}
		r.ImportState(ctx, req, resp)
		if resp.Diagnostics.HasError() {
			t.Fatalf("unexpected error: %v", resp.Diagnostics)
		}

		var projID, id types.String
		resp.State.GetAttribute(ctx, path.Root("project_id"), &projID)
		resp.State.GetAttribute(ctx, path.Root("id"), &id)
		if projID.ValueString() != "proj-1" || id.ValueString() != "feed-1" {
			t.Errorf("got projID=%s, id=%s; want proj-1, feed-1", projID.ValueString(), id.ValueString())
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		req := resource.ImportStateRequest{ID: "no-slash"}
		resp := &resource.ImportStateResponse{State: initImportState(ctx, sch)}
		r.ImportState(ctx, req, resp)
		if !resp.Diagnostics.HasError() {
			t.Fatal("expected error for invalid import id format")
		}
	})
}

func TestImportState_Dashboard(t *testing.T) {
	ctx := context.Background()
	r := NewDashboardResource().(resource.ResourceWithImportState)
	sch := DashboardResourceSchema(ctx)

	req := resource.ImportStateRequest{ID: "dash-123"}
	resp := &resource.ImportStateResponse{State: initImportState(ctx, sch)}
	r.ImportState(ctx, req, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected error: %v", resp.Diagnostics)
	}

	var id types.String
	resp.State.GetAttribute(ctx, path.Root("id"), &id)
	if id.ValueString() != "dash-123" {
		t.Errorf("got id=%s; want dash-123", id.ValueString())
	}
}

func TestImportState_DashboardWidget(t *testing.T) {
	ctx := context.Background()
	r := NewDashboardWidgetResource().(resource.ResourceWithImportState)
	sch := DashboardWidgetResourceSchema()

	t.Run("valid id", func(t *testing.T) {
		req := resource.ImportStateRequest{ID: "dash-1/widget-1"}
		resp := &resource.ImportStateResponse{State: initImportState(ctx, sch)}
		r.ImportState(ctx, req, resp)
		if resp.Diagnostics.HasError() {
			t.Fatalf("unexpected error: %v", resp.Diagnostics)
		}

		var dashID, id types.String
		resp.State.GetAttribute(ctx, path.Root("dashboard_id"), &dashID)
		resp.State.GetAttribute(ctx, path.Root("id"), &id)
		if dashID.ValueString() != "dash-1" || id.ValueString() != "widget-1" {
			t.Errorf("got dashID=%s, id=%s; want dash-1, widget-1", dashID.ValueString(), id.ValueString())
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		req := resource.ImportStateRequest{ID: "no-slash"}
		resp := &resource.ImportStateResponse{State: initImportState(ctx, sch)}
		r.ImportState(ctx, req, resp)
		if !resp.Diagnostics.HasError() {
			t.Fatal("expected error for invalid import id format")
		}
	})
}

func TestImportState_CustomDomain(t *testing.T) {
	ctx := context.Background()
	r := NewCustomDomainResource().(resource.ResourceWithImportState)
	sch := CustomDomainResourceSchema(ctx)

	t.Run("slash separated org and domain", func(t *testing.T) {
		req := resource.ImportStateRequest{ID: "org-1/example.com"}
		resp := &resource.ImportStateResponse{State: initImportState(ctx, sch)}
		r.ImportState(ctx, req, resp)
		if resp.Diagnostics.HasError() {
			t.Fatalf("unexpected error: %v", resp.Diagnostics)
		}

		var orgID, domain types.String
		resp.State.GetAttribute(ctx, path.Root("org_id"), &orgID)
		resp.State.GetAttribute(ctx, path.Root("domain"), &domain)
		if orgID.ValueString() != "org-1" || domain.ValueString() != "example.com" {
			t.Errorf("got orgID=%s, domain=%s; want org-1, example.com", orgID.ValueString(), domain.ValueString())
		}
	})

	t.Run("single id for org", func(t *testing.T) {
		req := resource.ImportStateRequest{ID: "org-1"}
		resp := &resource.ImportStateResponse{State: initImportState(ctx, sch)}
		r.ImportState(ctx, req, resp)
		if resp.Diagnostics.HasError() {
			t.Fatalf("unexpected error: %v", resp.Diagnostics)
		}

		var orgID, id types.String
		resp.State.GetAttribute(ctx, path.Root("org_id"), &orgID)
		resp.State.GetAttribute(ctx, path.Root("id"), &id)
		if orgID.ValueString() != "org-1" || id.ValueString() != "org-1" {
			t.Errorf("got orgID=%s, id=%s; want org-1, org-1", orgID.ValueString(), id.ValueString())
		}
	})
}

func TestImportState_DomainAutoJoin(t *testing.T) {
	ctx := context.Background()
	r := NewDomainAutoJoinResource().(resource.ResourceWithImportState)
	sch := DomainAutoJoinResourceSchema()

	req := resource.ImportStateRequest{ID: "org-1"}
	resp := &resource.ImportStateResponse{State: initImportState(ctx, sch)}
	r.ImportState(ctx, req, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected error: %v", resp.Diagnostics)
	}

	var orgID, id types.String
	resp.State.GetAttribute(ctx, path.Root("org_id"), &orgID)
	resp.State.GetAttribute(ctx, path.Root("id"), &id)
	if orgID.ValueString() != "org-1" || id.ValueString() != "org-1" {
		t.Errorf("got orgID=%s, id=%s; want org-1, org-1", orgID.ValueString(), id.ValueString())
	}
}

func TestImportState_BillingAddress(t *testing.T) {
	ctx := context.Background()
	r := NewBillingAddressResource().(resource.ResourceWithImportState)
	sch := BillingAddressResourceSchema(ctx)

	req := resource.ImportStateRequest{ID: "org-1"}
	resp := &resource.ImportStateResponse{State: initImportState(ctx, sch)}
	r.ImportState(ctx, req, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected error: %v", resp.Diagnostics)
	}

	var orgID, id types.String
	resp.State.GetAttribute(ctx, path.Root("org_id"), &orgID)
	resp.State.GetAttribute(ctx, path.Root("id"), &id)
	if orgID.ValueString() != "org-1" || id.ValueString() != "org-1" {
		t.Errorf("got orgID=%s, id=%s; want org-1, org-1", orgID.ValueString(), id.ValueString())
	}
}

func TestImportState_OrganizationMember(t *testing.T) {
	ctx := context.Background()
	r := NewOrganizationMemberResource().(resource.ResourceWithImportState)
	sch := OrganizationMemberResourceSchema()

	t.Run("valid id", func(t *testing.T) {
		req := resource.ImportStateRequest{ID: "org-1/user-1"}
		resp := &resource.ImportStateResponse{State: initImportState(ctx, sch)}
		r.ImportState(ctx, req, resp)
		if resp.Diagnostics.HasError() {
			t.Fatalf("unexpected error: %v", resp.Diagnostics)
		}

		var orgID, userID types.String
		resp.State.GetAttribute(ctx, path.Root("org_id"), &orgID)
		resp.State.GetAttribute(ctx, path.Root("user_id"), &userID)
		if orgID.ValueString() != "org-1" || userID.ValueString() != "user-1" {
			t.Errorf("got orgID=%s, userID=%s; want org-1, user-1", orgID.ValueString(), userID.ValueString())
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		req := resource.ImportStateRequest{ID: "no-slash"}
		resp := &resource.ImportStateResponse{State: initImportState(ctx, sch)}
		r.ImportState(ctx, req, resp)
		if !resp.Diagnostics.HasError() {
			t.Fatal("expected error for invalid import id format")
		}
	})
}
