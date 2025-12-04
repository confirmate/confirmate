package orchestratortest

import (
	"confirmate.io/core/api/assessment"
	"confirmate.io/core/api/orchestrator"
)

var (
	// Mock Metrics
	MockMetric1 = &assessment.Metric{
		Id:          "metric-1",
		Description: "Mock Metric 1",
	}
	MockMetric2 = &assessment.Metric{
		Id:          "metric-2",
		Description: "Mock Metric 2",
	}

	// Mock Metric Implementations
	MockMetricImplementation1 = &assessment.MetricImplementation{
		MetricId: "metric-1",
		Lang:     assessment.MetricImplementation_LANGUAGE_REGO,
		Code:     "mock implementation code",
	}

	// Mock Metric Configurations
	MockMetricConfiguration1 = &assessment.MetricConfiguration{
		TargetOfEvaluationId: "toe-1",
		MetricId:             "metric-1",
		IsDefault:            true,
	}
	MockMetricConfiguration2 = &assessment.MetricConfiguration{
		TargetOfEvaluationId: "toe-1",
		MetricId:             "metric-2",
		IsDefault:            false,
	}

	// Mock Targets of Evaluation
	MockTargetOfEvaluation1 = &orchestrator.TargetOfEvaluation{
		Id:   "toe-1",
		Name: "Mock TOE 1",
	}
	MockTargetOfEvaluation2 = &orchestrator.TargetOfEvaluation{
		Id:   "toe-2",
		Name: "Mock TOE 2",
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
		Id:          "cert-1",
		Name:        "Mock Certificate 1",
		Description: "Mock certificate description 1",
	}
	MockCertificate2 = &orchestrator.Certificate{
		Id:          "cert-2",
		Name:        "Mock Certificate 2",
		Description: "Mock certificate description 2",
	}

	// Mock Audit Scopes
	MockAuditScope1 = &orchestrator.AuditScope{
		Id:                   "scope-1",
		TargetOfEvaluationId: "toe-1",
		CatalogId:            "catalog-1",
	}
	MockAuditScope2 = &orchestrator.AuditScope{
		Id:                   "scope-2",
		TargetOfEvaluationId: "toe-2",
		CatalogId:            "catalog-2",
	}

	// Mock Assessment Results
	MockAssessmentResult1 = &assessment.AssessmentResult{
		Id:            "result-1",
		MetricId:      "metric-1",
		EvidenceId:    "evidence-1",
		ResourceId:    "resource-1",
		ResourceTypes: []string{"vm"},
		Compliant:     true,
	}
	MockAssessmentResult2 = &assessment.AssessmentResult{
		Id:            "result-2",
		MetricId:      "metric-2",
		EvidenceId:    "evidence-2",
		ResourceId:    "resource-2",
		ResourceTypes: []string{"storage"},
		Compliant:     false,
	}

	// Mock Assessment Results for Store tests (without ID, so service generates one)
	MockNewAssessmentResult = &assessment.AssessmentResult{
		MetricId:      "metric-1",
		EvidenceId:    "evidence-new",
		ResourceId:    "resource-new",
		ResourceTypes: []string{"vm"},
		Compliant:     true,
	}
	MockNewAssessmentResultWithId = &assessment.AssessmentResult{
		Id:            "custom-id",
		MetricId:      "metric-2",
		EvidenceId:    "evidence-custom",
		ResourceId:    "resource-custom",
		ResourceTypes: []string{"storage"},
		Compliant:     false,
	}
)
