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
	"confirmate.io/core/api/assessment"
	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/persistence"
	"confirmate.io/core/service"

	"connectrpc.com/connect"
	"github.com/xuri/excelize/v2"
	"google.golang.org/protobuf/types/known/structpb"
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

var reportFilenameSanitizer = regexp.MustCompile(`[^a-zA-Z0-9-]+`)

// reportMetricRow is the aggregated, latest-per-resource compliance status of a single metric
// for a control, as shown under that control's "Evaluation Result" in the report.
type reportMetricRow struct {
	name              string
	targetComponent   string
	condition         string
	evaluated         bool
	compliant         bool
	complianceComment string
}

// reportControlRow is one control ("Evaluation Result") of the audit scope compliance report,
// combining data from the catalog control, its ControlInScope record, and the aggregated
// compliance status of each of its metrics.
type reportControlRow struct {
	category                string
	shortName               string
	controlName             string
	assuranceLevel          string
	implementationState     string
	implementationStateEnum orchestrator.ControlInScopeState
	assignee                string
	implementationNotes     string
	metrics                 []reportMetricRow
}

// metricCounts returns the number of metrics that are compliant, evaluated at all, and
// configured in total for this control.
func (c reportControlRow) metricCounts() (compliant, evaluated, total int) {
	total = len(c.metrics)
	for _, m := range c.metrics {
		if !m.evaluated {
			continue
		}
		evaluated++
		if m.compliant {
			compliant++
		}
	}
	return
}

// statusLabel summarizes a control's metrics into a single status.
func (c reportControlRow) statusLabel() string {
	compliant, evaluated, total := c.metricCounts()
	switch {
	case total == 0:
		return "No Metrics"
	case evaluated == 0:
		return "Not Evaluated"
	case compliant == total:
		return "Passed"
	default:
		return "Action Required"
	}
}

// ExportAuditScopeReport generates a compliance report for the given audit scope, listing every
// in-scope control ("evaluation result") together with its implementation state and the
// aggregated compliance status of each of its metrics.
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

	// Preload the full control tree (not just top-level controls): autoCreateControlsInScope
	// creates a ControlInScope for every control in the catalog, including sub-controls, which
	// is also where metrics are typically attached (top-level controls are often just a grouping
	// with no metrics of their own). Catalogs are at most 2 levels deep in practice, but preload
	// one extra level as a safety margin; flattenControls below walks to any depth regardless.
	err = svc.db.Get(&catalog,
		persistence.WithPreload("Categories.Controls", "parent_control_id IS NULL"),
		persistence.WithPreload("Categories.Controls.Metrics"),
		persistence.WithPreload("Categories.Controls.Controls"),
		persistence.WithPreload("Categories.Controls.Controls.Metrics"),
		persistence.WithPreload("Categories.Controls.Controls.Controls"),
		persistence.WithPreload("Categories.Controls.Controls.Controls.Metrics"),
		"id = ?", auditScope.GetCatalogId())
	if err = service.HandleDatabaseError(err, service.ErrNotFound("catalog")); err != nil {
		return nil, err
	}

	controls, err := svc.buildReportControls(ctx, &auditScope, &catalog)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	var content []byte
	extension := "xlsx"
	if req.Msg.GetFormat() == orchestrator.ReportFormat_REPORT_FORMAT_PDF {
		extension = "pdf"
		content, err = renderAuditScopeReportPDF(&auditScope, &toe, &catalog, controls)
	} else {
		content, err = renderAuditScopeReportXLSX(&auditScope, &toe, controls)
	}
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("could not render report: %w", err))
	}

	filename := fmt.Sprintf("audit-scope-report-%s-%s.%s",
		reportFilenameSanitizer.ReplaceAllString(auditScope.GetName(), "-"),
		time.Now().Format("20060102"), extension)

	res = connect.NewResponse(&orchestrator.ExportAuditScopeReportResponse{
		Content:  content,
		Filename: filename,
	})
	return
}

// buildReportControls gathers the ControlInScope records, their catalog control metadata (with
// metrics), and the latest-per-resource assessment result for each metric, and combines them
// into report rows ordered by category (in catalog order) and then by control short name.
func (svc *Service) buildReportControls(ctx context.Context, auditScope *orchestrator.AuditScope, catalog *orchestrator.Catalog) ([]reportControlRow, error) {
	// Map every control ID (at any depth: top-level and sub-controls) to the name of the category
	// it belongs to, and collect the IDs of every metric attached to any of those controls.
	// ControlInScope records exist for controls at any depth (see autoCreateControlsInScope), and
	// metrics are typically attached to sub-controls rather than their top-level parent, so
	// restricting this to top-level controls would both drop rows and miss most metrics.
	categoryByControlId := make(map[string]string)
	controlById := make(map[string]*orchestrator.Control)
	var allMetricIds []string
	for _, cat := range catalog.GetCategories() {
		for _, c := range flattenControls(cat.GetControls()) {
			categoryByControlId[c.GetId()] = cat.GetName()
			controlById[c.GetId()] = c
			for _, m := range c.GetMetrics() {
				allMetricIds = append(allMetricIds, m.GetId())
			}
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

	// Fetch the latest assessment result per resource for every metric relevant to this catalog,
	// in one batched call, rather than querying per control or per metric.
	var assessmentResults []*assessment.AssessmentResult
	if len(allMetricIds) > 0 {
		toeId := auditScope.GetTargetOfEvaluationId()
		assessmentResults, err = api.ListAllPaginated(ctx, &orchestrator.ListAssessmentResultsRequest{
			Filter: &orchestrator.ListAssessmentResultsRequest_Filter{
				TargetOfEvaluationId: &toeId,
				MetricIds:            allMetricIds,
			},
			LatestByResourceId: new(true),
		}, func(ctx context.Context, req *orchestrator.ListAssessmentResultsRequest) (*orchestrator.ListAssessmentResultsResponse, error) {
			res, err := svc.ListAssessmentResults(ctx, connect.NewRequest(req))
			if err != nil {
				return nil, err
			}
			return res.Msg, nil
		}, func(res *orchestrator.ListAssessmentResultsResponse) []*assessment.AssessmentResult {
			return res.Results
		})
		if err != nil {
			return nil, fmt.Errorf("could not list assessment results: %w", err)
		}
	}

	resultsByMetric := make(map[string][]*assessment.AssessmentResult, len(allMetricIds))
	for _, r := range assessmentResults {
		resultsByMetric[r.GetMetricId()] = append(resultsByMetric[r.GetMetricId()], r)
	}

	var users []*orchestrator.User
	if err = svc.db.List(&users, "", true, 0, -1, persistence.WithoutPreload()); err != nil {
		return nil, fmt.Errorf("could not list users: %w", err)
	}
	nameByUserId := make(map[string]string, len(users))
	for _, u := range users {
		nameByUserId[u.GetId()] = userDisplayName(u)
	}

	controls := make([]reportControlRow, 0, len(cisList))
	for _, cis := range cisList {
		control := controlById[cis.GetControlId()]
		if control == nil {
			// Control is no longer part of the catalog's top-level structure (e.g. a sub-control,
			// or the catalog changed); skip it rather than showing an incomplete row.
			continue
		}

		row := reportControlRow{
			category:                categoryByControlId[control.GetId()],
			shortName:               control.GetShortName(),
			controlName:             control.GetName(),
			assuranceLevel:          control.GetAssuranceLevel(),
			implementationState:     controlInScopeStateLabels[cis.GetState()],
			implementationStateEnum: cis.GetState(),
			implementationNotes:     cis.GetImplementationDetails(),
		}
		if assigneeId := cis.GetAssigneeId(); assigneeId != "" {
			row.assignee = nameByUserId[assigneeId]
		}

		for _, m := range control.GetMetrics() {
			row.metrics = append(row.metrics, buildReportMetricRow(m, resultsByMetric[m.GetId()]))
		}

		controls = append(controls, row)
	}

	sort.SliceStable(controls, func(i, j int) bool {
		if controls[i].category != controls[j].category {
			return controls[i].category < controls[j].category
		}
		return controls[i].shortName < controls[j].shortName
	})

	return controls, nil
}

// buildReportMetricRow aggregates the latest-per-resource assessment results for a single metric
// into one report row. If more than one resource was assessed, the metric is only considered
// compliant if all of them are; the condition/observed value shown is taken from a non-compliant
// resource if one exists, so the row explains what needs fixing.
func buildReportMetricRow(m *assessment.Metric, results []*assessment.AssessmentResult) reportMetricRow {
	row := reportMetricRow{name: m.GetName()}
	if len(results) == 0 {
		row.targetComponent = "—"
		row.condition = "—"
		return row
	}

	row.evaluated = true

	sample := results[0]
	compliantCount := 0
	for _, r := range results {
		if r.GetCompliant() {
			compliantCount++
		} else if sample.GetCompliant() {
			sample = r
		}
	}
	row.compliant = compliantCount == len(results)
	row.complianceComment = sample.GetComplianceComment()
	row.condition = formatAssessmentCondition(sample)

	component := sample.GetResourceId()
	if types := sample.GetResourceTypes(); len(types) > 0 {
		component = types[0]
	}
	if len(results) > 1 {
		component = fmt.Sprintf("%s (×%d resources)", component, len(results))
	}
	row.targetComponent = component

	return row
}

// formatAssessmentCondition renders the evaluated condition and observed value of an assessment
// result as a single human-readable string, e.g. "tlsVersion >= 1.3 (Observed: 1.3)".
func formatAssessmentCondition(r *assessment.AssessmentResult) string {
	if details := r.GetComplianceDetails(); len(details) > 0 {
		d := details[0]
		return fmt.Sprintf("%s %s %s (Observed: %s)",
			d.GetProperty(), d.GetOperator(), formatStructValue(d.GetTargetValue()), formatStructValue(d.GetValue()))
	}
	if cfg := r.GetMetricConfiguration(); cfg != nil {
		return fmt.Sprintf("Expected %s %s", cfg.GetOperator(), formatStructValue(cfg.GetTargetValue()))
	}
	return "—"
}

func formatStructValue(v *structpb.Value) string {
	if v == nil {
		return "—"
	}
	return fmt.Sprint(v.AsInterface())
}

// flattenControls recursively flattens a control tree (including sub-controls at any depth) into
// a single slice, so every control that could have its own ControlInScope record is included.
func flattenControls(controls []*orchestrator.Control) []*orchestrator.Control {
	flat := make([]*orchestrator.Control, 0, len(controls))
	for _, c := range controls {
		flat = append(flat, c)
		flat = append(flat, flattenControls(c.GetControls())...)
	}
	return flat
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

// renderAuditScopeReportXLSX renders the given controls (and their metrics) into a single-sheet
// XLSX workbook, one row per (control, metric) pair, and returns its raw file content. Controls
// with no metrics still get a single row, so they aren't silently dropped from the export.
func renderAuditScopeReportXLSX(auditScope *orchestrator.AuditScope, toe *orchestrator.TargetOfEvaluation, controls []reportControlRow) ([]byte, error) {
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

	headers := []string{
		"Category", "Control ID", "Control Name", "Implementation State", "Assignee", "Implementation Notes",
		"Metric", "Target Component", "Evaluated Condition & Observed Value", "Metric Status", "Compliance Comment",
	}
	const headerRow = 5
	for i, h := range headers {
		cell, err := excelize.CoordinatesToCellName(i+1, headerRow)
		if err != nil {
			return nil, err
		}
		if err := f.SetCellStr(sheet, cell, h); err != nil {
			return nil, err
		}
	}
	lastHeaderCell, err := excelize.CoordinatesToCellName(len(headers), headerRow)
	if err != nil {
		return nil, err
	}
	if err := f.SetCellStyle(sheet, "A"+fmt.Sprint(headerRow), lastHeaderCell, headerStyle); err != nil {
		return nil, err
	}

	r := headerRow
	writeRow := func(values []string) error {
		r++
		for c, v := range values {
			cell, err := excelize.CoordinatesToCellName(c+1, r)
			if err != nil {
				return err
			}
			if err := f.SetCellStr(sheet, cell, v); err != nil {
				return err
			}
		}
		return nil
	}

	for _, ctrl := range controls {
		base := []string{ctrl.category, ctrl.shortName, ctrl.controlName, ctrl.implementationState, ctrl.assignee, ctrl.implementationNotes}
		if len(ctrl.metrics) == 0 {
			if err := writeRow(append(append([]string{}, base...), "", "", "", "No Metrics", "")); err != nil {
				return nil, err
			}
			continue
		}
		for _, m := range ctrl.metrics {
			status := "Not Evaluated"
			if m.evaluated {
				status = "Compliant"
				if !m.compliant {
					status = "Not Compliant"
				}
			}
			row := append(append([]string{}, base...), m.name, m.targetComponent, m.condition, status, m.complianceComment)
			if err := writeRow(row); err != nil {
				return nil, err
			}
		}
	}

	for i := range headers {
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
