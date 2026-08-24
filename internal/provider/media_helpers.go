package provider

import (
	"errors"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rixlhq/rixl-go/sdk/images"
	"github.com/rixlhq/rixl-go/sdk/models"
	"github.com/rixlhq/rixl-go/sdk/videos"
)

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

// isTransientWaitError reports whether an error from a media GET poll is
// expected to resolve if we keep waiting. 404 means the upload has not been
// registered yet; 5xx may be transient. Any other HTTP client error or
// non-HTTP error should fail fast.
func isTransientWaitError(err error) bool {
	if err == nil {
		return true
	}

	var imgErr *images.ClientHttpError[struct{}]
	if errors.As(err, &imgErr) {
		return imgErr.StatusCode == http.StatusNotFound || imgErr.StatusCode >= 500
	}

	var vidErr *videos.ClientHttpError[struct{}]
	if errors.As(err, &vidErr) {
		return vidErr.StatusCode == http.StatusNotFound || vidErr.StatusCode >= 500
	}

	return false
}
