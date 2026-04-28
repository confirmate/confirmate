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

import "slices"

// IsRelevantFor checks if the control is relevant for the given audit scope and catalog. This is determined by comparing the assurance levels of the control and the audit scope against the assurance levels defined in the catalog. If the control's assurance level is less than or equal to the audit scope's assurance level, then the control is considered relevant. In the future, this could also include checks, if the control is somehow out of scope.
func (c *Control) IsRelevantFor(auditScope *AuditScope, catalog *Catalog) bool {
	// If the catalog does not have an assurance level, we are good to go
	if len(catalog.AssuranceLevels) == 0 {
		return true
	}

	// If the control or the audit scope does not have an assurance level, we are good to go
	if c.AssuranceLevel == nil || auditScope.AssuranceLevel == nil {
		return true
	}

	// Otherwise, we need to retrieve the possible assurance levels (in order) from the catalogs and compare the
	// indices
	idxControl := slices.Index(catalog.AssuranceLevels, *c.AssuranceLevel)
	idxAuditScope := slices.Index(catalog.AssuranceLevels, *auditScope.AssuranceLevel)

	return idxControl <= idxAuditScope
}
