package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

var (
	dir  = flag.String("dir", "internal/provider", "directory containing generated data source files")
	suff = flag.String("suff", "_data_source_gen.go", "filename suffix to rewrite")
)

func main() {
	flag.Parse()

	entries, err := os.ReadDir(*dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	for _, ent := range entries {
		if ent.IsDir() || !strings.HasSuffix(ent.Name(), *suff) {
			continue
		}
		path := filepath.Join(*dir, ent.Name())
		if err := rewrite(path); err != nil {
			fmt.Fprintf(os.Stderr, "rewrite %s: %v\n", path, err)
			os.Exit(1)
		}
		fmt.Fprintln(os.Stderr, "rewrote", path)
	}
}

func rewrite(path string) error {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	prefix := filePrefix(path, *suff)
	mapping := buildMapping(f, prefix)
	if len(mapping) == 0 {
		return nil
	}

	ast.Inspect(f, func(n ast.Node) bool {
		ident, ok := n.(*ast.Ident)
		if !ok {
			return true
		}
		if newName, ok := mapping[ident.Name]; ok {
			ident.Name = newName
		}
		return true
	})

	out, err := formatFile(fset, f)
	if err != nil {
		return fmt.Errorf("format: %w", err)
	}

	return os.WriteFile(path, out, 0o644)
}

// filePrefix extracts the PascalCased data-source name from the filename.
// For example, "image_data_source_gen.go" yields "Image" and
// "api_keys_data_source_gen.go" yields "ApiKeys".
func filePrefix(path, suff string) string {
	base := filepath.Base(path)
	base = strings.TrimSuffix(base, suff)
	parts := strings.Split(base, "_")
	for i, p := range parts {
		if p == "" {
			continue
		}
		parts[i] = strings.ToUpper(p[:1]) + p[1:]
	}
	return strings.Join(parts, "")
}

func buildMapping(f *ast.File, prefix string) map[string]string {
	mapping := make(map[string]string)
	dataSourcePrefix := prefix + "DataSource"

	for _, decl := range f.Decls {
		gd, ok := decl.(*ast.GenDecl)
		if !ok || gd.Tok != token.TYPE {
			continue
		}
		for _, spec := range gd.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			old := ts.Name.Name

			// Idempotent: don't double-rename types that already carry the
			// DataSource prefix (e.g. FeedDataSourceModel from a prior run).
			if strings.HasPrefix(old, dataSourcePrefix) {
				continue
			}

			var newName string
			if strings.HasPrefix(old, prefix) {
				newName = dataSourcePrefix + strings.TrimPrefix(old, prefix)
			} else {
				newName = dataSourcePrefix + old
			}
			mapping[old] = newName

			for _, s := range []string{"Null", "Unknown", "Must"} {
				mapping["New"+old+s] = "New" + newName + s
			}
			mapping["New"+old] = "New" + newName
		}
	}

	return mapping
}

func formatFile(fset *token.FileSet, f *ast.File) ([]byte, error) {
	var buf strings.Builder
	if err := format.Node(&buf, fset, f); err != nil {
		return nil, err
	}
	return []byte(buf.String()), nil
}
