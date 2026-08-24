package provider

import "github.com/hashicorp/terraform-plugin-framework/resource"

// genericResourceConstructors returns the constructors for all generic
// managed resources. Manual resources are still returned directly from
// rixlProvider.Resources.
func genericResourceConstructors() []func() resource.Resource {
	return []func() resource.Resource{
		NewApiKeyResource,
		NewClientCredentialResource,
		NewPaymentMethodResource,
		NewPolicyAttachmentResource,
		NewProjectCustomDomainResource,
		NewSubscriptionResource,
	}
}

func NewApiKeyResource() resource.Resource {
	return newManagedResource(ResourceDescriptor{
		TypeName:            "api_key",
		SchemaFn:            ApiKeyResourceSchema,
		Model:               &ApiKeyModel{},
		ClientField:         "APIKeys",
		CreateMethod:        "CreateApiKey",
		ReadMethod:          "ListApiKeys",
		DeleteMethod:        "DeleteApiKey",
		PathParams:          []string{"org_id"},
		DeletePathParams:    []string{"org_id", "id"},
		CreateKeepPathKeys:  []string{},
		UpdateKeepPathKeys:  []string{},
		BodyRenames:         map[string]string{},
		ComputedBodyKeys:    []string{"id", "created_at", "last_used", "project_name", "secret"},
		CreateResponseField: "api_key",
		ReadListField:       "api_keys",
		ReadListIDField:     "id",
		ReadAfterCreate:     false,
		ReadAfterUpdate:     false,
	})
}

func NewClientCredentialResource() resource.Resource {
	return newManagedResource(ResourceDescriptor{
		TypeName:            "client_credential",
		SchemaFn:            ClientCredentialResourceSchema,
		Model:               &ClientCredentialModel{},
		ClientField:         "ClientCredentials",
		CreateMethod:        "CreateClientCredential",
		ReadMethod:          "ListClientCredentials",
		DeleteMethod:        "RevokeClientCredential",
		PathParams:          []string{},
		DeletePathParams:    []string{"id"},
		CreateKeepPathKeys:  []string{},
		UpdateKeepPathKeys:  []string{},
		BodyRenames:         map[string]string{},
		ComputedBodyKeys:    []string{"id", "created_at", "client_id", "kid", "last_used_at", "status", "client_secret"},
		CreateResponseField: "",
		CreateFlattenField:  "credential",
		ReadListField:       "credentials",
		ReadListIDField:     "id",
		ReadAfterCreate:     false,
		ReadAfterUpdate:     false,
	})
}

func NewPaymentMethodResource() resource.Resource {
	return newManagedResource(ResourceDescriptor{
		TypeName:            "payment_method",
		SchemaFn:            PaymentMethodResourceSchema,
		Model:               &PaymentMethodModel{},
		ClientField:         "Payments",
		CreateMethod:        "UpsertPaymentMethod",
		ReadMethod:          "ListPaymentMethods",
		UpdateMethod:        "UpsertPaymentMethod",
		DeleteMethod:        "DeletePaymentMethod",
		PathParams:          []string{},
		DeletePathParams:    []string{"id"},
		CreateKeepPathKeys:  []string{},
		UpdateKeepPathKeys:  []string{},
		BodyRenames:         map[string]string{},
		ComputedBodyKeys:    []string{"id", "type", "provider", "details", "is_default", "brand", "last4", "exp_month", "exp_year", "created_at"},
		CreateResponseField: "",
		ReadListField:       "payment_methods",
		ReadListIDField:     "id",
		ReadAfterCreate:     true,
		ReadAfterUpdate:     true,
	})
}

func NewPolicyAttachmentResource() resource.Resource {
	return newManagedResource(ResourceDescriptor{
		TypeName:            "policy_attachment",
		SchemaFn:            PolicyAttachmentResourceSchema,
		Model:               &PolicyAttachmentModel{},
		ClientField:         "AccessPolicies",
		CreateMethod:        "AttachPolicy",
		ReadMethod:          "ListPolicyAttachments",
		DeleteMethod:        "DetachPolicy",
		PathParams:          []string{"org_id", "policy_id"},
		DeletePathParams:    []string{"org_id", "id"},
		CreateKeepPathKeys:  []string{},
		UpdateKeepPathKeys:  []string{},
		BodyRenames:         map[string]string{},
		ComputedBodyKeys:    []string{"id", "created_at"},
		CreateResponseField: "",
		ReadListField:       "attachments",
		ReadListIDField:     "id",
		ReadAfterCreate:     true,
		ReadAfterUpdate:     false,
	})
}

func NewProjectCustomDomainResource() resource.Resource {
	return newManagedResource(ResourceDescriptor{
		TypeName:            "project_custom_domain",
		SchemaFn:            ProjectCustomDomainResourceSchema,
		Model:               &ProjectCustomDomainModel{},
		ClientField:         "Projects",
		CreateMethod:        "SetCustomDomain",
		ReadMethod:          "GetProject",
		UpdateMethod:        "SetCustomDomain",
		DeleteMethod:        "RemoveCustomDomain",
		PathParams:          []string{"org_id", "project_id"},
		CreateKeepPathKeys:  []string{},
		UpdateKeepPathKeys:  []string{},
		BodyRenames:         map[string]string{},
		ComputedBodyKeys:    []string{"id", "created_at", "updated_at"},
		CreateResponseField: "",
		ReadListField:       "",
		ReadListIDField:     "id",
		ReadAfterCreate:     false,
		ReadAfterUpdate:     false,
	})
}

func NewSubscriptionResource() resource.Resource {
	return newManagedResource(ResourceDescriptor{
		TypeName:               "subscription",
		SchemaFn:               SubscriptionResourceSchema,
		Model:                  &SubscriptionModel{},
		ClientField:            "Subscriptions",
		UpdateClientField:      "Payments",
		CreateMethod:           "CreateSubscription",
		ReadMethod:             "GetSubscription",
		UpdateMethod:           "UpgradeSubscription",
		DeleteMethod:           "CancelSubscription",
		PathParams:             []string{},
		CreateKeepPathKeys:     []string{},
		UpdateKeepPathKeys:     []string{},
		BodyRenames:            map[string]string{},
		ComputedBodyKeys:       []string{"id", "plan_id", "plan_name", "plan_type", "status", "current_period_end", "cancel_at_period_end", "stripe_customer_id", "stripe_subscription_id", "currency", "expiring_soon", "price", "trials_ending_soon"},
		UpdateComputedBodyKeys: []string{"id", "payment_method_id", "billing_address", "plan_id", "plan_name", "plan_type", "status", "current_period_end", "cancel_at_period_end", "stripe_customer_id", "stripe_subscription_id", "currency", "expiring_soon", "price", "trials_ending_soon"},
		CreateResponseField:    "",
		ReadListField:          "",
		ReadListIDField:        "id",
		ReadAfterCreate:        true,
		ReadAfterUpdate:        true,
	})
}
