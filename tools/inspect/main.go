package main

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/tools/go/packages"
)

func main() {
	cfg := &packages.Config{
		Mode: packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo | packages.NeedName | packages.NeedFiles,
		Dir:  "/Users/qeeqez/Projects/rixl/infra/terraform-provider-rixl",
	}
	pkgs, err := packages.Load(cfg, "github.com/rixlhq/rixl-go/sdk/feeds")
	if err != nil {
		fmt.Println("load error:", err)
		os.Exit(1)
	}
	if packages.PrintErrors(pkgs) > 0 {
		os.Exit(1)
	}
	for _, pkg := range pkgs {
		fmt.Println(pkg.Name, pkg.PkgPath)
		for _, name := range pkg.Types.Scope().Names() {
			if strings.HasPrefix(name, "SimpleClient") {
				fmt.Println("  type:", name)
			}
		}
		if obj := pkg.Types.Scope().Lookup("SimpleClient"); obj != nil {
			fmt.Println("  found SimpleClient")
		}
	}
}
