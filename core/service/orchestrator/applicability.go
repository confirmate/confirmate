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
	"fmt"
	"slices"

	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/persistence"
	"confirmate.io/core/service"
)

// scenarioReached tells whether the risk profile reaches the scenario of a requirement, so that the requirement
// applies. If the profile does not say something that the scenario depends on, the requirement applies: it is only
// left out if the profile shows that the risk is lower.
func scenarioReached(profile *orchestrator.RiskProfile, scenario *orchestrator.RiskScenario) bool {
	if scenario == nil {
		return true
	}

	if !envReached(profile.GetPhysicalAccess(), scenario.GetPhysicalAccess()) ||
		!envReached(profile.GetConnectivity(), scenario.GetConnectivity()) ||
		!envReached(profile.GetOperatorSkill(), scenario.GetOperatorSkill()) {
		return false
	}

	// One asset has to reach all the impacts of the scenario at the same time
	if scenario.GetMinConfidentiality() == 0 && scenario.GetMinIntegrity() == 0 && scenario.GetMinAvailability() == 0 {
		return true
	}
	if len(profile.GetAssets()) == 0 {
		return true
	}
	return slices.ContainsFunc(profile.GetAssets(), func(a *orchestrator.AssetImpact) bool {
		return a.GetConfidentiality() >= scenario.GetMinConfidentiality() &&
			a.GetIntegrity() >= scenario.GetMinIntegrity() &&
			a.GetAvailability() >= scenario.GetMinAvailability()
	})
}

// envReached tells whether the value of the profile is one of the values of the scenario. An empty list of values
// allows any value, and a value that is not set in the profile does not exclude anything.
func envReached[T comparable](value T, allowed []T) bool {
	var zero T
	return len(allowed) == 0 || value == zero || slices.Contains(allowed, value)
}

// deriveScopeDecisions derives from a risk profile which requirements of the catalog of the audit scope are out of
// scope. Only risk based requirements are decided. Conditional requirements stay with the person who knows whether the
// condition holds. The last requirement of a control is never taken out, so that a control cannot drop out as a whole.
func (svc *Service) deriveScopeDecisions(scope *orchestrator.AuditScope, profile *orchestrator.RiskProfile) (decisions []*orchestrator.ScopeDecision, err error) {
	var (
		controls []*orchestrator.Control
		// The IDs of the requirements that stay, by the ID of their parent control
		staying = make(map[string]int)
		below   []*orchestrator.Control
	)

	err = svc.db.List(&controls, "", true, 0, -1, persistence.WithoutPreload(), "catalog_id = ?", scope.CatalogId)
	if err = service.HandleDatabaseError(err); err != nil {
		return nil, err
	}

	for _, ctrl := range controls {
		if ctrl.ParentControlId == nil {
			continue
		}
		if ctrl.GetApplicability().GetKind() == orchestrator.ApplicabilityKind_APPLICABILITY_KIND_RISK_BASED &&
			!scenarioReached(profile, ctrl.GetApplicability().GetScenario()) {
			below = append(below, ctrl)
			continue
		}
		staying[*ctrl.ParentControlId]++
	}

	for _, ctrl := range below {
		if staying[*ctrl.ParentControlId] == 0 {
			continue
		}
		decisions = append(decisions, &orchestrator.ScopeDecision{
			ControlId:     ctrl.Id,
			InScope:       false,
			Reason:        orchestrator.NotApplicableReason_NOT_APPLICABLE_REASON_BELOW_RISK_LEVEL,
			Justification: fmt.Sprintf("The risk profile is below the level at which %s applies.", ctrl.GetShortName()),
			Source:        orchestrator.ScopeDecisionSource_SCOPE_DECISION_SOURCE_RULE,
		})
	}

	return decisions, nil
}
