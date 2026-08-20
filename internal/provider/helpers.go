package provider

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/rixlhq/terraform-provider-rixl/internal/rixlcommon"
)

var placeholderReg = regexp.MustCompile(`\{([^}]+)\}`)

func buildPath(value tftypes.Value, template string, aliases map[string]string) (string, error) {
	obj, err := asObject(value)
	if err != nil {
		return "", err
	}

	path := placeholderReg.ReplaceAllStringFunc(template, func(match string) string {
		name := match[1 : len(match)-1]
		attr := name
		if aliases != nil {
			if a, ok := aliases[name]; ok {
				attr = a
			}
		}
		v, ok := obj[attr]
		if !ok {
			return match
		}
		s, err := rixlcommon.ValueAsString(v)
		if err != nil {
			return match
		}
		return s
	})

	if strings.Contains(path, "{") {
		return "", fmt.Errorf("path %q still contains unresolved placeholders", path)
	}
	return path, nil
}

func asObject(v tftypes.Value) (map[string]tftypes.Value, error) {
	if v.IsNull() || !v.IsKnown() {
		return nil, errors.New("value is null or unknown")
	}
	objType, ok := v.Type().(tftypes.Object)
	if !ok {
		return nil, fmt.Errorf("expected object value, got %s", v.Type())
	}
	m := make(map[string]tftypes.Value, len(objType.AttributeTypes))
	if err := v.As(&m); err != nil {
		return nil, err
	}
	return m, nil
}

func responseToState(ctx context.Context, tfType tftypes.Type, body []byte, responseRoot string) (tftypes.Value, error) {
	decoded, err := rixlcommon.DecodeJSONResponse(body)
	if err != nil {
		return tftypes.Value{}, err
	}

	if responseRoot != "" {
		m, ok := decoded.(map[string]any)
		if !ok {
			return tftypes.Value{}, fmt.Errorf("expected object response for root %q", responseRoot)
		}
		for _, part := range strings.Split(responseRoot, ".") {
			if part == "" {
				continue
			}
			next, ok := m[part]
			if !ok {
				return rixlcommon.JSONToTfValue(ctx, tfType, decoded)
			}
			m, ok = next.(map[string]any)
			if !ok {
				return rixlcommon.JSONToTfValue(ctx, tfType, decoded)
			}
			decoded = m
		}
	}

	return rixlcommon.JSONToTfValue(ctx, tfType, decoded)
}
