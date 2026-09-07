//nolint:revive // generic SDK mapping handles known kinds; file is large due to shared mapping helpers
package provider

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rixlhq/rixl-go/sdk"
)

var _ resource.Resource = (*managedResource)(nil)

// ResourceDescriptor defines how a Terraform resource maps to the Rixl SDK.
type ResourceDescriptor struct {
	TypeName string

	// SchemaFn returns the resource schema. It must be a function with
	// signature func(context.Context) resource/schema.Schema.
	SchemaFn any

	// Model is a pointer to a zero value of the model struct.
	Model any

	// ClientField is the name of the typed client on *sdk.Client.
	ClientField string

	// CreateClientField, ReadClientField, UpdateClientField and DeleteClientField
	// override ClientField for a specific operation. They default to ClientField
	// when empty.
	CreateClientField string
	ReadClientField   string
	UpdateClientField string
	DeleteClientField string

	// SDK methods for CRUD. Empty UpdateMethod means updates are not
	// supported and the resource will be replaced instead.
	CreateMethod string
	ReadMethod   string
	UpdateMethod string
	DeleteMethod string

	// PathParams are the tfsdk field names, in order, that correspond to the
	// string path parameters of the SDK create/read/update/delete methods.
	// They are used for any method that does not set the method-specific
	// variant below.
	PathParams []string

	// CreatePathParams, ReadPathParams, UpdatePathParams and DeletePathParams
	// override PathParams for a specific operation. This is needed when an
	// SDK method has different path arguments than the others (e.g. create
	// needs a feed_id, but delete needs the post_id).
	CreatePathParams []string
	ReadPathParams   []string
	UpdatePathParams []string
	DeletePathParams []string

	// CreateKeepPathKeys lists tfsdk path params that should still be sent in
	// the create request body.
	CreateKeepPathKeys []string

	// UpdateKeepPathKeys lists tfsdk path params (or the resource id) that
	// should still be sent in the update request body. Values are taken from
	// state so that computed ids can be used.
	UpdateKeepPathKeys []string

	// BodyRenames renames body keys after the body map is built. The source is
	// a tfsdk name and the target is the body key. This is useful when the API
	// expects a different body key than the tfsdk name (e.g. id -> feed_id).
	BodyRenames map[string]string

	// ComputedBodyKeys are keys removed from create/update request bodies.
	// Defaults to id, created_at, updated_at and created_by.
	ComputedBodyKeys []string

	// UpdateComputedBodyKeys override ComputedBodyKeys for update request
	// bodies. If empty, ComputedBodyKeys is used for updates as well.
	UpdateComputedBodyKeys []string

	// PreserveMissing lists attribute names that should be preserved when they
	// are missing from an API response. Used for one-time secrets and other
	// computed values the API does not echo on read.
	PreserveMissing []string

	// CreateResponseField unwraps a nested response object. If the create
	// response has the shape { "api_key": { ... } }, set this to "api_key".
	CreateResponseField string

	// CreateFlattenField flattens a nested response object into the top-level
	// map, so fields like { "credential": { "id": ... }, "client_secret": ... }
	// become top-level attributes. Nested object fields are added only when the
	// top-level key does not already exist.
	CreateFlattenField string

	// ReadResponseField unwraps a nested response object returned by ReadMethod.
	// Use this when the read response has the shape { "post": { ... } }.
	ReadResponseField string

	// ReadFlattenField flattens a nested read response object into the top-level
	// map. Similar to CreateFlattenField but applied to the read response.
	ReadFlattenField string

	// ReadListField is set when ReadMethod returns a list response and the
	// resource must be selected from that list by id.
	ReadListField string

	// ReadListIDField is the attribute name used to identify the resource in a
	// list read. Defaults to "id".
	ReadListIDField string

	// ReadAfterCreate calls Read after Create to refresh computed attributes.
	ReadAfterCreate bool

	// ReadAfterUpdate calls Read after Update to refresh computed attributes.
	ReadAfterUpdate bool
}

type managedResource struct {
	client     *sdk.Client
	descriptor ResourceDescriptor
}

func newManagedResource(d ResourceDescriptor) resource.Resource {
	return &managedResource{descriptor: d}
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
	plan := newModel(r.descriptor.Model)
	resp.Diagnostics.Append(req.Plan.Get(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	attrTypes, preserve, d := r.preserveSet(ctx, "create")
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	clientVal, diags := r.clientValueMaybeOverride(r.descriptor.CreateClientField)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	method := clientVal.MethodByName(r.descriptor.CreateMethod)
	if !method.IsValid() {
		resp.Diagnostics.AddError("SDK create method not found", r.descriptor.CreateMethod)
		return
	}

	bodyMap, d := r.buildCreateBody(ctx, plan)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	response, d, _ := r.invokeResourceMethod(ctx, method, plan, bodyMap, r.methodPathParams("create"))
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	m, diags := r.responseToResourceMap(response)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if m != nil {
		resp.Diagnostics.Append(mapToModel(ctx, m, plan, attrTypes, preserve)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if r.descriptor.ReadAfterCreate {
		readM, diags := r.doRead(ctx, plan)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		if readM != nil {
			resp.Diagnostics.Append(mapToModel(ctx, readM, plan, attrTypes, preserve)...)
			if resp.Diagnostics.HasError() {
				return
			}
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *managedResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	state := newModel(r.descriptor.Model)
	resp.Diagnostics.Append(req.State.Get(ctx, state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	attrTypes, preserve, d := r.preserveSet(ctx, "read")
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	m, diags := r.doRead(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if m == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(mapToModel(ctx, m, state, attrTypes, preserve)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *managedResource) doRead(ctx context.Context, state any) (map[string]any, diag.Diagnostics) {
	var diags diag.Diagnostics

	clientVal, d := r.clientValueMaybeOverride(r.descriptor.ReadClientField)
	if d.HasError() {
		diags.Append(d...)
		return nil, diags
	}

	method := clientVal.MethodByName(r.descriptor.ReadMethod)
	if !method.IsValid() {
		diags.AddError("SDK read method not found", r.descriptor.ReadMethod)
		return nil, diags
	}

	response, d, err := r.invokeResourceMethod(ctx, method, state, nil, r.methodPathParams("read"))
	if err != nil {
		if isNotFound(err) {
			return nil, diags
		}
		diags.Append(d...)
		return nil, diags
	}
	if d.HasError() {
		diags.Append(d...)
		return nil, diags
	}

	return r.responseToReadMap(ctx, response, state)
}

func (r *managedResource) responseToReadMap(ctx context.Context, response reflect.Value, model any) (map[string]any, diag.Diagnostics) {
	var diags diag.Diagnostics
	if !response.IsValid() || !response.CanInterface() {
		return nil, diags
	}

	m, err := responseToMap(response.Interface())
	if err != nil {
		diags.AddError("Failed to convert read response", err.Error())
		return nil, diags
	}

	if r.descriptor.ReadListField != "" {
		selected, d := r.selectFromList(ctx, m, model)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		if selected == nil {
			return nil, diags
		}
		return r.applyReverseBodyRenames(selected), diags
	}

	if r.descriptor.ReadResponseField != "" {
		raw, ok := m[r.descriptor.ReadResponseField]
		if !ok || raw == nil {
			diags.AddError("Read response missing field", r.descriptor.ReadResponseField)
			return nil, diags
		}
		unwrapped, ok := raw.(map[string]any)
		if !ok {
			diags.AddError("Read response field is not an object", r.descriptor.ReadResponseField)
			return nil, diags
		}
		m = unwrapped
	}

	if r.descriptor.ReadFlattenField != "" {
		raw, ok := m[r.descriptor.ReadFlattenField]
		if !ok || raw == nil {
			diags.AddError("Read response missing flatten field", r.descriptor.ReadFlattenField)
			return nil, diags
		}
		flat, ok := raw.(map[string]any)
		if !ok {
			diags.AddError("Read response flatten field is not an object", r.descriptor.ReadFlattenField)
			return nil, diags
		}
		for k, v := range m {
			if k != r.descriptor.ReadFlattenField {
				flat[k] = v
			}
		}
		m = flat
	}

	return r.applyReverseBodyRenames(m), diags
}

func (r *managedResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	state := newModel(r.descriptor.Model)
	plan := newModel(r.descriptor.Model)
	resp.Diagnostics.Append(req.State.Get(ctx, state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	attrTypes, preserve, d := r.preserveSet(ctx, "update")
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	data, diags := mergeStateAndPlan(ctx, state, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if r.descriptor.UpdateMethod == "" {
		// No update method; all mutable fields should have RequiresReplace.
		// Refresh state from the API to keep computed attributes in sync.
		m, diags := r.doRead(ctx, data)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		if m == nil {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.Append(mapToModel(ctx, m, data, attrTypes, preserve)...)
		if resp.Diagnostics.HasError() {
			return
		}
		resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
		return
	}

	clientVal, diags := r.clientValueMaybeOverride(r.descriptor.UpdateClientField)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	method := clientVal.MethodByName(r.descriptor.UpdateMethod)
	if !method.IsValid() {
		resp.Diagnostics.AddError("SDK update method not found", r.descriptor.UpdateMethod)
		return
	}

	bodyMap, d := r.buildUpdateBody(ctx, plan, state)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	response, d, err := r.invokeResourceMethod(ctx, method, data, bodyMap, r.methodPathParams("update"))
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.Append(d...)
		return
	}
	if d.HasError() {
		resp.Diagnostics.Append(d...)
		return
	}

	m, diags := r.responseToResourceMap(response)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if m != nil {
		resp.Diagnostics.Append(mapToModel(ctx, m, data, attrTypes, preserve)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if r.descriptor.ReadAfterUpdate {
		readM, diags := r.doRead(ctx, data)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		if readM != nil {
			resp.Diagnostics.Append(mapToModel(ctx, readM, data, attrTypes, preserve)...)
			if resp.Diagnostics.HasError() {
				return
			}
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *managedResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	state := newModel(r.descriptor.Model)
	resp.Diagnostics.Append(req.State.Get(ctx, state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	clientVal, diags := r.clientValueMaybeOverride(r.descriptor.DeleteClientField)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	method := clientVal.MethodByName(r.descriptor.DeleteMethod)
	if !method.IsValid() {
		resp.Diagnostics.AddError("SDK delete method not found", r.descriptor.DeleteMethod)
		return
	}

	_, d, err := r.invokeResourceMethod(ctx, method, state, nil, r.methodPathParams("delete"))
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.Append(d...)
		return
	}
	if d.HasError() {
		resp.Diagnostics.Append(d...)
		return
	}
}

func (r *managedResource) clientValueMaybeOverride(override string) (reflect.Value, diag.Diagnostics) {
	if override == "" {
		return r.clientValueFor(r.descriptor.ClientField)
	}
	return r.clientValueFor(override)
}

func (r *managedResource) clientValueFor(field string) (reflect.Value, diag.Diagnostics) {
	var diags diag.Diagnostics
	v := reflect.ValueOf(r.client).Elem().FieldByName(field)
	if !v.IsValid() {
		diags.AddError("SDK client not found", field)
		return reflect.Value{}, diags
	}
	if v.Kind() == reflect.Pointer && v.IsNil() {
		diags.AddError("SDK client is nil", field)
		return reflect.Value{}, diags
	}
	return v, diags
}

func (r *managedResource) resourceSchema(ctx context.Context) rschema.Schema {
	fnVal := reflect.ValueOf(r.descriptor.SchemaFn)
	res := fnVal.Call([]reflect.Value{reflect.ValueOf(ctx)})
	return res[0].Interface().(rschema.Schema)
}

func (r *managedResource) preserveSet(ctx context.Context, kind string) (map[string]attr.Type, map[string]bool, diag.Diagnostics) {
	var diags diag.Diagnostics
	s := r.resourceSchema(ctx)
	attrTypes, err := attributeTypesForSchema(s.Attributes)
	if err != nil {
		diags.AddError("Failed to derive attribute types", err.Error())
		return nil, nil, diags
	}

	computedKeys := r.descriptor.ComputedBodyKeys
	if kind == "update" && len(r.descriptor.UpdateComputedBodyKeys) > 0 {
		computedKeys = r.descriptor.UpdateComputedBodyKeys
	}
	if len(computedKeys) == 0 {
		computedKeys = []string{"id", "created_at", "updated_at", "created_by"}
	}

	preserve, d := buildPreserveSet(s.Attributes, attrTypes, r.descriptor.PathParams, computedKeys, r.descriptor.PreserveMissing)
	return attrTypes, preserve, d
}

func (r *managedResource) methodPathParams(kind string) []string {
	switch kind {
	case "create":
		if len(r.descriptor.CreatePathParams) > 0 {
			return r.descriptor.CreatePathParams
		}
	case "read":
		if len(r.descriptor.ReadPathParams) > 0 {
			return r.descriptor.ReadPathParams
		}
	case "update":
		if len(r.descriptor.UpdatePathParams) > 0 {
			return r.descriptor.UpdatePathParams
		}
	case "delete":
		if len(r.descriptor.DeletePathParams) > 0 {
			return r.descriptor.DeletePathParams
		}
	}
	return r.descriptor.PathParams
}

func (r *managedResource) buildCreateBody(ctx context.Context, plan any) (map[string]any, diag.Diagnostics) {
	bodyMap, diags := r.buildBody(ctx, plan, r.methodPathParams("create"), r.descriptor.CreateKeepPathKeys, r.descriptor.ComputedBodyKeys)
	if diags.HasError() {
		return nil, diags
	}
	return r.applyBodyRenames(bodyMap), diags
}

func (r *managedResource) buildUpdateBody(ctx context.Context, plan, state any) (map[string]any, diag.Diagnostics) {
	var diags diag.Diagnostics

	computedKeys := r.descriptor.UpdateComputedBodyKeys
	if len(computedKeys) == 0 {
		computedKeys = r.descriptor.ComputedBodyKeys
	}
	bodyMap, d := r.buildBody(ctx, plan, r.methodPathParams("update"), r.descriptor.UpdateKeepPathKeys, computedKeys)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}

	// Fill keep-path keys from state so computed ids are available for renames.
	for _, k := range r.descriptor.UpdateKeepPathKeys {
		v, d := modelFieldValue(state, k)
		if d.HasError() {
			diags.Append(d...)
			continue
		}
		if v == nil || v.IsNull() || v.IsUnknown() {
			continue
		}
		native, d := toNative(ctx, v)
		if d.HasError() {
			diags.Append(d...)
			continue
		}
		bodyMap[k] = native
	}

	return r.applyBodyRenames(bodyMap), diags
}

func (r *managedResource) buildBody(ctx context.Context, model any, pathParams, keepKeys, computedBodyKeys []string) (map[string]any, diag.Diagnostics) {
	var diags diag.Diagnostics
	bodyMap, d := modelToMap(ctx, model)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}

	keep := make(map[string]bool, len(keepKeys))
	for _, k := range keepKeys {
		keep[k] = true
	}

	computed := make(map[string]bool)
	for _, k := range computedBodyKeys {
		computed[k] = true
	}
	if len(computedBodyKeys) == 0 {
		for _, k := range []string{"id", "created_at", "updated_at", "created_by"} {
			computed[k] = true
		}
	}

	for _, p := range pathParams {
		if !keep[p] {
			delete(bodyMap, p)
		}
	}
	for k := range bodyMap {
		if computed[k] && !keep[k] {
			delete(bodyMap, k)
		}
	}

	return bodyMap, diags
}

func (r *managedResource) applyBodyRenames(bodyMap map[string]any) map[string]any {
	for src, dst := range r.descriptor.BodyRenames {
		if v, ok := bodyMap[src]; ok {
			bodyMap[dst] = v
			delete(bodyMap, src)
		}
	}
	return bodyMap
}

func (r *managedResource) applyReverseBodyRenames(m map[string]any) map[string]any {
	if len(r.descriptor.BodyRenames) == 0 {
		return m
	}
	reverseRenameMap(m, r.descriptor.BodyRenames)
	if r.descriptor.ReadListField == "" {
		return m
	}
	raw, ok := m[r.descriptor.ReadListField]
	if !ok {
		return m
	}
	list, ok := raw.([]any)
	if !ok {
		return m
	}
	for _, item := range list {
		if itemMap, ok := item.(map[string]any); ok {
			reverseRenameMap(itemMap, r.descriptor.BodyRenames)
		}
	}
	return m
}

func reverseRenameMap(m map[string]any, renames map[string]string) {
	for src, dst := range renames {
		v, ok := m[dst]
		if !ok {
			continue
		}
		if _, exists := m[src]; !exists {
			m[src] = v
		}
		delete(m, dst)
	}
}

func (r *managedResource) invokeResourceMethod(ctx context.Context, method reflect.Value, model any, bodyMap map[string]any, pathParams []string) (reflect.Value, diag.Diagnostics, error) {
	var diags diag.Diagnostics

	mtype := method.Type()
	if mtype.NumIn() == 0 || mtype.In(0) != reflect.TypeFor[context.Context]() {
		diags.AddError("SDK method signature invalid", "first argument must be context.Context")
		return reflect.Value{}, diags, nil
	}

	args := []reflect.Value{reflect.ValueOf(ctx)}

	modelMap, d := modelToMap(ctx, model)
	if d.HasError() {
		diags.Append(d...)
		return reflect.Value{}, diags, nil
	}

	pathIdx := 0

	for i := 1; i < mtype.NumIn(); i++ {
		inType := mtype.In(i)

		if i == mtype.NumIn()-1 && method.Type().IsVariadic() {
			slice := reflect.MakeSlice(inType, 0, 0)
			args = append(args, slice)
			continue
		}

		if inType.Kind() == reflect.String {
			if pathIdx >= len(pathParams) {
				diags.AddError("SDK path params mismatch", fmt.Sprintf("missing path param for arg %d", i))
				return reflect.Value{}, diags, nil
			}
			v, d := modelFieldString(model, pathParams[pathIdx])
			if d.HasError() {
				diags.Append(d...)
				return reflect.Value{}, diags, nil
			}
			args = append(args, reflect.ValueOf(v).Convert(inType))
			pathIdx++
			continue
		}

		if inType.Kind() == reflect.Pointer && inType.Elem().Kind() == reflect.Struct {
			paramsPtr := reflect.New(inType.Elem())
			src := modelMap
			if bodyMap != nil {
				src = bodyMap
			}
			if err := buildStructFromModel(src, paramsPtr.Elem()); err != nil {
				diags.AddError("Failed to build request params", err.Error())
				return reflect.Value{}, diags, nil
			}
			args = append(args, paramsPtr)
			continue
		}

		if bodyMap == nil {
			bodyMap = map[string]any{}
		}
		bodyVal, d := buildBodyValue(bodyMap, inType)
		if d.HasError() {
			diags.Append(d...)
			return reflect.Value{}, diags, nil
		}
		args = append(args, bodyVal)
	}

	rets := method.Call(args)
	if len(rets) < 2 {
		diags.AddError("SDK method return mismatch", "expected (response, error)")
		return reflect.Value{}, diags, nil
	}

	errVal := rets[len(rets)-1]
	if !errVal.IsNil() {
		e := errVal.Interface().(error)
		diags.AddError("SDK request failed", e.Error())
		return reflect.Value{}, diags, e
	}

	return rets[0], diags, nil
}

func (r *managedResource) responseToResourceMap(response reflect.Value) (map[string]any, diag.Diagnostics) {
	var diags diag.Diagnostics
	if !response.IsValid() || !response.CanInterface() {
		return nil, diags
	}

	m, err := responseToMap(response.Interface())
	if err != nil {
		diags.AddError("Failed to convert response", err.Error())
		return nil, diags
	}

	m, d := r.unwrapCreateResponse(m)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}

	m, d = r.flattenCreateResponse(m)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}

	return r.applyReverseBodyRenames(m), diags
}

func (r *managedResource) unwrapCreateResponse(m map[string]any) (map[string]any, diag.Diagnostics) {
	var diags diag.Diagnostics
	if r.descriptor.CreateResponseField == "" {
		return m, diags
	}
	raw, ok := m[r.descriptor.CreateResponseField]
	if !ok || raw == nil {
		diags.AddError("Create response missing field", r.descriptor.CreateResponseField)
		return nil, diags
	}
	unwrapped, ok := raw.(map[string]any)
	if !ok {
		diags.AddError("Create response field is not an object", r.descriptor.CreateResponseField)
		return nil, diags
	}
	return unwrapped, diags
}

func (r *managedResource) flattenCreateResponse(m map[string]any) (map[string]any, diag.Diagnostics) {
	var diags diag.Diagnostics
	if r.descriptor.CreateFlattenField == "" {
		return m, diags
	}
	raw, ok := m[r.descriptor.CreateFlattenField]
	if !ok || raw == nil {
		diags.AddError("Create response missing flatten field", r.descriptor.CreateFlattenField)
		return nil, diags
	}
	nested, ok := raw.(map[string]any)
	if !ok {
		diags.AddError("Create response flatten field is not an object", r.descriptor.CreateFlattenField)
		return nil, diags
	}
	for k, v := range nested {
		if _, exists := m[k]; !exists {
			m[k] = v
		}
	}
	delete(m, r.descriptor.CreateFlattenField)
	return m, diags
}

func (r *managedResource) selectFromList(ctx context.Context, m map[string]any, model any) (map[string]any, diag.Diagnostics) {
	var diags diag.Diagnostics
	raw, ok := m[r.descriptor.ReadListField]
	if !ok || raw == nil {
		return nil, diags
	}
	list, ok := raw.([]any)
	if !ok {
		diags.AddError("Read list field is not a list", r.descriptor.ReadListField)
		return nil, diags
	}
	idField := r.descriptor.ReadListIDField
	if idField == "" {
		idField = "id"
	}
	id := r.modelID(ctx, model)
	if id == nil {
		return nil, diags
	}
	for _, raw := range list {
		item, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		if item[idField] == id {
			return item, diags
		}
	}
	return nil, diags
}

func (r *managedResource) modelID(_ context.Context, model any) any {
	if model == nil {
		return nil
	}
	idField := r.descriptor.ReadListIDField
	if idField == "" {
		idField = "id"
	}
	v, _ := modelFieldValue(model, idField)
	if v == nil || v.IsNull() || v.IsUnknown() {
		return nil
	}
	s, ok := v.(types.String)
	if ok {
		return s.ValueString()
	}
	i, ok := v.(types.Int64)
	if ok {
		return i.ValueInt64()
	}
	return v
}

func isNotFound(err error) bool {
	if err == nil {
		return false
	}

	// Fast path: all SDK ClientHttpError types implement Error() as "HTTP %d".
	if strings.Contains(err.Error(), "HTTP 404") {
		return true
	}

	// Walk the error chain and any wrapped errors, looking for a StatusCode
	// field equal to 404. We use errors.Unwrap instead of reflecting on the
	// Unwrap method so that pointer- and value-receiver Unwrap implementations
	// are both handled.
	for err != nil {
		if hasStatusCode(err, 404) {
			return true
		}
		u, ok := err.(interface{ Unwrap() error })
		if !ok {
			break
		}
		err = u.Unwrap()
	}
	return false
}

func hasStatusCode(err error, code int64) bool {
	if err == nil {
		return false
	}
	v := reflect.ValueOf(err)
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return false
		}
		v = v.Elem()
	}
	if v.Kind() == reflect.Struct {
		if f := v.FieldByName("StatusCode"); f.IsValid() && f.CanInt() {
			return f.Int() == code
		}
	}
	return false
}
