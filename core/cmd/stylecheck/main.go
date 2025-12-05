// Copyright 2016-2025 Fraunhofer AISEC
//
// SPDX-License-Identifier: Apache-2.0

// Command stylecheck runs the custom style checker for Confirmate.
package main

import (
	"confirmate.io/core/util/stylecheck"

	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(stylecheck.Analyzer)
}
