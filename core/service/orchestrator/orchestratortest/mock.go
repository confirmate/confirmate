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
	MockToeID3        = "00000000-0000-0000-0000-000000000003"
	MockScopeID1      = "00000000-0000-0000-0001-000000000001"
	MockScopeID2      = "00000000-0000-0000-0001-000000000002"
	MockResultID1     = "00000000-0000-0000-0002-000000000001"
	MockResultID2     = "00000000-0000-0000-0002-000000000002"
	MockResultID3     = "00000000-0000-0000-0002-000000000003"
	MockEvidenceID1   = "00000000-0000-0000-0003-000000000001"
	MockEvidenceID2   = "00000000-0000-0000-0003-000000000002"
	MockNonExistentID = "00000000-0000-0000-ffff-ffffffffffff"
	MockEmptyUUID     = "00000000-0000-0000-0000-000000000000"
)

// Mock string IDs for consistent testing
const (
	MockMetricID1       = "metric-1"
	MockMetricID2       = "metric-2"
	MockMetricID3       = "metric-3"
	MockMetricIDDefault = "metric-default"
	MockCatalogID1      = "catalog-1"
	MockCatalogID2      = "catalog-2"
	MockCategoryName1   = "category-1"
	MockCategoryName2   = "category-2"
	MockControlID1      = "control-1"
	MockControlID2      = "control-2"
	MockCertificateID1  = "certificate-1"
	MockCertificateID2  = "certificate-2"
	MockResourceID1     = "resource-1"
	MockResourceID2     = "resource-2"
	MockResourceIDNew   = "resource-new"
	MockToolID1         = "tool-1"
	MockTestCategory    = "test-category"
	MockDefaultVersion  = "1.0.0"
)

var (
	// Mock Metrics
	MockMetric1 = &assessment.Metric{
		Id:          MockMetricID1,
		Description: "Mock Metric 1",
		Version:     MockDefaultVersion,
		Category:    MockTestCategory,
	}
	MockMetric2 = &assessment.Metric{
		Id:          MockMetricID2,
		Description: "Mock Metric 2",
		Version:     MockDefaultVersion,
		Category:    MockTestCategory,
	}
	MockMetric3 = &assessment.Metric{
		Id:          MockMetricID3,
		Description: "Mock Metric 3",
		Version:     MockDefaultVersion,
		Category:    MockTestCategory,
	}
	MockMetricWithDefault = &assessment.Metric{
		Id:          MockMetricIDDefault,
		Description: "Mock Metric with Default Config",
		Version:     MockDefaultVersion,
		Category:    MockTestCategory,
	}

	// Mock Metric Implementations
	MockMetricImplementation1 = &assessment.MetricImplementation{
		MetricId: MockMetricID1,
		Lang:     assessment.MetricImplementation_LANGUAGE_REGO,
		Code:     "mock implementation code",
	}

	// Mock Metric Configurations
	MockMetricConfiguration1 = &assessment.MetricConfiguration{
		TargetOfEvaluationId: MockToeID1,
		MetricId:             MockMetricID1,
		Operator:             "==",
		TargetValue:          structpb.NewBoolValue(true),
		IsDefault:            true,
	}
	MockMetricConfiguration2 = &assessment.MetricConfiguration{
		TargetOfEvaluationId: MockToeID1,
		MetricId:             MockMetricID2,
		Operator:             "==",
		TargetValue:          structpb.NewBoolValue(true),
		IsDefault:            false,
	}
	MockMetricConfiguration3 = &assessment.MetricConfiguration{
		TargetOfEvaluationId: MockToeID1,
		MetricId:             MockMetricID3,
		Operator:             "==",
		TargetValue:          structpb.NewBoolValue(true),
		IsDefault:            false,
	}
	MockMetricConfigurationDefault = &assessment.MetricConfiguration{
		MetricId:    MockMetricIDDefault,
		Operator:    "==",
		TargetValue: structpb.NewBoolValue(true),
		IsDefault:   true,
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
		Id:          MockCatalogID1,
		Name:        "Mock Catalog 1",
		Description: "Mock catalog description 1",
	}
	MockCatalog2 = &orchestrator.Catalog{
		Id:          MockCatalogID2,
		Name:        "Mock Catalog 2",
		Description: "Mock catalog description 2",
	}

	// Mock Categories
	MockCategory1 = &orchestrator.Category{
		Name:      MockCategoryName1,
		CatalogId: MockCatalogID1,
	}
	MockCategory2 = &orchestrator.Category{
		Name:      MockCategoryName2,
		CatalogId: MockCatalogID2,
	}

	// Mock Controls
	MockControl1 = &orchestrator.Control{
		Id:                MockControlID1,
		CategoryName:      MockCategoryName1,
		CategoryCatalogId: MockCatalogID1,
	}
	MockControl2 = &orchestrator.Control{
		Id:                MockControlID2,
		CategoryName:      MockCategoryName2,
		CategoryCatalogId: MockCatalogID2,
	}

	// Mock Certificates
	MockCertificate1 = &orchestrator.Certificate{
		Id:                   MockCertificateID1,
		Name:                 "Mock Certificate 1",
		Description:          "Mock certificate description 1",
		TargetOfEvaluationId: MockToeID1,
	}
	MockCertificate2 = &orchestrator.Certificate{
		Id:                   MockCertificateID2,
		Name:                 "Mock Certificate 2",
		Description:          "Mock certificate description 2",
		TargetOfEvaluationId: MockToeID2,
	}

	// Mock Audit Scopes
	MockAuditScope1 = &orchestrator.AuditScope{
		Id:                   MockScopeID1,
		TargetOfEvaluationId: MockToeID1,
		CatalogId:            MockCatalogID1,
	}
	MockAuditScope2 = &orchestrator.AuditScope{
		Id:                   MockScopeID2,
		TargetOfEvaluationId: MockToeID2,
		CatalogId:            MockCatalogID2,
	}

	// Mock Assessment Results
	MockAssessmentResult1 = &assessment.AssessmentResult{
		Id:                   MockResultID1,
		CreatedAt:            timestamppb.Now(),
		MetricId:             MockMetricID1,
		MetricConfiguration:  MockMetricConfiguration1,
		Compliant:            true,
		EvidenceId:           MockEvidenceID1,
		ResourceId:           MockResourceID1,
		ResourceTypes:        []string{"vm"},
		ComplianceComment:    "Resource is compliant",
		TargetOfEvaluationId: MockToeID1,
		ToolId:               util.Ref(MockToolID1),
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
		MetricId:             MockMetricID2,
		MetricConfiguration:  MockMetricConfiguration2,
		Compliant:            false,
		EvidenceId:           MockEvidenceID2,
		ResourceId:           MockResourceID2,
		ResourceTypes:        []string{"storage"},
		ComplianceComment:    "Resource is not compliant",
		TargetOfEvaluationId: MockToeID1,
		ToolId:               util.Ref(MockToolID1),
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
		MetricId:             MockMetricID1,
		MetricConfiguration:  MockMetricConfiguration1,
		Compliant:            true,
		EvidenceId:           MockEvidenceID1,
		ResourceId:           MockResourceIDNew,
		ResourceTypes:        []string{"vm"},
		ComplianceComment:    "New resource is compliant",
		TargetOfEvaluationId: MockToeID1,
		ToolId:               util.Ref(MockToolID1),
		HistoryUpdatedAt:     timestamppb.Now(),
		History: []*assessment.Record{
			{
				EvidenceId:         MockEvidenceID1,
				EvidenceRecordedAt: timestamppb.Now(),
			},
		},
	}
)
