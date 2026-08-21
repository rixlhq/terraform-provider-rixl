package provider

import (
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

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
