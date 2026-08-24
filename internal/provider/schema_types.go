package provider

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// attributeTypesForSchema extracts a map of attr.Type for the top-level
// attributes of either a datasource or resource schema. It uses reflection so
// that it works with both schema packages.
func attributeTypesForSchema(attributes any) (map[string]attr.Type, error) {
	attrsVal := reflect.ValueOf(attributes)
	if attrsVal.Kind() != reflect.Map {
		return nil, fmt.Errorf("expected map of attributes, got %T", attributes)
	}

	out := make(map[string]attr.Type, attrsVal.Len())
	for _, key := range attrsVal.MapKeys() {
		name := key.String()
		at, err := attrTypeForAttribute(attrsVal.MapIndex(key).Interface())
		if err != nil {
			return nil, fmt.Errorf("attribute %q: %w", name, err)
		}
		out[name] = at
	}
	return out, nil
}

func attrTypeForAttribute(a any) (attr.Type, error) {
	val := reflect.ValueOf(a)
	if val.Kind() == reflect.Pointer {
		val = val.Elem()
	}
	typ := val.Type()

	// Custom type takes precedence for all attribute kinds.
	if custom := fieldValueByName(val, "CustomType"); custom.IsValid() && !custom.IsZero() {
		if at, ok := reflect.TypeAssert[attr.Type](custom); ok {
			return at, nil
		}
	}

	switch typ.Name() {
	case "StringAttribute":
		return types.StringType, nil
	case "Int64Attribute":
		return types.Int64Type, nil
	case "BoolAttribute":
		return types.BoolType, nil
	case "Float64Attribute":
		return types.Float64Type, nil
	case "NumberAttribute":
		return types.NumberType, nil
	case "DynamicAttribute":
		return types.DynamicType, nil
	case "ListAttribute", "SetAttribute", "MapAttribute":
		et, err := attrTypeForCollectionElement(val, typ.Name())
		if err != nil {
			return nil, err
		}
		switch typ.Name() {
		case "ListAttribute":
			return types.ListType{ElemType: et}, nil
		case "SetAttribute":
			return types.SetType{ElemType: et}, nil
		case "MapAttribute":
			return types.MapType{ElemType: et}, nil
		}
	case "ListNestedAttribute", "SetNestedAttribute", "SingleNestedAttribute":
		et, err := attrTypeForNestedObject(val)
		if err != nil {
			return nil, err
		}
		switch typ.Name() {
		case "ListNestedAttribute", "SetNestedAttribute":
			// Both list and set nested attributes result in a list/set of objects.
			if typ.Name() == "ListNestedAttribute" {
				return types.ListType{ElemType: et}, nil
			}
			return types.SetType{ElemType: et}, nil
		case "SingleNestedAttribute":
			return et, nil
		}
	case "ObjectAttribute":
		attrTypesVal := fieldValueByName(val, "AttributeTypes")
		if !attrTypesVal.IsValid() {
			return nil, errors.New("ObjectAttribute missing AttributeTypes")
		}
		attrTypes, err := mapAttrTypes(attrTypesVal)
		if err != nil {
			return nil, err
		}
		return types.ObjectType{AttrTypes: attrTypes}, nil
	}

	return nil, fmt.Errorf("unsupported attribute type %T", a)
}

func attrTypeForCollectionElement(val reflect.Value, kind string) (attr.Type, error) {
	etVal := fieldValueByName(val, "ElementType")
	if !etVal.IsValid() || etVal.IsZero() {
		return nil, fmt.Errorf("%s missing ElementType", kind)
	}
	at, ok := reflect.TypeAssert[attr.Type](etVal)
	if !ok {
		return nil, fmt.Errorf("%s ElementType is not an attr.Type", kind)
	}
	return at, nil
}

func attrTypeForNestedObject(val reflect.Value) (attr.Type, error) {
	// List/Set nested attributes store their object under NestedObject.
	nestedVal := fieldValueByName(val, "NestedObject")
	attributesVal := reflect.Value{}

	if nestedVal.IsValid() {
		if custom := fieldValueByName(nestedVal, "CustomType"); custom.IsValid() && !custom.IsZero() {
			if at, ok := reflect.TypeAssert[attr.Type](custom); ok {
				return at, nil
			}
		}
		attributesVal = fieldValueByName(nestedVal, "Attributes")
	}

	// Single nested attributes store their attributes directly on the attribute.
	if !attributesVal.IsValid() {
		attributesVal = fieldValueByName(val, "Attributes")
	}

	if !attributesVal.IsValid() {
		return nil, errors.New("nested object missing Attributes")
	}
	attrTypes, err := mapAttrTypes(attributesVal)
	if err != nil {
		return nil, err
	}
	return types.ObjectType{AttrTypes: attrTypes}, nil
}

func mapAttrTypes(val reflect.Value) (map[string]attr.Type, error) {
	if val.Kind() != reflect.Map {
		return nil, fmt.Errorf("expected attribute map, got %s", val.Kind())
	}
	out := make(map[string]attr.Type, val.Len())
	for _, key := range val.MapKeys() {
		name := key.String()
		at, err := attrTypeForAttribute(val.MapIndex(key).Interface())
		if err != nil {
			return nil, fmt.Errorf("nested attribute %q: %w", name, err)
		}
		out[name] = at
	}
	return out, nil
}

func fieldValueByName(val reflect.Value, name string) reflect.Value {
	field := val.FieldByName(name)
	return field
}

// userProvidedAttributes returns the names of top-level attributes that are
// user-provided (Required or Optional). These are inputs that the API may not
// echo, so they should be preserved when mapping a response into a model.
func userProvidedAttributes(attributes any) (map[string]bool, error) {
	attrsVal := reflect.ValueOf(attributes)
	if attrsVal.Kind() != reflect.Map {
		return nil, fmt.Errorf("expected map of attributes, got %T", attributes)
	}

	out := make(map[string]bool, attrsVal.Len())
	for _, key := range attrsVal.MapKeys() {
		name := key.String()
		attrVal := attrsVal.MapIndex(key)
		if attrVal.Kind() == reflect.Pointer {
			attrVal = attrVal.Elem()
		}
		if !attrVal.IsValid() {
			continue
		}

		required := fieldValueByName(attrVal, "Required")
		optional := fieldValueByName(attrVal, "Optional")
		if required.IsValid() && required.Bool() {
			out[name] = true
			continue
		}
		if optional.IsValid() && optional.Bool() {
			out[name] = true
		}
	}
	return out, nil
}
