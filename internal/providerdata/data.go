// Package providerdata holds dependencies shared between provider resources and data sources.
package providerdata

import "github.com/rixlhq/terraform-provider-rixl/internal/rixlclient"

// Data holds the configured provider dependencies shared with resources and data sources.
type Data struct {
	Client *rixlclient.Client
}
