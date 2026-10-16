// Copyright 2016-2026 Fraunhofer AISEC
//
// SPDX-License-Identifier: Apache-2.0
//
//                                 /$$$$$$  /$$                                     /$$
//                               /$$__  $$|__/                                    | $$
//   /$$$$$$$  /$$$$$$  /$$$$$$$ | $$  \__/ /$$  /$$$$$$  /$$$$$$/$$$$   /$$$$$$  /$$$$$$    /$$$$$$
//  /$$_____/ /$$__  $$| $$__  $$| $$$$    | $$ /$$__  $$| $$_  $$_  $$ |____  $$|_  $$_/   /$$__  $$
// | $$      | $$  \ $$| $$  \ $$| $$_/    | $$| $$  \__/| $$ \ $$ \ $$  /$$$$$$$  | $$    | $$$$$$$$
// | $$      | $$  | $$| $$  | $$| $$      | $$| $$      | $$ | $$ | $$ /$$__  $$  | $$ /$$| $$_____/
// |  $$$$$$$|  $$$$$$/| $$  | $$| $$      | $$| $$      | $$ | $$ | $$|  $$$$$$$  |  $$$$/|  $$$$$$$
// \_______/ \______/ |__/  |__/|__/      |__/|__/      |__/ |__/ |__/ \_______/   \___/   \_______/
//
// This file is part of Confirmate Core.

package orchestrator

import (
	"os"
	"testing"

	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/util/assert"
)

// The paths are relative to the root of the core module, which the tests of this package change to.
const (
	exampleCatalogsPath = "../example-data/catalogs"
	exampleMetricsPath  = "../example-data/metrics"
	securityMetricsPath = "./policies/security-metrics/metrics"
	craCatalogId        = "CRA_Catalog"
)

// loadCraCatalog loads the CRA catalog from the example data.
func loadCraCatalog(t *testing.T) (catalog *orchestrator.Catalog) {
	t.Helper()

	svc := &Service{}
	catalogs, err := svc.loadCatalogsFromFolder(exampleCatalogsPath)
	assert.NoError(t, err)

	for _, c := range catalogs {
		if c.GetId() == craCatalogId {
			return c
		}
	}

	t.Fatalf("catalog %q not found in %s", craCatalogId, exampleCatalogsPath)

	return nil
}

// leafControls returns all controls without sub-controls of the given controls.
func leafControls(controls []*orchestrator.Control) (leaves []*orchestrator.Control) {
	for _, c := range controls {
		if len(c.GetControls()) == 0 {
			leaves = append(leaves, c)
			continue
		}

		leaves = append(leaves, leafControls(c.GetControls())...)
	}

	return leaves
}

// allControls returns all controls of the given controls, including all sub-controls.
func allControls(controls []*orchestrator.Control) (all []*orchestrator.Control) {
	for _, c := range controls {
		all = append(all, c)
		all = append(all, allControls(c.GetControls())...)
	}

	return all
}

// TestExampleData_CraCatalog checks the consistency of the CRA catalog and the metrics it refers to. A control that
// refers to a metric that nobody defines can never be assessed, so this is verified here and not at runtime.
func TestExampleData_CraCatalog(t *testing.T) {
	if _, err := os.Stat(securityMetricsPath); err != nil {
		t.Skip("security-metrics submodule is not initialized")
	}

	var (
		catalog  = loadCraCatalog(t)
		controls []*orchestrator.Control
		defined  = map[string]struct{}{}
	)

	for _, cat := range catalog.GetCategories() {
		controls = append(controls, cat.GetControls()...)
	}

	// Metrics that are defined: the example metrics and the metrics of the metrics repository
	for _, path := range []string{exampleMetricsPath, securityMetricsPath} {
		svc := &Service{cfg: Config{DefaultMetricsPath: path}}
		metrics, err := svc.loadMetricsFromRepository()
		assert.NoError(t, err)

		for _, m := range metrics {
			defined[m.GetId()] = struct{}{}
		}
	}

	tests := []struct {
		name string
		want assert.Want[*orchestrator.Catalog]
	}{
		{
			name: "control IDs are unique",
			want: func(t *testing.T, _ *orchestrator.Catalog, msgAndArgs ...any) bool {
				seen := map[string]struct{}{}
				for _, c := range allControls(controls) {
					if _, dup := seen[c.GetId()]; dup {
						return assert.Fail(t, "duplicate control id", c.GetId())
					}
					seen[c.GetId()] = struct{}{}
				}

				return true
			},
		},
		{
			name: "every leaf control has at least one metric and parent controls have none",
			want: func(t *testing.T, _ *orchestrator.Catalog, msgAndArgs ...any) bool {
				ok := true
				for _, c := range allControls(controls) {
					if len(c.GetControls()) == 0 && len(c.GetMetrics()) == 0 {
						ok = assert.Fail(t, "leaf control without metrics", c.GetId()) && ok
					}
					if len(c.GetControls()) > 0 && len(c.GetMetrics()) > 0 {
						ok = assert.Fail(t, "parent control with metrics", c.GetId()) && ok
					}
				}

				return ok
			},
		},
		{
			name: "every referenced metric is defined",
			want: func(t *testing.T, _ *orchestrator.Catalog, msgAndArgs ...any) bool {
				ok := true
				for _, c := range leafControls(controls) {
					for _, m := range c.GetMetrics() {
						if _, found := defined[m.GetId()]; !found {
							ok = assert.Fail(t, "undefined metric", c.GetId()+": "+m.GetId()) && ok
						}
					}
				}

				return ok
			},
		},
		{
			name: "every control has a name and a description",
			want: func(t *testing.T, _ *orchestrator.Catalog, msgAndArgs ...any) bool {
				ok := true
				for _, c := range allControls(controls) {
					if c.GetName() == "" || c.GetDescription() == "" {
						ok = assert.Fail(t, "control without name or description", c.GetId()) && ok
					}
				}

				return ok
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.want(t, catalog)
		})
	}
}
