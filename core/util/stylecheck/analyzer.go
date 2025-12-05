package stylecheck
// Copyright 2016-2025 Fraunhofer AISEC
//
// SPDX-License-Identifier: Apache-2.0

// Package stylecheck provides a custom Go static analyzer that enforces
// the code style guidelines defined in CONTRIBUTING.md.
package stylecheck

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

// Analyzer checks that functions with error returns use named return values.
var Analyzer = &analysis.Analyzer{
	Name: "stylecheck",
	Doc:  "checks that functions returning error use named return values",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			fn, ok := n.(*ast.FuncDecl)
			if !ok {
				return true
			}

			// Skip test functions
			if fn.Name != nil && len(fn.Name.Name) > 4 && fn.Name.Name[:4] == "Test" {
				return true
			}

			// Check if function has return values
			if fn.Type.Results == nil || len(fn.Type.Results.List) == 0 {
				return true
			}

			// Check each return value
			for _, result := range fn.Type.Results.List {
				// Check if return type is error
				ident, ok := result.Type.(*ast.Ident)
				if ok && ident.Name == "error" {
					// Check if the error return has a name
					if len(result.Names) == 0 {
						pass.Reportf(fn.Pos(), "function %q should use named return value for error", fn.Name.Name)
					}
				}
			}

			return true
		})
	}

	return nil, nil
}
