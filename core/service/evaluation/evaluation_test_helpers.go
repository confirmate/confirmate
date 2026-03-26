package evaluation

import (
	"context"
	"net/http/httptest"
	"testing"

	"confirmate.io/core/api/assessment"
	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/api/orchestrator/orchestratorconnect"
	"confirmate.io/core/server"
	"confirmate.io/core/server/servertest"
	"confirmate.io/core/service/orchestrator/orchestratortest"
	"connectrpc.com/connect"
)

// mockOrchestratorHandler is a mock implementation of the orchestrator service for testing
type mockOrchestratorHandler struct {
	orchestratorconnect.UnimplementedOrchestratorHandler
	controls                   []*orchestrator.Control
	listError                  error
	listAssessmentResultError  error
	getAuditScopeNotFoundError error
	assessmentResults          []*assessment.AssessmentResult
	auditScope                 *orchestrator.AuditScope
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
	_ *connect.Request[orchestrator.ListAssessmentResultsRequest],
) (*connect.Response[orchestrator.ListAssessmentResultsResponse], error) {
	if m.listAssessmentResultError != nil {
		return nil, m.listAssessmentResultError
	}
	// Return configured assessment results, or empty list if none configured
	results := m.assessmentResults
	if results == nil {
		results = []*assessment.AssessmentResult{}
	}
	return connect.NewResponse(&orchestrator.ListAssessmentResultsResponse{
		Results: results,
	}), nil
}

// GetAuditScope returns audit scopes or an error if configured
func (m *mockOrchestratorHandler) GetAuditScope(
	_ context.Context,
	_ *connect.Request[orchestrator.GetAuditScopeRequest],
) (*connect.Response[orchestrator.AuditScope], error) {
	if m.auditScope == nil {
		return nil, m.getAuditScopeNotFoundError
	}
	return &connect.Response[orchestrator.AuditScope]{Msg: m.auditScope}, nil
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
