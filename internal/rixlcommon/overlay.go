package rixlcommon

import (
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// OverlayKnown returns a new tftypes.Value of the same type as response,
// replacing any null or unknown response attributes with the corresponding
// known value from base. This preserves configured path parameters and other
// fields the API response may omit.
func OverlayKnown(base, response tftypes.Value) tftypes.Value {
	if !response.IsKnown() || response.IsNull() {
		if base.IsKnown() && !base.IsNull() {
			return base
		}
		return response
	}

	t := response.Type()
	switch ty := t.(type) {
	case tftypes.Object:
		respObj := make(map[string]tftypes.Value)
		_ = response.As(&respObj)

		baseObj := make(map[string]tftypes.Value)
		if base.IsKnown() && !base.IsNull() {
			_ = base.As(&baseObj)
		}

		vals := make(map[string]tftypes.Value, len(ty.AttributeTypes))
		for attr, attrType := range ty.AttributeTypes {
			respAttr, ok := respObj[attr]
			if !ok || !respAttr.IsKnown() || respAttr.IsNull() {
				if baseAttr, ok := baseObj[attr]; ok && baseAttr.IsKnown() && !baseAttr.IsNull() {
					vals[attr] = baseAttr
					continue
				}
				vals[attr] = tftypes.NewValue(attrType, nil)
				continue
			}

			baseAttr := tftypes.NewValue(attrType, nil)
			if b, ok := baseObj[attr]; ok {
				baseAttr = b
			}
			vals[attr] = OverlayKnown(baseAttr, respAttr)
		}
		return tftypes.NewValue(t, vals)
	case tftypes.List, tftypes.Set, tftypes.Tuple, tftypes.Map:
		return response
	default:
		return response
	}
}
