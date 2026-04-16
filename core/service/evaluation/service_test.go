package evaluation

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"confirmate.io/core/api/assessment"
	"confirmate.io/core/api/evaluation"
	"confirmate.io/core/api/evaluation/evaluationconnect"
	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/api/orchestrator/orchestratorconnect"
	"google.golang.org/protobuf/testing/protocmp"

	"confirmate.io/core/persistence"
	"confirmate.io/core/server"
	"confirmate.io/core/server/servertest"
	"confirmate.io/core/service"
	"confirmate.io/core/service/evaluation/evaluationtest"
	"confirmate.io/core/service/evidence/evidencetest"
	"confirmate.io/core/service/orchestrator/orchestratortest"
	"confirmate.io/core/util/assert"
	"connectrpc.com/connect"
	"github.com/go-co-op/gocron"
)

func TestNewService(t *testing.T) {
	type args struct {
		opts []service.Option[Service]
	}
	tests := []struct {
		name    string
		args    args
		want    assert.Want[evaluationconnect.EvaluationHandler]
		wantErr assert.WantErr
	}{
		{
			name: "happy path with WithConfig option",
			args: args{
				opts: []service.Option[Service]{
					WithConfig(Config{
						PersistenceConfig:   persistence.DefaultConfig,
						OrchestratorClient:  http.DefaultClient,
						OrchestratorAddress: "http://testhost:8080",
					}),
				},
			},
			want: func(t *testing.T, got evaluationconnect.EvaluationHandler, msgAndArgs ...any) bool {
				assert.NotNil(t, got)
				svc, ok := got.(*Service)
				if !ok {
					t.Fatalf("expected *Service, got %T", got)
				}
				assert.Equal(t, Config{
					OrchestratorAddress: "http://testhost:8080",
					OrchestratorClient:  http.DefaultClient,
					PersistenceConfig:   persistence.DefaultConfig,
				}, svc.cfg)
				assert.NotEmpty(t, svc.scheduler)
				assert.NotEmpty(t, orchestratorconnect.NewOrchestratorClient(svc.cfg.OrchestratorClient, "http:://testhost:8080"), svc.orchestratorClient)
				assert.Equal(t, make(map[string]map[string]*orchestrator.Control), svc.catalogControls)
				return assert.NotNil(t, &svc.catalogsMutex)
			},
			wantErr: assert.NoError,
		},
		{
			name: "happy path",
			args: args{
				opts: []service.Option[Service]{},
			},
			want: func(t *testing.T, got evaluationconnect.EvaluationHandler, msgAndArgs ...any) bool {
				assert.NotNil(t, got)
				svc, ok := got.(*Service)
				if !ok {
					t.Fatalf("expected *Service, got %T", got)
				}
				assert.Equal(t, DefaultConfig, svc.cfg)
				assert.NotEmpty(t, svc.scheduler)
				assert.NotEmpty(t, orchestratorconnect.NewOrchestratorClient(svc.cfg.OrchestratorClient, svc.cfg.OrchestratorAddress), svc.orchestratorClient)
				assert.Equal(t, make(map[string]map[string]*orchestrator.Control), svc.catalogControls)
				return assert.NotNil(t, &svc.catalogsMutex)
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := NewService(tt.args.opts...)

			tt.want(t, got)
			tt.wantErr(t, gotErr)
		})
	}
}

func TestService_Shutdown(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "happy path",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewService()
			assert.NoError(t, err)

			svc, ok := got.(*Service)
			if !ok {
				t.Fatalf("expected *Service, got %T", got)
			}

			svc.Shutdown()
			assert.False(t, svc.scheduler.IsRunning())

		})
	}
}

func TestService_getAllMetricsFromControl(t *testing.T) {
	type fields struct {
		catalogControls map[string]map[string]*orchestrator.Control
	}
	type args struct {
		catalogId    string
		categoryName string
		controlId    string
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		wantMetrics []*assessment.Metric
		wantErr     assert.WantErr
	}{
		{
			name:        "Input empty",
			fields:      fields{},
			wantMetrics: nil,
			wantErr: func(t *testing.T, err error, i ...interface{}) bool {
				return assert.ErrorContains(t, err, "could not get control for control id")
			},
		},
		{
			name: "no sub-controls available",
			fields: fields{
				catalogControls: map[string]map[string]*orchestrator.Control{
					orchestratortest.MockCatalogId1: {
						fmt.Sprintf("%s-%s", orchestratortest.MockControl2.GetCategoryName(), orchestratortest.MockControl2.GetId()): orchestratortest.MockControl2,
					},
				},
			},
			args: args{
				catalogId:    orchestratortest.MockCatalogId1,
				categoryName: orchestratortest.MockCategoryName2,
				controlId:    orchestratortest.MockControlId2,
			},
			wantMetrics: nil,
			wantErr:     assert.NoError,
		},
		{
			name: "error getting metrics from sub-controls",
			fields: fields{
				catalogControls: map[string]map[string]*orchestrator.Control{
					orchestratortest.MockCatalogId1: {
						fmt.Sprintf("%s-%s", orchestratortest.MockControl1.GetCategoryName(), orchestratortest.MockControl1.GetId()): orchestratortest.MockControl1,
					},
				},
			},
			args: args{
				catalogId:    orchestratortest.MockCatalogId1,
				categoryName: orchestratortest.MockCategoryName1,
				controlId:    orchestratortest.MockControlId1,
			},
			wantMetrics: nil,
			wantErr: func(t *testing.T, err error, i ...interface{}) bool {
				return assert.ErrorContains(t, err, "error getting metrics from sub-controls")
			},
		},
		{
			name: "Happy path: control with metrics directly (no sub-controls)",
			fields: fields{
				catalogControls: map[string]map[string]*orchestrator.Control{
					orchestratortest.MockCatalogId2: {
						fmt.Sprintf("%s-%s", orchestratortest.MockControl2.GetCategoryName(), orchestratortest.MockControl2.GetId()): {
							Id:                orchestratortest.MockControlId2,
							CategoryName:      orchestratortest.MockCategoryName2,
							CategoryCatalogId: orchestratortest.MockCatalogId2,
							Name:              "Mock Control 2",
							Metrics: []*assessment.Metric{
								{
									Id:       orchestratortest.MockMetricId2,
									Version:  "2.0",
									Comments: "Direct metric on control",
								},
							},
						},
					},
				},
			},
			args: args{
				catalogId:    orchestratortest.MockCatalogId2,
				categoryName: orchestratortest.MockCategoryName2,
				controlId:    orchestratortest.MockControlId2,
			},
			wantMetrics: []*assessment.Metric{
				{
					Id:       orchestratortest.MockMetricId2,
					Version:  "2.0",
					Comments: "Direct metric on control",
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "Happy path: control with sub-controls",
			fields: fields{
				catalogControls: map[string]map[string]*orchestrator.Control{
					orchestratortest.MockControl1.GetCategoryCatalogId(): {
						fmt.Sprintf("%s-%s", orchestratortest.MockControl1.GetCategoryName(), orchestratortest.MockControl1.GetId()):       orchestratortest.MockControl1,
						fmt.Sprintf("%s-%s", orchestratortest.MockSubControl1.GetCategoryName(), orchestratortest.MockSubControl1.GetId()): orchestratortest.MockSubControl1,
					},
				},
			},
			args: args{
				catalogId:    orchestratortest.MockCatalogId1,
				categoryName: orchestratortest.MockCategoryName1,
				controlId:    orchestratortest.MockControlId1,
			},
			wantMetrics: []*assessment.Metric{
				orchestratortest.MockMetric1,
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{
				catalogControls: tt.fields.catalogControls,
			}
			gotMetrics, err := s.getAllMetricsFromControl(tt.args.catalogId, tt.args.categoryName, tt.args.controlId)
			tt.wantErr(t, err)

			if assert.Equal(t, len(gotMetrics), len(tt.wantMetrics)) {
				for i := range gotMetrics {
					assert.Equal(t, tt.wantMetrics[i], gotMetrics[i])
				}
			}
		})
	}
}

func TestService_getMetricsFromSubControls(t *testing.T) {
	type fields struct {
		catalogControls map[string]map[string]*orchestrator.Control
	}
	type args struct {
		control *orchestrator.Control
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		wantMetrics []*assessment.Metric
		wantErr     assert.WantErr
	}{
		{
			name:        "Control is missing",
			args:        args{},
			wantMetrics: nil,
			wantErr: func(t *testing.T, err error, i ...interface{}) bool {
				return assert.ErrorContains(t, err, "control is missing")
			},
		},
		{
			name: "Sub-control level and metrics missing",
			args: args{
				control: &orchestrator.Control{
					Id:                orchestratortest.MockControlId1,
					CategoryName:      orchestratortest.MockCategoryName1,
					CategoryCatalogId: orchestratortest.MockCatalogId1,
					Name:              "Mock Control 1",
					Description:       "Mock control description",
				},
			},
			wantMetrics: nil,
			wantErr:     assert.NoError,
		},
		{
			name:   "Error getting control",
			fields: fields{},
			args: args{
				control: &orchestrator.Control{
					Id:                "testId",
					CategoryName:      orchestratortest.MockCategoryName1,
					CategoryCatalogId: orchestratortest.MockCatalogId1,
					Name:              "testId",
					Controls: []*orchestrator.Control{
						{
							Id:                "testId-subcontrol",
							CategoryName:      orchestratortest.MockCategoryName1,
							CategoryCatalogId: orchestratortest.MockCatalogId1,
							Name:              "testId-subcontrol",
						},
					},
					Metrics: []*assessment.Metric{},
				},
			},
			wantMetrics: nil,
			wantErr: func(t *testing.T, err error, i ...interface{}) bool {
				return assert.ErrorIs(t, err, service.ErrControlNotAvailable)
			},
		},
		{
			name: "Happy path",
			fields: fields{
				catalogControls: map[string]map[string]*orchestrator.Control{
					orchestratortest.MockControl1.GetCategoryCatalogId(): {
						fmt.Sprintf("%s-%s", orchestratortest.MockControl1.GetCategoryName(), orchestratortest.MockControl1.GetId()):       orchestratortest.MockControl1,
						fmt.Sprintf("%s-%s", orchestratortest.MockSubControl1.GetCategoryName(), orchestratortest.MockSubControl1.GetId()): orchestratortest.MockSubControl1,
					},
				},
			},
			args: args{
				control: orchestratortest.MockControl1,
			},
			wantMetrics: orchestratortest.MockControl1.Controls[0].GetMetrics(),
			wantErr:     assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{
				catalogControls: tt.fields.catalogControls,
			}
			gotMetrics, err := s.getMetricsFromSubcontrols(tt.args.control)

			tt.wantErr(t, err)

			assert.Equal(t, len(gotMetrics), len(tt.wantMetrics))
			for i := range gotMetrics {
				assert.Equal(t, tt.wantMetrics[i], gotMetrics[i])
			}
		})
	}
}

func TestService_cacheControls(t *testing.T) {
	type fields struct {
		orchestratorClient orchestratorconnect.OrchestratorClient
		catalogControls    map[string]map[string]*orchestrator.Control
	}
	type args struct {
		catalogId string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantSvc assert.Want[*Service]
		wantErr assert.WantErr
	}{
		{
			name: "catalog_id missing",
			fields: fields{
				catalogControls: make(map[string]map[string]*orchestrator.Control),
			},
			args: args{
				catalogId: "",
			},
			wantErr: func(t *testing.T, err error, i ...interface{}) bool {
				return assert.ErrorContains(t, err, "catalog ID is missing")
			},
		},
		{
			name: "orchestrator client returns error",
			fields: func() fields {
				// Create test server that returns an error
				_, _, testSrv := newOrchestratorTestServerWithError(t, connect.NewError(connect.CodeInternal, fmt.Errorf("orchestrator service unavailable")))
				t.Cleanup(testSrv.Close)
				return fields{
					orchestratorClient: newOrchestratorClientForTest(testSrv),
					catalogControls:    make(map[string]map[string]*orchestrator.Control),
				}
			}(),
			args: args{
				catalogId: orchestratortest.MockCatalogId1,
			},
			wantErr: func(t *testing.T, err error, i ...interface{}) bool {
				return assert.ErrorContains(t, err, "orchestrator service unavailable")
			},
		},
		{
			name: "no controls available",
			fields: func() fields {
				// Create test server with empty control list
				_, _, testSrv := newOrchestratorTestServer(t, []*orchestrator.Control{})
				t.Cleanup(testSrv.Close)
				return fields{
					orchestratorClient: newOrchestratorClientForTest(testSrv),
					catalogControls:    make(map[string]map[string]*orchestrator.Control),
				}
			}(),
			args: args{
				catalogId: orchestratortest.MockCatalogId1,
			},
			wantErr: func(t *testing.T, err error, i ...interface{}) bool {
				return assert.ErrorContains(t, err, fmt.Sprintf("no controls for catalog '%s' available", orchestratortest.MockCatalogId1))
			},
		},
		{
			name: "Happy path",
			fields: func() fields {
				controls := mockControlsForCatalog(orchestratortest.MockCatalogId1)
				_, _, testSrv := newOrchestratorTestServer(t, controls)
				t.Cleanup(testSrv.Close)
				return fields{
					orchestratorClient: newOrchestratorClientForTest(testSrv),
					catalogControls:    make(map[string]map[string]*orchestrator.Control),
				}
			}(),
			args: args{
				catalogId: orchestratortest.MockCatalogId1,
			},
			wantSvc: func(t *testing.T, got *Service, msgAndArgs ...any) bool {
				assert.Equal(t, 1, len(got.catalogControls))
				return assert.Equal(t, 4, len(got.catalogControls[orchestratortest.MockCatalogId1]))
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{
				orchestratorClient: tt.fields.orchestratorClient,
				catalogControls:    tt.fields.catalogControls,
			}
			err := svc.cacheControls(tt.args.catalogId)
			tt.wantErr(t, err)
			assert.Optional(t, tt.wantSvc, svc)
		})
	}
}

func TestService_getControl(t *testing.T) {
	type fields struct {
		catalogControls map[string]map[string]*orchestrator.Control
	}
	type args struct {
		catalogId    string
		categoryName string
		controlId    string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    assert.Want[*orchestrator.Control]
		wantErr assert.WantErr
	}{
		{
			name:   "catalog_id is missing",
			fields: fields{},
			args: args{
				categoryName: evidencetest.MockCategoryName,
				controlId:    evidencetest.MockControlID1,
			},
			want: nil,
			wantErr: func(t *testing.T, err error, _ ...any) bool {
				return assert.ErrorContains(t, err, "catalog id is missing")
			},
		},
		{
			name:   "category_name is missing",
			fields: fields{},
			args: args{
				catalogId: evidencetest.MockCatalogID1,
				controlId: evidencetest.MockControlID1,
			},
			want: nil,
			wantErr: func(t *testing.T, err error, i ...interface{}) bool {
				return assert.ErrorContains(t, err, "category name is missing")
			},
		},
		{
			name:   "control_id is missing",
			fields: fields{},
			args: args{
				catalogId:    evidencetest.MockCatalogID1,
				categoryName: evidencetest.MockCategoryName,
			},
			want: nil,
			wantErr: func(t *testing.T, err error, i ...interface{}) bool {
				return assert.ErrorContains(t, err, "control id is missing")
			},
		},
		{
			name:   "control does not exist",
			fields: fields{},
			args: args{
				catalogId:    "wrong_catalog_id",
				categoryName: "wrong_category_id",
				controlId:    "wrong_control_id",
			},
			want: nil,
			wantErr: func(t *testing.T, err error, i ...interface{}) bool {
				return assert.ErrorIs(t, err, service.ErrControlNotAvailable)
			},
		},
		{
			name: "Happy path",
			fields: fields{
				catalogControls: map[string]map[string]*orchestrator.Control{
					orchestratortest.MockControl1.GetCategoryCatalogId(): {
						fmt.Sprintf("%s-%s", orchestratortest.MockControl1.GetCategoryName(), orchestratortest.MockControl1.GetId()): orchestratortest.MockControl1,
						fmt.Sprintf("%s-%s", orchestratortest.MockControl1.GetCategoryName(), orchestratortest.MockControl2.GetId()): orchestratortest.MockControl2,
					},
				},
			},
			args: args{
				catalogId:    orchestratortest.MockControl1.GetCategoryCatalogId(),
				categoryName: orchestratortest.MockControl1.GetCategoryName(),
				controlId:    orchestratortest.MockControl1.GetId(),
			},
			want: func(t *testing.T, got *orchestrator.Control, _ ...any) bool {
				// We need to truncate the metric from the control because the control is only returned with its
				// sub-control but without the sub-control's metric.
				// TODO(oxisto): Use ignore fields instead
				wantControl := orchestratortest.MockControl1
				tmpMetrics := wantControl.Controls[0].Metrics
				wantControl.Controls[0].Metrics = nil

				if !assert.Equal(t, wantControl, got) {
					t.Errorf("Service.GetControl() = %v, want %v", got, wantControl)
					wantControl.Controls[0].Metrics = tmpMetrics
					return false
				}

				wantControl.Controls[0].Metrics = tmpMetrics
				return true
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{
				catalogControls: tt.fields.catalogControls,
			}

			gotControl, err := s.getControl(tt.args.catalogId, tt.args.categoryName, tt.args.controlId)
			tt.wantErr(t, err)

			if gotControl != nil {
				tt.want(t, gotControl)
			}
		})
	}
}

func Test_handlePending(t *testing.T) {
	type args struct {
		eval *evaluation.EvaluationResult
	}
	tests := []struct {
		name string
		args args
		want assert.Want[evaluation.EvaluationStatus]
	}{
		{
			name: "Status: Pending",
			args: args{
				eval: &evaluation.EvaluationResult{
					Status: evaluation.EvaluationStatus_EVALUATION_STATUS_PENDING,
				},
			},
			want: func(t *testing.T, got evaluation.EvaluationStatus, _ ...any) bool {
				return assert.Equal(t, evaluation.EvaluationStatus_EVALUATION_STATUS_PENDING, got)
			},
		},
		{
			name: "Status: Compliant",
			args: args{
				eval: &evaluation.EvaluationResult{
					Status: evaluation.EvaluationStatus_EVALUATION_STATUS_COMPLIANT,
				},
			},
			want: func(t *testing.T, got evaluation.EvaluationStatus, _ ...any) bool {
				return assert.Equal(t, evaluation.EvaluationStatus_EVALUATION_STATUS_COMPLIANT, got)
			},
		},
		{
			name: "Status: Compliant manually",
			args: args{
				eval: &evaluation.EvaluationResult{
					Status: evaluation.EvaluationStatus_EVALUATION_STATUS_COMPLIANT_MANUALLY,
				},
			},
			want: func(t *testing.T, got evaluation.EvaluationStatus, _ ...any) bool {
				return assert.Equal(t, evaluation.EvaluationStatus_EVALUATION_STATUS_COMPLIANT, got)
			},
		},
		{
			name: "Status: Not compliant manually without failing assessment results",
			args: args{
				eval: &evaluation.EvaluationResult{
					Status: evaluation.EvaluationStatus_EVALUATION_STATUS_NOT_COMPLIANT_MANUALLY,
				},
			},
			want: func(t *testing.T, got evaluation.EvaluationStatus, _ ...any) bool {
				return assert.Equal(t, evaluation.EvaluationStatus_EVALUATION_STATUS_NOT_COMPLIANT, got)
			},
		},
		{
			name: "Status: Not compliant with failing assessment results",
			args: args{
				eval: &evaluation.EvaluationResult{
					Status: evaluation.EvaluationStatus_EVALUATION_STATUS_NOT_COMPLIANT,
				},
			},
			want: func(t *testing.T, got evaluation.EvaluationStatus, _ ...any) bool {
				return assert.Equal(t, evaluation.EvaluationStatus_EVALUATION_STATUS_NOT_COMPLIANT, got)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := handlePending(tt.args.eval)
			tt.want(t, got)
		})
	}
}

func Test_handleCompliant(t *testing.T) {
	type args struct {
		er *evaluation.EvaluationResult
	}
	tests := []struct {
		name string
		args args
		want assert.Want[evaluation.EvaluationStatus]
	}{
		{
			name: "Status: Pending",
			args: args{
				er: &evaluation.EvaluationResult{
					Status: evaluation.EvaluationStatus_EVALUATION_STATUS_PENDING,
				},
			},
			want: func(t *testing.T, got evaluation.EvaluationStatus, msgAndArgs ...any) bool {
				return assert.Equal(t, evaluation.EvaluationStatus_EVALUATION_STATUS_COMPLIANT, got)
			},
		},
		{
			name: "Status: Compliant",
			args: args{
				er: &evaluation.EvaluationResult{
					Status: evaluation.EvaluationStatus_EVALUATION_STATUS_COMPLIANT,
				},
			},
			want: func(t *testing.T, got evaluation.EvaluationStatus, msgAndArgs ...any) bool {
				return assert.Equal(t, evaluation.EvaluationStatus_EVALUATION_STATUS_COMPLIANT, got)
			},
		},
		{
			name: "Status: Compliant manually",
			args: args{
				er: &evaluation.EvaluationResult{
					Status: evaluation.EvaluationStatus_EVALUATION_STATUS_COMPLIANT_MANUALLY,
				},
			},
			want: func(t *testing.T, got evaluation.EvaluationStatus, msgAndArgs ...any) bool {
				return assert.Equal(t, evaluation.EvaluationStatus_EVALUATION_STATUS_COMPLIANT, got)
			},
		},
		{
			name: "Status: Not compliant manually",
			args: args{
				er: &evaluation.EvaluationResult{
					Status:              evaluation.EvaluationStatus_EVALUATION_STATUS_NOT_COMPLIANT_MANUALLY,
					AssessmentResultIds: []string{"fail1", "fail2"},
				},
			},
			want: func(t *testing.T, got evaluation.EvaluationStatus, msgAndArgs ...any) bool {
				return assert.Equal(t, evaluation.EvaluationStatus_EVALUATION_STATUS_NOT_COMPLIANT, got)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := handleCompliant(tt.args.er)
			tt.want(t, got)
		})
	}
}

func Test_getMetricIds(t *testing.T) {
	type args struct {
		metrics []*assessment.Metric
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "Empty input",
			args: args{},
			want: nil,
		},
		{
			name: "Happy path",
			args: args{
				metrics: []*assessment.Metric{
					{
						Id: evidencetest.MockSubControlID11,
					},
					{
						Id: evidencetest.MockSubControlID,
					},
				},
			},
			want: []string{evidencetest.MockSubControlID11, evidencetest.MockSubControlID},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getMetricIds(tt.args.metrics)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestService_StopEvaluation(t *testing.T) {
	type args struct {
		req *connect.Request[evaluation.StopEvaluationRequest]
	}
	type fields struct {
		orchestratorClient orchestratorconnect.OrchestratorClient
		scheduler          *gocron.Scheduler
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[*connect.Response[evaluation.StopEvaluationResponse]]
		wantErr assert.WantErr
	}{
		{
			name: "error: input empty",
			args: args{
				req: &connect.Request[evaluation.StopEvaluationRequest]{},
			},
			fields: fields{},
			want:   assert.Nil[*connect.Response[evaluation.StopEvaluationResponse]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument) &&
					assert.ErrorContains(t, err, "invalid_argument: empty request")
			},
		},
		{
			name: "error: audit scope found but has no schedule",
			args: args{
				req: connect.NewRequest(&evaluation.StopEvaluationRequest{
					AuditScopeId: evaluationtest.MockAuditScopeId1,
				}),
			},
			fields: func() fields {
				// Create test server that returns an Audit Scope
				handler := &mockOrchestratorHandler{
					auditScope: evaluationtest.MockAuditScope1,
				}
				_, testSrv := servertest.NewTestConnectServer(
					t,
					server.WithHandler(orchestratorconnect.NewOrchestratorHandler(handler)),
				)
				t.Cleanup(testSrv.Close)

				return fields{
					orchestratorClient: newOrchestratorClientForTest(testSrv),
					scheduler: func() *gocron.Scheduler {
						s := gocron.NewScheduler(time.UTC)

						return s
					}(),
				}
			}(),
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeFailedPrecondition) &&
					assert.ErrorContains(t, err, fmt.Sprintf("job for audit scope '%s' is not running", evaluationtest.MockAuditScopeId1))
			},
			want: assert.Nil[*connect.Response[evaluation.StopEvaluationResponse]],
		},
		{
			name: "Happy path",
			args: args{
				req: connect.NewRequest(&evaluation.StopEvaluationRequest{
					AuditScopeId: evaluationtest.MockAuditScopeId1,
				}),
			},
			fields: func() fields {
				// Create test server that returns an Audit Scope
				handler := &mockOrchestratorHandler{
					auditScope: evaluationtest.MockAuditScope1,
				}
				_, testSrv := servertest.NewTestConnectServer(
					t,
					server.WithHandler(orchestratorconnect.NewOrchestratorHandler(handler)),
				)
				t.Cleanup(testSrv.Close)

				return fields{
					orchestratorClient: newOrchestratorClientForTest(testSrv),
					scheduler: func() *gocron.Scheduler {
						s := gocron.NewScheduler(time.UTC)
						_, err := s.Every(1).Day().Tag(evaluationtest.MockAuditScopeId1).Do(func() {
							fmt.Println("Scheduler job executed")
						})
						assert.NoError(t, err)
						return s
					}(),
				}
			}(),
			wantErr: assert.NoError,
			want: func(t *testing.T, got *connect.Response[evaluation.StopEvaluationResponse], msgAndArgs ...any) bool {
				return assert.Empty(t, got.Msg)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{
				orchestratorClient: tt.fields.orchestratorClient,
				scheduler:          tt.fields.scheduler,
			}
			got, gotErr := svc.StopEvaluation(context.Background(), tt.args.req)

			tt.want(t, got)
			tt.wantErr(t, gotErr)
		})
	}
}

func TestService_addJobToScheduler(t *testing.T) {
	type fields struct {
		scheduler *gocron.Scheduler
	}
	type args struct {
		ctx        context.Context
		auditScope *orchestrator.AuditScope
		catalog    *orchestrator.Catalog
		interval   int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    assert.Want[*Service]
		wantErr assert.WantErr
	}{
		{
			name: "error: invalid input - missing audit scope",
			fields: fields{
				scheduler: gocron.NewScheduler(time.Local),
			},
			args: args{
				ctx:        context.Background(),
				auditScope: nil,
				catalog:    &orchestrator.Catalog{},
				interval:   5,
			},
			want: func(t *testing.T, got *Service, msgAndArgs ...any) bool {
				return assert.False(t, got.scheduler.IsRunning())
			},
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.ErrorContains(t, err, "evaluation cannot be scheduled due to invalid input")
			},
		},
		{
			name: "error: invalid input - invalid interval",
			fields: fields{
				scheduler: gocron.NewScheduler(time.Local),
			},
			args: args{
				ctx:        context.Background(),
				auditScope: evaluationtest.MockAuditScope1,
				catalog:    &orchestrator.Catalog{},
			},
			want: func(t *testing.T, got *Service, msgAndArgs ...any) bool {
				return assert.False(t, got.scheduler.IsRunning())
			},
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.ErrorContains(t, err, "evaluation cannot be scheduled due to invalid input")
			},
		},
		{
			name: "happy path: job added successfully",
			fields: fields{
				scheduler: gocron.NewScheduler(time.Local),
			},
			args: args{
				ctx:        context.Background(),
				auditScope: evaluationtest.MockAuditScope1,
				catalog:    &orchestrator.Catalog{},
				interval:   5,
			},
			want: func(t *testing.T, got *Service, msgAndArgs ...any) bool {
				assert.Equal(t, 1, len(got.scheduler.Jobs()))
				return assert.Equal(t, evaluationtest.MockAuditScope1.Id, got.scheduler.Jobs()[0].Tags()[0])
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{
				scheduler: tt.fields.scheduler,
			}
			err := svc.addJobToScheduler(tt.args.ctx, tt.args.auditScope, tt.args.catalog, tt.args.interval)

			tt.wantErr(t, err)
			tt.want(t, svc)
		})
	}
}

// TestService_evaluateControl currently covers mainly the different switch cases regarding control evaluation.
// TODO(all): We could add more tests for covering other scenarios (e.g. ignored controls) and errors.
func TestService_evaluateControl(t *testing.T) {
	type args struct {
		ctx        context.Context
		auditScope *orchestrator.AuditScope
		catalog    *orchestrator.Catalog
		control    *orchestrator.Control
		manual     []*evaluation.EvaluationResult
		interval   int
	}
	type fields struct {
		orchestratorClient orchestratorconnect.OrchestratorClient
		catalogControls    map[string]map[string]*orchestrator.Control
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		wantSvc assert.Want[*Service]
		wantErr assert.WantErr
	}{
		{
			name: "happy path - without manual results",
			args: args{
				ctx:        context.Background(),
				auditScope: evaluationtest.MockAuditScope1,
				catalog:    evaluationtest.MockCatalog1,
				control:    evaluationtest.MockControl1,
				interval:   5,
			},
			fields: fields{
				orchestratorClient: newOrchestratorClient(t,
					WithAssessmentResults([]*assessment.AssessmentResult{
						{
							Id:                   evaluationtest.MockAssessmentResultId1,
							MetricId:             evaluationtest.MockMetricId1,
							Compliant:            true,
							ResourceId:           "resource-1",
							TargetOfEvaluationId: "00000000-0000-0000-0000-000000000001",
						},
						{
							Id:                   evaluationtest.MockAssessmentResultId2,
							MetricId:             evaluationtest.MockMetricId1,
							Compliant:            true,
							ResourceId:           "resource-2",
							TargetOfEvaluationId: "00000000-0000-0000-0000-000000000001",
						},
						{
							Id:                   evaluationtest.MockAssessmentResultId3,
							MetricId:             evaluationtest.MockMetricId2,
							Compliant:            true,
							ResourceId:           "resource-3",
							TargetOfEvaluationId: "00000000-0000-0000-0000-000000000001",
						},
					}),
				),
				catalogControls: map[string]map[string]*orchestrator.Control{
					evaluationtest.MockCatalog1.Id: {
						fmt.Sprintf("%s-%s", evaluationtest.MockControl1.CategoryName, evaluationtest.MockControl1.Id):     evaluationtest.MockControl1,
						fmt.Sprintf("%s-%s", evaluationtest.MockControl1.CategoryName, evaluationtest.MockSubcontrol11.Id): evaluationtest.MockSubcontrol11,
						fmt.Sprintf("%s-%s", evaluationtest.MockControl1.CategoryName, evaluationtest.MockSubcontrol12.Id): evaluationtest.MockSubcontrol12,
						fmt.Sprintf("%s-%s", evaluationtest.MockControl2.CategoryName, evaluationtest.MockControl2.Id):     evaluationtest.MockControl2,
						fmt.Sprintf("%s-%s", evaluationtest.MockControl2.CategoryName, evaluationtest.MockSubcontrol21.Id): evaluationtest.MockSubcontrol21,
					},
				},
			},
			wantSvc: func(t *testing.T, got *Service, msgAndArgs ...any) bool {
				// Assert that evaluation results were stored in the orchestrator (one for control, one for subcontrol)
				evalResults, err := got.orchestratorClient.ListEvaluationResults(context.Background(), connect.NewRequest(&orchestrator.ListEvaluationResultsRequest{}))
				assert.NoError(t, err)

				// We should have 3 results: one for Control 1 and two for Control 1.1 and 1.2 (sub controls)
				assert.Equal(t, 3, len(evalResults.Msg.Results))

				// Find the result for the main control (Control 1)
				var mainControlResult *evaluation.EvaluationResult
				for _, result := range evalResults.Msg.Results {
					if result.ControlId == evaluationtest.MockControlId1 {
						mainControlResult = result
						break
					}
				}
				assert.NotNil(t, mainControlResult, "Should have result for main control")

				// Verify the main control result fields that are not deterministic (ID and Timestamp) and the number of assessment results, then compare the rest of the fields
				assert.NotEmpty(t, mainControlResult.Id)
				assert.NotNil(t, mainControlResult.Timestamp)
				assert.Equal(t, 3, len(mainControlResult.AssessmentResultIds))

				want := &evaluation.EvaluationResult{
					TargetOfEvaluationId: evaluationtest.MockToeId1,
					AuditScopeId:         evaluationtest.MockAuditScopeId1,
					ControlId:            evaluationtest.MockControlId1,
					ControlCategoryName:  evaluationtest.MockCategoryName1,
					ControlCatalogId:     evaluationtest.MockCatalogId1,
					ParentControlId:      nil,
					Status:               evaluation.EvaluationStatus_EVALUATION_STATUS_COMPLIANT,
					Comment:              nil,
					ValidUntil:           nil,
					Data:                 nil,
				}
				return assert.Equal(t, want, mainControlResult, protocmp.IgnoreFields(&evaluation.EvaluationResult{}, "id", "timestamp", "assessment_result_ids"))
			},
			wantErr: assert.NoError,
		},
		{
			name: "happy path - with non-compliant assessment results but manual compliant result",
			args: args{
				ctx:        context.Background(),
				auditScope: evaluationtest.MockAuditScope1,
				catalog:    evaluationtest.MockCatalog1,
				control:    evaluationtest.MockControl1,
				manual:     []*evaluation.EvaluationResult{evaluationtest.MockManualEvaluationResult1},
			},
			fields: fields{
				orchestratorClient: newOrchestratorClient(t,
					WithAssessmentResults([]*assessment.AssessmentResult{
						{
							Id:                   evaluationtest.MockAssessmentResultId1,
							MetricId:             evaluationtest.MockMetricId1,
							Compliant:            true,
							ResourceId:           "resource-1",
							TargetOfEvaluationId: "00000000-0000-0000-0000-000000000001",
						},
						{
							Id:                   evaluationtest.MockAssessmentResultId2,
							MetricId:             evaluationtest.MockMetricId1,
							Compliant:            true,
							ResourceId:           "resource-2",
							TargetOfEvaluationId: "00000000-0000-0000-0000-000000000001",
						},
						{
							Id:                   evaluationtest.MockAssessmentResultId3,
							MetricId:             evaluationtest.MockMetricId2,
							Compliant:            true,
							ResourceId:           "resource-3",
							TargetOfEvaluationId: "00000000-0000-0000-0000-000000000001",
						},
					}),
				),
				catalogControls: map[string]map[string]*orchestrator.Control{
					evaluationtest.MockCatalog1.Id: {
						fmt.Sprintf("%s-%s", evaluationtest.MockControl1.CategoryName, evaluationtest.MockControl1.Id):     evaluationtest.MockControl1,
						fmt.Sprintf("%s-%s", evaluationtest.MockControl1.CategoryName, evaluationtest.MockSubcontrol11.Id): evaluationtest.MockSubcontrol11,
						fmt.Sprintf("%s-%s", evaluationtest.MockControl1.CategoryName, evaluationtest.MockSubcontrol12.Id): evaluationtest.MockSubcontrol12,
						fmt.Sprintf("%s-%s", evaluationtest.MockControl2.CategoryName, evaluationtest.MockControl2.Id):     evaluationtest.MockControl2,
						fmt.Sprintf("%s-%s", evaluationtest.MockControl2.CategoryName, evaluationtest.MockSubcontrol21.Id): evaluationtest.MockSubcontrol21,
					},
				},
			},
			wantSvc: func(t *testing.T, got *Service, msgAndArgs ...any) bool {
				// Assert that evaluation results were stored in the database (one for control, one for subcontrol)
				evalResults, err := got.orchestratorClient.ListEvaluationResults(context.Background(), connect.NewRequest(&orchestrator.ListEvaluationResultsRequest{}))
				assert.NoError(t, err)

				// We should have 3 results: one for Control 1 and two for Control 1.1 and 1.2 (sub controls)
				assert.Equal(t, 3, len(evalResults.Msg.Results))

				// Find the result for the main control (Control 1)
				var mainControlResult *evaluation.EvaluationResult
				for _, result := range evalResults.Msg.Results {
					if result.ControlId == evaluationtest.MockControlId1 {
						mainControlResult = result
						break
					}
				}
				assert.NotNil(t, mainControlResult, "Should have result for main control")

				// Verify the main control result fields that are not deterministic (ID and Timestamp) and the number of assessment results, then compare the rest of the fields
				assert.NotEmpty(t, mainControlResult.Id)
				assert.NotNil(t, mainControlResult.Timestamp)
				assert.Equal(t, 3, len(mainControlResult.AssessmentResultIds))

				want := &evaluation.EvaluationResult{
					TargetOfEvaluationId: evaluationtest.MockToeId1,
					AuditScopeId:         evaluationtest.MockAuditScopeId1,
					ControlId:            evaluationtest.MockControlId1,
					ControlCategoryName:  evaluationtest.MockCategoryName1,
					ControlCatalogId:     evaluationtest.MockCatalogId1,
					ParentControlId:      nil,
					Status:               evaluation.EvaluationStatus_EVALUATION_STATUS_COMPLIANT,
					Comment:              nil,
					ValidUntil:           nil,
					Data:                 nil,
				}

				return assert.Equal(t, want, mainControlResult, protocmp.IgnoreFields(&evaluation.EvaluationResult{}, "id", "timestamp", "assessment_result_ids"))
			},
			wantErr: assert.NoError,
		},
		{
			name: "happy path: non-compliant subcontrols",
			args: args{
				ctx:        context.Background(),
				auditScope: evaluationtest.MockAuditScope1,
				catalog:    evaluationtest.MockCatalog1,
				control:    evaluationtest.MockControl1,
			},
			fields: fields{
				orchestratorClient: newOrchestratorClient(t,
					WithAssessmentResults([]*assessment.AssessmentResult{
						{
							Id:                   evaluationtest.MockAssessmentResultId1,
							MetricId:             evaluationtest.MockMetricId1,
							Compliant:            false,
							ResourceId:           "resource-1",
							TargetOfEvaluationId: "00000000-0000-0000-0000-000000000001",
						},
						{
							Id:                   evaluationtest.MockAssessmentResultId2,
							MetricId:             evaluationtest.MockMetricId2,
							Compliant:            false,
							ResourceId:           "resource-2",
							TargetOfEvaluationId: "00000000-0000-0000-0000-000000000001",
						},
					}),
				),
				catalogControls: map[string]map[string]*orchestrator.Control{
					evaluationtest.MockCatalog1.Id: {
						fmt.Sprintf("%s-%s", evaluationtest.MockControl1.CategoryName, evaluationtest.MockControl1.Id):     evaluationtest.MockControl1,
						fmt.Sprintf("%s-%s", evaluationtest.MockControl1.CategoryName, evaluationtest.MockSubcontrol11.Id): evaluationtest.MockSubcontrol11,
						fmt.Sprintf("%s-%s", evaluationtest.MockControl1.CategoryName, evaluationtest.MockSubcontrol12.Id): evaluationtest.MockSubcontrol12,
					},
				},
			},
			wantSvc: func(t *testing.T, got *Service, msgAndArgs ...any) bool {
				evalResults, err := got.orchestratorClient.ListEvaluationResults(context.Background(), connect.NewRequest(&orchestrator.ListEvaluationResultsRequest{}))
				assert.NoError(t, err)

				if !assert.Equal(t, 3, len(evalResults.Msg.Results)) {
					return false
				}

				var mainControlResult *evaluation.EvaluationResult
				for _, result := range evalResults.Msg.Results {
					if result.ControlId == evaluationtest.MockControlId1 {
						mainControlResult = result
						break
					}
				}

				if !assert.NotNil(t, mainControlResult) {
					return false
				}

				want := &evaluation.EvaluationResult{
					TargetOfEvaluationId: evaluationtest.MockToeId1,
					AuditScopeId:         evaluationtest.MockAuditScopeId1,
					ControlId:            evaluationtest.MockControlId1,
					ControlCategoryName:  evaluationtest.MockCategoryName1,
					ControlCatalogId:     evaluationtest.MockCatalogId1,
					ParentControlId:      nil,
					Status:               evaluation.EvaluationStatus_EVALUATION_STATUS_NOT_COMPLIANT,
					Comment:              nil,
					ValidUntil:           nil,
					Data:                 nil,
				}

				return assert.Equal(t, want, mainControlResult, protocmp.IgnoreFields(&evaluation.EvaluationResult{}, "id", "timestamp", "assessment_result_ids"))
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := Service{
				orchestratorClient: tt.fields.orchestratorClient,
				catalogControls:    tt.fields.catalogControls,
			}

			gotErr := svc.evaluateControl(tt.args.ctx, tt.args.auditScope, tt.args.catalog, tt.args.control, tt.args.manual)

			tt.wantErr(t, gotErr)
			tt.wantSvc(t, &svc)
		})
	}
}

// TestService_evaluateCatalog covers happy paths with different scenarios (with/without manual results).
// Error cases are not tested currently.
func TestService_evaluateCatalog(t *testing.T) {
	type args struct {
		ctx        context.Context
		auditScope *orchestrator.AuditScope
		catalog    *orchestrator.Catalog
		interval   int
	}
	type fields struct {
		orchestratorClient orchestratorconnect.OrchestratorClient
		catalogControls    map[string]map[string]*orchestrator.Control
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[*Service]
		wantErr assert.WantErr
	}{
		{
			name: "happy path - evaluates all relevant controls in catalog",
			args: args{
				ctx:        context.Background(),
				auditScope: evaluationtest.MockAuditScope1,
				catalog:    evaluationtest.MockCatalog1,
				interval:   5,
			},
			fields: fields{
				orchestratorClient: newOrchestratorClient(t,
					WithAssessmentResults([]*assessment.AssessmentResult{
						{
							Id:                   evaluationtest.MockAssessmentResultId1,
							MetricId:             evaluationtest.MockMetricId1,
							Compliant:            true,
							ResourceId:           "resource-1",
							TargetOfEvaluationId: evaluationtest.MockToeId1,
						},
						{
							Id:                   evaluationtest.MockAssessmentResultId2,
							MetricId:             evaluationtest.MockMetricId2,
							Compliant:            true,
							ResourceId:           "resource-2",
							TargetOfEvaluationId: evaluationtest.MockToeId1,
						},
					}),
				),
				catalogControls: map[string]map[string]*orchestrator.Control{
					evaluationtest.MockCatalog1.Id: {
						fmt.Sprintf("%s-%s", evaluationtest.MockControl1.CategoryName, evaluationtest.MockControl1.Id):     evaluationtest.MockControl1,
						fmt.Sprintf("%s-%s", evaluationtest.MockControl1.CategoryName, evaluationtest.MockSubcontrol11.Id): evaluationtest.MockSubcontrol11,
						fmt.Sprintf("%s-%s", evaluationtest.MockControl1.CategoryName, evaluationtest.MockSubcontrol12.Id): evaluationtest.MockSubcontrol12,
						fmt.Sprintf("%s-%s", evaluationtest.MockControl2.CategoryName, evaluationtest.MockControl2.Id):     evaluationtest.MockControl2,
						fmt.Sprintf("%s-%s", evaluationtest.MockControl2.CategoryName, evaluationtest.MockSubcontrol21.Id): evaluationtest.MockSubcontrol21,
					},
				},
			},
			want: func(t *testing.T, got *Service, msgAndArgs ...any) bool {
				// Assert that evaluation results were created in the database
				evalResults, err := got.orchestratorClient.ListEvaluationResults(context.Background(), connect.NewRequest(&orchestrator.ListEvaluationResultsRequest{}))
				assert.NoError(t, err)

				// We should have 5 results total:
				// - 1 for Control 1 (parent)
				// - 1 for Control 1.1 (subcontrol)
				// - 1 for Control 1.2 (subcontrol)
				// - 1 for Control 2 (parent)
				// - 1 for Control 2.1 (subcontrol)
				assert.Equal(t, 5, len(evalResults.Msg.Results))

				// Verify parent controls have correct evaluation status
				for _, result := range evalResults.Msg.Results {
					if result.ControlId == evaluationtest.MockControlId1 && result.ParentControlId == nil {
						assert.Equal(t, evaluation.EvaluationStatus_EVALUATION_STATUS_COMPLIANT, result.Status)
					}
					if result.ControlId == evaluationtest.MockControlId2 && result.ParentControlId == nil {
						assert.Equal(t, evaluation.EvaluationStatus_EVALUATION_STATUS_COMPLIANT, result.Status)
					}
				}

				return true
			},
			wantErr: assert.NoError,
		},
		{
			name: "happy path - with manual results ignores parent control",
			args: args{
				ctx:        context.Background(),
				auditScope: evaluationtest.MockAuditScope1,
				catalog:    evaluationtest.MockCatalog1,
				interval:   5,
			},
			fields: fields{
				orchestratorClient: newOrchestratorClient(t,
					WithAssessmentResults([]*assessment.AssessmentResult{
						{
							Id:                   evaluationtest.MockAssessmentResultId1,
							MetricId:             evaluationtest.MockMetricId1,
							Compliant:            true,
							ResourceId:           "resource-1",
							TargetOfEvaluationId: evaluationtest.MockToeId1,
						},
						{
							Id:                   evaluationtest.MockAssessmentResultId2,
							MetricId:             evaluationtest.MockMetricId2,
							Compliant:            true,
							ResourceId:           "resource-2",
							TargetOfEvaluationId: evaluationtest.MockToeId1,
						},
					}),
					WithEvaluationResults([]*evaluation.EvaluationResult{
						evaluationtest.MockManualEvaluationResult1,
					}),
				),
				catalogControls: map[string]map[string]*orchestrator.Control{
					evaluationtest.MockCatalog1.Id: {
						fmt.Sprintf("%s-%s", evaluationtest.MockControl1.CategoryName, evaluationtest.MockControl1.Id):     evaluationtest.MockControl1,
						fmt.Sprintf("%s-%s", evaluationtest.MockControl1.CategoryName, evaluationtest.MockSubcontrol11.Id): evaluationtest.MockSubcontrol11,
						fmt.Sprintf("%s-%s", evaluationtest.MockControl1.CategoryName, evaluationtest.MockSubcontrol12.Id): evaluationtest.MockSubcontrol12,
						fmt.Sprintf("%s-%s", evaluationtest.MockControl2.CategoryName, evaluationtest.MockControl2.Id):     evaluationtest.MockControl2,
						fmt.Sprintf("%s-%s", evaluationtest.MockControl2.CategoryName, evaluationtest.MockSubcontrol21.Id): evaluationtest.MockSubcontrol21,
					},
				},
			},
			want: func(t *testing.T, got *Service, msgAndArgs ...any) bool {
				evalResults, err := got.orchestratorClient.ListEvaluationResults(context.Background(), connect.NewRequest(&orchestrator.ListEvaluationResultsRequest{}))
				assert.NoError(t, err)

				// We should have 5 results total:
				// - 1 for Control 1 (parent) - compliant manually (due to manual result) -> subcontrols (Control 1.1 and Control 1.2) are ignored
				// - 1 for Control 2 (parent) - compliant
				// - 1 for Control 2.1 (subcontrol) - compliant
				assert.Equal(t, 3, len(evalResults.Msg.Results))

				// Extract control IDs from results
				controlIds := make([]string, len(evalResults.Msg.Results))
				for i, result := range evalResults.Msg.Results {
					controlIds[i] = result.ControlId
				}

				// Verify expected controls are present
				assert.Contains(t, controlIds, evaluationtest.MockControlId1) // Manual result
				assert.Contains(t, controlIds, evaluationtest.MockControlId2)
				assert.Contains(t, controlIds, evaluationtest.MockSubcontrolID21)

				return true
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := Service{
				orchestratorClient: tt.fields.orchestratorClient,
				catalogControls:    tt.fields.catalogControls,
			}

			gotErr := svc.evaluateCatalog(tt.args.ctx, tt.args.auditScope, tt.args.catalog, tt.args.interval)
			tt.wantErr(t, gotErr)
			tt.want(t, &svc)
		})
	}
}

func TestService_evaluateSubcontrol(t *testing.T) {
	type fields struct {
		orchestratorClient orchestratorconnect.OrchestratorClient
		catalogControls    map[string]map[string]*orchestrator.Control
	}
	type args struct {
		ctx        context.Context
		auditScope *orchestrator.AuditScope
		control    *orchestrator.Control
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    assert.Want[*evaluation.EvaluationResult]
		wantSvc assert.Want[*Service]
		wantErr assert.WantErr
	}{
		{
			name: "input missing - audit scope nil",
			fields: fields{
				orchestratorClient: nil,
				catalogControls:    map[string]map[string]*orchestrator.Control{},
			},
			args: args{
				ctx:        context.Background(),
				auditScope: nil,
				control:    orchestratortest.MockControl1,
			},
			want:    assert.Nil[*evaluation.EvaluationResult],
			wantErr: assert.NoError,
			wantSvc: assert.NotNil[*Service],
		},
		{
			name: "input missing - control nil",
			fields: fields{
				orchestratorClient: nil,
				catalogControls:    map[string]map[string]*orchestrator.Control{},
			},
			args: args{
				ctx: context.Background(),
				auditScope: &orchestrator.AuditScope{
					Id:                   evaluationtest.MockAuditScopeId1,
					TargetOfEvaluationId: evaluationtest.MockToeId1,
					CatalogId:            orchestratortest.MockCatalogId1,
				},
				control: nil,
			},
			want:    assert.Nil[*evaluation.EvaluationResult],
			wantErr: assert.NoError,
			wantSvc: assert.NotNil[*Service],
		},
		{
			name: "error - getAllMetricsFromControl fails (control not cached)",
			fields: fields{
				orchestratorClient: nil,
				catalogControls:    map[string]map[string]*orchestrator.Control{},
			},
			args: args{
				ctx: context.Background(),
				auditScope: &orchestrator.AuditScope{
					Id:                   evaluationtest.MockAuditScopeId1,
					TargetOfEvaluationId: evaluationtest.MockToeId1,
					CatalogId:            orchestratortest.MockCatalogId1,
				},
				control: orchestratortest.MockControl1,
			},
			want: assert.Nil[*evaluation.EvaluationResult],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.ErrorContains(t, err, "could not get control for control id {control-1}: ")
			},
			wantSvc: assert.NotNil[*Service],
		},
		{
			name: "happy path - no metrics",
			fields: func() fields {
				handler := &mockOrchestratorHandler{}
				_, testSrv := servertest.NewTestConnectServer(
					t,
					server.WithHandler(orchestratorconnect.NewOrchestratorHandler(handler)),
				)
				t.Cleanup(testSrv.Close)

				// Control with no metrics.
				ctrl := &orchestrator.Control{
					Id:                orchestratortest.MockControlId2,
					CategoryName:      orchestratortest.MockCategoryName2,
					CategoryCatalogId: orchestratortest.MockCatalogId2,
					Name:              "Mock Control 2",
					Metrics:           nil,
					Controls:          nil,
				}

				return fields{
					orchestratorClient: newOrchestratorClientForTest(testSrv),
					catalogControls: map[string]map[string]*orchestrator.Control{
						orchestratortest.MockCatalogId2: {
							fmt.Sprintf("%s-%s", ctrl.GetCategoryName(), ctrl.GetId()): ctrl,
						},
					},
				}
			}(),
			args: args{
				ctx: context.Background(),
				auditScope: &orchestrator.AuditScope{
					Id:                   evaluationtest.MockAuditScopeId2,
					TargetOfEvaluationId: evaluationtest.MockToeId2,
					CatalogId:            orchestratortest.MockCatalogId2,
				},
				control: &orchestrator.Control{
					Id:                orchestratortest.MockControlId2,
					CategoryName:      orchestratortest.MockCategoryName2,
					CategoryCatalogId: orchestratortest.MockCatalogId2,
					Name:              "Mock Control 2",
					Metrics:           nil,
					Controls:          nil,
				},
			},
			want: func(t *testing.T, got *evaluation.EvaluationResult, _ ...any) bool {
				assert.NotNil(t, got)
				assert.NotEmpty(t, got.GetId())
				assert.NotEmpty(t, got.GetTimestamp())
				assert.Empty(t, got.GetAssessmentResultIds())

				want := &evaluation.EvaluationResult{
					TargetOfEvaluationId: evaluationtest.MockToeId2,
					AuditScopeId:         evaluationtest.MockAuditScopeId2,
					ControlId:            orchestratortest.MockControlId2,
					ControlCategoryName:  orchestratortest.MockCategoryName2,
					ControlCatalogId:     orchestratortest.MockCatalogId2,
					ParentControlId:      nil,
					Status:               evaluation.EvaluationStatus_EVALUATION_STATUS_PENDING,
					Comment:              nil,
					ValidUntil:           nil,
					Data:                 nil,
				}

				return assert.Equal(t, want, got, protocmp.IgnoreFields(&evaluation.EvaluationResult{}, "id", "timestamp", "assessment_result_ids"))
			},
			wantSvc: func(t *testing.T, got *Service, _ ...any) bool {
				// Verify it was stored
				res, err := got.orchestratorClient.ListEvaluationResults(context.Background(), connect.NewRequest(&orchestrator.ListEvaluationResultsRequest{}))
				assert.NoError(t, err)
				assert.Equal(t, 1, len(res.Msg.Results))
				assert.NotNil(t, got)
				assert.NotEmpty(t, res.Msg.Results[0].GetId())
				assert.NotEmpty(t, res.Msg.Results[0].GetTimestamp())
				assert.Empty(t, res.Msg.Results[0].GetAssessmentResultIds())

				want := &evaluation.EvaluationResult{
					TargetOfEvaluationId: evaluationtest.MockToeId2,
					AuditScopeId:         evaluationtest.MockAuditScopeId2,
					ControlId:            orchestratortest.MockControlId2,
					ControlCategoryName:  orchestratortest.MockCategoryName2,
					ControlCatalogId:     orchestratortest.MockCatalogId2,
					ParentControlId:      nil,
					Status:               evaluation.EvaluationStatus_EVALUATION_STATUS_PENDING,
					Comment:              nil,
					ValidUntil:           nil,
					Data:                 nil,
				}

				return assert.Equal(t, want, res.Msg.Results[0], protocmp.IgnoreFields(&evaluation.EvaluationResult{}, "id", "timestamp", "assessment_result_ids"))
			},
			wantErr: assert.NoError,
		},
		{
			name: "happy path - metrics but ListAssessmentResults returns error => pending, stored",
			fields: func() fields {
				handler := &mockOrchestratorHandler{
					listAssessmentResultError: connect.NewError(connect.CodeInternal, fmt.Errorf("boom")),
				}
				_, testSrv := servertest.NewTestConnectServer(
					t,
					server.WithHandler(orchestratorconnect.NewOrchestratorHandler(handler)),
				)
				t.Cleanup(testSrv.Close)

				// Control with one metric.
				ctrl := &orchestrator.Control{
					Id:                orchestratortest.MockSubControlId1,
					CategoryName:      orchestratortest.MockCategoryName1,
					CategoryCatalogId: orchestratortest.MockCatalogId1,
					Name:              "Mock Subcontrol 1",
					Metrics: []*assessment.Metric{
						{Id: orchestratortest.MockMetricId1},
					},
				}

				return fields{
					orchestratorClient: newOrchestratorClientForTest(testSrv),
					catalogControls: map[string]map[string]*orchestrator.Control{
						orchestratortest.MockCatalogId1: {
							fmt.Sprintf("%s-%s", ctrl.GetCategoryName(), ctrl.GetId()): ctrl,
						},
					},
				}
			}(),
			args: args{
				ctx: context.Background(),
				auditScope: &orchestrator.AuditScope{
					Id:                   evaluationtest.MockAuditScopeId1,
					TargetOfEvaluationId: evaluationtest.MockToeId1,
					CatalogId:            orchestratortest.MockCatalogId1,
				},
				control: &orchestrator.Control{
					Id:                orchestratortest.MockSubControlId1,
					CategoryName:      orchestratortest.MockCategoryName1,
					CategoryCatalogId: orchestratortest.MockCatalogId1,
					Name:              "Mock Subcontrol 1",
					Metrics: []*assessment.Metric{
						{Id: orchestratortest.MockMetricId1},
					},
				},
			},
			want: func(t *testing.T, got *evaluation.EvaluationResult, _ ...any) bool {
				assert.NotNil(t, got)
				assert.NotEmpty(t, got.GetId())
				assert.NotEmpty(t, got.GetTimestamp())
				assert.Empty(t, got.GetAssessmentResultIds())

				want := &evaluation.EvaluationResult{
					TargetOfEvaluationId: evaluationtest.MockToeId1,
					AuditScopeId:         evaluationtest.MockAuditScopeId1,
					ControlId:            orchestratortest.MockSubControlId1,
					ControlCategoryName:  orchestratortest.MockCategoryName1,
					ControlCatalogId:     orchestratortest.MockCatalogId1,
					ParentControlId:      nil,
					Status:               evaluation.EvaluationStatus_EVALUATION_STATUS_PENDING,
					Comment:              nil,
					ValidUntil:           nil,
					Data:                 nil,
				}

				return assert.Equal(t, want, got, protocmp.IgnoreFields(&evaluation.EvaluationResult{}, "id", "timestamp", "assessment_result_ids"))
			},
			wantSvc: func(t *testing.T, got *Service, _ ...any) bool {
				// Verify it was stored
				res, err := got.orchestratorClient.ListEvaluationResults(context.Background(), connect.NewRequest(&orchestrator.ListEvaluationResultsRequest{}))
				assert.NoError(t, err)
				assert.Equal(t, 1, len(res.Msg.Results))
				assert.NotNil(t, got)
				assert.NotEmpty(t, res.Msg.Results[0].GetId())
				assert.NotEmpty(t, res.Msg.Results[0].GetTimestamp())
				assert.Empty(t, res.Msg.Results[0].GetAssessmentResultIds())

				want := &evaluation.EvaluationResult{
					TargetOfEvaluationId: evaluationtest.MockToeId1,
					AuditScopeId:         evaluationtest.MockAuditScopeId1,
					ControlId:            orchestratortest.MockSubControlId1,
					ControlCategoryName:  orchestratortest.MockCategoryName1,
					ControlCatalogId:     orchestratortest.MockCatalogId1,
					ParentControlId:      nil,
					Status:               evaluation.EvaluationStatus_EVALUATION_STATUS_PENDING,
					Comment:              nil,
					ValidUntil:           nil,
					Data:                 nil,
				}

				return assert.Equal(t, want, res.Msg.Results[0], protocmp.IgnoreFields(&evaluation.EvaluationResult{}, "id", "timestamp", "assessment_result_ids"))
			},
			wantErr: assert.NoError,
		},
		{
			name: "happy path - assessment results all compliant => compliant, includes assessment ids",
			fields: func() fields {
				return fields{
					orchestratorClient: newOrchestratorClient(t,
						WithAssessmentResults([]*assessment.AssessmentResult{
							{
								Id:                   "assessment-result-1",
								MetricId:             orchestratortest.MockMetricId1,
								Compliant:            true,
								ResourceId:           "resource-1",
								TargetOfEvaluationId: evaluationtest.MockToeId1,
							},
							{
								Id:                   "assessment-result-2",
								MetricId:             orchestratortest.MockMetricId1,
								Compliant:            true,
								ResourceId:           "resource-2",
								TargetOfEvaluationId: evaluationtest.MockToeId1,
							},
						}),
					),
					catalogControls: map[string]map[string]*orchestrator.Control{
						orchestratortest.MockCatalogId1: {
							fmt.Sprintf("%s-%s", orchestratortest.MockSubControl1.GetCategoryName(), orchestratortest.MockSubControl1.GetId()): orchestratortest.MockSubControl1,
						},
					},
				}
			}(),
			args: args{
				ctx: context.Background(),
				auditScope: &orchestrator.AuditScope{
					Id:                   evaluationtest.MockAuditScopeId1,
					TargetOfEvaluationId: evaluationtest.MockToeId1,
					CatalogId:            orchestratortest.MockCatalogId1,
				},
				control: orchestratortest.MockSubControl1,
			},
			want: func(t *testing.T, got *evaluation.EvaluationResult, _ ...any) bool {
				assert.NotNil(t, got)
				assert.NotEmpty(t, got.GetId())
				assert.NotEmpty(t, got.GetTimestamp())
				assert.Equal(t, 2, len(got.AssessmentResultIds))
				assert.Contains(t, got.AssessmentResultIds, "assessment-result-1")
				assert.Contains(t, got.AssessmentResultIds, "assessment-result-2")

				want := &evaluation.EvaluationResult{
					TargetOfEvaluationId: evaluationtest.MockToeId1,
					AuditScopeId:         evaluationtest.MockAuditScopeId1,
					ControlId:            orchestratortest.MockSubControlId1,
					ControlCategoryName:  orchestratortest.MockCategoryName1,
					ControlCatalogId:     orchestratortest.MockCatalogId1,
					ParentControlId:      new(orchestratortest.MockControlId1),
					Status:               evaluation.EvaluationStatus_EVALUATION_STATUS_COMPLIANT,
					Comment:              nil,
					ValidUntil:           nil,
					Data:                 nil,
				}

				return assert.Equal(t, want, got, protocmp.IgnoreFields(&evaluation.EvaluationResult{}, "id", "timestamp", "assessment_result_ids"))
			},
			wantErr: assert.NoError,
			wantSvc: func(t *testing.T, got *Service, _ ...any) bool {
				// Verify it was stored
				res, err := got.orchestratorClient.ListEvaluationResults(context.Background(), connect.NewRequest(&orchestrator.ListEvaluationResultsRequest{}))
				assert.NoError(t, err)
				assert.Equal(t, 1, len(res.Msg.Results))
				assert.NotNil(t, got)
				assert.NotEmpty(t, res.Msg.Results[0].GetId())
				assert.NotEmpty(t, res.Msg.Results[0].GetTimestamp())
				assert.Equal(t, 2, len(res.Msg.Results[0].GetAssessmentResultIds()))

				want := &evaluation.EvaluationResult{
					TargetOfEvaluationId: evaluationtest.MockToeId1,
					AuditScopeId:         evaluationtest.MockAuditScopeId1,
					ControlId:            orchestratortest.MockSubControlId1,
					ControlCategoryName:  orchestratortest.MockCategoryName1,
					ControlCatalogId:     orchestratortest.MockCatalogId1,
					ParentControlId:      new(orchestratortest.MockControlId1),
					Status:               evaluation.EvaluationStatus_EVALUATION_STATUS_COMPLIANT,
					Comment:              nil,
					ValidUntil:           nil,
					Data:                 nil,
				}

				return assert.Equal(t, want, res.Msg.Results[0], protocmp.IgnoreFields(&evaluation.EvaluationResult{}, "id", "timestamp", "assessment_result_ids"))
			},
		},
		{
			name: "happy path - assessment results include non-compliant => not compliant",
			fields: func() fields {
				return fields{
					orchestratorClient: newOrchestratorClient(t,
						WithAssessmentResults([]*assessment.AssessmentResult{
							{
								Id:                   "assessment-result-1",
								MetricId:             orchestratortest.MockMetricId1,
								Compliant:            true,
								ResourceId:           "resource-1",
								TargetOfEvaluationId: evaluationtest.MockToeId1,
							},
							{
								Id:                   "assessment-result-2",
								MetricId:             orchestratortest.MockMetricId1,
								Compliant:            false,
								ResourceId:           "resource-2",
								TargetOfEvaluationId: evaluationtest.MockToeId1,
							},
						}),
					),
					catalogControls: map[string]map[string]*orchestrator.Control{
						orchestratortest.MockCatalogId1: {
							fmt.Sprintf("%s-%s", orchestratortest.MockSubControl1.GetCategoryName(), orchestratortest.MockSubControl1.GetId()): orchestratortest.MockSubControl1,
						},
					},
				}
			}(),
			args: args{
				ctx: context.Background(),
				auditScope: &orchestrator.AuditScope{
					Id:                   evaluationtest.MockAuditScopeId1,
					TargetOfEvaluationId: evaluationtest.MockToeId1,
					CatalogId:            orchestratortest.MockCatalogId1,
				},
				control: orchestratortest.MockSubControl1,
			},
			want: func(t *testing.T, got *evaluation.EvaluationResult, _ ...any) bool {
				assert.NotNil(t, got)
				assert.NotEmpty(t, got.GetId())
				assert.NotEmpty(t, got.GetTimestamp())
				assert.Equal(t, 2, len(got.AssessmentResultIds))

				want := &evaluation.EvaluationResult{
					TargetOfEvaluationId: orchestratortest.MockToeId1,
					AuditScopeId:         evaluationtest.MockAuditScopeId1,
					ControlId:            orchestratortest.MockSubControlId1,
					ControlCategoryName:  orchestratortest.MockCategoryName1,
					ControlCatalogId:     orchestratortest.MockCatalogId1,
					ParentControlId:      orchestratortest.MockSubControl1.ParentControlId,
					Status:               evaluation.EvaluationStatus_EVALUATION_STATUS_NOT_COMPLIANT,
					Comment:              nil,
					ValidUntil:           nil,
					Data:                 nil,
				}
				return assert.Equal(t, want, got, protocmp.IgnoreFields(&evaluation.EvaluationResult{}, "id", "timestamp", "assessment_result_ids"))
			},
			wantSvc: func(t *testing.T, got *Service, _ ...any) bool {
				res, err := got.orchestratorClient.ListEvaluationResults(context.Background(), connect.NewRequest(&orchestrator.ListEvaluationResultsRequest{}))
				assert.NoError(t, err)

				assert.NotNil(t, got)
				assert.NotEmpty(t, res.Msg.Results[0].GetId())
				assert.NotEmpty(t, res.Msg.Results[0].GetTimestamp())
				assert.Equal(t, 2, len(res.Msg.Results[0].AssessmentResultIds))

				want := &evaluation.EvaluationResult{
					TargetOfEvaluationId: orchestratortest.MockToeId1,
					AuditScopeId:         evaluationtest.MockAuditScopeId1,
					ControlId:            orchestratortest.MockSubControlId1,
					ControlCategoryName:  orchestratortest.MockCategoryName1,
					ControlCatalogId:     orchestratortest.MockCatalogId1,
					ParentControlId:      orchestratortest.MockSubControl1.ParentControlId,
					Status:               evaluation.EvaluationStatus_EVALUATION_STATUS_NOT_COMPLIANT,
					Comment:              nil,
					ValidUntil:           nil,
					Data:                 nil,
				}

				return assert.Equal(t, 1, len(res.Msg.Results)) &&
					assert.Equal(t, want, res.Msg.Results[0], protocmp.IgnoreFields(&evaluation.EvaluationResult{}, "id", "timestamp", "assessment_result_ids"))
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{
				orchestratorClient: tt.fields.orchestratorClient,
				catalogControls:    tt.fields.catalogControls,
			}

			got, gotErr := svc.evaluateSubcontrol(tt.args.ctx, tt.args.auditScope, tt.args.control)

			tt.want(t, got)
			tt.wantErr(t, gotErr)
			tt.wantSvc(t, svc)
		})
	}
}

func TestService_StartEvaluation(t *testing.T) {
	type args struct {
		ctx context.Context
		req *connect.Request[evaluation.StartEvaluationRequest]
	}
	type fields struct {
		orchestratorClient orchestratorconnect.OrchestratorClient
		catalogControls    map[string]map[string]*orchestrator.Control
		scheduler          *gocron.Scheduler
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[*connect.Response[evaluation.StartEvaluationResponse]]
		wantSvc assert.Want[*Service]
		wantErr assert.WantErr
	}{
		{
			name: "err: evaluation already started for audit scope",
			args: args{
				ctx: context.Background(),
				req: connect.NewRequest(&evaluation.StartEvaluationRequest{
					AuditScopeId: evaluationtest.MockAuditScopeId1,
				}),
			},
			fields: fields{
				orchestratorClient: newOrchestratorClient(t,
					WithAuditScope(evaluationtest.MockAuditScope1),
					WithControls(
						evaluationtest.MockControl1.Controls,
						evaluationtest.MockControl2.Controls,
						[]*orchestrator.Control{evaluationtest.MockControl1, evaluationtest.MockControl2},
					),
					WithCatalog(evaluationtest.MockCatalog1),
				),
				scheduler: func() *gocron.Scheduler {
					s := gocron.NewScheduler(time.Local)
					_, err := s.Every(1).
						Day().
						Tag(evaluationtest.MockAuditScopeId1).
						Do(func() { fmt.Println("Scheduler") })
					assert.NoError(t, err)

					return s
				}(),
				catalogControls: map[string]map[string]*orchestrator.Control{
					evaluationtest.MockCatalog1.Id: {
						fmt.Sprintf("%s-%s", evaluationtest.MockControl1.CategoryName, evaluationtest.MockControl1.Id):     evaluationtest.MockControl1,
						fmt.Sprintf("%s-%s", evaluationtest.MockControl1.CategoryName, evaluationtest.MockSubcontrol11.Id): evaluationtest.MockSubcontrol11,
						fmt.Sprintf("%s-%s", evaluationtest.MockControl1.CategoryName, evaluationtest.MockSubcontrol12.Id): evaluationtest.MockSubcontrol12,
						fmt.Sprintf("%s-%s", evaluationtest.MockControl2.CategoryName, evaluationtest.MockControl2.Id):     evaluationtest.MockControl2,
						fmt.Sprintf("%s-%s", evaluationtest.MockControl2.CategoryName, evaluationtest.MockSubcontrol21.Id): evaluationtest.MockSubcontrol21,
					},
				},
			},
			want:    assert.Nil[*connect.Response[evaluation.StartEvaluationResponse]],
			wantSvc: assert.NotNil[*Service],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeAlreadyExists) &&
					assert.ErrorContains(t, err, "evaluation already started for the given audit scope")
			},
		},
		{
			name: "err: gettging catalog error",
			args: args{
				ctx: context.Background(),
				req: connect.NewRequest(&evaluation.StartEvaluationRequest{
					AuditScopeId: evaluationtest.MockAuditScopeId1,
				}),
			},
			fields: fields{
				orchestratorClient: newOrchestratorClient(t,
					WithAuditScope(evaluationtest.MockAuditScope1),
					WithControls(
						evaluationtest.MockControl1.Controls,
						evaluationtest.MockControl2.Controls,
						[]*orchestrator.Control{evaluationtest.MockControl1, evaluationtest.MockControl2},
					),
					WithGetCatalogNotFoundError(fmt.Errorf("catalog not found")),
				),
				scheduler: gocron.NewScheduler(time.Local),
				catalogControls: map[string]map[string]*orchestrator.Control{
					evaluationtest.MockCatalog1.Id: {
						fmt.Sprintf("%s-%s", evaluationtest.MockControl1.CategoryName, evaluationtest.MockControl1.Id):     evaluationtest.MockControl1,
						fmt.Sprintf("%s-%s", evaluationtest.MockControl1.CategoryName, evaluationtest.MockSubcontrol11.Id): evaluationtest.MockSubcontrol11,
						fmt.Sprintf("%s-%s", evaluationtest.MockControl1.CategoryName, evaluationtest.MockSubcontrol12.Id): evaluationtest.MockSubcontrol12,
						fmt.Sprintf("%s-%s", evaluationtest.MockControl2.CategoryName, evaluationtest.MockControl2.Id):     evaluationtest.MockControl2,
						fmt.Sprintf("%s-%s", evaluationtest.MockControl2.CategoryName, evaluationtest.MockSubcontrol21.Id): evaluationtest.MockSubcontrol21,
					},
				},
			},
			want:    assert.Nil[*connect.Response[evaluation.StartEvaluationResponse]],
			wantSvc: assert.NotNil[*Service],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInternal) &&
					assert.ErrorContains(t, err, "could not get catalog from the orchestrator")
			},
		},
		{
			name: "err: getting controls from orchestrator returns error",
			args: args{
				ctx: context.Background(),
				req: connect.NewRequest(&evaluation.StartEvaluationRequest{
					AuditScopeId: evaluationtest.MockAuditScopeId1,
				}),
			},
			fields: fields{
				orchestratorClient: newOrchestratorClient(t,
					WithAuditScope(&orchestrator.AuditScope{
						Id:                   evaluationtest.MockAuditScopeId1,
						TargetOfEvaluationId: evaluationtest.MockToeId1,
					})),
				scheduler: gocron.NewScheduler(time.Local),
			},
			want:    assert.Nil[*connect.Response[evaluation.StartEvaluationResponse]],
			wantSvc: assert.NotNil[*Service],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInternal) &&
					assert.ErrorContains(t, err, "could not cache controls")
			},
		},
		{
			name: "err: GetAuditScope returns error",
			args: args{
				ctx: context.Background(),
				req: connect.NewRequest(&evaluation.StartEvaluationRequest{
					AuditScopeId: evaluationtest.MockAuditScopeId1,
				}),
			},
			fields: fields{
				orchestratorClient: newOrchestratorClient(t),
			},
			want: assert.Nil[*connect.Response[evaluation.StartEvaluationResponse]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeNotFound) &&
					assert.ErrorContains(t, err, "could not get audit scope from orchestrator")
			},
			wantSvc: assert.NotNil[*Service],
		},
		{
			name: "err: invalid request - empty request",
			args: args{
				ctx: context.Background(),
				req: connect.NewRequest(&evaluation.StartEvaluationRequest{}),
			},
			want: assert.Nil[*connect.Response[evaluation.StartEvaluationResponse]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument) &&
					assert.ErrorContains(t, err, "invalid request")
			},
			wantSvc: assert.NotNil[*Service],
		},
		{
			name: "happy path: with interval argument",
			args: args{
				ctx: context.Background(),
				req: connect.NewRequest(&evaluation.StartEvaluationRequest{
					AuditScopeId: evaluationtest.MockAuditScopeId1,
					Interval:     new((int32(10))),
				}),
			},
			fields: fields{
				orchestratorClient: newOrchestratorClient(t,
					WithAuditScope(evaluationtest.MockAuditScope1),
					WithControls(
						evaluationtest.MockControl1.Controls,
						evaluationtest.MockControl2.Controls,
						[]*orchestrator.Control{evaluationtest.MockControl1, evaluationtest.MockControl2},
					),
					WithCatalog(evaluationtest.MockCatalog1),
				),
				scheduler: gocron.NewScheduler(time.Local),
				catalogControls: map[string]map[string]*orchestrator.Control{
					evaluationtest.MockCatalog1.Id: {
						fmt.Sprintf("%s-%s", evaluationtest.MockControl1.CategoryName, evaluationtest.MockControl1.Id):     evaluationtest.MockControl1,
						fmt.Sprintf("%s-%s", evaluationtest.MockControl1.CategoryName, evaluationtest.MockSubcontrol11.Id): evaluationtest.MockSubcontrol11,
						fmt.Sprintf("%s-%s", evaluationtest.MockControl1.CategoryName, evaluationtest.MockSubcontrol12.Id): evaluationtest.MockSubcontrol12,
						fmt.Sprintf("%s-%s", evaluationtest.MockControl2.CategoryName, evaluationtest.MockControl2.Id):     evaluationtest.MockControl2,
						fmt.Sprintf("%s-%s", evaluationtest.MockControl2.CategoryName, evaluationtest.MockSubcontrol21.Id): evaluationtest.MockSubcontrol21,
					},
				},
			},
			want: func(t *testing.T, got *connect.Response[evaluation.StartEvaluationResponse], _ ...any) bool {
				assert.NotNil(t, got)
				return assert.True(t, got.Msg.GetSuccessful())
			},
			wantErr: assert.NoError,
			wantSvc: func(t *testing.T, got *Service, msgAndArgs ...any) bool {
				return assert.Equal(t, 10, got.scheduler.Jobs()[0].ScheduledInterval())
			},
		},
		{
			name: "happy path",
			args: args{
				ctx: context.Background(),
				req: connect.NewRequest(&evaluation.StartEvaluationRequest{
					AuditScopeId: evaluationtest.MockAuditScopeId1,
				}),
			},
			fields: fields{
				orchestratorClient: newOrchestratorClient(t,
					WithAuditScope(evaluationtest.MockAuditScope1),
					WithControls(
						evaluationtest.MockControl1.Controls,
						evaluationtest.MockControl2.Controls,
						[]*orchestrator.Control{evaluationtest.MockControl1, evaluationtest.MockControl2},
					),
					WithCatalog(evaluationtest.MockCatalog1),
				),
				scheduler: gocron.NewScheduler(time.Local),
				catalogControls: map[string]map[string]*orchestrator.Control{
					evaluationtest.MockCatalog1.Id: {
						fmt.Sprintf("%s-%s", evaluationtest.MockControl1.CategoryName, evaluationtest.MockControl1.Id):     evaluationtest.MockControl1,
						fmt.Sprintf("%s-%s", evaluationtest.MockControl1.CategoryName, evaluationtest.MockSubcontrol11.Id): evaluationtest.MockSubcontrol11,
						fmt.Sprintf("%s-%s", evaluationtest.MockControl1.CategoryName, evaluationtest.MockSubcontrol12.Id): evaluationtest.MockSubcontrol12,
						fmt.Sprintf("%s-%s", evaluationtest.MockControl2.CategoryName, evaluationtest.MockControl2.Id):     evaluationtest.MockControl2,
						fmt.Sprintf("%s-%s", evaluationtest.MockControl2.CategoryName, evaluationtest.MockSubcontrol21.Id): evaluationtest.MockSubcontrol21,
					},
				},
			},
			want: func(t *testing.T, got *connect.Response[evaluation.StartEvaluationResponse], _ ...any) bool {
				assert.NotNil(t, got)
				return assert.True(t, got.Msg.GetSuccessful())
			},
			wantSvc: assert.NotNil[*Service],
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			svc := Service{
				orchestratorClient: tt.fields.orchestratorClient,
				catalogControls:    tt.fields.catalogControls,
				scheduler:          tt.fields.scheduler,
			}
			got, gotErr := svc.StartEvaluation(tt.args.ctx, tt.args.req)

			tt.want(t, got)
			tt.wantErr(t, gotErr)
			tt.wantSvc(t, &svc)
		})
	}
}
