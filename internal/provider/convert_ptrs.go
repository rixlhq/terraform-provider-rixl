package provider

import (
	"time"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ptr returns a pointer to v. Use this for any literal or value that needs to
// be passed as *T to the SDK.
func ptr[T any](v T) *T {
	return &v
}

// tfStringPtr converts a types.String to *string, returning nil for null/unknown.
func tfStringPtr(v types.String) *string {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	s := v.ValueString()
	return &s
}

// tfInt32Ptr converts a types.Int64 to *int32, returning nil for null/unknown.
func tfInt32Ptr(v types.Int64) *int32 {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	n := int32(v.ValueInt64())
	return &n
}

// ptrString converts *string to types.String (nil → null).
func ptrString(s *string) types.String {
	if s == nil {
		return types.StringNull()
	}
	return types.StringValue(*s)
}

// ptrInt64 converts *int64 to types.Int64 (nil → null).
func ptrInt64(i *int64) types.Int64 {
	if i == nil {
		return types.Int64Null()
	}
	return types.Int64Value(*i)
}

// ptrInt32 converts *int32 to types.Int64 (nil → null).
func ptrInt32(i *int32) types.Int64 {
	if i == nil {
		return types.Int64Null()
	}
	return types.Int64Value(int64(*i))
}

// ptrBool converts *bool to types.Bool (nil → null).
func ptrBool(b *bool) types.Bool {
	if b == nil {
		return types.BoolNull()
	}
	return types.BoolValue(*b)
}

// ptrTime converts *time.Time to types.String as RFC3339 (nil → null).
func ptrTime(t *time.Time) types.String {
	if t == nil {
		return types.StringNull()
	}
	return types.StringValue(t.Format(time.RFC3339Nano))
}
