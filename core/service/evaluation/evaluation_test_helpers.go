package evaluation

import (
	"context"
	"net/http/httptest"
	"sync"
	"testing"

	"confirmate.io/core/api/assessment"
	"confirmate.io/core/api/evaluation"
	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/api/orchestrator/orchestratorconnect"
	"confirmate.io/core/server"
	"confirmate.io/core/server/servertest"
	"confirmate.io/core/service/orchestrator/orchestratortest"
	"connectrpc.com/connect"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// mockOrchestratorHandler is a mock implementation of the orchestrator service for testing
type mockOrchestratorHandler struct {
	orchestratorconnect.UnimplementedOrchestratorHandler
	// ListControls support
	controls  []*orchestrator.Control
	listError error

	// ListAssessmentResults support
	assessmentResults         []*assessment.AssessmentResult
	listAssessmentResultError error

	// GetAuditScope support
	auditScope                 *orchestrator.AuditScope
	getAuditScopeNotFoundError error
	getAuditScopeError         error

	// GetCatalog support
	catalog                 *orchestrator.Catalog
	getCatalogNotFoundError error
	getCatalogError         error

	// ListEvaluation support
	evaluationResults []*evaluation.EvaluationResult
	listEvalError     error
	storeEvalError    error
	mu                sync.Mutex
}

// ListControls returns the mocked controls or an error if configured
func (m *mockOrchestratorHandler) ListControls(
	_ context.Context,
	_ *connect.Request[orchestrator.ListControlsRequest],
) (*connect.Response[orchestrator.ListControlsResponse], error) {
	if m.listError != nil {
		return nil, m.listError
	}
	return connect.NewResponse(&orchestrator.ListControlsResponse{
		Controls: m.controls,
	}), nil
}

// ListAssessmentResults returns assessment results or an error if configured
func (m *mockOrchestratorHandler) ListAssessmentResults(
	_ context.Context,
	req *connect.Request[orchestrator.ListAssessmentResultsRequest],
) (*connect.Response[orchestrator.ListAssessmentResultsResponse], error) {
	if m.listAssessmentResultError != nil {
		return nil, m.listAssessmentResultError
	}
	// Return configured assessment results, or empty list if none configured
	results := m.assessmentResults
	if results == nil {
		results = []*assessment.AssessmentResult{}
	}

	// Filter by target of evaluation id and metric ids if provided
	if req.Msg.Filter != nil {
		filtered := []*assessment.AssessmentResult{}
		for _, result := range results {
			// Filter by target of evaluation id
			if req.Msg.Filter.TargetOfEvaluationId != nil &&
				result.TargetOfEvaluationId != *req.Msg.Filter.TargetOfEvaluationId {
				continue
			}
			// Filter by metric ids
			if len(req.Msg.Filter.MetricIds) > 0 {
				metricMatch := false
				for _, metricId := range req.Msg.Filter.MetricIds {
					if result.MetricId == metricId {
						metricMatch = true
						break
					}
				}
				if !metricMatch {
					continue
				}
			}
			filtered = append(filtered, result)
		}
		results = filtered
	}

	// If LatestByResourceId is true, group by resource_id and keep only the latest for each
	if req.Msg.LatestByResourceId != nil && *req.Msg.LatestByResourceId {
		latestByResource := make(map[string]*assessment.AssessmentResult)
		for _, result := range results {
			existing, found := latestByResource[result.ResourceId]
			if !found || result.CreatedAt.AsTime().After(existing.CreatedAt.AsTime()) {
				latestByResource[result.ResourceId] = result
			}
		}
		// Convert map back to slice
		results = make([]*assessment.AssessmentResult, 0, len(latestByResource))
		for _, result := range latestByResource {
			results = append(results, result)
		}
	}

	return connect.NewResponse(&orchestrator.ListAssessmentResultsResponse{
		Results: results,
	}), nil
}

// StoreEvaluationResult stores the result in-memory so tests can verify it via ListEvaluationResults.
func (m *mockOrchestratorHandler) StoreEvaluationResult(
	_ context.Context,
	req *connect.Request[orchestrator.StoreEvaluationResultRequest],
) (*connect.Response[evaluation.EvaluationResult], error) {
	if m.storeEvalError != nil {
		return nil, m.storeEvalError
	}

	if r := req.Msg.GetResult(); r != nil {
		m.mu.Lock()
		m.evaluationResults = append(m.evaluationResults, r)
		m.mu.Unlock()
	}

	eval := &evaluation.EvaluationResult{
		Id:                   uuid.NewString(),
		Timestamp:            timestamppb.Now(),
		ControlCategoryName:  req.Msg.GetResult().GetControlCategoryName(),
		ControlCatalogId:     req.Msg.GetResult().GetControlCatalogId(),
		ControlId:            req.Msg.GetResult().GetControlId(),
		ParentControlId:      req.Msg.GetResult().ParentControlId,
		TargetOfEvaluationId: req.Msg.GetResult().GetTargetOfEvaluationId(),
		AuditScopeId:         req.Msg.GetResult().GetAuditScopeId(),
		Status:               req.Msg.GetResult().GetStatus(),
		AssessmentResultIds:  req.Msg.GetResult().GetAssessmentResultIds(),
	}

	return connect.NewResponse(eval), nil
}

// ListEvaluationResults returns the evaluation results stored via StoreEvaluationResult.
func (m *mockOrchestratorHandler) ListEvaluationResults(
	_ context.Context,
	_ *connect.Request[orchestrator.ListEvaluationResultsRequest],
) (*connect.Response[orchestrator.ListEvaluationResultsResponse], error) {
	if m.listEvalError != nil {
		return nil, m.listEvalError
	}

	m.mu.Lock()
	out := make([]*evaluation.EvaluationResult, len(m.evaluationResults))
	copy(out, m.evaluationResults)
	m.mu.Unlock()

	return connect.NewResponse(&orchestrator.ListEvaluationResultsResponse{
		Results: out,
	}), nil
}

// GetAuditScope returns audit scope or an error if configured
func (m *mockOrchestratorHandler) GetAuditScope(
	_ context.Context,
	_ *connect.Request[orchestrator.GetAuditScopeRequest],
) (*connect.Response[orchestrator.AuditScope], error) {
	// 1) allow forcing an arbitrary error (e.g. internal)
	if m.getAuditScopeError != nil {
		return nil, m.getAuditScopeError
	}

	// 2) simulate "not found"
	if m.auditScope == nil {
		return nil, m.getAuditScopeNotFoundError
	}

	return connect.NewResponse(m.auditScope), nil
}

// GetCatalog returns catalog or an error if configured
func (m *mockOrchestratorHandler) GetCatalog(
	_ context.Context,
	_ *connect.Request[orchestrator.GetCatalogRequest],
) (*connect.Response[orchestrator.Catalog], error) {
	// 1) allow forcing an arbitrary error (e.g. internal)
	if m.getCatalogError != nil {
		return nil, m.getCatalogError
	}

	// 2) simulate "not found"
	if m.catalog == nil {
		return nil, m.getCatalogNotFoundError
	}

	return connect.NewResponse(m.catalog), nil
}

// newOrchestratorTestServer creates a mock orchestrator server for testing
func newOrchestratorTestServer(t *testing.T, controls []*orchestrator.Control) (
	*mockOrchestratorHandler,
	*server.Server,
	*httptest.Server,
) {
	t.Helper()
	handler := &mockOrchestratorHandler{
		controls: controls,
	}
	srv, testSrv := servertest.NewTestConnectServer(
		t,
		server.WithHandler(orchestratorconnect.NewOrchestratorHandler(handler)),
	)
	return handler, srv, testSrv
}

// newOrchestratorTestServerWithError creates a mock orchestrator server that returns an error
func newOrchestratorTestServerWithError(t *testing.T, err error) (
	*mockOrchestratorHandler,
	*server.Server,
	*httptest.Server,
) {
	t.Helper()
	handler := &mockOrchestratorHandler{
		listError: err,
	}
	srv, testSrv := servertest.NewTestConnectServer(
		t,
		server.WithHandler(orchestratorconnect.NewOrchestratorHandler(handler)),
	)
	return handler, srv, testSrv
}

// newOrchestratorClientForTest creates a Connect-based orchestrator client for testing
func newOrchestratorClientForTest(testSrv *httptest.Server) orchestratorconnect.OrchestratorClient {
	return orchestratorconnect.NewOrchestratorClient(testSrv.Client(), testSrv.URL)
}

// newOrchestratorClient creates a mock orchestrator client that can serve BOTH assessmentResults and evaluationResults.
func newOrchestratorClient(
	t *testing.T,
	opts ...func(*mockOrchestratorHandler),
) orchestratorconnect.OrchestratorClient {
	t.Helper()

	handler := &mockOrchestratorHandler{}
	for _, opt := range opts {
		opt(handler)
	}

	_, testSrv := servertest.NewTestConnectServer(
		t,
		server.WithHandler(orchestratorconnect.NewOrchestratorHandler(handler)),
	)
	t.Cleanup(testSrv.Close)

	return newOrchestratorClientForTest(testSrv)
}

// WithAssessmentResults seeds the handler with assessment results.
func WithAssessmentResults(results []*assessment.AssessmentResult) func(*mockOrchestratorHandler) {
	return func(h *mockOrchestratorHandler) { h.assessmentResults = results }
}

// WithEvaluationResults seeds the handler with evaluation results (visible via ListEvaluationResults).
func WithEvaluationResults(results []*evaluation.EvaluationResult) func(*mockOrchestratorHandler) {
	return func(h *mockOrchestratorHandler) { h.evaluationResults = results }
}

// WithControls seeds the handler with controls. It accepts one or more control lists and flattens them.
func WithControls(lists ...[]*orchestrator.Control) func(*mockOrchestratorHandler) {
	return func(h *mockOrchestratorHandler) {
		var combined []*orchestrator.Control
		for _, l := range lists {
			combined = append(combined, l...)
		}
		h.controls = combined
	}
}

// WithAdditionalControls appends controls to any already configured controls.
func WithAdditionalControls(controls ...*orchestrator.Control) func(*mockOrchestratorHandler) {
	return func(h *mockOrchestratorHandler) {
		h.controls = append(h.controls, controls...)
	}
}

// WithAuditScope seeds the handler with an audit scope returned by GetAuditScope.
func WithAuditScope(scope *orchestrator.AuditScope) func(*mockOrchestratorHandler) {
	return func(h *mockOrchestratorHandler) { h.auditScope = scope }
}

// WithGetAuditScopeError forces GetAuditScope to return the given error.
func WithGetAuditScopeError(err error) func(*mockOrchestratorHandler) {
	return func(h *mockOrchestratorHandler) { h.getAuditScopeError = err }
}

// WithGetAuditScopeNotFoundError configures the error returned when auditScope is nil.
func WithGetAuditScopeNotFoundError(err error) func(*mockOrchestratorHandler) {
	return func(h *mockOrchestratorHandler) { h.getAuditScopeNotFoundError = err }
}

// WithCatalog seeds the handler with a catalog returned by GetCatalog.
func WithCatalog(catalog *orchestrator.Catalog) func(*mockOrchestratorHandler) {
	return func(h *mockOrchestratorHandler) { h.catalog = catalog }
}

// WithGetCatalogError forces GetCatalog to return the given error.
func WithGetCatalogError(err error) func(*mockOrchestratorHandler) {
	return func(h *mockOrchestratorHandler) { h.getCatalogError = err }
}

// WithGetCatalogNotFoundError configures the error returned when catalog is nil.
func WithGetCatalogNotFoundError(err error) func(*mockOrchestratorHandler) {
	return func(h *mockOrchestratorHandler) { h.getCatalogNotFoundError = err }
}

// mockControlsForCatalog returns mock controls for a catalog
func mockControlsForCatalog(catalogID string) []*orchestrator.Control {
	// Return 4 controls as expected by the test
	control1 := &orchestrator.Control{
		Id:                orchestratortest.MockControlId1,
		CategoryName:      orchestratortest.MockCategoryName1,
		CategoryCatalogId: catalogID,
		Name:              "Mock Control 1",
	}
	control2 := &orchestrator.Control{
		Id:                orchestratortest.MockControlId2,
		CategoryName:      orchestratortest.MockCategoryName1,
		CategoryCatalogId: catalogID,
		Name:              "Mock Control 2",
	}
	control3 := &orchestrator.Control{
		Id:                "control-3",
		CategoryName:      orchestratortest.MockCategoryName1,
		CategoryCatalogId: catalogID,
		Name:              "Mock Control 3",
	}
	control4 := &orchestrator.Control{
		Id:                "control-4",
		CategoryName:      orchestratortest.MockCategoryName2,
		CategoryCatalogId: catalogID,
		Name:              "Mock Control 4",
	}
	return []*orchestrator.Control{control1, control2, control3, control4}
}
