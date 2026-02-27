package orchestrator

import "slices"

// IsRelevantFor checks, whether this control is relevant for the given audit scope. For now this mainly
// checks, whether the assurance level matches, if the Audit Scope has one. In the future, this could also include checks, if
// the control is somehow out of scope.
func (c *Control) IsRelevantFor(auditScope *AuditScope, catalog *Catalog) bool {
	// If the catalog does not have an assurance level, we are good to go
	if len(catalog.AssuranceLevels) == 0 {
		return true
	}

	// If the control does not explicitly specify an assurance level, we are also ok
	if c.AssuranceLevel == nil || auditScope.AssuranceLevel == nil {
		return true
	}

	// Otherwise, we need to retrieve the possible assurance levels (in order) from the catalogs and compare the
	// indices
	idxControl := slices.Index(catalog.AssuranceLevels, *c.AssuranceLevel)
	idxToe := slices.Index(catalog.AssuranceLevels, *auditScope.AssuranceLevel)

	return idxControl <= idxToe
}
