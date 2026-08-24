package provider

import (
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func fileResourceAttributes() map[string]rschema.Attribute {
	return map[string]rschema.Attribute{
		"created_at": rschema.StringAttribute{Computed: true},
		"format":     rschema.StringAttribute{Computed: true},
		"id":         rschema.StringAttribute{Computed: true},
		"name":       rschema.StringAttribute{Computed: true},
		"project_id": rschema.StringAttribute{Computed: true},
		"size":       rschema.Int64Attribute{Computed: true},
		"status":     rschema.StringAttribute{Computed: true},
		"updated_at": rschema.StringAttribute{Computed: true},
		"url":        rschema.StringAttribute{Computed: true},
	}
}

func imageResourceAttributes() map[string]rschema.Attribute {
	return map[string]rschema.Attribute{
		"attached_to_video": rschema.BoolAttribute{Computed: true},
		"file": rschema.SingleNestedAttribute{
			Computed:   true,
			Attributes: fileResourceAttributes(),
		},
		"height":     rschema.Int64Attribute{Computed: true},
		"id":         rschema.StringAttribute{Computed: true},
		"thumbhash":  rschema.StringAttribute{Computed: true},
		"visibility": rschema.StringAttribute{Computed: true},
		"width":      rschema.Int64Attribute{Computed: true},
	}
}

func fileDataSourceAttributes() map[string]dschema.Attribute {
	return map[string]dschema.Attribute{
		"created_at": dschema.StringAttribute{Computed: true},
		"format":     dschema.StringAttribute{Computed: true},
		"id":         dschema.StringAttribute{Computed: true},
		"name":       dschema.StringAttribute{Computed: true},
		"project_id": dschema.StringAttribute{Computed: true},
		"size":       dschema.Int64Attribute{Computed: true},
		"status":     dschema.StringAttribute{Computed: true},
		"updated_at": dschema.StringAttribute{Computed: true},
		"url":        dschema.StringAttribute{Computed: true},
	}
}

func imageDataSourceAttributes() map[string]dschema.Attribute {
	return map[string]dschema.Attribute{
		"attached_to_video": dschema.BoolAttribute{Computed: true},
		"file": dschema.SingleNestedAttribute{
			Computed:   true,
			Attributes: fileDataSourceAttributes(),
		},
		"height":     dschema.Int64Attribute{Computed: true},
		"id":         dschema.StringAttribute{Computed: true},
		"thumbhash":  dschema.StringAttribute{Computed: true},
		"visibility": dschema.StringAttribute{Computed: true},
		"width":      dschema.Int64Attribute{Computed: true},
	}
}
