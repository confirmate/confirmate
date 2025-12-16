package orchestratortest

import (
	"confirmate.io/core/api/assessment"
	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/util"

	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Mock UUIDs for consistent testing
const (
	MockToeID1        = "00000000-0000-0000-0000-000000000001"
	MockToeID2        = "00000000-0000-0000-0000-000000000002"
	MockScopeID1      = "00000000-0000-0000-0001-000000000001"
	MockScopeID2      = "00000000-0000-0000-0001-000000000002"
	MockResultID1     = "00000000-0000-0000-0002-000000000001"
	MockResultID2     = "00000000-0000-0000-0002-000000000002"
	MockResultID3     = "00000000-0000-0000-0002-000000000003"
	MockEvidenceID1   = "00000000-0000-0000-0003-000000000001"
	MockEvidenceID2   = "00000000-0000-0000-0003-000000000002"
	MockNonExistentID = "00000000-0000-0000-ffff-ffffffffffff"
)

var (
	// Mock Metrics
	MockMetric1 = &assessment.Metric{
		Id:          "metric-1",
		Description: "Mock Metric 1",
		Version:     "1.0.0",
		Category:    "test-category",
	}
	MockMetric2 = &assessment.Metric{
		Id:          "metric-2",
		Description: "Mock Metric 2",
		Version:     "1.0.0",
		Category:    "test-category",
	}

	// Mock Metric Implementations
	MockMetricImplementation1 = &assessment.MetricImplementation{
		MetricId: "metric-1",
		Lang:     assessment.MetricImplementation_LANGUAGE_REGO,
		Code:     "mock implementation code",
	}

	// Mock Metric Configurations
	MockMetricConfiguration1 = &assessment.MetricConfiguration{
		TargetOfEvaluationId: MockToeID1,
		MetricId:             "metric-1",
		Operator:             "==",
		TargetValue:          structpb.NewBoolValue(true),
		IsDefault:            true,
	}
	MockMetricConfiguration2 = &assessment.MetricConfiguration{
		TargetOfEvaluationId: MockToeID1,
		MetricId:             "metric-2",
		Operator:             "==",
		TargetValue:          structpb.NewBoolValue(true),
		IsDefault:            false,
	}

	// Mock Targets of Evaluation
	MockTargetOfEvaluation1 = &orchestrator.TargetOfEvaluation{
		Id:         MockToeID1,
		Name:       "Mock TOE 1",
		TargetType: orchestrator.TargetOfEvaluation_TARGET_TYPE_CLOUD,
	}
	MockTargetOfEvaluation2 = &orchestrator.TargetOfEvaluation{
		Id:         MockToeID2,
		Name:       "Mock TOE 2",
		TargetType: orchestrator.TargetOfEvaluation_TARGET_TYPE_CLOUD,
	}

	// Mock Catalogs
	MockCatalog1 = &orchestrator.Catalog{
		Id:          "catalog-1",
		Name:        "Mock Catalog 1",
		Description: "Mock catalog description 1",
	}
	MockCatalog2 = &orchestrator.Catalog{
		Id:          "catalog-2",
		Name:        "Mock Catalog 2",
		Description: "Mock catalog description 2",
	}

	// Mock Categories
	MockCategory1 = &orchestrator.Category{
		Name:      "category-1",
		CatalogId: "catalog-1",
	}
	MockCategory2 = &orchestrator.Category{
		Name:      "category-2",
		CatalogId: "catalog-2",
	}

	// Mock Controls
	MockControl1 = &orchestrator.Control{
		Id:                "control-1",
		CategoryName:      "category-1",
		CategoryCatalogId: "catalog-1",
	}
	MockControl2 = &orchestrator.Control{
		Id:                "control-2",
		CategoryName:      "category-2",
		CategoryCatalogId: "catalog-2",
	}

	// Mock Certificates
	MockCertificate1 = &orchestrator.Certificate{
		Id:                   "cert-1",
		Name:                 "Mock Certificate 1",
		Description:          "Mock certificate description 1",
		TargetOfEvaluationId: MockToeID1,
	}
	MockCertificate2 = &orchestrator.Certificate{
		Id:                   "cert-2",
		Name:                 "Mock Certificate 2",
		Description:          "Mock certificate description 2",
		TargetOfEvaluationId: MockToeID2,
	}

	// Mock Audit Scopes
	MockAuditScope1 = &orchestrator.AuditScope{
		Id:                   MockScopeID1,
		TargetOfEvaluationId: MockToeID1,
		CatalogId:            "catalog-1",
	}
	MockAuditScope2 = &orchestrator.AuditScope{
		Id:                   MockScopeID2,
		TargetOfEvaluationId: MockToeID2,
		CatalogId:            "catalog-2",
	}

	// Mock Assessment Results
	MockAssessmentResult1 = &assessment.AssessmentResult{
		Id:                   MockResultID1,
		CreatedAt:            timestamppb.Now(),
		MetricId:             "metric-1",
		MetricConfiguration:  MockMetricConfiguration1,
		Compliant:            true,
		EvidenceId:           MockEvidenceID1,
		ResourceId:           "resource-1",
		ResourceTypes:        []string{"vm"},
		ComplianceComment:    "Resource is compliant",
		TargetOfEvaluationId: MockToeID1,
		ToolId:               util.Ref("tool-1"),
		HistoryUpdatedAt:     timestamppb.Now(),
		History: []*assessment.Record{
			{
				EvidenceId:         MockEvidenceID1,
				EvidenceRecordedAt: timestamppb.Now(),
			},
		},
	}
	MockAssessmentResult2 = &assessment.AssessmentResult{
		Id:                   MockResultID2,
		CreatedAt:            timestamppb.Now(),
		MetricId:             "metric-2",
		MetricConfiguration:  MockMetricConfiguration2,
		Compliant:            false,
		EvidenceId:           MockEvidenceID2,
		ResourceId:           "resource-2",
		ResourceTypes:        []string{"storage"},
		ComplianceComment:    "Resource is not compliant",
		TargetOfEvaluationId: MockToeID1,
		ToolId:               util.Ref("tool-1"),
		HistoryUpdatedAt:     timestamppb.Now(),
		History: []*assessment.Record{
			{
				EvidenceId:         MockEvidenceID2,
				EvidenceRecordedAt: timestamppb.Now(),
			},
		},
	}

	// Mock Assessment Results for Store tests
	MockNewAssessmentResult = &assessment.AssessmentResult{
		Id:                   MockResultID3,
		CreatedAt:            timestamppb.Now(),
		MetricId:             "metric-1",
		MetricConfiguration:  MockMetricConfiguration1,
		Compliant:            true,
		EvidenceId:           MockEvidenceID1,
		ResourceId:           "resource-new",
		ResourceTypes:        []string{"vm"},
		ComplianceComment:    "New resource is compliant",
		TargetOfEvaluationId: MockToeID1,
		ToolId:               util.Ref("tool-1"),
		HistoryUpdatedAt:     timestamppb.Now(),
		History: []*assessment.Record{
			{
				EvidenceId:         MockEvidenceID1,
				EvidenceRecordedAt: timestamppb.Now(),
			},
		},
	}
)
