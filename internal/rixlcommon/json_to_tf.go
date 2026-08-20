// Package rixlcommon contains helpers for converting between JSON and Terraform values.
package rixlcommon

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var (
	lowerToUpperReg     = regexp.MustCompile(`([a-z])[A-Z]`)
	unsupportedCharsReg = regexp.MustCompile(`[^a-zA-Z0-9_]+`)
	leadingNumbersReg   = regexp.MustCompile(`^(\d+)`)
)

// TerraformIdentifier converts a name to the snake_case form used by Terraform.
func TerraformIdentifier(original string) string {
	if len(original) == 0 {
		return original
	}
	removed := unsupportedCharsReg.ReplaceAllString(original, "")
	noLeading := leadingNumbersReg.ReplaceAllString(removed, "")
	inserted := lowerToUpperReg.ReplaceAllStringFunc(noLeading, func(s string) string {
		firstRune, size := utf8.DecodeRuneInString(s)
		return fmt.Sprintf("%s_%s", string(firstRune), strings.ToLower(s[size:]))
	})
	return strings.ToLower(inserted)
}

// DecodeJSONResponse decodes JSON using json.Number to preserve numeric precision.
func DecodeJSONResponse(body []byte) (any, error) {
	var v any
	dec := json.NewDecoder(strings.NewReader(string(body)))
	dec.UseNumber()
	if err := dec.Decode(&v); err != nil {
		return nil, err
	}
	return v, nil
}

// JSONToTfValue converts a JSON-decoded value into a tftypes.Value that matches t.
func JSONToTfValue(ctx context.Context, t tftypes.Type, v any) (tftypes.Value, error) {
	return jsonToTfValue(ctx, t, v)
}

func jsonToTfValue(ctx context.Context, t tftypes.Type, v any) (tftypes.Value, error) {
	if v == nil {
		return tftypes.NewValue(t, nil), nil
	}

	if t.Is(tftypes.String) {
		return tftypes.NewValue(tftypes.String, toString(v)), nil
	}
	if t.Is(tftypes.Number) {
		n, err := toBigFloat(v)
		if err != nil {
			return tftypes.Value{}, err
		}
		return tftypes.NewValue(tftypes.Number, n), nil
	}
	if t.Is(tftypes.Bool) {
		b, ok := v.(bool)
		if !ok {
			return tftypes.Value{}, fmt.Errorf("expected bool, got %T", v)
		}
		return tftypes.NewValue(tftypes.Bool, b), nil
	}
	if t.Is(tftypes.DynamicPseudoType) {
		return dynamicJSONToTfValue(v)
	}

	switch ty := t.(type) {
	case tftypes.Object:
		return jsonToTfObject(ctx, t, ty, v)
	case tftypes.List:
		return jsonToTfList(ctx, t, ty.ElementType, v)
	case tftypes.Set:
		return jsonToTfList(ctx, t, ty.ElementType, v)
	case tftypes.Map:
		return jsonToTfMap(ctx, t, ty, v)
	case tftypes.Tuple:
		return jsonToTfTuple(ctx, t, ty, v)
	default:
		return tftypes.Value{}, fmt.Errorf("unsupported tftypes.Type %T", t)
	}
}

func jsonToTfObject(ctx context.Context, t tftypes.Type, ty tftypes.Object, v any) (tftypes.Value, error) {
	m, ok := v.(map[string]any)
	if !ok {
		return tftypes.Value{}, fmt.Errorf("expected object for type %s, got %T", t, v)
	}

	normalized := make(map[string]any, len(m))
	for k, val := range m {
		normalized[TerraformIdentifier(k)] = val
	}

	vals := make(map[string]tftypes.Value, len(ty.AttributeTypes))
	for attr, attrType := range ty.AttributeTypes {
		attrVal, ok := normalized[attr]
		if !ok {
			vals[attr] = tftypes.NewValue(attrType, nil)
			continue
		}
		converted, err := jsonToTfValue(ctx, attrType, attrVal)
		if err != nil {
			return tftypes.Value{}, err
		}
		vals[attr] = converted
	}
	return tftypes.NewValue(t, vals), nil
}

func jsonToTfList(ctx context.Context, t tftypes.Type, elemType tftypes.Type, v any) (tftypes.Value, error) {
	s, ok := v.([]any)
	if !ok {
		return tftypes.Value{}, fmt.Errorf("expected list for type %s, got %T", t, v)
	}
	vals := make([]tftypes.Value, 0, len(s))
	for _, elem := range s {
		converted, err := jsonToTfValue(ctx, elemType, elem)
		if err != nil {
			return tftypes.Value{}, err
		}
		vals = append(vals, converted)
	}
	return tftypes.NewValue(t, vals), nil
}

func jsonToTfMap(ctx context.Context, t tftypes.Type, ty tftypes.Map, v any) (tftypes.Value, error) {
	m, ok := v.(map[string]any)
	if !ok {
		return tftypes.Value{}, fmt.Errorf("expected map for type %s, got %T", t, v)
	}
	vals := make(map[string]tftypes.Value, len(m))
	for k, elem := range m {
		converted, err := jsonToTfValue(ctx, ty.ElementType, elem)
		if err != nil {
			return tftypes.Value{}, err
		}
		vals[k] = converted
	}
	return tftypes.NewValue(t, vals), nil
}

func jsonToTfTuple(ctx context.Context, t tftypes.Type, ty tftypes.Tuple, v any) (tftypes.Value, error) {
	s, ok := v.([]any)
	if !ok {
		return tftypes.Value{}, fmt.Errorf("expected tuple for type %s, got %T", t, v)
	}
	if len(s) != len(ty.ElementTypes) {
		return tftypes.Value{}, fmt.Errorf("expected tuple of length %d, got %d", len(ty.ElementTypes), len(s))
	}
	vals := make([]tftypes.Value, 0, len(s))
	for i, elem := range s {
		converted, err := jsonToTfValue(ctx, ty.ElementTypes[i], elem)
		if err != nil {
			return tftypes.Value{}, err
		}
		vals = append(vals, converted)
	}
	return tftypes.NewValue(t, vals), nil
}

func toString(v any) string {
	switch val := v.(type) {
	case string:
		return val
	case json.Number:
		return val.String()
	case nil:
		return ""
	default:
		return fmt.Sprintf("%v", val)
	}
}

func toBigFloat(v any) (*big.Float, error) {
	switch val := v.(type) {
	case json.Number:
		f, _, err := big.NewFloat(0).SetPrec(128).Parse(val.String(), 10)
		if err != nil {
			return nil, err
		}
		return f, nil
	case float64:
		return big.NewFloat(val).SetPrec(128), nil
	case int:
		return big.NewFloat(float64(val)).SetPrec(128), nil
	case int64:
		return big.NewFloat(float64(val)).SetPrec(128), nil
	case string:
		f, _, err := big.NewFloat(0).SetPrec(128).Parse(val, 10)
		if err != nil {
			return nil, err
		}
		return f, nil
	default:
		return nil, fmt.Errorf("cannot convert %T to number", v)
	}
}

func dynamicJSONToTfValue(v any) (tftypes.Value, error) {
	switch val := v.(type) {
	case nil:
		return tftypes.NewValue(tftypes.DynamicPseudoType, nil), nil
	case bool:
		return tftypes.NewValue(tftypes.DynamicPseudoType, val), nil
	case json.Number:
		n, err := toBigFloat(val)
		if err != nil {
			return tftypes.Value{}, err
		}
		return tftypes.NewValue(tftypes.DynamicPseudoType, n), nil
	case float64:
		return tftypes.NewValue(tftypes.DynamicPseudoType, big.NewFloat(val).SetPrec(128)), nil
	case int:
		return tftypes.NewValue(tftypes.DynamicPseudoType, big.NewFloat(float64(val)).SetPrec(128)), nil
	case string:
		return tftypes.NewValue(tftypes.DynamicPseudoType, val), nil
	case []any:
		elems := make([]tftypes.Value, 0, len(val))
		for _, e := range val {
			el, err := dynamicJSONToTfValue(e)
			if err != nil {
				return tftypes.Value{}, err
			}
			elems = append(elems, el)
		}
		return tftypes.NewValue(tftypes.DynamicPseudoType, elems), nil
	case map[string]any:
		vals := make(map[string]tftypes.Value, len(val))
		for k, e := range val {
			el, err := dynamicJSONToTfValue(e)
			if err != nil {
				return tftypes.Value{}, err
			}
			vals[k] = el
		}
		return tftypes.NewValue(tftypes.DynamicPseudoType, vals), nil
	default:
		return tftypes.Value{}, fmt.Errorf("unsupported dynamic value type %T", val)
	}
}

// ValueAsString extracts a string, number, or bool value from a tftypes.Value.
func ValueAsString(v tftypes.Value) (string, error) {
	if v.IsNull() || !v.IsKnown() {
		return "", errors.New("value is null or unknown")
	}
	t := v.Type()
	if t.Is(tftypes.String) {
		var s string
		if err := v.As(&s); err != nil {
			return "", err
		}
		return s, nil
	}
	if t.Is(tftypes.Number) {
		var n big.Float
		if err := v.As(&n); err != nil {
			return "", err
		}
		return n.Text('f', 0), nil
	}
	if t.Is(tftypes.Bool) {
		var b bool
		if err := v.As(&b); err != nil {
			return "", err
		}
		return strconv.FormatBool(b), nil
	}
	return "", fmt.Errorf("cannot use %s as path parameter", t)
}
