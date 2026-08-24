//nolint:exhaustive,revive // generic SDK mapping handles known kinds; file is large due to shared mapping helpers
package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"math"
	"reflect"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rixlhq/rixl-go/sdk"
)

var _ datasource.DataSource = (*managedDataSource)(nil)

// DataSourceDescriptor defines how a Terraform data source maps to the Rixl SDK.
type DataSourceDescriptor struct {
	TypeName string

	// SchemaFn returns the data source schema. It must be a function with
	// signature func(context.Context) datasource/schema.Schema.
	SchemaFn any

	// Model is a pointer to a zero value of the model struct.
	Model any

	// ClientField is the name of the typed client on *sdk.Client.
	ClientField string

	// ReadMethod is the SDK method used to read data (e.g. "GetFeed" or
	// "ListFeeds").
	ReadMethod string

	// PathParams are the tfsdk field names, in order, that correspond to the
	// string path parameters of the SDK read method.
	PathParams []string

	// ReadResponseField unwraps a nested read response object. Use this when
	// the read response has the shape { "post": { ... } }.
	ReadResponseField string

	// PreserveMissing lists attribute names that should be preserved when they
	// are missing from an API response. Used for one-time secrets and other
	// computed values the API does not echo on read.
	PreserveMissing []string
}

type managedDataSource struct {
	client     *sdk.Client
	descriptor DataSourceDescriptor
}

func newManagedDataSource(d DataSourceDescriptor) datasource.DataSource {
	return &managedDataSource{descriptor: d}
}

func (d *managedDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.descriptor.TypeName
}

func (d *managedDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = d.dataSourceSchema(ctx)
}

func (d *managedDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*sdk.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data type", fmt.Sprintf("expected *sdk.Client, got %T", req.ProviderData))
		return
	}
	d.client = client
}

func (d *managedDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	data := newModel(d.descriptor.Model)
	resp.Diagnostics.Append(req.Config.Get(ctx, data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	clientVal, diags := d.clientValue()
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	method := clientVal.MethodByName(d.descriptor.ReadMethod)
	if !method.IsValid() {
		resp.Diagnostics.AddError("SDK method not found", d.descriptor.ReadMethod)
		return
	}

	response, diags := invokeSDKMethod(ctx, method, data, d.descriptor.PathParams)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	if response.IsValid() && response.CanInterface() {
		s := d.dataSourceSchema(ctx)
		attrs, err := attributeTypesForSchema(s.Attributes)
		if err != nil {
			resp.Diagnostics.AddError("Failed to derive attribute types", err.Error())
			return
		}

		m, err := responseToMap(response.Interface())
		if err != nil {
			resp.Diagnostics.AddError("Failed to convert response", err.Error())
			return
		}

		if d.descriptor.ReadResponseField != "" {
			raw, ok := m[d.descriptor.ReadResponseField]
			if !ok || raw == nil {
				resp.Diagnostics.AddError("Read response missing field", d.descriptor.ReadResponseField)
				return
			}
			unwrapped, ok := raw.(map[string]any)
			if !ok {
				resp.Diagnostics.AddError("Read response field is not an object", d.descriptor.ReadResponseField)
				return
			}
			m = unwrapped
		}

		preserve, d := buildPreserveSet(s.Attributes, attrs, d.descriptor.PathParams, nil, d.descriptor.PreserveMissing)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}

		resp.Diagnostics.Append(mapToModel(ctx, m, data, attrs, preserve)...)
		if resp.Diagnostics.HasError() {
			return
		}
		resp.Diagnostics.Append(mapResponsePagination(ctx, m, data)...)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (d *managedDataSource) clientValue() (reflect.Value, diag.Diagnostics) {
	var diags diag.Diagnostics
	v := reflect.ValueOf(d.client).Elem().FieldByName(d.descriptor.ClientField)
	if !v.IsValid() {
		diags.AddError("SDK client not found", d.descriptor.ClientField)
		return reflect.Value{}, diags
	}
	if v.Kind() == reflect.Pointer && v.IsNil() {
		diags.AddError("SDK client is nil", d.descriptor.ClientField)
		return reflect.Value{}, diags
	}
	return v, diags
}

func (d *managedDataSource) dataSourceSchema(ctx context.Context) dschema.Schema {
	fnVal := reflect.ValueOf(d.descriptor.SchemaFn)
	res := fnVal.Call([]reflect.Value{reflect.ValueOf(ctx)})
	return res[0].Interface().(dschema.Schema)
}

// mapResponsePagination maps response pagination fields (limit/offset) to model
// fields that may be named "limit"/"offset" or "paginationlimit"/"paginationoffset".
func mapResponsePagination(_ context.Context, m map[string]any, data any) diag.Diagnostics {
	var diags diag.Diagnostics

	val := reflect.ValueOf(data)
	if val.Kind() == reflect.Pointer {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return diags
	}

	type fieldRef struct {
		idx   int
		match string
	}
	fields := make(map[string]fieldRef)
	typ := val.Type()
	for i := range typ.NumField() {
		name := typ.Field(i).Tag.Get("tfsdk")
		if name == "" {
			continue
		}
		fields[name] = fieldRef{idx: i, match: typ.Field(i).Type.Name()}
	}

	for src, targets := range map[string][]string{
		"limit":          {"limit", "paginationlimit"},
		"offset":         {"offset", "paginationoffset"},
		"page":           {"page"},
		"page_size":      {"page_size"},
		"total":          {"total"},
		"sort_field":     {"sort_field"},
		"sort_direction": {"sort_direction"},
	} {
		raw, ok := m[src]
		if !ok || raw == nil {
			continue
		}
		for _, t := range targets {
			f, ok := fields[t]
			if !ok {
				continue
			}
			field := val.Field(f.idx)
			base, _ := strings.CutSuffix(f.match, "Value")
			switch base {
			case "Int64":
				v, d := anyToInt64(raw)
				if d.HasError() {
					diags.Append(d...)
					continue
				}
				field.Set(reflect.ValueOf(types.Int64Value(v)))
			case "String":
				field.Set(reflect.ValueOf(types.StringValue(nativeToString(raw))))
			}
		}
	}

	return diags
}

func anyToInt64(v any) (int64, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch t := v.(type) {
	case int64:
		return t, diags
	case int:
		return int64(t), diags
	case int32:
		return int64(t), diags
	case json.Number:
		i, err := strconv.ParseInt(string(t), 10, 64)
		if err != nil {
			diags.AddError("Invalid integer", fmt.Sprintf("cannot parse %q as int64: %s", t, err))
			return 0, diags
		}
		return i, diags
	case float64:
		if t == math.Trunc(t) && t >= math.MinInt64 && t <= math.MaxInt64 {
			return int64(t), diags
		}
		diags.AddError("Invalid integer", fmt.Sprintf("cannot convert %v to int64", v))
		return 0, diags
	case string:
		i, err := strconv.ParseInt(t, 10, 64)
		if err != nil {
			diags.AddError("Invalid integer", fmt.Sprintf("cannot parse %q as int64: %s", t, err))
			return 0, diags
		}
		return i, diags
	default:
		diags.AddError("Invalid integer", fmt.Sprintf("cannot convert %T to int64", v))
		return 0, diags
	}
}

// Shared helpers.

func newModel(model any) any {
	return reflect.New(reflect.TypeOf(model).Elem()).Interface()
}

func invokeSDKMethod(ctx context.Context, method reflect.Value, data any, pathParams []string) (reflect.Value, diag.Diagnostics) {
	var diags diag.Diagnostics

	response, err := invokeSDKMethodErr(ctx, method, data, pathParams)
	if err != nil {
		diags.AddError("SDK request failed", err.Error())
		return reflect.Value{}, diags
	}
	return response, diags
}

func invokeSDKMethodErr(ctx context.Context, method reflect.Value, data any, pathParams []string) (reflect.Value, error) {
	mtype := method.Type()
	if mtype.NumIn() == 0 || mtype.In(0) != reflect.TypeFor[context.Context]() {
		return reflect.Value{}, errors.New("SDK method signature invalid: first argument must be context.Context")
	}

	args := []reflect.Value{reflect.ValueOf(ctx)}

	modelMap, diags := modelToMap(ctx, data)
	if diags.HasError() {
		return reflect.Value{}, fmt.Errorf("%s", diags)
	}

	pathIdx := 0
	bodyMap := make(map[string]any, len(modelMap))
	maps.Copy(bodyMap, modelMap)
	for _, p := range pathParams {
		delete(bodyMap, p)
	}
	// Timestamps and ids are typically read-only.
	delete(bodyMap, "id")
	delete(bodyMap, "created_at")
	delete(bodyMap, "updated_at")

	for i := 1; i < mtype.NumIn(); i++ {
		inType := mtype.In(i)

		// Variadic RequestEditorFn at the end; pass empty slice.
		if i == mtype.NumIn()-1 && method.Type().IsVariadic() {
			slice := reflect.MakeSlice(inType, 0, 0)
			args = append(args, slice)
			continue
		}

		if inType.Kind() == reflect.String {
			if pathIdx >= len(pathParams) {
				return reflect.Value{}, fmt.Errorf("SDK path params mismatch: missing path param for arg %d", i)
			}
			v, d := modelFieldString(data, pathParams[pathIdx])
			if d.HasError() {
				return reflect.Value{}, fmt.Errorf("%s", d)
			}
			args = append(args, reflect.ValueOf(v).Convert(inType))
			pathIdx++
			continue
		}

		// Pointer to a parameter struct (e.g. *ListFeedsParams).
		if inType.Kind() == reflect.Pointer && inType.Elem().Kind() == reflect.Struct {
			paramsPtr := reflect.New(inType.Elem())
			if err := buildStructFromModel(modelMap, paramsPtr.Elem()); err != nil {
				return reflect.Value{}, fmt.Errorf("failed to build request params: %w", err)
			}
			args = append(args, paramsPtr)
			continue
		}

		// Body argument: any, map[string]any, or a typed struct.
		bodyVal, d := buildBodyValue(bodyMap, inType)
		if d.HasError() {
			return reflect.Value{}, fmt.Errorf("%s", d)
		}
		args = append(args, bodyVal)
	}

	rets := method.Call(args)
	if len(rets) < 2 {
		return reflect.Value{}, errors.New("SDK method return mismatch: expected (response, error)")
	}

	errVal := rets[len(rets)-1]
	if !errVal.IsNil() {
		return reflect.Value{}, errVal.Interface().(error)
	}

	return rets[0], nil
}

func modelFieldString(data any, tfsdk string) (string, diag.Diagnostics) {
	var diags diag.Diagnostics
	v, d := modelFieldValue(data, tfsdk)
	diags.Append(d...)
	if diags.HasError() {
		return "", diags
	}
	if v == nil || v.IsNull() || v.IsUnknown() {
		return "", diags
	}
	s, ok := v.(types.String)
	if !ok {
		diags.AddError("Path param type mismatch", tfsdk)
		return "", diags
	}
	return s.ValueString(), diags
}

func modelFieldValue(data any, tfsdk string) (attr.Value, diag.Diagnostics) {
	var diags diag.Diagnostics
	val := reflect.ValueOf(data)
	if val.Kind() == reflect.Pointer {
		val = val.Elem()
	}
	typ := val.Type()
	for i := range val.NumField() {
		if f := typ.Field(i).Tag.Get("tfsdk"); f == tfsdk {
			return val.Field(i).Interface().(attr.Value), diags
		}
	}
	diags.AddError("Model field not found", tfsdk)
	return nil, diags
}

func buildBodyValue(bodyMap map[string]any, target reflect.Type) (reflect.Value, diag.Diagnostics) {
	var diags diag.Diagnostics

	switch target.Kind() {
	case reflect.Interface:
		// "any" / interface{}: pass the map directly.
		return reflect.ValueOf(bodyMap), diags
	case reflect.Map:
		m := reflect.MakeMapWithSize(target, len(bodyMap))
		for k, v := range bodyMap {
			m.SetMapIndex(reflect.ValueOf(k), reflect.ValueOf(v))
		}
		return m, diags
	case reflect.Struct:
		v := reflect.New(target).Interface()
		b, err := json.Marshal(bodyMap)
		if err != nil {
			diags.AddError("Marshal body failed", err.Error())
			return reflect.Value{}, diags
		}
		if err := json.Unmarshal(b, v); err != nil {
			diags.AddError("Unmarshal body failed", err.Error())
			return reflect.Value{}, diags
		}
		return reflect.ValueOf(v).Elem(), diags
	case reflect.Pointer:
		if target.Elem().Kind() == reflect.Struct {
			vv, d := buildBodyValue(bodyMap, target.Elem())
			diags.Append(d...)
			if diags.HasError() {
				return reflect.Value{}, diags
			}
			return vv.Addr(), diags
		}
	}

	diags.AddError("Unsupported body argument type", target.String())
	return reflect.Value{}, diags
}

// buildStructFromModel fills a struct from a model map by matching the
// normalized json/form tags to the map keys.
func buildStructFromModel(modelMap map[string]any, target reflect.Value) error {
	typeOfT := target.Type()
	out := make(map[string]any, typeOfT.NumField())

	for field := range typeOfT.Fields() {
		key := structFieldKey(field)
		if key == "" {
			continue
		}
		norm := normalizeKey(key)
		for mk, mv := range modelMap {
			if normalizeKey(mk) == norm {
				out[key] = mv
				break
			}
		}
	}

	b, err := json.Marshal(out)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, target.Addr().Interface())
}

func structFieldKey(field reflect.StructField) string {
	for _, tag := range []string{"json", "form", "url"} {
		if t := field.Tag.Get(tag); t != "" {
			if before, _, ok := strings.Cut(t, ","); ok {
				return before
			}
			return t
		}
	}
	return ""
}

func normalizeKey(k string) string {
	return strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(k, ".", ""), "_", ""))
}
