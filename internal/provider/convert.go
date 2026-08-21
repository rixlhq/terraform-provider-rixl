package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"reflect"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// modelToMap converts a Terraform model struct (with tfsdk tags) into a native
// Go map. Null or unknown values are omitted, making the result suitable for
// JSON request bodies.
func modelToMap(ctx context.Context, model any) (map[string]any, diag.Diagnostics) {
	var diags diag.Diagnostics
	out := map[string]any{}

	val := reflect.ValueOf(model)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		name := fieldName(field)
		if name == "" {
			continue
		}

		v, d := toNative(ctx, val.Field(i).Interface().(attr.Value))
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		// Omit null, unknown or zero values for request bodies.
		if v == nil {
			continue
		}

		out[name] = v
	}

	return out, diags
}

// mapToModel sets the fields of a Terraform model struct from a native Go map
// using the provided schema for type conversion.
func mapToModel(ctx context.Context, m map[string]any, model any, attributes map[string]attr.Type) diag.Diagnostics {
	var diags diag.Diagnostics

	val := reflect.ValueOf(model)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		name := fieldName(field)
		if name == "" {
			continue
		}

		attrType, ok := attributes[name]
		if !ok {
			continue
		}

		raw, ok := m[name]
		if !ok || raw == nil {
			// Leave field at zero (null) value.
			continue
		}

		av, d := fromNative(ctx, raw, attrType)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		val.Field(i).Set(reflect.ValueOf(av))
	}

	return diags
}

// toNative converts an attr.Value into a native Go value (string, bool, int64,
// float64, []any, map[string]any) by walking its tftypes representation.
func toNative(ctx context.Context, v attr.Value) (any, diag.Diagnostics) {
	var diags diag.Diagnostics

	if v.IsNull() || v.IsUnknown() {
		return nil, diags
	}

	tv, err := v.ToTerraformValue(ctx)
	if err != nil {
		diags.AddError("Terraform value conversion failed", err.Error())
		return nil, diags
	}

	return tftypesToNative(tv)
}

func tftypesToNative(tv tftypes.Value) (any, diag.Diagnostics) {
	var diags diag.Diagnostics

	if tv.IsNull() || !tv.IsKnown() {
		return nil, diags
	}

	typ := tv.Type()
	switch {
	case typ.Is(tftypes.String):
		var s string
		if err := tv.As(&s); err != nil {
			diags.AddError("String conversion failed", err.Error())
			return nil, diags
		}
		return s, diags
	case typ.Is(tftypes.Bool):
		var b bool
		if err := tv.As(&b); err != nil {
			diags.AddError("Bool conversion failed", err.Error())
			return nil, diags
		}
		return b, diags
	case typ.Is(tftypes.Number):
		var n *big.Float
		if err := tv.As(&n); err != nil {
			diags.AddError("Number conversion failed", err.Error())
			return nil, diags
		}
		if n.IsInt() {
			i, _ := n.Int64()
			return i, diags
		}
		f, _ := n.Float64()
		return f, diags
	case typ.Is(tftypes.List{}), typ.Is(tftypes.Set{}), typ.Is(tftypes.Tuple{}):
		var elements []tftypes.Value
		if err := tv.As(&elements); err != nil {
			diags.AddError("List conversion failed", err.Error())
			return nil, diags
		}
		out := make([]any, 0, len(elements))
		for _, e := range elements {
			nv, d := tftypesToNative(e)
			diags.Append(d...)
			if diags.HasError() {
				return nil, diags
			}
			out = append(out, nv)
		}
		return out, diags
	case typ.Is(tftypes.Object{}):
		var attrs map[string]tftypes.Value
		if err := tv.As(&attrs); err != nil {
			diags.AddError("Object conversion failed", err.Error())
			return nil, diags
		}
		// Preserve stable ordering to make maps deterministic.
		keys := make([]string, 0, len(attrs))
		for k := range attrs {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		out := make(map[string]any, len(attrs))
		for _, k := range keys {
			nv, d := tftypesToNative(attrs[k])
			diags.Append(d...)
			if diags.HasError() {
				return nil, diags
			}
			out[k] = nv
		}
		return out, diags
	default:
		diags.AddError("Unsupported terraform type", fmt.Sprintf("cannot convert %s", typ.String()))
		return nil, diags
	}
}

// fromNative converts a native Go value into an attr.Value for a given
// attr.Type by building the appropriate tftypes.Value and delegating to the
// attr.Type.
func fromNative(ctx context.Context, v any, attrType attr.Type) (attr.Value, diag.Diagnostics) {
	var diags diag.Diagnostics

	if v == nil {
		av, err := attrType.ValueFromTerraform(ctx, tftypes.NewValue(attrType.TerraformType(ctx), nil))
		if err != nil {
			diags.AddError("ValueFromTerraform failed", err.Error())
			return nil, diags
		}
		return av, diags
	}

	tfType := attrType.TerraformType(ctx)
	tv, d := nativeToTftypes(ctx, v, tfType)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}

	av, err := attrType.ValueFromTerraform(ctx, tv)
	if err != nil {
		diags.AddError("ValueFromTerraform failed", err.Error())
		return nil, diags
	}

	return av, diags
}

func nativeToTftypes(_ context.Context, v any, tfType tftypes.Type) (tftypes.Value, diag.Diagnostics) {
	var diags diag.Diagnostics

	switch {
	case tfType.Is(tftypes.String):
		s, ok := v.(string)
		if !ok {
			s = fmt.Sprint(v)
		}
		return tftypes.NewValue(tfType, s), diags
	case tfType.Is(tftypes.Bool):
		b, err := toBool(v)
		if err != nil {
			diags.AddError("Type mismatch", err.Error())
			return tftypes.Value{}, diags
		}
		return tftypes.NewValue(tfType, b), diags
	case tfType.Is(tftypes.Number):
		var n *big.Float
		switch t := v.(type) {
		case int:
			n = big.NewFloat(float64(t))
		case int8:
			n = big.NewFloat(float64(t))
		case int16:
			n = big.NewFloat(float64(t))
		case int32:
			n = big.NewFloat(float64(t))
		case int64:
			n = big.NewFloat(float64(t))
		case uint:
			n = big.NewFloat(float64(t))
		case uint8:
			n = big.NewFloat(float64(t))
		case uint16:
			n = big.NewFloat(float64(t))
		case uint32:
			n = big.NewFloat(float64(t))
		case uint64:
			n = big.NewFloat(float64(t))
		case float32:
			n = big.NewFloat(float64(t))
		case float64:
			n = big.NewFloat(t)
		case *big.Float:
			n = t
		case *big.Rat:
			if t != nil {
				n = new(big.Float).SetRat(t)
			}
		case string:
			var ok bool
			n, ok = new(big.Float).SetString(t)
			if !ok {
				diags.AddError("Invalid number", fmt.Sprintf("cannot parse %q as number", t))
				return tftypes.Value{}, diags
			}
		default:
			diags.AddError("Type mismatch", fmt.Sprintf("expected number, got %T", v))
			return tftypes.Value{}, diags
		}
		return tftypes.NewValue(tfType, n), diags
	case tfType.Is(tftypes.List{}), tfType.Is(tftypes.Set{}), tfType.Is(tftypes.Tuple{}):
		slice, ok := v.([]any)
		if !ok {
			diags.AddError("Type mismatch", fmt.Sprintf("expected list, got %T", v))
			return tftypes.Value{}, diags
		}
		elements := make([]tftypes.Value, 0, len(slice))
		var elemType tftypes.Type
		switch t := tfType.(type) {
		case tftypes.List:
			elemType = t.ElementType
		case tftypes.Set:
			elemType = t.ElementType
		case tftypes.Tuple:
			if len(slice) != len(t.ElementTypes) {
				diags.AddError("Tuple length mismatch", fmt.Sprintf("expected %d elements, got %d", len(t.ElementTypes), len(slice)))
				return tftypes.Value{}, diags
			}
		}
		for i, e := range slice {
			if tup, ok := tfType.(tftypes.Tuple); ok {
				elemType = tup.ElementTypes[i]
			}
			ev, d := nativeToTftypes(nil, e, elemType)
			diags.Append(d...)
			if diags.HasError() {
				return tftypes.Value{}, diags
			}
			elements = append(elements, ev)
		}
		return tftypes.NewValue(tfType, elements), diags
	case tfType.Is(tftypes.Object{}):
		m, ok := v.(map[string]any)
		if !ok {
			diags.AddError("Type mismatch", fmt.Sprintf("expected object, got %T", v))
			return tftypes.Value{}, diags
		}
		objType := tfType.(tftypes.Object)
		attrs := make(map[string]tftypes.Value, len(objType.AttributeTypes))
		for name, at := range objType.AttributeTypes {
			raw, ok := m[name]
			if !ok {
				raw = nil
			}
			ev, d := nativeToTftypes(nil, raw, at)
			diags.Append(d...)
			if diags.HasError() {
				return tftypes.Value{}, diags
			}
			attrs[name] = ev
		}
		return tftypes.NewValue(tfType, attrs), diags
	default:
		diags.AddError("Unsupported terraform type", fmt.Sprintf("cannot build tftypes.Value for %s", tfType.String()))
		return tftypes.Value{}, diags
	}
}

func toBool(v any) (bool, error) {
	switch t := v.(type) {
	case bool:
		return t, nil
	case string:
		return t == "true" || t == "1" || t == "yes" || t == "on", nil
	case int:
		return t != 0, nil
	case int8:
		return t != 0, nil
	case int16:
		return t != 0, nil
	case int32:
		return t != 0, nil
	case int64:
		return t != 0, nil
	case uint:
		return t != 0, nil
	case uint8:
		return t != 0, nil
	case uint16:
		return t != 0, nil
	case uint32:
		return t != 0, nil
	case uint64:
		return t != 0, nil
	case float32:
		return t != 0, nil
	case float64:
		return t != 0, nil
	default:
		return false, fmt.Errorf("expected bool, got %T", v)
	}
}

func fieldName(field reflect.StructField) string {
	return field.Tag.Get("tfsdk")
}

// mapResponseToModel converts a strongly-typed SDK response value into a
// Terraform model struct using the provided schema attributes.
func mapResponseToModel(ctx context.Context, response any, model any, attributes any) diag.Diagnostics {
	m, err := responseToMap(response)
	if err != nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic("Failed to convert response", err.Error())}
	}

	attrTypes, err := attributeTypesForSchema(attributes)
	if err != nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic("Failed to derive attribute types", err.Error())}
	}

	return mapToModel(ctx, m, model, attrTypes)
}

// responseToMap converts a strongly-typed SDK response value into a native Go
// map by serializing it as JSON. This produces the same shape that Terraform
// configurations use.
func responseToMap(v any) (map[string]any, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	return m, nil
}
