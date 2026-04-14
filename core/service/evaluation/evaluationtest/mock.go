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
	MockEvaluationResultId1 = "00000000-0000-0000-0000-000000000001"
	MockEvaluationResultId2 = "00000000-0000-0000-0000-000000000002"
	MockEvaluationResultId3 = "00000000-0000-0000-0000-000000000003"
	MockEvaluationResultId4 = "00000000-0000-0000-0000-000000000004"
	MockEvaluationResultId5 = "00000000-0000-0000-0000-000000000005"

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

	MockControlId1     = "Control 1"
	MockControlId2     = "Control 2"
	MockSubcontrolId11 = "Control 1.1"
	MockSubcontrolId12 = "Control 1.2"
	MockSubcontrolID21 = "Control 2.1"

	MockControlName1     = "Control Name 1"
	MockControlName2     = "Control Name 2"
	MockSubcontrolName11 = "Control Name 1.1"
	MockSubcontrolName12 = "Control Name 1.2"
	MockSubcontrolName21 = "Control Name 2.1"

	MockControlDescription1     = "Description for Control 1"
	MockControlDescription2     = "Description for Control 2"
	MockSubcontrolDescription11 = "Description for Control 1.1"
	MockSubcontrolDescription12 = "Description for Control 1.2"
	MockSubcontrolDescription21 = "Description for Control 2.1"

	// Mock Metrics
	MockMetricId1 = "Metric 1"
	MockMetricId2 = "Metric 2"

	MockMetricName1 = "Metric Name 1"
	MockMetricName2 = "Metric Name 2"

	MockMetricDescription1 = "Description for Metric 1"
	MockMetricDescription2 = "Description for Metric 2"

	MockMetricCategory1 = "Metric Category 1"
	MockMetricCategory2 = "Metric Category 2"

	MockMetricComments1 = "Comments for Metric 1"
	MockMetricComments2 = "Comments for Metric 2"

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
		ControlCategoryName:  MockCategoryName1,
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
		ControlCategoryName:  MockCategoryName2,
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
		ControlId:            MockSubcontrolId11,
		ParentControlId:      new(MockControlId1),
		ControlCategoryName:  MockCategoryName1,
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
		AuditScopeId:         MockAuditScopeId1,
		ControlId:            MockControlId1,
		ControlCategoryName:  MockCategoryName1,
		ControlCatalogId:     MockCatalogId1,
		Status:               evaluation.EvaluationStatus_EVALUATION_STATUS_COMPLIANT_MANUALLY,
		Timestamp:            timestamppb.New(MockEvaluationResult1.Timestamp.AsTime().Add(15 * time.Minute)),
		ValidUntil:           timestamppb.New(time.Now().Add(24 * time.Hour)),
		AssessmentResultIds:  []string{MockAssessmentResultId1},
		Comment:              new("Mock evaluation result 4"),
		Data:                 []byte{},
	}
	MockManualEvaluationResult1 = &evaluation.EvaluationResult{
		Id:                   MockEvaluationResultId5,
		TargetOfEvaluationId: MockToeId1,
		AuditScopeId:         MockAuditScopeId1,
		ControlId:            MockControlId1,
		ControlCategoryName:  MockCategoryName1,
		ControlCatalogId:     MockCatalogId1,
		Status:               evaluation.EvaluationStatus_EVALUATION_STATUS_COMPLIANT_MANUALLY,
		Timestamp:            timestamppb.New(MockEvaluationResult1.Timestamp.AsTime().Add(20 * time.Minute)),
		AssessmentResultIds:  []string{MockAssessmentResultId1, MockAssessmentResultId2},
		ValidUntil:           timestamppb.New(time.Now().Add(48 * time.Hour)),
		Comment:              new("Mock manual evaluation result 1"),
		Data:                 make([]byte, 2*2), // small blob
	}
	// MockManualEvaluationResult2 is identical to MockManualEvaluationResult1 except for the ID. The ID is missing.
	MockManualEvaluationResult2 = &evaluation.EvaluationResult{
		TargetOfEvaluationId: MockToeId1,
		AuditScopeId:         MockAuditScopeId1,
		ControlId:            MockSubcontrolId11,
		ControlCategoryName:  MockCategoryName1,
		ControlCatalogId:     MockCatalogId1,
		Status:               evaluation.EvaluationStatus_EVALUATION_STATUS_COMPLIANT_MANUALLY,
		Timestamp:            timestamppb.New(MockEvaluationResult1.Timestamp.AsTime().Add(25 * time.Minute)),
		AssessmentResultIds:  []string{MockAssessmentResultId1, MockAssessmentResultId2},
		ValidUntil:           timestamppb.New(time.Now().Add(48 * time.Hour)),
		ParentControlId:      new(MockControlId1),
		Comment:              new("Mock manual evaluation result 1"),
		Data:                 make([]byte, 2*2), // small blob
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
	MockCatalog1 = &orchestrator.Catalog{
		Id:          MockCatalogId1,
		Name:        MockCatalogName1,
		Description: MockCatalogDescription1,
		AllInScope:  false,
		Categories: []*orchestrator.Category{
			{
				Name: MockCategoryName1,
				Controls: []*orchestrator.Control{
					MockControl1,
					// MockSubcontrol11,
					// MockControl2,
					// MockSubcontrol21,
				},
			},
		},
	}
)

// Mock Controls
var (
	MockControl1 = &orchestrator.Control{
		Id:                MockControlId1,
		Name:              MockControlName1,
		CategoryName:      MockCategoryName1,
		CategoryCatalogId: MockCatalogId1,
		Description:       MockControlDescription1,
		Controls: []*orchestrator.Control{
			MockSubcontrol11,
			MockSubcontrol12,
			// {
			// Id:   MockSubcontrolId11,
			// Name: MockSubcontrolName11,
			// 	CategoryName:                   MockCategoryName1,
			// 	CategoryCatalogId:              MockCatalogId1,
			// 	Description:                    MockSubcontrolDescription11,
			// 	AssuranceLevel:                 new("basic"),
			// 	ParentControlId:                new(MockControlId1),
			// 	ParentControlCategoryName:      new(MockCategoryName1),
			// 	ParentControlCategoryCatalogId: new(MockCatalogId1),
			// 	Metrics: []*assessment.Metric{{
			// 		Id:          MockMetricId1,
			// 		Name:        MockMetricName1,
			// 		Description: MockMetricDescription1,
			// 		Category:    MockMetricCategory1,
			// 		Version:     MockDefaultVersion,
			// 		Comments:    new(MockMetricComments1),
			// }}},
		}}
	MockSubcontrol11 = &orchestrator.Control{
		Id:                MockSubcontrolId11,
		Name:              MockSubcontrolName11,
		CategoryName:      MockCategoryName1,
		CategoryCatalogId: MockCatalogId1,
		Description:       MockSubcontrolDescription11,
		// AssuranceLevel:                 new("basic"),
		ParentControlId:                new(MockControlId1),
		ParentControlCategoryName:      new(MockCategoryName1),
		ParentControlCategoryCatalogId: new(MockCatalogId1),
		Metrics: []*assessment.Metric{{
			Id:          MockMetricId1,
			Name:        MockMetricName1,
			Description: MockMetricDescription1,
			Category:    MockMetricCategory1,
			Version:     MockDefaultVersion,
			Comments:    new(MockMetricComments1),
		},
		}}
	MockSubcontrol12 = &orchestrator.Control{
		Id:                MockSubcontrolId12,
		Name:              MockSubcontrolName12,
		CategoryName:      MockCategoryName1,
		CategoryCatalogId: MockCatalogId1,
		Description:       MockSubcontrolDescription12,
		// AssuranceLevel:                 new("basic"),
		ParentControlId:                new(MockControlId1),
		ParentControlCategoryName:      new(MockCategoryName1),
		ParentControlCategoryCatalogId: new(MockCatalogId1),
		Metrics: []*assessment.Metric{{
			Id:          MockMetricId2,
			Name:        MockMetricName2,
			Description: MockMetricDescription2,
			Category:    MockMetricCategory2,
			Version:     MockDefaultVersion,
			Comments:    new(MockMetricComments2),
		},
		}}
	MockControl2 = &orchestrator.Control{
		Id:                MockControlId2,
		Name:              MockControlName2,
		CategoryName:      MockCategoryName1,
		CategoryCatalogId: MockCatalogId1,
		Description:       MockControlDescription2,
		Controls: []*orchestrator.Control{
			MockSubcontrol21,
			// {
			// Id:                             MockSubcontrolID21,
			// Name:                           MockControlName2,
			// CategoryName:                   MockCategoryName1,
			// CategoryCatalogId:              MockCatalogId1,
			// Description:                    MockControlDescription2,
			// AssuranceLevel:                 new("basic"),
			// ParentControlId:                new(MockControlId2),
			// ParentControlCategoryName:      new(MockCategoryName1),
			// ParentControlCategoryCatalogId: new(MockCatalogId1),
			// Metrics: []*assessment.Metric{{
			// 	Id:          MockMetricId2,
			// 	Name:        MockMetricName2,
			// 	Description: MockMetricDescription2,
			// 	Category:    MockMetricCategory2,
			// 	Version:     MockDefaultVersion,
			// 	Comments:    new("This is a comment"),
			// }},
			// },
		},
	}
	MockSubcontrol21 = &orchestrator.Control{
		Id:                             MockSubcontrolID21,
		Name:                           MockSubcontrolName21,
		CategoryName:                   MockCategoryName1,
		CategoryCatalogId:              MockCatalogId1,
		Description:                    MockSubcontrolDescription21,
		AssuranceLevel:                 new("basic"),
		ParentControlId:                new(MockControlId2),
		ParentControlCategoryName:      new(MockCategoryName1),
		ParentControlCategoryCatalogId: new(MockCatalogId1),
		Metrics: []*assessment.Metric{{
			Id:          MockMetricId2,
			Name:        MockMetricName2,
			Description: MockMetricDescription2,
			Category:    MockMetricCategory2,
			Version:     MockDefaultVersion,
			Comments:    new(MockMetricComments2),
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
