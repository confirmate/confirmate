package evaluation

import (
	"context"
	"net/http/httptest"
	"testing"

	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/api/orchestrator/orchestratorconnect"
	"confirmate.io/core/server"
	"confirmate.io/core/server/servertest"
	"confirmate.io/core/service/orchestrator/orchestratortest"
	"confirmate.io/core/util/assert"
	"connectrpc.com/connect"
)

// mockOrchestratorHandler is a mock implementation of the orchestrator service for testing
type mockOrchestratorHandler struct {
	orchestratorconnect.UnimplementedOrchestratorHandler
	controls []*orchestrator.Control
}

// ListControls returns the mocked controls
func (m *mockOrchestratorHandler) ListControls(
	_ context.Context,
	req *connect.Request[orchestrator.ListControlsRequest],
) (*connect.Response[orchestrator.ListControlsResponse], error) {
	return connect.NewResponse(&orchestrator.ListControlsResponse{
		Controls: m.controls,
	}), nil
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

// assertNoTestServerError ensures test server creation succeeds and fails test otherwise
func assertNoTestServerError(t *testing.T, srv *server.Server, testSrv *httptest.Server) {
	t.Helper()
	assert.NotNil(t, srv)
	assert.NotNil(t, testSrv)
}
