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

package evaluationtest

import (
	"time"

	"confirmate.io/core/api/assessment"
	"confirmate.io/core/api/evaluation"
	"confirmate.io/core/api/orchestrator"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	// Mock IDs for evaluation results
	MockEvaluationResultId1   = "00000000-0000-0000-0000-000000000001"
	MockEvaluationResultId2   = "00000000-0000-0000-0000-000000000002"
	MockEvaluationResultId3   = "00000000-0000-0000-0000-000000000003"
	MockEvaluationResultId4   = "00000000-0000-0000-0000-000000000004"
	MockEvaluationResultId101 = "00000000-0000-0000-0000-000000000101"
	MockEvaluationResultId102 = "00000000-0000-0000-0000-000000000102"
	MockEvaluationResultId103 = "00000000-0000-0000-0000-000000000103"
	MockEvaluationResultId104 = "00000000-0000-0000-0000-000000000104"

	// Mock IDs for audit scopes
	MockAuditScopeId1 = "00000000-0000-0000-0001-000000000001"
	MockAuditScopeId2 = "00000000-0000-0000-0001-000000000002"

	// Mock IDs for target of evaluations
	MockToeId1 = "00000000-0000-0000-0000-000000000001"
	MockToeId2 = "00000000-0000-0000-0000-000000000002"

	// Mock Catalogs
	MockCatalogId1 = "Catalog 1"
	MockCatalogId2 = "Catalog 2"

	MockCatalogName1 = "Catalog 1"
	MockCatalogName2 = "Catalog 2"

	MockCatalogDescription1 = "Description for Catalog 1"
	MockCatalogDescription2 = "Description for Catalog 2"

	// Mock Controls
	MockCategoryName1 = "Category 1"
	MockCategoryName2 = "Category 2"
	MockCategoryName3 = "Category 3"

	MockControlId1                   = "Control 1"
	MockControlName1                 = "Control Name 1"
	MockControlShortName1            = "control-1"
	MockControlDescription1          = "Description for Control 1"
	MockControl1SubControlShortName1 = "subcontrol-1"
	MockControl1SubcontrolId11       = "Control 1.1"
	MockControl1SubcontrolName11     = "Control Name 1.1"
	MockSubcontrolDescription11      = "Description for Control 1.1"
	MockControl1SubcontrolId12       = "Control 1.2"
	MockControl1SubcontrolName12     = "Subcontrol Name 1.2"
	MockSubcontrolDescription12      = "Description for Control 1.2"

	MockControlId2                      = "Control 2"
	MockControlName2                    = "Control Name 2"
	MockControlShortName2               = "control-2"
	MockControlDescription2             = "Description for Control 2"
	MockControl2SubcontrolID21          = "Control 2.1"
	MockControl2SubControlShortName2    = "subcontrol-2"
	MockControl2SubcontrolName21        = "Control Name 2.1"
	MockControl2SubcontrolDescription21 = "Description for Control 2.1"

	// Mock Metrics
	MockMetricId1 = "Metric 1"
	MockMetricId2 = "Metric 2"
	MockMetricId3 = "Metric 3"

	MockMetricName1 = "Metric Name 1"
	MockMetricName2 = "Metric Name 2"
	MockMetricName3 = "Metric Name 3"

	MockMetricDescription1 = "Description for Metric 1"
	MockMetricDescription2 = "Description for Metric 2"
	MockMetricDescription3 = "Description for Metric 3"

	MockMetricCategory1 = "Metric Category 1"
	MockMetricCategory2 = "Metric Category 2"
	MockMetricCategory3 = "Metric Category 3"

	MockMetricComments1 = "Comments for Metric 1"
	MockMetricComments2 = "Comments for Metric 2"
	MockMetricComments3 = "Comments for Metric 3"

	// Mock Assessment Results
	MockAssessmentResultId1 = "00000000-0000-0000-0000-000000000001"
	MockAssessmentResultId2 = "00000000-0000-0000-0000-000000000002"
	MockAssessmentResultId3 = "00000000-0000-0000-0000-000000000003"

	MockDefaultVersion = "v1"
)

// Mock Evaluation Results
var (
	MockEvaluationResult1 = &evaluation.EvaluationResult{
		Id:                   MockEvaluationResultId1,
		TargetOfEvaluationId: MockToeId1,
		AuditScopeId:         MockAuditScopeId1,
		ControlId:            MockControlId1,
		ControlCatalogId:     MockCatalogId1,
		Status:               evaluation.EvaluationStatus_EVALUATION_STATUS_COMPLIANT,
		Timestamp:            timestamppb.New(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)),
		AssessmentResultIds:  []string{MockAssessmentResultId1, MockAssessmentResultId2},
		Comment:              new("Mock evaluation result 1"),
		Data:                 []byte{},
	}

	MockEvaluationResult2 = &evaluation.EvaluationResult{
		Id:                   MockEvaluationResultId2,
		TargetOfEvaluationId: MockToeId2,
		AuditScopeId:         MockAuditScopeId2,
		ControlId:            MockControlId2,
		ControlCatalogId:     MockCatalogId2,
		Status:               evaluation.EvaluationStatus_EVALUATION_STATUS_NOT_COMPLIANT,
		Timestamp:            timestamppb.New(MockEvaluationResult1.Timestamp.AsTime().Add(5 * time.Minute)),
		AssessmentResultIds:  []string{MockAssessmentResultId3},
		Comment:              new("Mock evaluation result 2"),
		Data:                 []byte{},
	}

	MockEvaluationResult3 = &evaluation.EvaluationResult{
		Id:                   MockEvaluationResultId3,
		TargetOfEvaluationId: MockToeId1,
		AuditScopeId:         MockAuditScopeId1,
		ControlId:            MockControl1SubcontrolId11,
		ParentControlId:      new(MockControlId1),
		ControlCatalogId:     MockCatalogId1,
		Status:               evaluation.EvaluationStatus_EVALUATION_STATUS_COMPLIANT,
		Timestamp:            timestamppb.New(MockEvaluationResult1.Timestamp.AsTime().Add(10 * time.Minute)),
		AssessmentResultIds:  []string{MockAssessmentResultId1},
		Comment:              new("Mock evaluation result 3"),
		Data:                 []byte{},
	}
	MockEvaluationResult4 = &evaluation.EvaluationResult{
		Id:                   MockEvaluationResultId4,
		TargetOfEvaluationId: MockToeId1,
		AuditScopeId:         MockAuditScopeId2,
		ControlId:            MockControlId1,
		ControlCatalogId:     MockCatalogId1,
		Status:               evaluation.EvaluationStatus_EVALUATION_STATUS_COMPLIANT,
		Timestamp:            timestamppb.New(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)),
		AssessmentResultIds:  []string{MockAssessmentResultId1, MockAssessmentResultId2},
		Comment:              new("Mock evaluation result 1"),
		Data:                 []byte{},
	}

	MockManualEvaluationResult1 = &evaluation.EvaluationResult{
		Id:                   MockEvaluationResultId101,
		TargetOfEvaluationId: MockToeId1,
		AuditScopeId:         MockAuditScopeId1,
		ControlId:            MockControlId1,
		ControlCatalogId:     MockCatalogId1,
		ParentControlId:      new(""),
		Status:               evaluation.EvaluationStatus_EVALUATION_STATUS_COMPLIANT_MANUALLY,
		Timestamp:            timestamppb.New(MockEvaluationResult1.Timestamp.AsTime().Add(20 * time.Minute)),
		AssessmentResultIds:  []string{MockAssessmentResultId1, MockAssessmentResultId2},
		ValidUntil:           timestamppb.New(time.Now().Add(48 * time.Hour)),
		Comment:              new("Mock manual evaluation result 1"),
		Data:                 make([]byte, 2*2), // small blob
	}

	MockManualEvaluationResult2 = &evaluation.EvaluationResult{
		Id:                   MockEvaluationResultId102,
		TargetOfEvaluationId: MockToeId2,
		AuditScopeId:         MockAuditScopeId1,
		ControlId:            MockControl1SubcontrolId11,
		ControlCatalogId:     MockCatalogId1,
		Status:               evaluation.EvaluationStatus_EVALUATION_STATUS_COMPLIANT_MANUALLY,
		Timestamp:            timestamppb.New(MockEvaluationResult1.Timestamp.AsTime().Add(25 * time.Minute)),
		AssessmentResultIds:  []string{MockAssessmentResultId1, MockAssessmentResultId2},
		ValidUntil:           timestamppb.New(time.Now().Add(48 * time.Hour)),
		ParentControlId:      new(MockControlId1),
		Comment:              new("Mock manual evaluation result 2"),
		Data:                 make([]byte, 2*2), // small blob
	}

	MockManualEvaluationResult3 = &evaluation.EvaluationResult{
		Id:                   MockEvaluationResultId103,
		TargetOfEvaluationId: MockToeId1,
		AuditScopeId:         MockAuditScopeId2,
		ControlId:            MockControlId1,
		ControlCatalogId:     MockCatalogId1,
		Status:               evaluation.EvaluationStatus_EVALUATION_STATUS_COMPLIANT_MANUALLY,
		Timestamp:            timestamppb.New(MockEvaluationResult1.Timestamp.AsTime().Add(15 * time.Minute)),
		ValidUntil:           timestamppb.New(time.Now().Add(24 * time.Hour)),
		AssessmentResultIds:  []string{MockAssessmentResultId1},
		Comment:              new("Mock evaluation result 3"),
		Data:                 []byte{},
	}

	// Evaluation Result with valid until in the past, should be filtered out when fetching valid evaluation results
	MockManualEvaluationResult4 = &evaluation.EvaluationResult{
		Id:                   MockEvaluationResultId104,
		TargetOfEvaluationId: MockToeId1,
		AuditScopeId:         MockAuditScopeId1,
		ControlId:            MockControlId1,
		ControlCatalogId:     MockCatalogId1,
		Status:               evaluation.EvaluationStatus_EVALUATION_STATUS_COMPLIANT_MANUALLY,
		Timestamp:            timestamppb.New(MockEvaluationResult1.Timestamp.AsTime().Add(-15 * time.Minute)),
		ValidUntil:           timestamppb.New(time.Now().Add(-24 * time.Hour)),
		AssessmentResultIds:  []string{MockAssessmentResultId1},
		Comment:              new("Mock evaluation result 4"),
		Data:                 []byte{},
	}
)

// Mock Assessment Results
var (
	MockAuditScope1 = &orchestrator.AuditScope{
		Id:                   MockAuditScopeId1,
		TargetOfEvaluationId: MockToeId1,
		CatalogId:            MockCatalogId1,
	}
	MockAuditScope2 = &orchestrator.AuditScope{
		Id:                   MockAuditScopeId2,
		TargetOfEvaluationId: MockToeId2,
		CatalogId:            MockCatalogId2,
	}
)

// Mock Catalogs
var (
	// Mock Catalogs
	// MockCatalog1 contains 2 Categories
	// * category1 with 1 control and 2 sub-controls with one metric each
	// * category2 with 1 control and 1 sub-control with one metric
	MockCatalog1 = &orchestrator.Catalog{
		Id:          MockCatalogId1,
		Name:        MockCatalogName1,
		Description: MockCatalogDescription1,
		Categories: []*orchestrator.Category{
			{
				Name:      MockCategoryName1,
				CatalogId: MockCatalogId1,
				Controls: []*orchestrator.Control{
					{
						Id:        MockControlId1,
						Name:      MockControlName1,
						ShortName: MockControlShortName1,
						CatalogId: MockCatalogId1,
						Controls: []*orchestrator.Control{
							{
								Id:              MockControl1SubcontrolId11,
								Name:            MockControl1SubcontrolName11,
								ShortName:       MockControl1SubControlShortName1,
								Metrics:         []*assessment.Metric{MockMetric1},
								ParentControlId: new(MockControlId1),
								AssuranceLevel:  new("high"),
								CatalogId:       MockCatalogId1,
							},
							{
								Id:              MockControl1SubcontrolId12,
								Name:            MockControl1SubcontrolName12,
								ShortName:       MockControl2SubControlShortName2,
								Metrics:         []*assessment.Metric{MockMetric2},
								ParentControlId: new(MockControlId1),
								AssuranceLevel:  new("medium"),
								CatalogId:       MockCatalogId1,
							},
						},
					},
				},
			},
			{
				Name:      MockCategoryName2,
				CatalogId: MockCatalogId1,
				Controls: []*orchestrator.Control{
					{
						Id:        MockControlId2,
						Name:      MockControlName2,
						ShortName: MockControlShortName2,
						CatalogId: MockCatalogId1,
						Controls: []*orchestrator.Control{
							{
								Id:              MockControl2SubcontrolID21,
								Name:            MockControl2SubcontrolName21,
								ShortName:       MockControl1SubControlShortName1,
								Metrics:         []*assessment.Metric{MockMetric3},
								ParentControlId: new(MockControlId2),
								CatalogId:       MockCatalogId1,
							},
						},
					},
				},
			},
		},
	}

	// Mock Metrics
	MockMetric1 = &assessment.Metric{
		Id:          MockMetricId1,
		Name:        MockMetricName1,
		Description: MockMetricDescription1,
		Version:     MockDefaultVersion,
		Category:    MockCategoryName1,
	}
	MockMetric2 = &assessment.Metric{
		Id:          MockMetricId2,
		Name:        MockMetricName2,
		Description: MockMetricDescription2,
		Version:     MockDefaultVersion,
		Category:    MockCategoryName1,
	}
	MockMetric3 = &assessment.Metric{
		Id:          MockMetricId3,
		Name:        MockMetricName3,
		Description: MockMetricDescription3,
		Version:     MockDefaultVersion,
		Category:    MockCategoryName3,
	}

	// MockCatalog1 = &orchestrator.Catalog{
	// 	Id:          MockCatalogId1,
	// 	Name:        MockCatalogName1,
	// 	Description: MockCatalogDescription1,
	// 	AllInScope:  false,
	// 	Categories: []*orchestrator.Category{
	// 		{
	// 			Name: MockCategoryName1,
	// 			Controls: []*orchestrator.Control{
	// 				MockControl1,
	// 				// MockSubcontrol11,
	// 				// MockControl2,
	// 				// MockSubcontrol21,
	// 			},
	// 		},
	// 	},
	// }
)

// Mock Controls
var (
	MockControl1 = &orchestrator.Control{
		Id:          MockControlId1,
		Name:        MockControlName1,
		Description: MockControlDescription1,
		Controls: []*orchestrator.Control{
			MockSubcontrol11,
			MockSubcontrol12,
		}}
	MockSubcontrol11 = &orchestrator.Control{
		Id:          MockControl1SubcontrolId11,
		Name:        MockControl1SubcontrolName11,
		Description: MockSubcontrolDescription11,
		// AssuranceLevel:                 new("basic"),
		ParentControlId: new(MockControlId1),
		Metrics: []*assessment.Metric{{
			Id:          MockMetricId1,
			Name:        MockMetricName1,
			Description: MockMetricDescription1,
			Category:    MockMetricCategory1,
			Version:     MockDefaultVersion,
			Comments:    MockMetricComments1,
		},
		}}
	MockSubcontrol12 = &orchestrator.Control{
		Id:          MockControl1SubcontrolId12,
		Name:        MockControl1SubcontrolName12,
		Description: MockSubcontrolDescription12,
		// AssuranceLevel:                 new("basic"),
		ParentControlId: new(MockControlId1),
		Metrics: []*assessment.Metric{{
			Id:          MockMetricId2,
			Name:        MockMetricName2,
			Description: MockMetricDescription2,
			Category:    MockMetricCategory2,
			Version:     MockDefaultVersion,
			Comments:    MockMetricComments2,
		},
		}}
	MockControl2 = &orchestrator.Control{
		Id:          MockControlId2,
		Name:        MockControlName2,
		Description: MockControlDescription2,
		Controls: []*orchestrator.Control{
			MockSubcontrol21,
		},
	}
	MockSubcontrol21 = &orchestrator.Control{
		Id:              MockControl2SubcontrolID21,
		Name:            MockControl2SubcontrolName21,
		Description:     MockControl2SubcontrolDescription21,
		AssuranceLevel:  new("basic"),
		ParentControlId: new(MockControlId2),
		Metrics: []*assessment.Metric{{
			Id:          MockMetricId2,
			Name:        MockMetricName2,
			Description: MockMetricDescription2,
			Category:    MockMetricCategory2,
			Version:     MockDefaultVersion,
			Comments:    MockMetricComments2,
		},
		}}
)

// types contains all types that we need to auto-migrate the catalogs and controls into database tables
var TypesCatalog = []any{
	&orchestrator.Catalog{},
	&orchestrator.Category{},
	&orchestrator.Catalog_Metadata{},
	&orchestrator.Control{},
	&assessment.Metric{},
	&evaluation.EvaluationResult{},
}
