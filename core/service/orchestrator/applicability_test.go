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

	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/persistence"
	"confirmate.io/core/persistence/persistencetest"
	"confirmate.io/core/service/orchestrator/orchestratortest"
	"confirmate.io/core/util/assert"
)

// riskScenarioHigh is the scenario of a requirement that applies to assets with a high impact on confidentiality that
// can be reached from the internet.
var riskScenarioHigh = &orchestrator.RiskScenario{
	MinConfidentiality: 4,
	Connectivity:       []orchestrator.Connectivity{orchestrator.Connectivity_CONNECTIVITY_INTERNET},
}

func Test_scenarioReached(t *testing.T) {
	type args struct {
		profile  *orchestrator.RiskProfile
		scenario *orchestrator.RiskScenario
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "no scenario",
			args: args{profile: &orchestrator.RiskProfile{}},
			want: true,
		},
		{
			name: "impact and environment reached",
			args: args{
				profile: &orchestrator.RiskProfile{
					Assets:       []*orchestrator.AssetImpact{{Name: "Credentials", Confidentiality: 5}},
					Connectivity: orchestrator.Connectivity_CONNECTIVITY_INTERNET,
				},
				scenario: riskScenarioHigh,
			},
			want: true,
		},
		{
			name: "impact too low",
			args: args{
				profile: &orchestrator.RiskProfile{
					Assets:       []*orchestrator.AssetImpact{{Name: "Settings", Confidentiality: 2}},
					Connectivity: orchestrator.Connectivity_CONNECTIVITY_INTERNET,
				},
				scenario: riskScenarioHigh,
			},
			want: false,
		},
		{
			name: "environment not reached",
			args: args{
				profile: &orchestrator.RiskProfile{
					Assets:       []*orchestrator.AssetImpact{{Name: "Credentials", Confidentiality: 5}},
					Connectivity: orchestrator.Connectivity_CONNECTIVITY_LOCAL,
				},
				scenario: riskScenarioHigh,
			},
			want: false,
		},
		{
			name: "the impacts have to be reached by one asset",
			args: args{
				profile: &orchestrator.RiskProfile{
					Assets: []*orchestrator.AssetImpact{
						{Name: "Credentials", Confidentiality: 5},
						{Name: "Firmware", Integrity: 5},
					},
				},
				scenario: &orchestrator.RiskScenario{MinConfidentiality: 4, MinIntegrity: 4},
			},
			want: false,
		},
		{
			name: "environment not given in the profile",
			args: args{
				profile: &orchestrator.RiskProfile{
					Assets: []*orchestrator.AssetImpact{{Name: "Credentials", Confidentiality: 5}},
				},
				scenario: riskScenarioHigh,
			},
			want: true,
		},
		{
			name: "no assets in the profile",
			args: args{
				profile:  &orchestrator.RiskProfile{Connectivity: orchestrator.Connectivity_CONNECTIVITY_INTERNET},
				scenario: riskScenarioHigh,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := scenarioReached(tt.args.profile, tt.args.scenario)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestService_deriveScopeDecisions(t *testing.T) {
	type args struct {
		profile *orchestrator.RiskProfile
	}
	type fields struct {
		db persistence.DB
	}
	var (
		// seedApplicability adds a risk based sub-control to MockControlId1, next to its sub-control that always applies
		seedApplicability = func(sub2Scenario *orchestrator.RiskScenario) func(persistence.DB) {
			return func(d persistence.DB) {
				seedRiskAssessmentScope(t, d)
				assert.NoError(t, d.Create(&orchestrator.Control{
					Id:              orchestratortest.MockControl1SubControlId2,
					CatalogId:       orchestratortest.MockCatalogId1,
					Name:            orchestratortest.MockSubControlName2,
					ShortName:       orchestratortest.MockSubControlShortName2,
					ParentControlId: new(orchestratortest.MockControlId1),
					Applicability: &orchestrator.Applicability{
						Kind:     orchestrator.ApplicabilityKind_APPLICABILITY_KIND_RISK_BASED,
						Scenario: sub2Scenario,
					},
				}))
			}
		}
	)
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[[]*orchestrator.ScopeDecision]
		wantErr assert.WantErr
	}{
		{
			name: "risk below the scenario takes the requirement out",
			args: args{profile: &orchestrator.RiskProfile{
				Assets:       []*orchestrator.AssetImpact{{Name: "Settings", Confidentiality: 1}},
				Connectivity: orchestrator.Connectivity_CONNECTIVITY_LOCAL,
			}},
			fields: fields{db: persistencetest.NewInMemoryDB(t, types, joinTables, seedApplicability(riskScenarioHigh))},
			want: func(t *testing.T, got []*orchestrator.ScopeDecision, msgAndArgs ...any) bool {
				return assert.Equal(t, 1, len(got)) &&
					assert.Equal(t, orchestratortest.MockControl1SubControlId2, got[0].GetControlId()) &&
					assert.Equal(t, orchestrator.NotApplicableReason_NOT_APPLICABLE_REASON_BELOW_RISK_LEVEL, got[0].GetReason()) &&
					assert.Equal(t, orchestrator.ScopeDecisionSource_SCOPE_DECISION_SOURCE_RULE, got[0].GetSource())
			},
			wantErr: assert.NoError,
		},
		{
			name: "risk reaching the scenario keeps the requirement",
			args: args{profile: &orchestrator.RiskProfile{
				Assets:       []*orchestrator.AssetImpact{{Name: "Credentials", Confidentiality: 5}},
				Connectivity: orchestrator.Connectivity_CONNECTIVITY_INTERNET,
			}},
			fields: fields{db: persistencetest.NewInMemoryDB(t, types, joinTables, seedApplicability(riskScenarioHigh))},
			want: func(t *testing.T, got []*orchestrator.ScopeDecision, msgAndArgs ...any) bool {
				return assert.Equal(t, 0, len(got))
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{db: tt.fields.db}

			got, err := svc.deriveScopeDecisions(orchestratortest.MockAuditScope1, tt.args.profile)

			tt.wantErr(t, err)
			tt.want(t, got)
		})
	}
}
