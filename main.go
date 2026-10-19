package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/rixlhq/terraform-provider-rixl/internal/provider"
	providerversion "github.com/rixlhq/terraform-provider-rixl/internal/version"
)

var (
	// version is populated at build time by ldflags (-X main.version=...) or falls back to internal/version.Version.
	version = providerversion.Version
)

func main() {
	var debug bool
	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/rixlhq/rixl",
		Debug:   debug,
	}

	err := providerserver.Serve(context.Background(), provider.New(version), opts)
	if err != nil {
		log.Fatal(err)
	}
}
