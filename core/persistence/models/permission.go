package models

import "time"

type PermissionRole string

const (
	RoleOwner                    PermissionRole = "OWNER"
	RoleContributor              PermissionRole = "CONTRIBUTOR"
	RoleReader                   PermissionRole = "READER"
	ResourceTypeToE              string         = "TOE"
	ResourceTypeCatalog          string         = "CATALOG"
	ResourceTypeAuditScope       string         = "AUDIT_SCOPE"
	ResourceTypeCertificate      string         = "CERTIFICATE"
	ResourceTypeAssessmentResult string         = "ASSESSMENT_RESULT"
	ResourceTypeEvidence         string         = "EVIDENCE"
)

// PersmissionBase contains the common fields for all permission models join tables.
type PermissionBase struct {
	UserID       string         `gorm:"column:user_id;primaryKey;not null"`
	ResourceID   string         `gorm:"column:resource_id;primaryKey;not null"`
	ResourceType string         `gorm:"column:resource_type;primaryKey;type:text;not null"`
	Role         PermissionRole `gorm:"column:role;type:text;not null"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// type ToEPermission struct {
// 	PermissionBase
// 	ToEID string `gorm:"column:toe_id;primaryKey;not null;index"`
// }

// func (ToEPermission) TableName() string { return "toe_permissions" }

// type CatalogPermission struct {
// 	PermissionBase
// 	CatalogID string `gorm:"column:catalog_id;primaryKey;not null;index"`
// }

// func (CatalogPermission) TableName() string { return "catalog_permissions" }

// type AuditScopePermission struct {
// 	PermissionBase
// 	AuditScopeID string `gorm:"column:audit_scope_id;primaryKey;not null;index"`
// }

// func (AuditScopePermission) TableName() string { return "audit_scope_permissions" }

// type CertificatePermission struct {
// 	PermissionBase
// 	CertificateID string `gorm:"column:certificate_id;primaryKey;not null;index"`
// }

// func (CertificatePermission) TableName() string { return "certificate_permissions" }

// type AssessmentResultPermission struct {
// 	PermissionBase
// 	AssessmentResultID string `gorm:"column:assessment_result_id;primaryKey;not null;index"`
// }

// func (AssessmentResultPermission) TableName() string { return "assessment_result_permissions" }

// type EvidencePermission struct {
// 	PermissionBase
// 	EvidenceID string `gorm:"column:evidence_id;primaryKey;not null;index"`
// }

// func (EvidencePermission) TableName() string { return "evidence_permissions" }
