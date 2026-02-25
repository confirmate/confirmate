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

package orchestratortest

import (
	"strconv"

	"confirmate.io/core/api/assessment"
	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/util"

	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Mock UUIDs for consistent testing
const (
	MockToeId1        = "00000000-0000-0000-0000-000000000001"
	MockToeId2        = "00000000-0000-0000-0000-000000000002"
	MockToeId3        = "00000000-0000-0000-0000-000000000003"
	MockScopeId1      = "00000000-0000-0000-0001-000000000001"
	MockScopeId2      = "00000000-0000-0000-0001-000000000002"
	MockResultId1     = "00000000-0000-0000-0002-000000000001"
	MockResultId2     = "00000000-0000-0000-0002-000000000002"
	MockResultId3     = "00000000-0000-0000-0002-000000000003"
	MockEvidenceId1   = "00000000-0000-0000-0003-000000000001"
	MockEvidenceId2   = "00000000-0000-0000-0003-000000000002"
	MockNonExistentId = "00000000-0000-0000-ffff-ffffffffffff"
	MockEmptyUuid     = "00000000-0000-0000-0000-000000000000"
	MockMetricId1     = "00000000-0000-0000-0000-000000000001"
	MockMetricId2     = "00000000-0000-0000-0000-000000000002"
	MockMetricId3     = "00000000-0000-0000-0000-000000000003"
)

// Mock strings for consistent testing
const (
	MockMetricName1             = "Mock Metric 1"
	MockMetricName2             = "Mock Metric 2"
	MockMetricName3             = "Mock Metric 3"
	MockMetricDescription1      = "Mock Metric Description 1"
	MockMetricDescription2      = "Mock Metric Description 2"
	MockMetricDescription3      = "Mock Metric Description 3"
	MockMetricIdDefault         = "metric-default"
	MockCatalogId1              = "catalog-1"
	MockCatalogId2              = "catalog-2"
	MockCatalogId3              = "catalog-3"
	MockCatalogName1            = "Mock Catalog 1"
	MockCatalogName2            = "Mock Catalog 2"
	MockCatalogName3            = "Mock Catalog 3"
	MockCatalogDescription1     = "Mock catalog description 1"
	MockCatalogDescription2     = "Mock catalog description 2"
	MockCatalogDescription3     = "Mock catalog description 3"
	MockCategoryName1           = "category-1"
	MockCategoryName2           = "category-2"
	MockControlId1              = "control-1"
	MockControlId2              = "control-2"
	MockSubControlId1           = "control-1-1"
	MockSubControlName1         = "Mock Sub-Control 1"
	MockCertificateId1          = "certificate-1"
	MockCertificateId2          = "certificate-2"
	MockCertifiateName1         = "Mock Certificate 1"
	MockCertifiateName2         = "Mock Certificate 2"
	MockCertificateDescription1 = "Mock certificate description 1"
	MockCertificateDescription2 = "Mock certificate description 2"
	MockResourceId1             = "resource-1"
	MockResourceId2             = "resource-2"
	MockResourceIdNew           = "resource-new"
	MockResourceId3             = "resource-3"
	MockToolId1                 = "tool-1"
	MockToolId2                 = "tool-2"
	MockToolName1               = "Mock Tool 1"
	MockToolName2               = "Mock Tool 2"
	MockToolDescription1        = "Mock assessment tool"
	MockToolDescription2        = "Mock assessment tool"
	MockToolIdConcurrent        = "tool-concurrent"
	MockTestCategory            = "test-category"
	MockCompliantComment        = "Resource is compliant"
	MockNotCompliantComment     = "Resource is not compliant"
	MockDefaultVersion          = "v1"
)

var (
	// Mock Metrics
	MockMetric1 = &assessment.Metric{
		Id:          MockMetricId1,
		Name:        MockMetricName1,
		Description: MockMetricDescription1,
		Version:     MockDefaultVersion,
		Category:    MockTestCategory,
	}
	MockMetric2 = &assessment.Metric{
		Id:          MockMetricId2,
		Name:        MockMetricName2,
		Description: MockMetricDescription2,
		Version:     MockDefaultVersion,
		Category:    MockTestCategory,
	}
	MockMetric3 = &assessment.Metric{
		Id:          MockMetricId3,
		Name:        MockMetricName3,
		Description: MockMetricDescription3,
		Version:     MockDefaultVersion,
		Category:    MockTestCategory,
	}
	MockMetricWithDefault = &assessment.Metric{
		Id:          MockMetricIdDefault,
		Description: "Mock Metric with Default Config",
		Version:     MockDefaultVersion,
		Category:    MockTestCategory,
	}

	// Mock Metric Implementations
	MockMetricImplementation1 = &assessment.MetricImplementation{
		MetricId: MockMetricId1,
		Lang:     assessment.MetricImplementation_LANGUAGE_REGO,
		Code:     "mock implementation code",
	}

	// Mock Metric Configurations
	MockMetricConfiguration1 = &assessment.MetricConfiguration{
		TargetOfEvaluationId: MockToeId1,
		MetricId:             MockMetricId1,
		Operator:             "==",
		TargetValue:          structpb.NewBoolValue(true),
		IsDefault:            true,
	}
	MockMetricConfiguration2 = &assessment.MetricConfiguration{
		TargetOfEvaluationId: MockToeId1,
		MetricId:             MockMetricId2,
		Operator:             "==",
		TargetValue:          structpb.NewBoolValue(true),
		IsDefault:            false,
	}
	MockMetricConfiguration3 = &assessment.MetricConfiguration{
		TargetOfEvaluationId: MockToeId1,
		MetricId:             MockMetricId3,
		Operator:             "==",
		TargetValue:          structpb.NewBoolValue(true),
		IsDefault:            false,
	}
	MockMetricConfigurationDefault = &assessment.MetricConfiguration{
		MetricId:    MockMetricIdDefault,
		Operator:    "==",
		TargetValue: structpb.NewBoolValue(true),
		IsDefault:   true,
	}

	// Mock Targets of Evaluation
	MockTargetOfEvaluation1 = &orchestrator.TargetOfEvaluation{
		Id:         MockToeId1,
		Name:       "Mock TOE 1",
		TargetType: orchestrator.TargetOfEvaluation_TARGET_TYPE_CLOUD,
	}
	MockTargetOfEvaluation2 = &orchestrator.TargetOfEvaluation{
		Id:         MockToeId2,
		Name:       "Mock TOE 2",
		TargetType: orchestrator.TargetOfEvaluation_TARGET_TYPE_CLOUD,
	}

	// Mock Catalogs
	MockCatalog1 = &orchestrator.Catalog{
		Id:          MockCatalogId1,
		Name:        MockCatalogName1,
		Description: MockCatalogDescription1,
	}
	MockCatalog2 = &orchestrator.Catalog{
		Id:          MockCatalogId2,
		Name:        MockCatalogName2,
		Description: MockCatalogDescription2,
	}
	MockCatalog3 = &orchestrator.Catalog{
		Id:          MockCatalogId3,
		Name:        MockCatalogName3,
		Description: MockCatalogDescription3,
	}

	// Mock Categories
	MockCategory1 = &orchestrator.Category{
		Name:      MockCategoryName1,
		CatalogId: MockCatalogId1,
	}
	MockCategory2 = &orchestrator.Category{
		Name:      MockCategoryName2,
		CatalogId: MockCatalogId2,
	}

	// Mock Controls
	MockControl1 = &orchestrator.Control{
		Id:                MockControlId1,
		CategoryName:      MockCategoryName1,
		CategoryCatalogId: MockCatalogId1,
		Controls:          []*orchestrator.Control{MockSubControl1},
	}
	MockSubControl1 = &orchestrator.Control{
		Id:                MockSubControlId1,
		CategoryName:      MockCategoryName1,
		CategoryCatalogId: MockCatalogId1,
		Name:              MockSubControlName1,
		Metrics: []*assessment.Metric{{
			Id:       MockMetricId1,
			Version:  MockDefaultVersion,
			Comments: "This is a comment",
		}},
		ParentControlId:                util.Ref(MockControlId1),
		ParentControlCategoryName:      util.Ref(MockCategoryName1),
		ParentControlCategoryCatalogId: util.Ref(MockCatalogId1),
	}
	MockControl2 = &orchestrator.Control{
		Id:                MockControlId2,
		CategoryName:      MockCategoryName2,
		CategoryCatalogId: MockCatalogId2,
	}

	// Mock Certificates
	MockCertificate1 = &orchestrator.Certificate{
		Id:                   MockCertificateId1,
		Name:                 MockCertifiateName1,
		Description:          MockCertificateDescription1,
		TargetOfEvaluationId: MockToeId1,
	}
	MockCertificate2 = &orchestrator.Certificate{
		Id:                   MockCertificateId2,
		Name:                 MockCertifiateName2,
		Description:          MockCertificateDescription2,
		TargetOfEvaluationId: MockToeId2,
	}

	// Mock Assessment Tools
	MockAssessmentTool1 = &orchestrator.AssessmentTool{
		Id:          MockToolId1,
		Name:        MockToolName1,
		Description: MockToolDescription1,
		AvailableMetrics: []string{
			MockMetricId1,
		},
	}
	MockAssessmentTool2 = &orchestrator.AssessmentTool{
		Id:          MockToolId2,
		Name:        MockToolName2,
		Description: MockToolDescription2,
		AvailableMetrics: []string{
			MockMetricId2,
		},
	}

	// Mock Audit Scopes
	MockAuditScope1 = &orchestrator.AuditScope{
		Id:                   MockScopeId1,
		TargetOfEvaluationId: MockToeId1,
		CatalogId:            MockCatalogId1,
	}
	MockAuditScope2 = &orchestrator.AuditScope{
		Id:                   MockScopeId2,
		TargetOfEvaluationId: MockToeId2,
		CatalogId:            MockCatalogId2,
	}

	// Mock Assessment Results
	MockAssessmentResult1 = &assessment.AssessmentResult{
		Id:                   MockResultId1,
		CreatedAt:            timestamppb.Now(),
		MetricId:             MockMetricId1,
		MetricConfiguration:  MockMetricConfiguration1,
		Compliant:            true,
		EvidenceId:           MockEvidenceId1,
		ResourceId:           MockResourceId1,
		ResourceTypes:        []string{"vm"},
		ComplianceComment:    MockCompliantComment,
		TargetOfEvaluationId: MockToeId1,
		ToolId:               util.Ref(MockToolId1),
		HistoryUpdatedAt:     timestamppb.Now(),
		History: []*assessment.Record{
			{
				EvidenceId:         MockEvidenceId1,
				EvidenceRecordedAt: timestamppb.Now(),
			},
		},
	}
	MockAssessmentResult2 = &assessment.AssessmentResult{
		Id:                   MockResultId2,
		CreatedAt:            timestamppb.Now(),
		MetricId:             MockMetricId2,
		MetricConfiguration:  MockMetricConfiguration2,
		Compliant:            false,
		EvidenceId:           MockEvidenceId2,
		ResourceId:           MockResourceId2,
		ResourceTypes:        []string{"storage"},
		ComplianceComment:    MockNotCompliantComment,
		TargetOfEvaluationId: MockToeId1,
		ToolId:               util.Ref(MockToolId1),
		HistoryUpdatedAt:     timestamppb.Now(),
		History: []*assessment.Record{
			{
				EvidenceId:         MockEvidenceId2,
				EvidenceRecordedAt: timestamppb.Now(),
			},
		},
	}

	// Mock Assessment Results for Store tests
	MockNewAssessmentResult = &assessment.AssessmentResult{
		Id:                   MockResultId3,
		CreatedAt:            timestamppb.Now(),
		MetricId:             MockMetricId1,
		MetricConfiguration:  MockMetricConfiguration1,
		Compliant:            true,
		EvidenceId:           MockEvidenceId1,
		ResourceId:           MockResourceIdNew,
		ResourceTypes:        []string{"vm"},
		ComplianceComment:    MockCompliantComment,
		TargetOfEvaluationId: MockToeId1,
		ToolId:               util.Ref(MockToolId1),
		HistoryUpdatedAt:     timestamppb.Now(),
		History: []*assessment.Record{
			{
				EvidenceId:         MockEvidenceId1,
				EvidenceRecordedAt: timestamppb.Now(),
			},
		},
	}

	// MockAssessmentResult3 for integration testing - can be reused for additional result in streams
	MockAssessmentResult3 = &assessment.AssessmentResult{
		Id:                   "00000000-0000-0000-0003-000000000005",
		CreatedAt:            timestamppb.Now(),
		MetricId:             MockMetricId3,
		MetricConfiguration:  MockMetricConfiguration3,
		Compliant:            true,
		EvidenceId:           MockEvidenceId1,
		ResourceId:           MockResourceId3,
		ResourceTypes:        []string{"compute"},
		ComplianceComment:    "Third resource test",
		TargetOfEvaluationId: MockToeId1,
		ToolId:               util.Ref(MockToolId1),
		HistoryUpdatedAt:     timestamppb.Now(),
		History: []*assessment.Record{
			{
				EvidenceId:         MockEvidenceId1,
				EvidenceRecordedAt: timestamppb.Now(),
			},
		},
	}

	// MockAssessmentResultForDuplicate - result with id that can be pre-created to test duplicate errors
	MockAssessmentResultForDuplicate = &assessment.AssessmentResult{
		Id:                   "00000000-0000-0000-0002-000000000004",
		CreatedAt:            timestamppb.Now(),
		MetricId:             MockMetricId2,
		MetricConfiguration:  MockMetricConfiguration2,
		Compliant:            false,
		EvidenceId:           MockEvidenceId2,
		ResourceId:           MockResourceId2,
		ResourceTypes:        []string{"vm"},
		ComplianceComment:    "Duplicate test",
		TargetOfEvaluationId: MockToeId1,
		ToolId:               util.Ref(MockToolId1),
		HistoryUpdatedAt:     timestamppb.Now(),
		History: []*assessment.Record{
			{
				EvidenceId:         MockEvidenceId2,
				EvidenceRecordedAt: timestamppb.Now(),
			},
		},
	}
)

// NewMockAssessmentResultForConcurrentStream creates a unique assessment result for concurrent stream testing
// with a unique id based on the stream id
func NewMockAssessmentResultForConcurrentStream(streamID int) *assessment.AssessmentResult {
	// Create a valid UUID by using stream id as the last part
	// Format: 00000000-0000-0000-00cc-0000000000XX where XX is the stream id
	idSuffix := strconv.Itoa(streamID)
	if len(idSuffix) == 1 {
		idSuffix = "0" + idSuffix
	}
	validUUID := "00000000-0000-0000-00cc-0000000000" + idSuffix

	return &assessment.AssessmentResult{
		Id:                   validUUID,
		CreatedAt:            timestamppb.Now(),
		MetricId:             MockMetricId1,
		MetricConfiguration:  MockMetricConfiguration1,
		Compliant:            true,
		EvidenceId:           MockEvidenceId1,
		ResourceId:           "resource-concurrent-" + idSuffix,
		ResourceTypes:        []string{"compute"},
		ComplianceComment:    "Concurrent stream test",
		TargetOfEvaluationId: MockToeId1,
		ToolId:               util.Ref(MockToolIdConcurrent),
		HistoryUpdatedAt:     timestamppb.Now(),
		History: []*assessment.Record{
			{
				EvidenceId:         MockEvidenceId1,
				EvidenceRecordedAt: timestamppb.Now(),
			},
		},
	}
}
