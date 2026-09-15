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
	"bytes"
	"context"
	"testing"
	"time"

	"confirmate.io/core/api/assessment"
	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/persistence"
	"confirmate.io/core/persistence/persistencetest"
	"confirmate.io/core/service"
	"confirmate.io/core/util/assert"

	"connectrpc.com/connect"
	pdfapi "github.com/pdfcpu/pdfcpu/pkg/api"
	pdfmodel "github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/xuri/excelize/v2"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestService_ExportAuditScopeReport(t *testing.T) {
	const (
		toeId      = "10000000-0000-0000-0000-000000000001"
		scopeId    = "20000000-0000-0000-0000-000000000001"
		catalogId  = "report-test-catalog"
		control1Id = "30000000-0000-0000-0000-000000000001"
		control2Id = "30000000-0000-0000-0000-000000000002"
		metric1Id  = "40000000-0000-0000-0000-000000000001"
		metric2Id  = "40000000-0000-0000-0000-000000000002"
		metric3Id  = "40000000-0000-0000-0000-000000000003"
		userId     = "50000000-0000-0000-0000-000000000001"
	)

	catalog := &orchestrator.Catalog{
		Id:   catalogId,
		Name: "Report Test Catalog",
		Categories: []*orchestrator.Category{
			{
				Name:      "Access Control",
				CatalogId: catalogId,
				Controls: []*orchestrator.Control{
					{
						Id: control1Id, Name: "Multi-Factor Authentication", ShortName: "AC-1", CatalogId: catalogId,
						Metrics: []*assessment.Metric{{Id: metric1Id, Name: "MFAEnforcedForAdmin"}},
					},
				},
			},
			{
				Name:      "Data Protection",
				CatalogId: catalogId,
				Controls: []*orchestrator.Control{
					{
						Id: control2Id, Name: "Encryption at Rest", ShortName: "DP-1", CatalogId: catalogId,
						Metrics: []*assessment.Metric{
							{Id: metric2Id, Name: "DiskEncryptionEnabled"},
							{Id: metric3Id, Name: "KeyRotationEnabled"},
						},
					},
				},
			},
		},
	}
	toe := &orchestrator.TargetOfEvaluation{Id: toeId, Name: "Report Test ToE"}
	auditScope := &orchestrator.AuditScope{Id: scopeId, Name: "Mock Audit Scope 1", TargetOfEvaluationId: toeId, CatalogId: catalogId}
	user := &orchestrator.User{Id: userId, FirstName: new("Test"), LastName: new("User")}

	evalTime := time.Date(2026, 1, 15, 10, 30, 0, 0, time.UTC)

	db := persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
		assert.NoError(t, d.Create(catalog))
		assert.NoError(t, d.Create(toe))
		assert.NoError(t, d.Create(auditScope))
		assert.NoError(t, d.Create(user))

		// Control 1 (category "Access Control"): open, no assignee. Its one metric is never
		// assessed, so it should show up as "Not Evaluated".
		assert.NoError(t, d.Create(&orchestrator.ControlInScope{
			Id:                   "60000000-0000-0000-0000-000000000001",
			AuditScopeId:         scopeId,
			TargetOfEvaluationId: toeId,
			ControlId:            control1Id,
			State:                orchestrator.ControlInScopeState_CONTROL_IN_SCOPE_STATE_OPEN,
		}))

		// Control 2 (category "Data Protection"): accepted, assigned. One metric compliant, one
		// not compliant, so the control should show up as "Action Required".
		assert.NoError(t, d.Create(&orchestrator.ControlInScope{
			Id:                    "60000000-0000-0000-0000-000000000002",
			AuditScopeId:          scopeId,
			TargetOfEvaluationId:  toeId,
			ControlId:             control2Id,
			State:                 orchestrator.ControlInScopeState_CONTROL_IN_SCOPE_STATE_ACCEPTED,
			AssigneeId:            new(userId),
			ImplementationDetails: new("Rolled out via IaC pipeline."),
		}))

		assert.NoError(t, d.Create(&assessment.AssessmentResult{
			Id:                   "70000000-0000-0000-0000-000000000001",
			CreatedAt:            timestamppb.New(evalTime),
			MetricId:             metric2Id,
			TargetOfEvaluationId: toeId,
			Compliant:            true,
			EvidenceId:           "80000000-0000-0000-0000-000000000001",
			ResourceId:           "disk-1",
			ResourceTypes:        []string{"BlockStorage"},
			ComplianceComment:    "Resource is compliant",
			ComplianceDetails: []*assessment.ComparisonResult{{
				Property:    "encryptionEnabled",
				Operator:    "==",
				TargetValue: structpb.NewBoolValue(true),
				Value:       structpb.NewBoolValue(true),
			}},
		}))
		assert.NoError(t, d.Create(&assessment.AssessmentResult{
			Id:                   "70000000-0000-0000-0000-000000000002",
			CreatedAt:            timestamppb.New(evalTime),
			MetricId:             metric3Id,
			TargetOfEvaluationId: toeId,
			Compliant:            false,
			EvidenceId:           "80000000-0000-0000-0000-000000000002",
			ResourceId:           "disk-1",
			ResourceTypes:        []string{"BlockStorage"},
			ComplianceComment:    "Key rotation is disabled, enable automatic rotation.",
			ComplianceDetails: []*assessment.ComparisonResult{{
				Property:    "keyRotationEnabled",
				Operator:    "==",
				TargetValue: structpb.NewBoolValue(true),
				Value:       structpb.NewBoolValue(false),
			}},
		}))
	})

	svc := &Service{
		db:    db,
		authz: &service.AuthorizationStrategyAllowAll{},
	}

	t.Run("happy path", func(t *testing.T) {
		res, err := svc.ExportAuditScopeReport(context.Background(), connect.NewRequest(&orchestrator.ExportAuditScopeReportRequest{
			AuditScopeId: scopeId,
		}))
		assert.NoError(t, err)
		if !assert.NotNil(t, res) {
			return
		}
		assert.NotEmpty(t, res.Msg.GetContent())
		assert.Equal(t, "audit-scope-report-Mock-Audit-Scope-1-"+time.Now().Format("20060102")+".xlsx",
			res.Msg.GetFilename())

		f, err := excelize.OpenReader(bytes.NewReader(res.Msg.GetContent()))
		if !assert.NoError(t, err) {
			return
		}
		defer f.Close()

		rows, err := f.GetRows("Report")
		assert.NoError(t, err)

		// Row 1-3 are the title block, row 5 is the header. Control 1 has 1 metric row, control 2
		// has 2 metric rows: 3 data rows total, sorted by category then control short name.
		if !assert.Equal(t, 8, len(rows)) {
			return
		}

		header := rows[4]
		assert.Equal(t, []string{
			"Category", "Control ID", "Control Name", "Implementation State", "Assignee", "Implementation Notes",
			"Metric", "Target Component", "Evaluated Condition & Observed Value", "Metric Status", "Compliance Comment",
		}, header)

		control1MetricRow := rows[5]
		assert.Equal(t, "Access Control", control1MetricRow[0])
		assert.Equal(t, "AC-1", control1MetricRow[1])
		assert.Equal(t, "Open", control1MetricRow[3])
		assert.Equal(t, "MFAEnforcedForAdmin", control1MetricRow[6])
		assert.Equal(t, "Not Evaluated", control1MetricRow[9])

		control2CompliantRow := rows[6]
		assert.Equal(t, "Data Protection", control2CompliantRow[0])
		assert.Equal(t, "DP-1", control2CompliantRow[1])
		assert.Equal(t, "Accepted", control2CompliantRow[3])
		assert.Equal(t, "Test User", control2CompliantRow[4])
		assert.Equal(t, "Rolled out via IaC pipeline.", control2CompliantRow[5])
		assert.Equal(t, "DiskEncryptionEnabled", control2CompliantRow[6])
		assert.Equal(t, "BlockStorage", control2CompliantRow[7])
		assert.Equal(t, "encryptionEnabled == true (Observed: true)", control2CompliantRow[8])
		assert.Equal(t, "Compliant", control2CompliantRow[9])

		control2NonCompliantRow := rows[7]
		assert.Equal(t, "KeyRotationEnabled", control2NonCompliantRow[6])
		assert.Equal(t, "Not Compliant", control2NonCompliantRow[9])
		assert.Equal(t, "Key rotation is disabled, enable automatic rotation.", control2NonCompliantRow[10])
	})

	t.Run("happy path: PDF format", func(t *testing.T) {
		res, err := svc.ExportAuditScopeReport(context.Background(), connect.NewRequest(&orchestrator.ExportAuditScopeReportRequest{
			AuditScopeId: scopeId,
			Format:       orchestrator.ReportFormat_REPORT_FORMAT_PDF,
		}))
		assert.NoError(t, err)
		if !assert.NotNil(t, res) {
			return
		}
		assert.NotEmpty(t, res.Msg.GetContent())
		assert.Equal(t, "audit-scope-report-Mock-Audit-Scope-1-"+time.Now().Format("20060102")+".pdf",
			res.Msg.GetFilename())

		content := res.Msg.GetContent()
		assert.True(t, bytes.HasPrefix(content, []byte("%PDF-")), "content does not start with a PDF header")

		r := bytes.NewReader(content)
		assert.NoError(t, pdfapi.Validate(r, pdfmodel.NewDefaultConfiguration()))

		_, err = r.Seek(0, 0)
		assert.NoError(t, err)
		pages, err := pdfapi.PageCount(r, pdfmodel.NewDefaultConfiguration())
		assert.NoError(t, err)
		assert.Equal(t, 1, pages)
	})

	t.Run("err: audit scope not found", func(t *testing.T) {
		_, err := svc.ExportAuditScopeReport(context.Background(), connect.NewRequest(&orchestrator.ExportAuditScopeReportRequest{
			AuditScopeId: "00000000-0000-0000-0000-000000000000",
		}))
		assert.IsConnectError(t, err, connect.CodeNotFound)
	})

	t.Run("err: permission denied", func(t *testing.T) {
		denySvc := &Service{
			db:    db,
			authz: &denyAuthorizationStrategy{},
		}
		_, err := denySvc.ExportAuditScopeReport(context.Background(), connect.NewRequest(&orchestrator.ExportAuditScopeReportRequest{
			AuditScopeId: scopeId,
		}))
		assert.IsConnectError(t, err, connect.CodePermissionDenied)
	})
}
