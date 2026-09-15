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

	"confirmate.io/core/api/evaluation"
	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/persistence"
	"confirmate.io/core/persistence/persistencetest"
	"confirmate.io/core/service"
	"confirmate.io/core/service/orchestrator/orchestratortest"
	"confirmate.io/core/util/assert"

	"connectrpc.com/connect"
	pdfapi "github.com/pdfcpu/pdfcpu/pkg/api"
	pdfmodel "github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/xuri/excelize/v2"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestService_ExportAuditScopeReport(t *testing.T) {
	evalTime := time.Date(2026, 1, 15, 10, 30, 0, 0, time.UTC)

	db := persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
		assert.NoError(t, d.Create(orchestratortest.MockCatalog1))
		assert.NoError(t, d.Create(orchestratortest.MockTargetOfEvaluation1))
		assert.NoError(t, d.Create(orchestratortest.MockAuditScope1))
		assert.NoError(t, d.Create(orchestratortest.MockUser1))

		// Control 1 (category 1): open, no assignee, no evaluation result yet.
		assert.NoError(t, d.Create(&orchestrator.ControlInScope{
			Id:                   "00000000-0000-0000-0004-000000000101",
			AuditScopeId:         orchestratortest.MockScopeId1,
			TargetOfEvaluationId: orchestratortest.MockToeId1,
			ControlId:            orchestratortest.MockControlId1,
			State:                orchestrator.ControlInScopeState_CONTROL_IN_SCOPE_STATE_OPEN,
		}))

		// Control 2 (category 2): accepted, assigned, with a latest evaluation result.
		assert.NoError(t, d.Create(&orchestrator.ControlInScope{
			Id:                    "00000000-0000-0000-0004-000000000102",
			AuditScopeId:          orchestratortest.MockScopeId1,
			TargetOfEvaluationId:  orchestratortest.MockToeId1,
			ControlId:             orchestratortest.MockControlId2,
			State:                 orchestrator.ControlInScopeState_CONTROL_IN_SCOPE_STATE_ACCEPTED,
			AssigneeId:            new(orchestratortest.MockUser1.Id),
			ImplementationDetails: new("Rolled out via IaC pipeline."),
		}))

		assert.NoError(t, d.Create(&evaluation.EvaluationResult{
			Id:                   "00000000-0000-0000-0002-000000000101",
			TargetOfEvaluationId: orchestratortest.MockToeId1,
			AuditScopeId:         orchestratortest.MockScopeId1,
			ControlId:            orchestratortest.MockControlId2,
			ControlCatalogId:     orchestratortest.MockCatalogId1,
			Status:               evaluation.EvaluationStatus_EVALUATION_STATUS_COMPLIANT,
			Timestamp:            timestamppb.New(evalTime),
			AssessmentResultIds:  []string{"some-assessment-result-id"},
		}))
	})

	svc := &Service{
		db:    db,
		authz: &service.AuthorizationStrategyAllowAll{},
	}

	t.Run("happy path", func(t *testing.T) {
		res, err := svc.ExportAuditScopeReport(context.Background(), connect.NewRequest(&orchestrator.ExportAuditScopeReportRequest{
			AuditScopeId: orchestratortest.MockScopeId1,
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

		// Row 1-3 are the title block, row 5 is the header, data starts at row 6.
		// Control 1 (category-1) must sort before control 2 (category-2).
		if !assert.Equal(t, 7, len(rows)) {
			return
		}

		header := rows[4]
		assert.Equal(t, []string{
			"Category", "Control ID", "Control Name", "Assurance Level",
			"Implementation State", "Assignee", "Implementation Notes",
			"Evaluation Status", "Evaluation Timestamp", "Evaluation Comment",
		}, header)

		control1Row := rows[5]
		assert.Equal(t, orchestratortest.MockCategoryName1, control1Row[0])
		assert.Equal(t, orchestratortest.MockControlShortName1, control1Row[1])
		assert.Equal(t, "Open", control1Row[4])

		control2Row := rows[6]
		assert.Equal(t, orchestratortest.MockCategoryName2, control2Row[0])
		assert.Equal(t, orchestratortest.MockControlShortName2, control2Row[1])
		assert.Equal(t, "Accepted", control2Row[4])
		assert.Equal(t, "Test User", control2Row[5])
		assert.Equal(t, "Rolled out via IaC pipeline.", control2Row[6])
		assert.Equal(t, "Compliant", control2Row[7])
		assert.Equal(t, evalTime.Local().Format("2006-01-02 15:04"), control2Row[8])
	})

	t.Run("happy path: PDF format", func(t *testing.T) {
		res, err := svc.ExportAuditScopeReport(context.Background(), connect.NewRequest(&orchestrator.ExportAuditScopeReportRequest{
			AuditScopeId: orchestratortest.MockScopeId1,
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
			AuditScopeId: orchestratortest.MockNonExistentId,
		}))
		assert.IsConnectError(t, err, connect.CodeNotFound)
	})

	t.Run("err: permission denied", func(t *testing.T) {
		denySvc := &Service{
			db:    db,
			authz: &denyAuthorizationStrategy{},
		}
		_, err := denySvc.ExportAuditScopeReport(context.Background(), connect.NewRequest(&orchestrator.ExportAuditScopeReportRequest{
			AuditScopeId: orchestratortest.MockScopeId1,
		}))
		assert.IsConnectError(t, err, connect.CodePermissionDenied)
	})
}
