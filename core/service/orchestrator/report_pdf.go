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
	"time"

	"confirmate.io/core/api/evaluation"
	"confirmate.io/core/api/orchestrator"

	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/line"
	"github.com/johnfercher/maroto/v2/pkg/components/row"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/border"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/core"
	"github.com/johnfercher/maroto/v2/pkg/props"
)

// Report palette. brandBlue matches the confirmate UI's --color-confirmate; the rest are chosen
// to read clearly as semantic status colors (green/red/amber/gray) against white and each other.
var (
	pdfBrandBlue  = &props.Color{Red: 0, Green: 91, Blue: 153}
	pdfInk        = &props.Color{Red: 31, Green: 41, Blue: 55}
	pdfMuted      = &props.Color{Red: 107, Green: 114, Blue: 128}
	pdfWhite      = &props.Color{Red: 255, Green: 255, Blue: 255}
	pdfZebra      = &props.Color{Red: 249, Green: 250, Blue: 251}
	pdfHeaderBg   = &props.Color{Red: 243, Green: 244, Blue: 246}
	pdfCategoryBg = &props.Color{Red: 227, Green: 239, Blue: 249}

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

// pdfChip is a text/background color pair used to render a status or state as a small colored badge.
type pdfChip struct {
	text *props.Color
	bg   *props.Color
}

var evaluationStatusChips = map[evaluation.EvaluationStatus]pdfChip{
	evaluation.EvaluationStatus_EVALUATION_STATUS_COMPLIANT:              {pdfGreen, pdfGreenTint},
	evaluation.EvaluationStatus_EVALUATION_STATUS_COMPLIANT_MANUALLY:     {pdfGreen, pdfGreenTint},
	evaluation.EvaluationStatus_EVALUATION_STATUS_NOT_COMPLIANT:          {pdfRed, pdfRedTint},
	evaluation.EvaluationStatus_EVALUATION_STATUS_NOT_COMPLIANT_MANUALLY: {pdfRed, pdfRedTint},
	evaluation.EvaluationStatus_EVALUATION_STATUS_PENDING:                {pdfAmber, pdfAmberTint},
}

var implementationStateChips = map[orchestrator.ControlInScopeState]pdfChip{
	orchestrator.ControlInScopeState_CONTROL_IN_SCOPE_STATE_OPEN:             {pdfGray, pdfGrayTint},
	orchestrator.ControlInScopeState_CONTROL_IN_SCOPE_STATE_IN_PROGRESS:      {pdfBlue, pdfBlueTint},
	orchestrator.ControlInScopeState_CONTROL_IN_SCOPE_STATE_IMPLEMENTED:      {pdfBlue, pdfBlueTint},
	orchestrator.ControlInScopeState_CONTROL_IN_SCOPE_STATE_READY_FOR_REVIEW: {pdfAmber, pdfAmberTint},
	orchestrator.ControlInScopeState_CONTROL_IN_SCOPE_STATE_ACCEPTED:         {pdfGreen, pdfGreenTint},
}

// truncate shortens s to at most n runes, appending an ellipsis if it was cut. Keeping report
// table rows to a single line avoids relying on maroto's auto row height, which undersizes rows
// with wrapped text and lets the second line bleed into the row below.
func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n-1]) + "…"
}

func evaluationChip(s evaluation.EvaluationStatus) pdfChip {
	if c, ok := evaluationStatusChips[s]; ok {
		return c
	}
	return pdfChip{pdfGray, pdfGrayTint}
}

func implementationChip(s orchestrator.ControlInScopeState) pdfChip {
	if c, ok := implementationStateChips[s]; ok {
		return c
	}
	return pdfChip{pdfGray, pdfGrayTint}
}

// reportSummary holds the control counts shown in the PDF's summary cards.
type reportSummary struct {
	total, compliant, notCompliant, outstanding int
}

func summarizeReportRows(rows []reportRow) reportSummary {
	var s reportSummary
	s.total = len(rows)
	for _, r := range rows {
		switch r.evaluationStatusEnum {
		case evaluation.EvaluationStatus_EVALUATION_STATUS_COMPLIANT,
			evaluation.EvaluationStatus_EVALUATION_STATUS_COMPLIANT_MANUALLY:
			s.compliant++
		case evaluation.EvaluationStatus_EVALUATION_STATUS_NOT_COMPLIANT,
			evaluation.EvaluationStatus_EVALUATION_STATUS_NOT_COMPLIANT_MANUALLY:
			s.notCompliant++
		default:
			s.outstanding++
		}
	}
	return s
}

// renderAuditScopeReportPDF renders the given rows into a styled, single-document PDF report and
// returns its raw file content.
func renderAuditScopeReportPDF(auditScope *orchestrator.AuditScope, toe *orchestrator.TargetOfEvaluation, catalog *orchestrator.Catalog, rows []reportRow) ([]byte, error) {
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

	if err := m.RegisterHeader(pdfHeaderRow(auditScope)); err != nil {
		return nil, err
	}
	if err := m.RegisterFooter(pdfFooterRow()); err != nil {
		return nil, err
	}

	m.AddRows(pdfInfoRow(auditScope, toe, catalog))
	m.AddRows(pdfSummaryRow(summarizeReportRows(rows)))
	m.AddRows(line.NewRow(6, props.Line{Color: pdfGrayTint, Thickness: 0.4}))

	category := ""
	i := 0
	for _, r := range rows {
		if r.category != category {
			category = r.category
			i = 0
			m.AddRows(pdfCategoryHeaderRow(category))
			m.AddRows(pdfTableHeaderRow())
		}
		m.AddRows(pdfControlRow(r, i%2 == 1))
		i++
	}

	document, err := m.Generate()
	if err != nil {
		return nil, err
	}
	return document.GetBytes(), nil
}

func pdfHeaderRow(auditScope *orchestrator.AuditScope) core.Row {
	return row.New(22).Add(
		col.New(7).Add(
			text.New("CONFIRMATE", props.Text{Top: 5, Left: 2, Size: 16, Style: fontstyle.Bold, Color: pdfWhite}),
			text.New("Audit Scope Compliance Report", props.Text{Top: 13, Left: 2, Size: 9, Color: pdfWhite}),
		),
		col.New(5).Add(
			text.New(auditScope.GetName(), props.Text{Top: 6, Right: 2, Size: 11, Style: fontstyle.Bold, Align: align.Right, Color: pdfWhite}),
			text.New("Generated "+time.Now().Local().Format("2006-01-02 15:04"), props.Text{Top: 13, Right: 2, Size: 8, Align: align.Right, Color: pdfWhite}),
		),
	).WithStyle(&props.Cell{BackgroundColor: pdfBrandBlue})
}

func pdfFooterRow() core.Row {
	return row.New(14).Add(
		col.New(12).Add(
			text.New("Confirmate · Continuous Compliance Platform", props.Text{Top: 5, Size: 7, Color: pdfMuted}),
		),
	).WithStyle(&props.Cell{BorderType: border.Top, BorderColor: pdfGrayTint, BorderThickness: 0.3})
}

func pdfInfoField(size int, label, value string) core.Col {
	return col.New(size).Add(
		text.New(label, props.Text{Top: 3, Size: 7.5, Style: fontstyle.Bold, Color: pdfMuted}),
		text.New(value, props.Text{Top: 8, Size: 11, Style: fontstyle.Bold, Color: pdfInk}),
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
	)
}

func pdfSummaryCard(size int, label string, value int, chip pdfChip) core.Col {
	return col.New(size).Add(
		text.New(fmt.Sprint(value), props.Text{Top: 4, Size: 20, Style: fontstyle.Bold, Align: align.Center, Color: chip.text}),
		text.New(label, props.Text{Top: 13, Size: 7.5, Align: align.Center, Color: pdfMuted}),
	).WithStyle(&props.Cell{BackgroundColor: chip.bg, BorderType: border.Full, BorderColor: pdfWhite, BorderThickness: 1.2})
}

func pdfSummaryRow(s reportSummary) core.Row {
	return row.New(20).Add(
		pdfSummaryCard(3, "TOTAL CONTROLS", s.total, pdfChip{pdfInk, pdfGrayTint}),
		pdfSummaryCard(3, "COMPLIANT", s.compliant, pdfChip{pdfGreen, pdfGreenTint}),
		pdfSummaryCard(3, "NOT COMPLIANT", s.notCompliant, pdfChip{pdfRed, pdfRedTint}),
		pdfSummaryCard(3, "OUTSTANDING", s.outstanding, pdfChip{pdfAmber, pdfAmberTint}),
	)
}

func pdfCategoryHeaderRow(category string) core.Row {
	return row.New(9).Add(
		col.New(12).Add(
			text.New(category, props.Text{Top: 3, Left: 2, Size: 10.5, Style: fontstyle.Bold, Color: pdfBrandBlue}),
		),
	).WithStyle(&props.Cell{BackgroundColor: pdfCategoryBg})
}

func pdfTableHeaderRow() core.Row {
	headerText := props.Text{Top: 2.5, Size: 7.5, Style: fontstyle.Bold, Color: pdfMuted}
	return row.New(6).Add(
		text.NewCol(5, "CONTROL", headerText),
		text.NewCol(2, "ASSIGNEE", headerText),
		text.NewCol(2, "IMPLEMENTATION", headerText),
		text.NewCol(3, "EVALUATION", headerText),
	).WithStyle(&props.Cell{BackgroundColor: pdfHeaderBg})
}

func pdfControlRow(r reportRow, zebra bool) core.Row {
	implChip := implementationChip(r.implementationStateEnum)
	evalChip := evaluationChip(r.evaluationStatusEnum)

	evaluationValue := r.evaluationStatus
	if evaluationValue == "" {
		evaluationValue = "Not Evaluated"
	}
	assignee := r.assignee
	if assignee == "" {
		assignee = "—"
	}

	control := col.New(5).Add(
		text.New(fmt.Sprintf("%s  ·  %s", r.shortName, truncate(r.controlName, 42)), props.Text{Top: 2.5, Left: 2, Size: 8.5, Style: fontstyle.Bold, Color: pdfInk}),
	)
	assigneeCol := col.New(2).Add(
		text.New(assignee, props.Text{Top: 2.5, Size: 8, Color: pdfInk}),
	)
	implCol := col.New(2).Add(
		text.New(r.implementationState, props.Text{Top: 2.5, Size: 8, Style: fontstyle.Bold, Align: align.Center, Color: implChip.text}),
	).WithStyle(&props.Cell{BackgroundColor: implChip.bg})

	evalTexts := []core.Component{
		text.New(evaluationValue, props.Text{Top: 2.5, Size: 8, Style: fontstyle.Bold, Align: align.Center, Color: evalChip.text}),
	}
	if r.evaluationTimestamp != "" {
		evalTexts[0] = text.New(evaluationValue, props.Text{Top: 1.5, Size: 8, Style: fontstyle.Bold, Align: align.Center, Color: evalChip.text})
		evalTexts = append(evalTexts, text.New(r.evaluationTimestamp, props.Text{Top: 6, Size: 6.5, Align: align.Center, Color: evalChip.text}))
	}
	evalCol := col.New(3).Add(evalTexts...).WithStyle(&props.Cell{BackgroundColor: evalChip.bg})

	rw := row.New(9).Add(control, assigneeCol, implCol, evalCol)
	if zebra {
		rw.WithStyle(&props.Cell{BackgroundColor: pdfZebra})
	}
	return rw
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
