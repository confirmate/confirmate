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

	"confirmate.io/core/api/orchestrator"

	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/row"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/border"
	"github.com/johnfercher/maroto/v2/pkg/consts/breakline"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontfamily"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/core"
	"github.com/johnfercher/maroto/v2/pkg/props"
)

// Report palette. pdfCodeBlue matches the confirmate UI's --color-confirmate-light and is reused
// for code-styled identifiers and section accents; the rest read clearly as semantic status
// colors (green/red/amber/gray) against white and each other.
var (
	pdfNavy       = &props.Color{Red: 15, Green: 23, Blue: 42}
	pdfCodeBlue   = &props.Color{Red: 26, Green: 122, Blue: 191}
	pdfInk        = &props.Color{Red: 31, Green: 41, Blue: 55}
	pdfMuted      = &props.Color{Red: 107, Green: 114, Blue: 128}
	pdfWhite      = &props.Color{Red: 255, Green: 255, Blue: 255}
	pdfZebra      = &props.Color{Red: 249, Green: 250, Blue: 251}
	pdfHeaderBg   = &props.Color{Red: 243, Green: 244, Blue: 246}
	pdfBorderGray = &props.Color{Red: 224, Green: 227, Blue: 231}

	pdfGreen     = &props.Color{Red: 21, Green: 128, Blue: 61}
	pdfGreenTint = &props.Color{Red: 220, Green: 252, Blue: 231}
	pdfRed       = &props.Color{Red: 185, Green: 28, Blue: 28}
	pdfRedTint   = &props.Color{Red: 254, Green: 226, Blue: 226}
	pdfAmber     = &props.Color{Red: 180, Green: 83, Blue: 9}
	pdfAmberTint = &props.Color{Red: 254, Green: 243, Blue: 199}
	pdfBlue      = &props.Color{Red: 29, Green: 78, Blue: 145}
	pdfBlueTint  = &props.Color{Red: 219, Green: 234, Blue: 254}
	pdfGray      = &props.Color{Red: 75, Green: 85, Blue: 99}
	pdfGrayTint  = &props.Color{Red: 229, Green: 231, Blue: 235}
)

// pdfChip is a text/background color pair used to render a status as a small colored badge.
type pdfChip struct {
	text *props.Color
	bg   *props.Color
}

var implementationStateChips = map[orchestrator.ControlInScopeState]pdfChip{
	orchestrator.ControlInScopeState_CONTROL_IN_SCOPE_STATE_OPEN:             {pdfGray, pdfGrayTint},
	orchestrator.ControlInScopeState_CONTROL_IN_SCOPE_STATE_IN_PROGRESS:      {pdfBlue, pdfBlueTint},
	orchestrator.ControlInScopeState_CONTROL_IN_SCOPE_STATE_IMPLEMENTED:      {pdfBlue, pdfBlueTint},
	orchestrator.ControlInScopeState_CONTROL_IN_SCOPE_STATE_READY_FOR_REVIEW: {pdfAmber, pdfAmberTint},
	orchestrator.ControlInScopeState_CONTROL_IN_SCOPE_STATE_ACCEPTED:         {pdfGreen, pdfGreenTint},
}

func implementationChip(s orchestrator.ControlInScopeState) pdfChip {
	if c, ok := implementationStateChips[s]; ok {
		return c
	}
	return pdfChip{pdfGray, pdfGrayTint}
}

// controlStatusChip returns the chip and left-border accent color for a control's aggregated
// metric status label (see [reportControlRow.statusLabel]).
func controlStatusChip(status string) pdfChip {
	switch status {
	case "Passed":
		return pdfChip{pdfGreen, pdfGreenTint}
	case "Action Required":
		return pdfChip{pdfAmber, pdfAmberTint}
	default: // "Not Evaluated", "No Metrics"
		return pdfChip{pdfGray, pdfGrayTint}
	}
}

func metricStatusChip(m reportMetricRow) pdfChip {
	switch {
	case !m.evaluated:
		return pdfChip{pdfGray, pdfGrayTint}
	case m.compliant:
		return pdfChip{pdfGreen, pdfGreenTint}
	default:
		return pdfChip{pdfRed, pdfRedTint}
	}
}

func metricStatusLabel(m reportMetricRow) string {
	switch {
	case !m.evaluated:
		return "Not Evaluated"
	case m.compliant:
		return "Compliant"
	default:
		return "Not Compliant"
	}
}

// truncate shortens s to at most n runes, appending an ellipsis if it was cut. Report table rows
// are kept to a single line so we don't depend on maroto's auto row height, which undersizes
// rows with wrapped text and lets later lines bleed into the row below.
func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n-1]) + "…"
}

// reportSummary holds the metric counts shown in the PDF's summary cards and header status pill.
type reportSummary struct {
	controls, totalMetrics, compliantMetrics, evaluatedMetrics, actionRequired int
}

func summarizeReportControls(controls []reportControlRow) reportSummary {
	s := reportSummary{controls: len(controls)}
	for _, c := range controls {
		compliant, evaluated, total := c.metricCounts()
		s.totalMetrics += total
		s.compliantMetrics += compliant
		s.evaluatedMetrics += evaluated
		s.actionRequired += evaluated - compliant
	}
	return s
}

// compliancePercent returns the share of evaluated metrics that are compliant, ignoring metrics
// that haven't been evaluated yet (so a freshly-scoped audit doesn't score as failing).
func (s reportSummary) compliancePercent() float64 {
	if s.evaluatedMetrics == 0 {
		return 0
	}
	return 100 * float64(s.compliantMetrics) / float64(s.evaluatedMetrics)
}

// overallStatusPassThreshold mirrors the confirmate demo report template, which still labels a
// 90%-compliant assessment as an overall "PASS".
const overallStatusPassThreshold = 90.0

// renderAuditScopeReportPDF renders the given controls (and their metrics) into a styled,
// single-document PDF report and returns its raw file content.
func renderAuditScopeReportPDF(auditScope *orchestrator.AuditScope, toe *orchestrator.TargetOfEvaluation, catalog *orchestrator.Catalog, controls []reportControlRow) ([]byte, error) {
	cfg := config.NewBuilder().
		WithLeftMargin(12).
		WithTopMargin(10).
		WithRightMargin(12).
		WithBottomMargin(15).
		WithPageNumber(props.PageNumber{
			Pattern: "Page {current} of {total}",
			Place:   props.RightBottom,
			Size:    8,
			Color:   pdfMuted,
		}).
		WithTitle(fmt.Sprintf("Compliance Report - %s", auditScope.GetName()), false).
		Build()

	m := maroto.New(cfg)

	summary := summarizeReportControls(controls)

	if err := m.RegisterHeader(pdfHeaderRow(auditScope, summary)); err != nil {
		return nil, err
	}
	if err := m.RegisterFooter(pdfFooterRow()); err != nil {
		return nil, err
	}

	m.AddRows(pdfInfoRow(auditScope, toe, catalog))
	m.AddRows(pdfSummaryRow(summary))

	m.AddRows(pdfSectionHeaderRow("1. Evaluation Results Overview"))
	m.AddRows(pdfOverviewTableHeaderRow())
	for _, ctrl := range controls {
		m.AddRows(pdfOverviewRow(ctrl))
	}

	m.AddRows(row.New(6))
	m.AddRows(pdfSectionHeaderRow("2. Metrics by Evaluation Result"))
	for _, ctrl := range controls {
		for _, r := range pdfControlCardRows(ctrl) {
			m.AddRows(r)
		}
		m.AddRows(row.New(4))
	}

	document, err := m.Generate()
	if err != nil {
		return nil, err
	}
	return document.GetBytes(), nil
}

func pdfHeaderRow(auditScope *orchestrator.AuditScope, summary reportSummary) core.Row {
	pct := summary.compliancePercent()
	status := "ACTION REQUIRED"
	pillChip := pdfChip{pdfWhite, pdfAmber}
	if pct >= overallStatusPassThreshold {
		status = "PASS"
		pillChip = pdfChip{pdfWhite, pdfGreen}
	}

	return row.New(24).Add(
		col.New(7).Add(
			text.New("Compliance Report", props.Text{Top: 5, Left: 2, Size: 16, Style: fontstyle.Bold, Color: pdfCodeBlue}),
			text.New("CONFIRMATE Full Compliance Breakdown", props.Text{Top: 13, Left: 2, Size: 9, Color: pdfWhite}),
			text.New(auditScope.GetName(), props.Text{Top: 18, Left: 2, Size: 8, Color: pdfMuted}),
		),
		col.New(5).Add(
			text.New(fmt.Sprintf("STATUS: %s (%.1f%%)", status, pct), props.Text{Top: 7, Right: 2, Size: 10, Style: fontstyle.Bold, Align: align.Right, Color: pillChip.text}),
		).WithStyle(&props.Cell{BackgroundColor: pillChip.bg, BorderType: border.Full, BorderColor: pillChip.bg}),
	).WithStyle(&props.Cell{BackgroundColor: pdfNavy})
}

func pdfFooterRow() core.Row {
	return row.New(14).Add(
		col.New(12).Add(
			text.New("Confirmate Compliance Report · Confidential", props.Text{Top: 5, Size: 7, Color: pdfMuted}),
		),
	).WithStyle(&props.Cell{BorderType: border.Top, BorderColor: pdfGrayTint, BorderThickness: 0.3})
}

func pdfInfoField(size int, label, value string) core.Col {
	return col.New(size).Add(
		text.New(label, props.Text{Top: 3, Left: 2, Size: 7.5, Style: fontstyle.Bold, Color: pdfMuted}),
		text.New(truncate(value, 34), props.Text{Top: 8, Left: 2, Size: 10.5, Style: fontstyle.Bold, Color: pdfInk}),
	)
}

func pdfInfoRow(auditScope *orchestrator.AuditScope, toe *orchestrator.TargetOfEvaluation, catalog *orchestrator.Catalog) core.Row {
	assurance := auditScope.GetAssuranceLevel()
	if assurance == "" {
		assurance = "—"
	}
	return row.New(18).Add(
		pdfInfoField(4, "TARGET OF EVALUATION", toe.GetName()),
		pdfInfoField(4, "CATALOG", catalog.GetName()),
		pdfInfoField(2, "ASSURANCE LEVEL", assurance),
		pdfInfoField(2, "STATUS", auditScopeStatusLabels[auditScope.GetStatus()]),
	).WithStyle(&props.Cell{BorderType: border.Bottom, BorderColor: pdfBorderGray, BorderThickness: 0.4})
}

func pdfSummaryCard(size int, label string, value int, color *props.Color) core.Col {
	return col.New(size).Add(
		text.New(fmt.Sprint(value), props.Text{Top: 4, Size: 20, Style: fontstyle.Bold, Align: align.Center, Color: color}),
		text.New(label, props.Text{Top: 13, Size: 7, Align: align.Center, Color: pdfMuted}),
	).WithStyle(&props.Cell{BorderType: border.Full, BorderColor: pdfBorderGray, BorderThickness: 0.4})
}

func pdfSummaryRow(s reportSummary) core.Row {
	return row.New(20).Add(
		pdfSummaryCard(3, "EVALUATION RESULTS", s.controls, pdfInk),
		pdfSummaryCard(3, "TOTAL METRICS", s.totalMetrics, pdfInk),
		pdfSummaryCard(3, "COMPLIANT METRICS", s.compliantMetrics, pdfGreen),
		pdfSummaryCard(3, "ACTION REQUIRED", s.actionRequired, pdfAmber),
	)
}

func pdfSectionHeaderRow(title string) core.Row {
	return row.New(10).Add(
		col.New(1).Add(
			text.New(" ", props.Text{}),
		).WithStyle(&props.Cell{BackgroundColor: pdfCodeBlue}),
		col.New(11).Add(
			text.New(title, props.Text{Top: 2, Left: 2, Size: 11, Style: fontstyle.Bold, Color: pdfInk}),
		),
	)
}

func pdfOverviewTableHeaderRow() core.Row {
	headerText := props.Text{Top: 2.5, Size: 7.5, Style: fontstyle.Bold, Color: pdfMuted}
	return row.New(6).Add(
		text.NewCol(2, "CONTROL", headerText),
		text.NewCol(3, "NAME / FOCUS AREA", headerText),
		text.NewCol(2, "CATEGORY", headerText),
		text.NewCol(2, "IMPLEMENTATION", headerText),
		text.NewCol(1, "METRICS", headerText),
		text.NewCol(2, "STATUS", headerText),
	).WithStyle(&props.Cell{BackgroundColor: pdfHeaderBg})
}

func pdfOverviewRow(ctrl reportControlRow) core.Row {
	compliant, _, total := ctrl.metricCounts()
	statusChip := controlStatusChip(ctrl.statusLabel())
	implChip := implementationChip(ctrl.implementationStateEnum)

	idCol := col.New(2).Add(
		text.New(ctrl.shortName, props.Text{Top: 2.5, Left: 2, Size: 8.5, Family: fontfamily.Courier, Style: fontstyle.Bold, Color: pdfCodeBlue}),
	)
	nameCol := col.New(3).Add(
		text.New(truncate(ctrl.controlName, 28), props.Text{Top: 2.5, Size: 8.5, Color: pdfInk}),
	)
	categoryCol := col.New(2).Add(
		text.New(truncate(ctrl.category, 18), props.Text{Top: 2.5, Size: 8, Color: pdfMuted}),
	)
	implCol := col.New(2).Add(
		text.New(ctrl.implementationState, props.Text{Top: 2.5, Size: 7.5, Style: fontstyle.Bold, Align: align.Center, Color: implChip.text}),
	).WithStyle(&props.Cell{BackgroundColor: implChip.bg})
	metricsCol := col.New(1).Add(
		text.New(fmt.Sprint(total), props.Text{Top: 2.5, Size: 8.5, Align: align.Center, Color: pdfInk}),
	)
	statusCol := col.New(2).Add(
		text.New(fmt.Sprintf("%s (%d/%d)", ctrl.statusLabel(), compliant, total), props.Text{Top: 2.5, Size: 7.5, Style: fontstyle.Bold, Align: align.Center, Color: statusChip.text}),
	).WithStyle(&props.Cell{BackgroundColor: statusChip.bg})

	return row.New(8).Add(idCol, nameCol, categoryCol, implCol, metricsCol, statusCol)
}

// pdfControlCardRows renders one control's card: a header row (control ID/name and aggregated
// status), a metrics table (if any), and a remediation callout for any non-compliant metric that
// has a compliance comment. Every row shares a colored left border matching the control's status,
// approximating a bordered card since maroto rows can't span a single box with rounded corners.
func pdfControlCardRows(ctrl reportControlRow) []core.Row {
	statusChip := controlStatusChip(ctrl.statusLabel())
	accent := &props.Cell{BorderType: border.Left, BorderColor: statusChip.text, BorderThickness: 1.5}

	compliant, _, total := ctrl.metricCounts()
	rows := []core.Row{
		row.New(10).Add(
			col.New(8).Add(
				text.New(fmt.Sprintf("Control: %s", ctrl.shortName), props.Text{Top: 2.5, Left: 3, Size: 9.5, Style: fontstyle.Bold, Color: pdfInk}),
				text.New(truncate(ctrl.controlName, 55), props.Text{Top: 2.5, Left: 30, Size: 9.5, Color: pdfMuted}),
			),
			col.New(4).Add(
				text.New(fmt.Sprintf("%s (%d/%d)", ctrl.statusLabel(), compliant, total), props.Text{Top: 3, Size: 8, Style: fontstyle.Bold, Align: align.Right, Right: 3, Color: statusChip.text}),
			),
		).WithStyle(mergeCellStyle(accent, &props.Cell{BackgroundColor: pdfZebra})),
	}

	if len(ctrl.metrics) == 0 {
		rows = append(rows, row.New(8).Add(
			col.New(12).Add(text.New("No metrics configured for this control.", props.Text{Top: 2.5, Left: 3, Size: 8, Color: pdfMuted})),
		).WithStyle(accent))
		return rows
	}

	headerText := props.Text{Top: 2.5, Size: 7, Style: fontstyle.Bold, Color: pdfMuted}
	firstHeaderText := headerText
	firstHeaderText.Left = 3
	rows = append(rows, row.New(6).Add(
		text.NewCol(3, "METRIC", firstHeaderText),
		text.NewCol(3, "TARGET COMPONENT", headerText),
		text.NewCol(4, "EVALUATED CONDITION & OBSERVED VALUE", headerText),
		text.NewCol(2, "STATUS", headerText),
	).WithStyle(mergeCellStyle(accent, &props.Cell{BackgroundColor: pdfHeaderBg})))

	for i, m := range ctrl.metrics {
		chip := metricStatusChip(m)
		rw := row.New().Add(
			col.New(3).Add(text.New(m.name, props.Text{Top: 2.5, Left: 3, Size: 7.5, Family: fontfamily.Courier, Color: pdfCodeBlue, BreakLineStrategy: breakline.DashStrategy})),
			col.New(3).Add(text.New(m.targetComponent, props.Text{Top: 2.5, Size: 7.5, Color: pdfInk})),
			col.New(4).Add(text.New(m.condition, props.Text{Top: 2.5, Size: 7.5, Family: fontfamily.Courier, Color: pdfInk, BreakLineStrategy: breakline.DashStrategy})),
			col.New(2).Add(text.New(metricStatusLabel(m), props.Text{Top: 2.5, Size: 7.5, Style: fontstyle.Bold, Align: align.Center, Color: chip.text})).WithStyle(&props.Cell{BackgroundColor: chip.bg}),
		)
		style := *accent
		if i%2 == 1 {
			style.BackgroundColor = pdfZebra
		}
		rows = append(rows, rw.WithStyle(&style))

		// Attach the remediation callout directly under its own metric, rather than batching all
		// of a control's remediations together at the end of its card.
		if m.evaluated && !m.compliant && m.complianceComment != "" {
			rows = append(rows, row.New().Add(
				col.New(12).Add(
					text.New(fmt.Sprintf("Remediation (%s): %s", m.name, m.complianceComment), props.Text{Top: 2.5, Left: 3, Size: 7.5, Color: pdfAmber}),
				),
			).WithStyle(mergeCellStyle(accent, &props.Cell{BackgroundColor: pdfAmberTint})))
		}
	}

	return rows
}

func mergeCellStyle(a, b *props.Cell) *props.Cell {
	merged := *a
	if b.BackgroundColor != nil {
		merged.BackgroundColor = b.BackgroundColor
	}
	return &merged
}

// auditScopeStatusLabels maps [orchestrator.AuditScopeStatus] values to short, human-readable
// labels for display in the report.
var auditScopeStatusLabels = map[orchestrator.AuditScopeStatus]string{
	orchestrator.AuditScopeStatus_AUDIT_SCOPE_STATUS_UNSPECIFIED:                      "—",
	orchestrator.AuditScopeStatus_AUDIT_SCOPE_STATUS_SETUP:                            "Setup",
	orchestrator.AuditScopeStatus_AUDIT_SCOPE_STATUS_INTERNAL_REVIEW:                  "Internal Review",
	orchestrator.AuditScopeStatus_AUDIT_SCOPE_STATUS_AUDITOR_REVIEW:                   "Auditor Review",
	orchestrator.AuditScopeStatus_AUDIT_SCOPE_STATUS_CONTINUOUS_COMPLIANCE_MANAGEMENT: "Continuous Compliance",
	orchestrator.AuditScopeStatus_AUDIT_SCOPE_STATUS_FIXED:                            "Fixed",
}
