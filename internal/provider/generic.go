//nolint:revive,exhaustive // generic SDK mapping is long and handles known kinds
package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rixlhq/rixl-go/sdk"
)

// ResourceDescriptor defines how a Terraform resource maps to the Rixl SDK.
type ResourceDescriptor struct {
	TypeName string

	// SchemaFn returns the resource schema. It must be a function with signature
	// func(context.Context) resource/schema.Schema.
	SchemaFn any

	// Model is a pointer to a zero value of the model struct.
	Model any

	// ClientField is the name of the typed client on *sdk.Client, e.g. "Feeds".
	ClientField string

	// SDK method names. Empty methods are skipped.
	CreateMethod string
	ReadMethod   string
	UpdateMethod string
	DeleteMethod string

	// PathParams are the tfsdk field names, in order, that correspond to the
	// string path parameters of the SDK methods.
	PathParams []string
}

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
}

type managedResource struct {
	client     *sdk.Client
	descriptor ResourceDescriptor
}

type managedDataSource struct {
	client     *sdk.Client
	descriptor DataSourceDescriptor
}

func newManagedResource(d ResourceDescriptor) resource.Resource {
	return &managedResource{descriptor: d}
}

func newManagedDataSource(d DataSourceDescriptor) datasource.DataSource {
	return &managedDataSource{descriptor: d}
}

func (r *managedResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + r.descriptor.TypeName
}

func (r *managedResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = r.resourceSchema(ctx)
}

func (r *managedResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*sdk.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data type", fmt.Sprintf("expected *sdk.Client, got %T", req.ProviderData))
		return
	}
	r.client = client
}

func (r *managedResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.descriptor.CreateMethod == "" {
		resp.Diagnostics.AddError("Create not supported", "This resource does not support create operations.")
		return
	}
	r.callSDK(ctx, r.descriptor.CreateMethod, &req.Plan, &resp.State, &resp.Diagnostics)
}

func (r *managedResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.descriptor.ReadMethod == "" {
		resp.Diagnostics.AddError("Read not supported", "This resource does not support read operations.")
		return
	}
	r.callSDK(ctx, r.descriptor.ReadMethod, &req.State, &resp.State, &resp.Diagnostics)
}

func (r *managedResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.descriptor.UpdateMethod == "" {
		resp.Diagnostics.AddError("Update not supported", "This resource does not support update operations.")
		return
	}
	r.callSDK(ctx, r.descriptor.UpdateMethod, &req.Plan, &resp.State, &resp.Diagnostics)
}

func (r *managedResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.descriptor.DeleteMethod == "" {
		resp.Diagnostics.AddError("Delete not supported", "This resource does not support delete operations.")
		return
	}
	data := newModel(r.descriptor.Model)
	resp.Diagnostics.Append(req.State.Get(ctx, data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	clientVal, d := r.clientValue()
	if d.HasError() {
		resp.Diagnostics.Append(d...)
		return
	}

	method := clientVal.MethodByName(r.descriptor.DeleteMethod)
	if !method.IsValid() {
		resp.Diagnostics.AddError("SDK method not found", r.descriptor.DeleteMethod)
		return
	}

	_, d = invokeSDKMethod(ctx, method, data, r.descriptor.PathParams)
	resp.Diagnostics.Append(d...)
}

func (r *managedResource) callSDK(ctx context.Context, methodName string, from tfsdkGetter, to tfsdkSetter, diags *diag.Diagnostics) {
	data := newModel(r.descriptor.Model)
	diags.Append(from.Get(ctx, data)...)
	if diags.HasError() {
		return
	}

	clientVal, d := r.clientValue()
	if d.HasError() {
		diags.Append(d...)
		return
	}

	method := clientVal.MethodByName(methodName)
	if !method.IsValid() {
		diags.AddError("SDK method not found", methodName)
		return
	}

	response, d := invokeSDKMethod(ctx, method, data, r.descriptor.PathParams)
	if d.HasError() {
		diags.Append(d...)
		return
	}

	if response.IsValid() && response.CanInterface() {
		attrs, err := attributeTypesForSchema(r.resourceSchema(ctx).Attributes)
		if err != nil {
			diags.AddError("Failed to derive attribute types", err.Error())
			return
		}
		diags.Append(mapToModel(ctx, mustResponseToMap(response.Interface()), data, attrs)...)
		if diags.HasError() {
			return
		}
	}

	diags.Append(to.Set(ctx, data)...)
}

func (r *managedResource) clientValue() (reflect.Value, diag.Diagnostics) {
	var diags diag.Diagnostics
	v := reflect.ValueOf(r.client).Elem().FieldByName(r.descriptor.ClientField)
	if !v.IsValid() {
		diags.AddError("SDK client not found", r.descriptor.ClientField)
		return reflect.Value{}, diags
	}
	if v.Kind() == reflect.Pointer && v.IsNil() {
		diags.AddError("SDK client is nil", r.descriptor.ClientField)
		return reflect.Value{}, diags
	}
	return v, diags
}

func (r *managedResource) resourceSchema(ctx context.Context) rschema.Schema {
	fnVal := reflect.ValueOf(r.descriptor.SchemaFn)
	res := fnVal.Call([]reflect.Value{reflect.ValueOf(ctx)})
	return res[0].Interface().(rschema.Schema)
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
		resp.Diagnostics.Append(mapToModel(ctx, mustResponseToMap(response.Interface()), data, attrs)...)
		if resp.Diagnostics.HasError() {
			return
		}
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

// Shared helpers.

type tfsdkGetter interface {
	Get(context.Context, any) diag.Diagnostics
}

type tfsdkSetter interface {
	Set(context.Context, any) diag.Diagnostics
}

func newModel(model any) any {
	return reflect.New(reflect.TypeOf(model).Elem()).Interface()
}

func mustResponseToMap(v any) map[string]any {
	m, err := responseToMap(v)
	if err != nil {
		panic(err)
	}
	return m
}

func invokeSDKMethod(ctx context.Context, method reflect.Value, data any, pathParams []string) (reflect.Value, diag.Diagnostics) {
	var diags diag.Diagnostics

	mtype := method.Type()
	if mtype.NumIn() == 0 || mtype.In(0) != reflect.TypeOf((*context.Context)(nil)).Elem() {
		diags.AddError("SDK method signature invalid", "first argument must be context.Context")
		return reflect.Value{}, diags
	}

	args := []reflect.Value{reflect.ValueOf(ctx)}

	modelMap, d := modelToMap(ctx, data)
	diags.Append(d...)
	if diags.HasError() {
		return reflect.Value{}, diags
	}

	pathIdx := 0
	bodyMap := make(map[string]any, len(modelMap))
	for k, v := range modelMap {
		bodyMap[k] = v
	}
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
				diags.AddError("SDK path params mismatch", fmt.Sprintf("missing path param for arg %d", i))
				return reflect.Value{}, diags
			}
			v, d := modelFieldString(data, pathParams[pathIdx])
			diags.Append(d...)
			if diags.HasError() {
				return reflect.Value{}, diags
			}
			args = append(args, reflect.ValueOf(v))
			pathIdx++
			continue
		}

		// Pointer to a parameter struct (e.g. *ListFeedsParams).
		if inType.Kind() == reflect.Pointer && inType.Elem().Kind() == reflect.Struct {
			paramsPtr := reflect.New(inType.Elem())
			if err := buildStructFromModel(modelMap, paramsPtr.Elem()); err != nil {
				diags.AddError("Failed to build request params", err.Error())
				return reflect.Value{}, diags
			}
			args = append(args, paramsPtr)
			continue
		}

		// Body argument: any, map[string]any, or a typed struct.
		bodyVal, d := buildBodyValue(bodyMap, inType)
		diags.Append(d...)
		if diags.HasError() {
			return reflect.Value{}, diags
		}
		args = append(args, bodyVal)
	}

	rets := method.Call(args)
	if len(rets) < 2 {
		diags.AddError("SDK method return mismatch", "expected (response, error)")
		return reflect.Value{}, diags
	}

	errVal := rets[len(rets)-1]
	if !errVal.IsNil() {
		diags.AddError("SDK request failed", errVal.Interface().(error).Error())
		return reflect.Value{}, diags
	}

	return rets[0], diags
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

	for i := range typeOfT.NumField() {
		field := typeOfT.Field(i)
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
			if i := strings.Index(t, ","); i >= 0 {
				return t[:i]
			}
			return t
		}
	}
	return ""
}

func normalizeKey(k string) string {
	return strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(k, ".", ""), "_", ""))
}
