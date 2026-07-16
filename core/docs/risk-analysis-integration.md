# Risk Analysis Integration Design

## 1. Overview

This document outlines the design for integrating risk-based control scoping into Confirmate, inspired by the BSI TR-03183-1 (Cloud Computing Compliance Criteria Catalogue - C5) and the CRA (Cyber Resilience Act) approach. The goal is to allow users to define risk levels for their cloud services and automatically select appropriate controls from a catalog based on those risk levels.

## 2. Background

### 2.1 BSI TR-03183-1 / CRA Scoping

The BSI Technical Guideline TR-03183-1 defines different assurance levels (Basic, Substantial, High) for cloud services. Each level requires different controls to be implemented, with higher assurance levels requiring more stringent security measures. The CRA (Cyber Resilience Act) extends this concept by tying control requirements to risk levels.

Key concepts:
- **Risk Level**: Classification of the cloud service's risk profile (e.g., Low, Medium, High, Critical)
- **Assurance Level**: The required level of security assurance (e.g., Basic, Substantial, High)
- **Control Mapping**: Controls are associated with specific risk/assurance level thresholds

### 2.2 Current State in Confirmate

Currently, Confirmate supports:
- **Catalogs**: Contain controls organized by categories
- **Controls**: Have optional `assurance_level` field
- **AuditScopes**: Link a TargetOfEvaluation to a Catalog with optional `assurance_level`
- **Filtering**: Controls can be filtered by assurance levels when listing

Missing pieces:
- No explicit risk level model
- No mapping between risk levels and controls
- No automatic control selection based on risk analysis

## 3. Requirements

### 3.1 Core Requirements

1. **Risk Level Management**
   - Define risk levels (e.g., Low, Medium, High, Critical)
   - Associate risk levels with TargetOfEvaluation
   - Support custom risk level definitions per organization

2. **Control-Risk Mapping**
   - Define which controls apply to which risk levels
   - Support threshold-based mapping (e.g., "all controls up to risk level X")
   - Allow manual override of automatic selection

3. **Automatic Control Selection**
   - When creating an AuditScope, automatically select controls based on risk level
   - Provide UI/API to review and modify selected controls
   - Support incremental updates when risk level changes

### 3.2 API Requirements

1. CRUD operations for RiskLevels
2. CRUD operations for Control-Risk mappings (at catalog level)
3. New endpoint to generate AuditScope with controls based on risk level
4. Endpoints to update risk level for a TargetOfEvaluation

## 4. Data Model Design

### 4.1 New Messages

```protobuf
// RiskLevel represents a risk classification
message RiskLevel {
  string id = 1 [
    (tagger.tags) = "gorm:\"primaryKey\"",
    (buf.validate.field).string.uuid = true,
    (google.api.field_behavior) = REQUIRED
  ];
  string name = 2 [  // e.g., "Low", "Medium", "High", "Critical"
    (buf.validate.field).string.min_len = 1,
    (google.api.field_behavior) = REQUIRED
  ];
  string description = 3;
  int32 priority = 4;  // Higher = more severe, used for threshold-based selection
  optional string catalog_id = 5;  // Optional: risk level specific to a catalog
}

// ControlRiskMapping defines which controls apply to which risk levels
message ControlRiskMapping {
  string id = 1 [
    (tagger.tags) = "gorm:\"primaryKey\"",
    (buf.validate.field).string.uuid = true,
    (google.api.field_behavior) = REQUIRED
  ];
  string control_id = 2 [
    (buf.validate.field).string.min_len = 1,
    (google.api.field_behavior) = REQUIRED
  ];
  string category_name = 3 [
    (buf.validate.field).string.min_len = 1,
    (google.api.field_behavior) = REQUIRED
  ];
  string category_catalog_id = 4 [
    (buf.validate.field).string.min_len = 1,
    (google.api.field_behavior) = REQUIRED
  ];
  string risk_level_id = 5 [
    (buf.validate.field).string.uuid = true,
    (google.api.field_behavior) = REQUIRED
  ];
  // Threshold: if true, this control applies to this risk level AND all higher levels
  bool threshold = 6;
}

// RiskAssessment ties a risk level to a TargetOfEvaluation
message RiskAssessment {
  string id = 1 [
    (tagger.tags) = "gorm:\"primaryKey\"",
    (buf.validate.field).string.uuid = true,
    (google.api.field_behavior) = REQUIRED
  ];
  string target_of_evaluation_id = 2 [
    (buf.validate.field).string.uuid = true,
    (google.api.field_behavior) = REQUIRED
  ];
  string risk_level_id = 3 [
    (buf.validate.field).string.uuid = true,
    (google.api.field_behavior) = REQUIRED
  ];
  string justification = 4;  // Documentation of risk assessment reasoning
  google.protobuf.Timestamp assessed_at = 5 [(tagger.tags) = "gorm:\"serializer:timestamppb;type:timestamp\""];
}
```

### 4.2 Extended Existing Messages

```protobuf
// Extend TargetOfEvaluation with risk assessment reference
message TargetOfEvaluation {
  // ... existing fields ...
  optional string current_risk_assessment_id = 16;
}

// Extend AuditScope to track risk-based generation
message AuditScope {
  // ... existing fields ...
  optional string risk_assessment_id = 9;  // Link to the risk assessment used
  bool auto_generated = 10;  // Whether controls were auto-selected based on risk
}
```

## 5. API Endpoints Design

### 5.1 Risk Level Management

```
POST   /v1/orchestrator/risk_levels
GET    /v1/orchestrator/risk_levels
GET    /v1/orchestrator/risk_levels/{risk_level_id}
PUT    /v1/orchestrator/risk_levels/{risk_level_id}
DELETE /v1/orchestrator/risk_levels/{risk_level_id}
```

### 5.2 Control-Risk Mapping Management

```
POST   /v1/orchestrator/catalogs/{catalog_id}/control_risk_mappings
GET    /v1/orchestrator/catalogs/{catalog_id}/control_risk_mappings
PUT    /v1/orchestrator/catalogs/{catalog_id}/control_risk_mappings/{mapping_id}
DELETE /v1/orchestrator/catalogs/{catalog_id}/control_risk_mappings/{mapping_id}
```

### 5.3 Risk Assessment Management

```
POST   /v1/orchestrator/targets_of_evaluation/{target_of_evaluation_id}/risk_assessment
GET    /v1/orchestrator/targets_of_evaluation/{target_of_evaluation_id}/risk_assessment
GET    /v1/orchestrator/targets_of_evaluation/{target_of_evaluation_id}/risk_assessment/history
PUT    /v1/orchestrator/targets_of_evaluation/{target_of_evaluation_id}/risk_assessment/{assessment_id}
```

### 5.4 Audit Scope Generation

```
POST   /v1/orchestrator/targets_of_evaluation/{target_of_evaluation_id}/audit_scopes/from_risk
```

Request body:
```json
{
  "catalog_id": "cra-catalog",
  "name": "Audit Scope based on Risk Assessment",
  "override_existing": false
}
```

Response: Returns the created AuditScope with automatically selected controls

## 6. Control Selection Algorithm

When generating an AuditScope from risk assessment:

1. Retrieve the RiskAssessment for the TargetOfEvaluation
2. Get the associated RiskLevel and its priority
3. Query all ControlRiskMappings for the catalog where:
   - `risk_level_id` matches the current risk level, OR
   - `threshold` is true AND the mapping's risk level priority <= current priority
4. Create the AuditScope with the selected control IDs
5. Store reference to the RiskAssessment used

## 7. Database Schema Changes

New tables required:

| Table | Description |
|-------|-------------|
| risk_levels | Store risk level definitions |
| control_risk_mappings | Map controls to risk levels |
| risk_assessments | Track risk assessments per ToE |

## 8. UI/UX Considerations

### 8.1 Risk Assessment Flow

1. User selects a TargetOfEvaluation
2. User navigates to "Risk Assessment" section
3. User selects or creates a RiskLevel
4. User provides justification for the risk level
5. System saves the RiskAssessment

### 8.2 Audit Scope Creation Flow

1. User initiates "Create Audit Scope"
2. User selects TargetOfEvaluation and Catalog
3. If a RiskAssessment exists, system shows:
   - Current risk level
   - Estimated number of controls that will be selected
   - Option to proceed with risk-based selection or manual selection
4. User confirms and AuditScope is created with appropriate controls
5. User can review and manually add/remove controls

### 8.3 Catalog Management

1. Admin users can manage ControlRiskMappings
2. UI should show controls grouped by risk level
3. Support bulk import/export of mappings (CSV/JSON)

## 9. Integration with QuBA

The design allows for future integration with QuBA-libre:
- RiskAssessment can store the raw assessment data
- The justification field can link to external risk analysis documents
- The CyberSecurityRiskAssessmentDocument ontology type can be referenced

## 10. Backwards Compatibility

- Existing AuditScopes continue to work unchanged
- Risk level is optional - users can still create scopes manually
- Control filtering by assurance level remains available
- No changes to existing API responses (new fields are optional)

## 11. Implementation Roadmap

### Phase 1: Core Data Models
- Add RiskLevel, ControlRiskMapping, RiskAssessment messages
- Create database migrations
- Implement basic CRUD APIs

### Phase 2: Audit Scope Integration
- Extend TargetOfEvaluation with risk assessment reference
- Implement "generate from risk" endpoint
- Add control selection algorithm

### Phase 3: UI Implementation
- Risk assessment management UI
- Audit scope creation wizard with risk-based selection
- Control-Risk mapping management UI

### Phase 4: Advanced Features
- Bulk import/export of mappings
- Risk assessment history and audit trail
- Integration with QuBA-libre

## 12. Open Questions

1. Should risk levels be global or catalog-specific?
   - Recommendation: Allow catalog-specific risk levels with fallback to global

2. How to handle controls that don't have risk mappings?
   - Recommendation: Require explicit mapping; un-mapped controls are not auto-selected

3. Should risk assessment be required before creating an AuditScope?
   - Recommendation: No - make it optional to maintain backwards compatibility

4. How to handle changes in risk level after AuditScope is created?
   - Recommendation: Show warning when opening scope, offer re-generation option

## 13. References

- BSI TR-03183-1: Cloud Computing Compliance Criteria Catalogue (C5)
- EU Cyber Resilience Act (CRA)
- QuBA-libre: https://github.com/SICKAG/QuBA-libre