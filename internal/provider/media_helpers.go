package provider

import (
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rixlhq/rixl-go/sdk/models"
)

func ptrString(s *string) types.String {
	if s == nil {
		return types.StringNull()
	}
	return types.StringValue(*s)
}

func stringPtr(s string) *string {
	return &s
}

func ptrInt64(i *int64) types.Int64 {
	if i == nil {
		return types.Int64Null()
	}
	return types.Int64Value(*i)
}

func ptrInt32(i *int32) types.Int64 {
	if i == nil {
		return types.Int64Null()
	}
	return types.Int64Value(int64(*i))
}

func ptrBool(b *bool) types.Bool {
	if b == nil {
		return types.BoolNull()
	}
	return types.BoolValue(*b)
}

func ptrTime(t *time.Time) types.String {
	if t == nil {
		return types.StringNull()
	}
	return types.StringValue(t.Format(time.RFC3339Nano))
}

func commonV1FileAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"created_at": types.StringType,
		"format":     types.StringType,
		"id":         types.StringType,
		"name":       types.StringType,
		"project_id": types.StringType,
		"size":       types.Int64Type,
		"status":     types.StringType,
		"updated_at": types.StringType,
		"url":        types.StringType,
	}
}

func commonV1FileToObject(f *models.CommonV1File) (types.Object, diag.Diagnostics) {
	if f == nil {
		return types.ObjectNull(commonV1FileAttrTypes()), nil
	}
	status := types.StringNull()
	if f.Status != nil {
		status = types.StringValue(string(*f.Status))
	}
	return types.ObjectValue(commonV1FileAttrTypes(), map[string]attr.Value{
		"created_at": ptrTime(f.CreatedAt),
		"format":     ptrString(f.Format),
		"id":         ptrString(f.ID),
		"name":       ptrString(f.Name),
		"project_id": ptrString(f.ProjectID),
		"size":       ptrInt64(f.Size),
		"status":     status,
		"updated_at": ptrTime(f.UpdatedAt),
		"url":        ptrString(f.URL),
	})
}

func imageV1AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"attached_to_video": types.BoolType,
		"file":              types.ObjectType{AttrTypes: commonV1FileAttrTypes()},
		"height":            types.Int64Type,
		"id":                types.StringType,
		"thumbhash":         types.StringType,
		"visibility":        types.StringType,
		"width":             types.Int64Type,
	}
}

func imageV1ToObject(img *models.ImagesV1Image) (types.Object, diag.Diagnostics) {
	if img == nil {
		return types.ObjectNull(imageV1AttrTypes()), nil
	}
	file, diags := commonV1FileToObject(img.File)
	if diags.HasError() {
		return types.ObjectNull(imageV1AttrTypes()), diags
	}
	return types.ObjectValue(imageV1AttrTypes(), map[string]attr.Value{
		"attached_to_video": ptrBool(img.AttachedToVideo),
		"file":              file,
		"height":            ptrInt32(img.Height),
		"id":                ptrString(img.ID),
		"thumbhash":         ptrString(img.Thumbhash),
		"visibility":        ptrString((*string)(img.Visibility)),
		"width":             ptrInt32(img.Width),
	})
}

func videoV1AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"bitrate":    types.Int64Type,
		"codec":      types.StringType,
		"duration":   types.StringType,
		"file":       types.ObjectType{AttrTypes: commonV1FileAttrTypes()},
		"framerate":  types.StringType,
		"hdr":        types.BoolType,
		"height":     types.Int64Type,
		"id":         types.StringType,
		"poster":     types.ObjectType{AttrTypes: imageV1AttrTypes()},
		"visibility": types.StringType,
		"width":      types.Int64Type,
	}
}

func videoV1ToObject(vid *models.VideosV1Video) (types.Object, diag.Diagnostics) {
	if vid == nil {
		return types.ObjectNull(videoV1AttrTypes()), nil
	}
	file, diags := commonV1FileToObject(vid.File)
	if diags.HasError() {
		return types.ObjectNull(videoV1AttrTypes()), diags
	}
	poster, d := imageV1ToObject(vid.Poster)
	if d.HasError() {
		return types.ObjectNull(videoV1AttrTypes()), d
	}
	return types.ObjectValue(videoV1AttrTypes(), map[string]attr.Value{
		"bitrate":    ptrInt32(vid.Bitrate),
		"codec":      ptrString(vid.Codec),
		"duration":   ptrString(vid.Duration),
		"file":       file,
		"framerate":  ptrString(vid.Framerate),
		"hdr":        ptrBool(vid.Hdr),
		"height":     ptrInt32(vid.Height),
		"id":         ptrString(vid.ID),
		"poster":     poster,
		"visibility": ptrString((*string)(vid.Visibility)),
		"width":      ptrInt32(vid.Width),
	})
}
