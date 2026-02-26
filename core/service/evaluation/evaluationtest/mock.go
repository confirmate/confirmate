package evaluationtest

import (
	"time"

	"confirmate.io/core/api/evaluation"
	"confirmate.io/core/util"
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

	// Mock IDs for catalogs
	MockCatalogId1 = "catalog-1"
	MockCatalogId2 = "catalog-2"

	// Mock names for control category
	MockCategoryName1 = "category-1"
	MockCategoryName2 = "category-2"

	// Mock IDs for controls
	MockControlId1  = "control-1"
	MockControlId11 = "control-1.1"
	MockControlId2  = "control-2"

	// Mock Ids for assessment results
	MockAssessmentResultId1 = "00000000-0000-0000-0000-000000000001"
	MockAssessmentResultId2 = "00000000-0000-0000-0000-000000000002"
	MockAssessmentResultId3 = "00000000-0000-0000-0000-000000000003"
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
		Timestamp:            timestamppb.Now(),
		AssessmentResultIds:  []string{MockAssessmentResultId1, MockAssessmentResultId2},
		Comment:              util.Ref("Mock evaluation result 1"),
		Data:                 make([]byte, 1024*1024), // 1MB data blob
	}

	MockEvaluationResult2 = &evaluation.EvaluationResult{
		Id:                   MockEvaluationResultId2,
		TargetOfEvaluationId: MockToeId2,
		AuditScopeId:         MockAuditScopeId2,
		ControlId:            MockControlId2,
		ControlCategoryName:  MockCategoryName2,
		ControlCatalogId:     MockCatalogId2,
		Status:               evaluation.EvaluationStatus_EVALUATION_STATUS_NOT_COMPLIANT,
		Timestamp:            timestamppb.Now(),
		AssessmentResultIds:  []string{MockAssessmentResultId3},
		Comment:              util.Ref("Mock evaluation result 2"),
		Data:                 []byte{},
	}

	MockEvaluationResult3 = &evaluation.EvaluationResult{
		Id:                   MockEvaluationResultId3,
		TargetOfEvaluationId: MockToeId1,
		AuditScopeId:         MockAuditScopeId1,
		ControlId:            MockControlId11,
		ParentControlId:      util.Ref(MockControlId1),
		ControlCategoryName:  MockCategoryName1,
		ControlCatalogId:     MockCatalogId1,
		Status:               evaluation.EvaluationStatus_EVALUATION_STATUS_COMPLIANT,
		Timestamp:            timestamppb.Now(),
		AssessmentResultIds:  []string{MockAssessmentResultId1},
		Comment:              util.Ref("Mock evaluation result 3"),
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
		Timestamp:            timestamppb.Now(),
		ValidUntil:           timestamppb.New(time.Now().Add(24 * time.Hour)),
		AssessmentResultIds:  []string{MockAssessmentResultId1},
		Comment:              util.Ref("Mock evaluation result 4"),
		Data:                 []byte{},
	}
	MockManualEvaluationResult1 = &evaluation.EvaluationResult{
		Id:                   MockEvaluationResultId5,
		TargetOfEvaluationId: MockToeId1,
		AuditScopeId:         MockAuditScopeId1,
		ControlId:            MockControlId11,
		ControlCategoryName:  MockCategoryName1,
		ControlCatalogId:     MockCatalogId1,
		Status:               evaluation.EvaluationStatus_EVALUATION_STATUS_COMPLIANT_MANUALLY,
		Timestamp:            timestamppb.Now(),
		AssessmentResultIds:  []string{MockAssessmentResultId1, MockAssessmentResultId2},
		ValidUntil:           timestamppb.New(time.Now().Add(48 * time.Hour)),
		ParentControlId:      util.Ref(MockControlId1),
		Comment:              util.Ref("Mock manual evaluation result 1"),
		Data:                 make([]byte, 1024*1024), // 1MB data blob
	}
	// MockManualEvaluationResult2 is identical to MockManualEvaluationResult1 except for the ID. The ID is missing.
	MockManualEvaluationResult2 = &evaluation.EvaluationResult{
		TargetOfEvaluationId: MockToeId1,
		AuditScopeId:         MockAuditScopeId1,
		ControlId:            MockControlId11,
		ControlCategoryName:  MockCategoryName1,
		ControlCatalogId:     MockCatalogId1,
		Status:               evaluation.EvaluationStatus_EVALUATION_STATUS_COMPLIANT_MANUALLY,
		Timestamp:            timestamppb.Now(),
		AssessmentResultIds:  []string{MockAssessmentResultId1, MockAssessmentResultId2},
		ValidUntil:           timestamppb.New(time.Now().Add(48 * time.Hour)),
		ParentControlId:      util.Ref(MockControlId1),
		Comment:              util.Ref("Mock manual evaluation result 1"),
		Data:                 make([]byte, 1024*1024), // 1MB data blob
	}
)
