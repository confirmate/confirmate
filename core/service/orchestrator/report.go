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
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"confirmate.io/core/api"
	"confirmate.io/core/api/evaluation"
	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/persistence"
	"confirmate.io/core/service"

	"connectrpc.com/connect"
	"github.com/xuri/excelize/v2"
)

// controlInScopeStateLabels maps [orchestrator.ControlInScopeState] values to short,
// human-readable labels for display in the report.
var controlInScopeStateLabels = map[orchestrator.ControlInScopeState]string{
	orchestrator.ControlInScopeState_CONTROL_IN_SCOPE_STATE_UNSPECIFIED:      "",
	orchestrator.ControlInScopeState_CONTROL_IN_SCOPE_STATE_OPEN:             "Open",
	orchestrator.ControlInScopeState_CONTROL_IN_SCOPE_STATE_IN_PROGRESS:      "In Progress",
	orchestrator.ControlInScopeState_CONTROL_IN_SCOPE_STATE_IMPLEMENTED:      "Implemented",
	orchestrator.ControlInScopeState_CONTROL_IN_SCOPE_STATE_READY_FOR_REVIEW: "Ready for Review",
	orchestrator.ControlInScopeState_CONTROL_IN_SCOPE_STATE_ACCEPTED:         "Accepted",
}

// evaluationStatusLabels maps [evaluation.EvaluationStatus] values to short, human-readable
// labels for display in the report.
var evaluationStatusLabels = map[evaluation.EvaluationStatus]string{
	evaluation.EvaluationStatus_EVALUATION_STATUS_UNSPECIFIED:            "",
	evaluation.EvaluationStatus_EVALUATION_STATUS_COMPLIANT:              "Compliant",
	evaluation.EvaluationStatus_EVALUATION_STATUS_COMPLIANT_MANUALLY:     "Compliant (manual)",
	evaluation.EvaluationStatus_EVALUATION_STATUS_NOT_COMPLIANT:          "Not Compliant",
	evaluation.EvaluationStatus_EVALUATION_STATUS_NOT_COMPLIANT_MANUALLY: "Not Compliant (manual)",
	evaluation.EvaluationStatus_EVALUATION_STATUS_PENDING:                "Pending",
}

var reportFilenameSanitizer = regexp.MustCompile(`[^a-zA-Z0-9-]+`)

// reportRow is one row of the audit scope compliance report, combining data from the catalog
// control, its ControlInScope record, and its latest evaluation result.
type reportRow struct {
	category            string
	shortName           string
	controlName         string
	assuranceLevel      string
	implementationState string
	assignee            string
	implementationNotes string
	evaluationStatus    string
	evaluationTimestamp string
	evaluationComment   string
}

// ExportAuditScopeReport generates an XLSX compliance report for the given audit scope, listing
// every in-scope control together with its implementation state and latest evaluation status.
func (svc *Service) ExportAuditScopeReport(
	ctx context.Context,
	req *connect.Request[orchestrator.ExportAuditScopeReportRequest],
) (res *connect.Response[orchestrator.ExportAuditScopeReportResponse], err error) {
	var (
		auditScope orchestrator.AuditScope
		toe        orchestrator.TargetOfEvaluation
		catalog    orchestrator.Catalog
		allowed    bool
	)

	// Validate the request
	if err = service.Validate(req); err != nil {
		return nil, err
	}

	// Check access via the configured auth strategy
	allowed, _, err = CheckAccess(ctx, svc.authz, svc, orchestrator.RequestType_REQUEST_TYPE_GET, req.Msg.GetAuditScopeId(), orchestrator.ObjectType_OBJECT_TYPE_AUDIT_SCOPE)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if !allowed {
		return nil, service.ErrPermissionDenied
	}

	err = svc.db.Get(&auditScope, persistence.WithoutPreload(), "id = ?", req.Msg.GetAuditScopeId())
	if err = service.HandleDatabaseError(err, service.ErrNotFound("audit scope")); err != nil {
		return nil, err
	}

	err = svc.db.Get(&toe, persistence.WithoutPreload(), "id = ?", auditScope.GetTargetOfEvaluationId())
	if err = service.HandleDatabaseError(err, service.ErrNotFound("target of evaluation")); err != nil {
		return nil, err
	}

	err = svc.db.Get(&catalog,
		persistence.WithPreload("Categories.Controls", "parent_control_id IS NULL"),
		"id = ?", auditScope.GetCatalogId())
	if err = service.HandleDatabaseError(err, service.ErrNotFound("catalog")); err != nil {
		return nil, err
	}

	rows, err := svc.buildReportRows(ctx, &auditScope, &catalog)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	content, err := renderAuditScopeReportXLSX(&auditScope, &toe, rows)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("could not render report: %w", err))
	}

	filename := fmt.Sprintf("audit-scope-report-%s-%s.xlsx",
		reportFilenameSanitizer.ReplaceAllString(auditScope.GetName(), "-"),
		time.Now().Format("20060102"))

	res = connect.NewResponse(&orchestrator.ExportAuditScopeReportResponse{
		Content:  content,
		Filename: filename,
	})
	return
}

// buildReportRows gathers the ControlInScope records, their catalog control metadata, and their
// latest evaluation result for the given audit scope, and combines them into report rows ordered
// by category (in catalog order) and then by control short name.
func (svc *Service) buildReportRows(ctx context.Context, auditScope *orchestrator.AuditScope, catalog *orchestrator.Catalog) ([]reportRow, error) {
	// Map every top-level control ID to the name of the category it belongs to.
	categoryByControlId := make(map[string]string)
	controlById := make(map[string]*orchestrator.Control)
	for _, cat := range catalog.GetCategories() {
		for _, c := range cat.GetControls() {
			categoryByControlId[c.GetId()] = cat.GetName()
			controlById[c.GetId()] = c
		}
	}

	auditScopeId := auditScope.GetId()
	cisList, err := api.ListAllPaginated(ctx, &orchestrator.ListControlsInScopeRequest{
		Filter: &orchestrator.ListControlsInScopeRequest_Filter{AuditScopeId: &auditScopeId},
	}, func(ctx context.Context, req *orchestrator.ListControlsInScopeRequest) (*orchestrator.ListControlsInScopeResponse, error) {
		res, err := svc.ListControlsInScope(ctx, connect.NewRequest(req))
		if err != nil {
			return nil, err
		}
		return res.Msg, nil
	}, func(res *orchestrator.ListControlsInScopeResponse) []*orchestrator.ControlInScope {
		return res.ControlsInScope
	})
	if err != nil {
		return nil, fmt.Errorf("could not list controls in scope: %w", err)
	}

	evalRes, err := svc.ListEvaluationResults(ctx, connect.NewRequest(&orchestrator.ListEvaluationResultsRequest{
		Filter: &orchestrator.ListEvaluationResultsRequest_Filter{
			AuditScopeId: &auditScopeId,
			ParentsOnly:  new(true),
		},
		LatestByControlId: new(true),
	}))
	if err != nil {
		return nil, fmt.Errorf("could not list evaluation results: %w", err)
	}
	evalByControlId := make(map[string]*evaluation.EvaluationResult, len(evalRes.Msg.GetResults()))
	for _, r := range evalRes.Msg.GetResults() {
		evalByControlId[r.GetControlId()] = r
	}

	var users []*orchestrator.User
	if err = svc.db.List(&users, "", true, 0, -1, persistence.WithoutPreload()); err != nil {
		return nil, fmt.Errorf("could not list users: %w", err)
	}
	nameByUserId := make(map[string]string, len(users))
	for _, u := range users {
		nameByUserId[u.GetId()] = userDisplayName(u)
	}

	rows := make([]reportRow, 0, len(cisList))
	for _, cis := range cisList {
		control := controlById[cis.GetControlId()]
		if control == nil {
			// Control is no longer part of the catalog's top-level structure (e.g. a sub-control,
			// or the catalog changed); skip it rather than showing an incomplete row.
			continue
		}

		row := reportRow{
			category:            categoryByControlId[control.GetId()],
			shortName:           control.GetShortName(),
			controlName:         control.GetName(),
			assuranceLevel:      control.GetAssuranceLevel(),
			implementationState: controlInScopeStateLabels[cis.GetState()],
			implementationNotes: cis.GetImplementationDetails(),
		}
		if assigneeId := cis.GetAssigneeId(); assigneeId != "" {
			row.assignee = nameByUserId[assigneeId]
		}
		if evalResult := evalByControlId[control.GetId()]; evalResult != nil {
			row.evaluationStatus = evaluationStatusLabels[evalResult.GetStatus()]
			if ts := evalResult.GetTimestamp(); ts != nil {
				row.evaluationTimestamp = ts.AsTime().Local().Format("2006-01-02 15:04")
			}
			row.evaluationComment = evalResult.GetComment()
		}

		rows = append(rows, row)
	}

	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].category != rows[j].category {
			return rows[i].category < rows[j].category
		}
		return rows[i].shortName < rows[j].shortName
	})

	return rows, nil
}

// userDisplayName returns the best available human-readable name for a user.
func userDisplayName(u *orchestrator.User) string {
	if name := strings.TrimSpace(fmt.Sprintf("%s %s", u.GetFirstName(), u.GetLastName())); name != "" {
		return name
	}
	if u.GetUsername() != "" {
		return u.GetUsername()
	}
	return u.GetId()
}

// reportColumns describes each column of the report sheet, in order.
var reportColumns = []struct {
	header string
	value  func(reportRow) string
}{
	{"Category", func(r reportRow) string { return r.category }},
	{"Control ID", func(r reportRow) string { return r.shortName }},
	{"Control Name", func(r reportRow) string { return r.controlName }},
	{"Assurance Level", func(r reportRow) string { return r.assuranceLevel }},
	{"Implementation State", func(r reportRow) string { return r.implementationState }},
	{"Assignee", func(r reportRow) string { return r.assignee }},
	{"Implementation Notes", func(r reportRow) string { return r.implementationNotes }},
	{"Evaluation Status", func(r reportRow) string { return r.evaluationStatus }},
	{"Evaluation Timestamp", func(r reportRow) string { return r.evaluationTimestamp }},
	{"Evaluation Comment", func(r reportRow) string { return r.evaluationComment }},
}

// renderAuditScopeReportXLSX renders the given rows into a single-sheet XLSX workbook and returns
// its raw file content.
func renderAuditScopeReportXLSX(auditScope *orchestrator.AuditScope, toe *orchestrator.TargetOfEvaluation, rows []reportRow) ([]byte, error) {
	const sheet = "Report"

	f := excelize.NewFile()
	defer f.Close()

	if err := f.SetSheetName(f.GetSheetName(0), sheet); err != nil {
		return nil, err
	}

	titleStyle, err := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true, Size: 14}})
	if err != nil {
		return nil, err
	}
	headerStyle, err := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})
	if err != nil {
		return nil, err
	}

	if err := f.SetCellStr(sheet, "A1", fmt.Sprintf("Compliance Report: %s", auditScope.GetName())); err != nil {
		return nil, err
	}
	if err := f.SetCellStyle(sheet, "A1", "A1", titleStyle); err != nil {
		return nil, err
	}
	if err := f.SetCellStr(sheet, "A2", fmt.Sprintf("Target of Evaluation: %s", toe.GetName())); err != nil {
		return nil, err
	}
	if err := f.SetCellStr(sheet, "A3", fmt.Sprintf("Generated: %s", time.Now().Local().Format("2006-01-02 15:04"))); err != nil {
		return nil, err
	}

	const headerRow = 5
	for i, col := range reportColumns {
		cell, err := excelize.CoordinatesToCellName(i+1, headerRow)
		if err != nil {
			return nil, err
		}
		if err := f.SetCellStr(sheet, cell, col.header); err != nil {
			return nil, err
		}
	}
	lastHeaderCell, err := excelize.CoordinatesToCellName(len(reportColumns), headerRow)
	if err != nil {
		return nil, err
	}
	if err := f.SetCellStyle(sheet, "A"+fmt.Sprint(headerRow), lastHeaderCell, headerStyle); err != nil {
		return nil, err
	}

	for r, row := range rows {
		for c, col := range reportColumns {
			cell, err := excelize.CoordinatesToCellName(c+1, headerRow+1+r)
			if err != nil {
				return nil, err
			}
			if err := f.SetCellStr(sheet, cell, col.value(row)); err != nil {
				return nil, err
			}
		}
	}

	for i := range reportColumns {
		col, err := excelize.ColumnNumberToName(i + 1)
		if err != nil {
			return nil, err
		}
		if err := f.SetColWidth(sheet, col, col, 22); err != nil {
			return nil, err
		}
	}

	var buf bytes.Buffer
	if _, err := f.WriteTo(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
