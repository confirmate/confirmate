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
	"testing"

	"confirmate.io/core/api/assessment"
	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/persistence"
	"confirmate.io/core/persistence/persistencetest"
	"confirmate.io/core/service/orchestrator/orchestratortest"
	"confirmate.io/core/util/assert"
)

func TestService_loadControlTrees(t *testing.T) {
	var (
		// root -> {leafA, sub -> {leafB}}; leafA sorts before sub by its short name
		root = &orchestrator.Control{
			Id:        orchestratortest.MockControlId1,
			ShortName: "R",
			CatalogId: orchestratortest.MockCatalogId1,
		}
		sub = &orchestrator.Control{
			Id:              orchestratortest.MockControl1SubControlId1,
			ShortName:       "R.2",
			CatalogId:       orchestratortest.MockCatalogId1,
			ParentControlId: new(orchestratortest.MockControlId1),
		}
		leafA = &orchestrator.Control{
			Id:              orchestratortest.MockControl2SubControlId1,
			ShortName:       "R.1",
			CatalogId:       orchestratortest.MockCatalogId1,
			ParentControlId: new(orchestratortest.MockControlId1),
		}
		leafB = &orchestrator.Control{
			Id:              orchestratortest.MockControl1SubControlId2,
			ShortName:       "R.2.1",
			CatalogId:       orchestratortest.MockCatalogId1,
			ParentControlId: new(orchestratortest.MockControl1SubControlId1),
			Metrics:         []*assessment.Metric{orchestratortest.MockMetric1},
		}
	)

	type args struct {
		roots           []*orchestrator.Control
		withLeafMetrics bool
	}
	tests := []struct {
		name    string
		args    args
		want    assert.Want[[]*orchestrator.Control]
		wantErr assert.WantErr
	}{
		{
			name:    "no controls",
			args:    args{},
			want:    assert.Empty[[]*orchestrator.Control],
			wantErr: assert.NoError,
		},
		{
			name: "loads all levels ordered by short name, metrics of leaves only on request",
			args: args{
				roots:           []*orchestrator.Control{{Id: root.Id}},
				withLeafMetrics: false,
			},
			want: func(t *testing.T, got []*orchestrator.Control, args ...any) bool {
				r := got[0]
				return assert.Equal(t, 2, len(r.Controls)) &&
					assert.Equal(t, leafA.Id, r.Controls[0].Id) &&
					assert.Equal(t, sub.Id, r.Controls[1].Id) &&
					assert.Equal(t, 1, len(r.Controls[1].Controls)) &&
					assert.Equal(t, leafB.Id, r.Controls[1].Controls[0].Id) &&
					assert.Equal(t, 0, len(r.Controls[1].Controls[0].Metrics))
			},
			wantErr: assert.NoError,
		},
		{
			name: "with leaf metrics",
			args: args{
				roots:           []*orchestrator.Control{{Id: root.Id}},
				withLeafMetrics: true,
			},
			want: func(t *testing.T, got []*orchestrator.Control, args ...any) bool {
				r := got[0]
				return assert.Equal(t, 0, len(r.Controls[0].Metrics)) && // leafA has no metrics
					assert.Equal(t, 1, len(r.Controls[1].Controls[0].Metrics)) &&
					assert.Equal(t, orchestratortest.MockMetricId1, r.Controls[1].Controls[0].Metrics[0].Id)
			},
			wantErr: assert.NoError,
		},
		{
			name: "a control without children is a leaf",
			args: args{
				roots:           []*orchestrator.Control{{Id: leafB.Id}},
				withLeafMetrics: true,
			},
			want: func(t *testing.T, got []*orchestrator.Control, args ...any) bool {
				return assert.Equal(t, 0, len(got[0].Controls)) &&
					assert.Equal(t, 1, len(got[0].Metrics))
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					assert.NoError(t, d.Create(&orchestrator.Catalog{Id: orchestratortest.MockCatalogId1, Name: orchestratortest.MockCatalog1.Name}))
					assert.NoError(t, d.Create(orchestratortest.MockMetric1))
					assert.NoError(t, d.Create(root))
					assert.NoError(t, d.Create(sub))
					assert.NoError(t, d.Create(leafA))
					assert.NoError(t, d.Create(leafB))
				}),
			}

			err := svc.loadControlTrees(tt.args.roots, tt.args.withLeafMetrics)

			tt.wantErr(t, err)
			tt.want(t, tt.args.roots)
		})
	}
}
