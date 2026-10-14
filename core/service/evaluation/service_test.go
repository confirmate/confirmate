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

package evaluation

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"testing"
	"time"

	"confirmate.io/core/api/assessment"
	"confirmate.io/core/api/evaluation"
	"confirmate.io/core/api/evaluation/evaluationconnect"
	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/api/orchestrator/orchestratorconnect"
	"confirmate.io/core/auth"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/go-cmp/cmp/cmpopts"
	"golang.org/x/oauth2/clientcredentials"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/timestamppb"

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

// denyAuthorizationStrategy is a test strategy that denies all access.
type denyAuthorizationStrategy struct{}

func (*denyAuthorizationStrategy) CheckAccess(_ context.Context, _ string, _ orchestrator.RequestType, _ orchestrator.UserPermission_Permission, _ string, _ orchestrator.ObjectType) (bool, []string) {
	return false, nil
}

func (*denyAuthorizationStrategy) AllowedTargetOfEvaluations(_ context.Context) (bool, []string) {
	return false, nil
}

func (*denyAuthorizationStrategy) AllowedAuditScopes(_ context.Context) (bool, []string) {
	return false, nil
}

func (*denyAuthorizationStrategy) AllowedUserPermission(_ context.Context) (bool, []string) {
	return false, nil
}

// partialScopeAuthorizationStrategy is a test strategy that grants access only
// to the given audit scope IDs (but not to all scopes).
type partialScopeAuthorizationStrategy struct {
	scopeIds []string
}

func (s *partialScopeAuthorizationStrategy) CheckAccess(_ context.Context, _ string, _ orchestrator.RequestType, _ orchestrator.UserPermission_Permission, _ string, _ orchestrator.ObjectType) (bool, []string) {
	return false, s.scopeIds
}

func (s *partialScopeAuthorizationStrategy) AllowedTargetOfEvaluations(_ context.Context) (bool, []string) {
	return false, s.scopeIds
}

func (s *partialScopeAuthorizationStrategy) AllowedAuditScopes(_ context.Context) (bool, []string) {
	return false, s.scopeIds
}

func (s *partialScopeAuthorizationStrategy) AllowedUserPermission(_ context.Context) (bool, []string) {
	return false, nil
}

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
				}, svc.cfg)
				assert.NotEmpty(t, svc.scheduler)
				assert.NotEmpty(t, orchestratorconnect.NewOrchestratorClient(svc.cfg.OrchestratorClient, "http:://testhost:8080"), svc.orchestratorClient)
				assert.Equal(t, make(map[string]map[string]*orchestrator.Control), svc.catalogControls)
				return assert.NotNil(t, &svc.catalogsMutex)
			},
			wantErr: assert.NoError,
		},
		{
			name: "happy path with service OAuth2 config: the orchestrator client authenticates",
			args: args{
				opts: []service.Option[Service]{
					WithConfig(Config{
						OrchestratorClient:  http.DefaultClient,
						OrchestratorAddress: "http://testhost:8080",
						ServiceOAuth2Config: &clientcredentials.Config{TokenURL: "http://testhost:8080/token"},
					}),
					WithAuthorizationStrategy(&service.AuthorizationStrategyAllowAll{}),
				},
			},
			want: func(t *testing.T, got evaluationconnect.EvaluationHandler, msgAndArgs ...any) bool {
				svc, ok := got.(*Service)
				if !assert.True(t, ok) {
					return false
				}
				return assert.NotNil(t, svc.orchestratorClient) &&
					assert.NotNil(t, svc.authz)
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
				assert.Equal(t, DefaultConfig, svc.cfg, cmpopts.IgnoreFields(Config{}, "OrchestratorClient"))
				assert.True(t, svc.cfg.OrchestratorClient == DefaultConfig.OrchestratorClient)
				assert.NotEmpty(t, svc.scheduler)
				assert.NotEmpty(t, orchestratorconnect.NewOrchestratorClient(svc.cfg.OrchestratorClient, svc.cfg.OrchestratorAddress), svc.orchestratorClient)
				assert.Equal(t, make(map[string]map[string]*orchestrator.Control), svc.catalogControls)
				return assert.NotNil(t, &svc.catalogsMutex)
			},
			wantErr: assert.NoError,
		},
		{
			name: "WithAuthorizationStrategyPermissionStore wires up orchestrator permission store",
			args: args{
				opts: []service.Option[Service]{
					WithAuthorizationStrategyPermissionStore(),
				},
			},
			want: func(t *testing.T, got evaluationconnect.EvaluationHandler, msgAndArgs ...any) bool {
				svc, ok := got.(*Service)
				if !ok {
					t.Fatalf("expected *Service, got %T", got)
				}
				permStrat, ok := svc.authz.(*service.AuthorizationStrategyPermissionStore)
				if !assert.True(t, ok, "authz should be *AuthorizationStrategyPermissionStore") {
					return false
				}
				_, ok = permStrat.Permissions.(*service.OrchestratorPermissionStore)
				return assert.True(t, ok, "Permissions should be *OrchestratorPermissionStore")
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
				assert.Equal(t, 4, len(got.catalogControls[orchestratortest.MockCatalogId1]))
				return assert.Equal(t, orchestratortest.MockControl1, got.catalogControls[orchestratortest.MockCatalogId1][orchestratortest.MockControlId1])
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

func Test_preferAutomaticResults(t *testing.T) {
	now := time.Now()
	auto := &assessment.AssessmentResult{Id: "auto", MetricId: "m1"}
	manual := &assessment.AssessmentResult{Id: "manual", MetricId: "m1", Manual: true}
	manualOtherMetric := &assessment.AssessmentResult{Id: "manual2", MetricId: "m2", Manual: true}
	manualExpired := &assessment.AssessmentResult{Id: "expired", MetricId: "m2", Manual: true, ValidUntil: timestamppb.New(now.Add(-time.Hour))}
	manualValid := &assessment.AssessmentResult{Id: "valid", MetricId: "m3", Manual: true, ValidUntil: timestamppb.New(now.Add(time.Hour))}

	type args struct {
		results []*assessment.AssessmentResult
	}
	tests := []struct {
		name string
		args args
		want []*assessment.AssessmentResult
	}{
		{
			name: "no results",
			args: args{},
			want: nil,
		},
		{
			name: "only automatic results",
			args: args{results: []*assessment.AssessmentResult{auto}},
			want: []*assessment.AssessmentResult{auto},
		},
		{
			name: "manual result is dropped if an automatic result of the same metric exists",
			args: args{results: []*assessment.AssessmentResult{manual, auto}},
			want: []*assessment.AssessmentResult{auto},
		},
		{
			name: "manual result of another metric is kept",
			args: args{results: []*assessment.AssessmentResult{auto, manualOtherMetric}},
			want: []*assessment.AssessmentResult{auto, manualOtherMetric},
		},
		{
			name: "expired manual result is dropped, valid one is kept",
			args: args{results: []*assessment.AssessmentResult{manualExpired, manualValid}},
			want: []*assessment.AssessmentResult{manualValid},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := preferAutomaticResults(tt.args.results, now)

			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_getMetricsFromControl(t *testing.T) {
	type args struct {
		control *orchestrator.Control
	}
	tests := []struct {
		name string
		args args
		want assert.Want[[]*assessment.Metric]
	}{
		{
			name: "no metrics",
			args: args{control: &orchestrator.Control{Id: evaluationtest.MockControlId1}},
			want: assert.Empty[[]*assessment.Metric],
		},
		{
			name: "metrics of the control itself",
			args: args{control: &orchestrator.Control{Metrics: []*assessment.Metric{evaluationtest.MockMetric1}}},
			want: func(t *testing.T, got []*assessment.Metric, msgAndArgs ...any) bool {
				return assert.Equal(t, 1, len(got))
			},
		},
		{
			name: "metrics of the control and of its sub-controls",
			args: args{control: &orchestrator.Control{
				Metrics: []*assessment.Metric{evaluationtest.MockMetric1},
				Controls: []*orchestrator.Control{
					{Metrics: []*assessment.Metric{evaluationtest.MockMetric1}},
					{Metrics: []*assessment.Metric{evaluationtest.MockMetric1}},
				},
			}},
			want: func(t *testing.T, got []*assessment.Metric, msgAndArgs ...any) bool {
				return assert.Equal(t, 3, len(got))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getMetricsFromControl(tt.args.control)

			tt.want(t, got)
		})
	}
}

func TestService_StopEvaluation(t *testing.T) {
	type args struct {
		ctx context.Context
		req *connect.Request[evaluation.StopEvaluationRequest]
	}
	type fields struct {
		orchestratorClient orchestratorconnect.OrchestratorClient
		scheduler          *gocron.Scheduler
		authz              service.AuthorizationStrategy
	}
	// Authorization with the permission store that is backed by the orchestrator, as in a real deployment
	newPermissionStoreFields := func(permissions []*orchestrator.UserPermission, listError error) fields {
		client := newOrchestratorClient(t, WithUserPermissions(permissions), func(m *mockOrchestratorHandler) {
			m.listUserPermissionsError = listError
		})

		return fields{
			orchestratorClient: client,
			scheduler:          gocron.NewScheduler(time.UTC),
			authz:              &service.AuthorizationStrategyPermissionStore{Permissions: &service.OrchestratorPermissionStore{Client: client}},
		}
	}
	userCtx := auth.WithClaims(context.Background(), &auth.OAuthClaims{
		RegisteredClaims: jwt.RegisteredClaims{Subject: orchestratortest.MockUserId1, Issuer: orchestratortest.MockUserIssuer1},
	})
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
				ctx: context.Background(),
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
			name: "error: permission denied",
			args: args{
				ctx: context.Background(),
				req: connect.NewRequest(&evaluation.StopEvaluationRequest{
					AuditScopeId: evaluationtest.MockAuditScopeId1,
				}),
			},
			fields: fields{
				authz: &denyAuthorizationStrategy{},
			},
			want: assert.Nil[*connect.Response[evaluation.StopEvaluationResponse]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodePermissionDenied)
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
		{
			name: "permission store: user with a permission for the audit scope has access, but no evaluation is running",
			args: args{
				ctx: userCtx,
				req: connect.NewRequest(&evaluation.StopEvaluationRequest{AuditScopeId: evaluationtest.MockAuditScopeId1}),
			},
			fields: newPermissionStoreFields([]*orchestrator.UserPermission{orchestratortest.MockUserPermissionsAuditScopeAdmin}, nil),
			want:   assert.Nil[*connect.Response[evaluation.StopEvaluationResponse]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeFailedPrecondition)
			},
		},
		{
			name: "permission store: user without any permission is denied",
			args: args{
				ctx: userCtx,
				req: connect.NewRequest(&evaluation.StopEvaluationRequest{AuditScopeId: evaluationtest.MockAuditScopeId1}),
			},
			fields: newPermissionStoreFields(nil, nil),
			want:   assert.Nil[*connect.Response[evaluation.StopEvaluationResponse]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodePermissionDenied)
			},
		},
		{
			name: "permission store: only another user has a permission, the user is denied",
			args: args{
				ctx: userCtx,
				req: connect.NewRequest(&evaluation.StopEvaluationRequest{AuditScopeId: evaluationtest.MockAuditScopeId1}),
			},
			fields: newPermissionStoreFields([]*orchestrator.UserPermission{{
				UserId:     orchestratortest.GetConfirmateUserID(orchestratortest.MockUserIssuer1, orchestratortest.MockNonExistentId),
				ObjectId:   evaluationtest.MockAuditScopeId1,
				ObjectType: orchestrator.ObjectType_OBJECT_TYPE_AUDIT_SCOPE,
				Permission: orchestrator.UserPermission_PERMISSION_ADMIN,
			}}, nil),
			want: assert.Nil[*connect.Response[evaluation.StopEvaluationResponse]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodePermissionDenied)
			},
		},
		{
			name: "permission store: the permissions cannot be listed, the user is denied",
			args: args{
				ctx: userCtx,
				req: connect.NewRequest(&evaluation.StopEvaluationRequest{AuditScopeId: evaluationtest.MockAuditScopeId1}),
			},
			fields: newPermissionStoreFields([]*orchestrator.UserPermission{orchestratortest.MockUserPermissionsAuditScopeAdmin}, errors.New("boom")),
			want:   assert.Nil[*connect.Response[evaluation.StopEvaluationResponse]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodePermissionDenied)
			},
		},
		{
			name: "permission store: a request without a user is denied",
			args: args{
				ctx: context.Background(),
				req: connect.NewRequest(&evaluation.StopEvaluationRequest{AuditScopeId: evaluationtest.MockAuditScopeId1}),
			},
			fields: newPermissionStoreFields([]*orchestrator.UserPermission{orchestratortest.MockUserPermissionsAuditScopeAdmin}, nil),
			want:   assert.Nil[*connect.Response[evaluation.StopEvaluationResponse]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodePermissionDenied)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.args.ctx
			if ctx == nil {
				ctx = context.Background()
			}
			svc := &Service{
				orchestratorClient: tt.fields.orchestratorClient,
				scheduler:          tt.fields.scheduler,
				authz:              tt.fields.authz,
			}
			got, gotErr := svc.StopEvaluation(ctx, tt.args.req)

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
		// latest is the snapshot of the latest assessment results; if it is nil, the orchestrator is asked
		latest   *latestAssessments
		interval int
	}
	type fields struct {
		orchestratorClient orchestratorconnect.OrchestratorClient
		catalogControls    map[string]map[string]*orchestrator.Control
	}
	var (
		compliantResult = &assessment.AssessmentResult{
			Id:                   orchestratortest.MockResultId1,
			MetricId:             evaluationtest.MockMetricId1,
			ResourceId:           orchestratortest.MockResourceId1,
			Compliant:            true,
			TargetOfEvaluationId: evaluationtest.MockToeId1,
		}
		nonCompliantResult = &assessment.AssessmentResult{
			Id:                   orchestratortest.MockResultId2,
			MetricId:             evaluationtest.MockMetricId1,
			ResourceId:           orchestratortest.MockResourceId2,
			Compliant:            false,
			TargetOfEvaluationId: evaluationtest.MockToeId1,
		}
		// A control without sub-controls: its metrics are attached to the control itself
		leafWithMetrics = &orchestrator.Control{
			Id:      evaluationtest.MockControlId1,
			Metrics: []*assessment.Metric{evaluationtest.MockMetric1},
		}
		leafWithoutMetrics = &orchestrator.Control{
			Id: evaluationtest.MockControlId1,
		}
		// A control whose only sub-control is out of scope (the sub-control requires a higher assurance level than the
		// one of the audit scope) must not be evaluated by its own metrics
		parentWithoutRelevantSubcontrols = &orchestrator.Control{
			Id:      evaluationtest.MockControlId1,
			Metrics: []*assessment.Metric{evaluationtest.MockMetric1},
			Controls: []*orchestrator.Control{{
				Id:              evaluationtest.MockControl1SubcontrolId11,
				ParentControlId: new(evaluationtest.MockControlId1),
				Metrics:         []*assessment.Metric{evaluationtest.MockMetric1},
				AssuranceLevel:  new("high"),
			}},
		}
		basicAuditScope = &orchestrator.AuditScope{
			Id:                   evaluationtest.MockAuditScopeId1,
			TargetOfEvaluationId: evaluationtest.MockToeId1,
			CatalogId:            evaluationtest.MockCatalogId1,
			AssuranceLevel:       new("basic"),
		}
		leveledCatalog = &orchestrator.Catalog{
			Id:              evaluationtest.MockCatalogId1,
			AssuranceLevels: []string{"basic", "high"},
		}
	)
	// wantOneResult checks that exactly one evaluation result was stored, and applies the given checks to it
	wantOneResult := func(check func(t *testing.T, got *evaluation.EvaluationResult) bool) assert.Want[*Service] {
		return func(t *testing.T, got *Service, msgAndArgs ...any) bool {
			res, err := got.orchestratorClient.ListEvaluationResults(context.Background(), connect.NewRequest(&orchestrator.ListEvaluationResultsRequest{}))
			assert.NoError(t, err)

			return assert.Equal(t, 1, len(res.Msg.Results)) && check(t, res.Msg.Results[0])
		}
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
						evaluationtest.MockControl1.Id: evaluationtest.MockControl1,
						// evaluationtest.MockSubcontrol11.Id: evaluationtest.MockSubcontrol11,
						// evaluationtest.MockSubcontrol12.Id: evaluationtest.MockSubcontrol12,
						evaluationtest.MockControl2.Id: evaluationtest.MockControl2,
						// evaluationtest.MockSubcontrol21.Id: evaluationtest.MockSubcontrol21,
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
						evaluationtest.MockControl1.Id:     evaluationtest.MockControl1,
						evaluationtest.MockSubcontrol11.Id: evaluationtest.MockSubcontrol11,
						evaluationtest.MockSubcontrol12.Id: evaluationtest.MockSubcontrol12,
						evaluationtest.MockControl2.Id:     evaluationtest.MockControl2,
						evaluationtest.MockSubcontrol21.Id: evaluationtest.MockSubcontrol21,
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
						evaluationtest.MockControl1.Id:     evaluationtest.MockControl1,
						evaluationtest.MockSubcontrol11.Id: evaluationtest.MockSubcontrol11,
						evaluationtest.MockSubcontrol12.Id: evaluationtest.MockSubcontrol12,
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
		{
			name: "leaf control (no sub-controls): evaluated by its own metrics, compliant",
			args: args{
				ctx:        context.Background(),
				auditScope: evaluationtest.MockAuditScope1,
				catalog:    evaluationtest.MockCatalog1,
				control:    leafWithMetrics,
				latest:     newLatestAssessments([]*assessment.AssessmentResult{compliantResult}),
			},
			fields: fields{
				orchestratorClient: newOrchestratorClient(t),
			},
			wantSvc: wantOneResult(func(t *testing.T, got *evaluation.EvaluationResult) bool {
				return assert.Equal(t, evaluation.EvaluationStatus_EVALUATION_STATUS_COMPLIANT, got.GetStatus()) &&
					assert.Equal(t, []string{compliantResult.Id}, got.GetAssessmentResultIds()) &&
					assert.Nil(t, got.ParentControlId)
			}),
			wantErr: assert.NoError,
		},
		{
			name: "leaf control: one non-compliant assessment result makes the control not compliant",
			args: args{
				ctx:        context.Background(),
				auditScope: evaluationtest.MockAuditScope1,
				catalog:    evaluationtest.MockCatalog1,
				control:    leafWithMetrics,
				latest:     newLatestAssessments([]*assessment.AssessmentResult{compliantResult, nonCompliantResult}),
			},
			fields: fields{
				orchestratorClient: newOrchestratorClient(t),
			},
			wantSvc: wantOneResult(func(t *testing.T, got *evaluation.EvaluationResult) bool {
				return assert.Equal(t, evaluation.EvaluationStatus_EVALUATION_STATUS_NOT_COMPLIANT, got.GetStatus()) &&
					assert.Equal(t, []string{compliantResult.Id, nonCompliantResult.Id}, got.GetAssessmentResultIds())
			}),
			wantErr: assert.NoError,
		},
		{
			name: "leaf control without the snapshot of the assessment results: the orchestrator is asked",
			args: args{
				ctx:        context.Background(),
				auditScope: evaluationtest.MockAuditScope1,
				catalog:    evaluationtest.MockCatalog1,
				control:    leafWithMetrics,
			},
			fields: fields{
				orchestratorClient: newOrchestratorClient(t, WithAssessmentResults([]*assessment.AssessmentResult{compliantResult, nonCompliantResult})),
			},
			wantSvc: wantOneResult(func(t *testing.T, got *evaluation.EvaluationResult) bool {
				return assert.Equal(t, evaluation.EvaluationStatus_EVALUATION_STATUS_NOT_COMPLIANT, got.GetStatus()) &&
					assert.Equal(t, []string{compliantResult.Id, nonCompliantResult.Id}, got.GetAssessmentResultIds())
			}),
			wantErr: assert.NoError,
		},
		{
			name: "leaf control without assessment results: pending",
			args: args{
				ctx:        context.Background(),
				auditScope: evaluationtest.MockAuditScope1,
				catalog:    evaluationtest.MockCatalog1,
				control:    leafWithMetrics,
				latest:     newLatestAssessments(nil),
			},
			fields: fields{
				orchestratorClient: newOrchestratorClient(t),
			},
			wantSvc: wantOneResult(func(t *testing.T, got *evaluation.EvaluationResult) bool {
				return assert.Equal(t, evaluation.EvaluationStatus_EVALUATION_STATUS_PENDING, got.GetStatus())
			}),
			wantErr: assert.NoError,
		},
		{
			name: "leaf control without metrics: pending",
			args: args{
				ctx:        context.Background(),
				auditScope: evaluationtest.MockAuditScope1,
				catalog:    evaluationtest.MockCatalog1,
				control:    leafWithoutMetrics,
				latest:     newLatestAssessments([]*assessment.AssessmentResult{compliantResult}),
			},
			fields: fields{
				orchestratorClient: newOrchestratorClient(t),
			},
			wantSvc: wantOneResult(func(t *testing.T, got *evaluation.EvaluationResult) bool {
				return assert.Equal(t, evaluation.EvaluationStatus_EVALUATION_STATUS_PENDING, got.GetStatus())
			}),
			wantErr: assert.NoError,
		},
		{
			name: "all sub-controls out of scope: the metrics of the control itself are not used",
			args: args{
				ctx:        context.Background(),
				auditScope: basicAuditScope,
				catalog:    leveledCatalog,
				control:    parentWithoutRelevantSubcontrols,
				latest:     newLatestAssessments([]*assessment.AssessmentResult{compliantResult}),
			},
			fields: fields{
				orchestratorClient: newOrchestratorClient(t),
			},
			wantSvc: wantOneResult(func(t *testing.T, got *evaluation.EvaluationResult) bool {
				return assert.Equal(t, evaluation.EvaluationStatus_EVALUATION_STATUS_PENDING, got.GetStatus())
			}),
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := Service{
				orchestratorClient: tt.fields.orchestratorClient,
				catalogControls:    tt.fields.catalogControls,
			}

			gotErr := svc.evaluateControl(tt.args.ctx, tt.args.auditScope, tt.args.catalog, tt.args.control, tt.args.manual, tt.args.latest)

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
	// automatedResults returns the stored evaluation results of the given control that were not created manually
	automatedResults := func(t *testing.T, got *Service, controlId string) (results []*evaluation.EvaluationResult) {
		res, err := got.orchestratorClient.ListEvaluationResults(context.Background(), connect.NewRequest(&orchestrator.ListEvaluationResultsRequest{}))
		assert.NoError(t, err)

		for _, r := range res.Msg.Results {
			if r.GetControlId() == controlId && r.GetStatus() != evaluation.EvaluationStatus_EVALUATION_STATUS_COMPLIANT_MANUALLY {
				results = append(results, r)
			}
		}

		return results
	}
	catalogControls := map[string]map[string]*orchestrator.Control{
		evaluationtest.MockCatalog1.Id: {
			evaluationtest.MockControl1.Id:     evaluationtest.MockControl1,
			evaluationtest.MockSubcontrol11.Id: evaluationtest.MockSubcontrol11,
			evaluationtest.MockSubcontrol12.Id: evaluationtest.MockSubcontrol12,
		},
	}
	// The manual result of sub-control 1.1 (audit scope 1) of the mocks, but created in audit scope 2
	manualSubcontrolOtherScope := &evaluation.EvaluationResult{
		Id:                   evaluationtest.MockManualEvaluationResult2.Id,
		TargetOfEvaluationId: evaluationtest.MockToeId1,
		AuditScopeId:         evaluationtest.MockAuditScopeId2,
		ControlId:            evaluationtest.MockControl1SubcontrolId11,
		ControlCatalogId:     evaluationtest.MockCatalogId1,
		ParentControlId:      new(evaluationtest.MockControlId1),
		Status:               evaluation.EvaluationStatus_EVALUATION_STATUS_COMPLIANT_MANUALLY,
		ValidUntil:           evaluationtest.MockManualEvaluationResult2.ValidUntil,
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
						evaluationtest.MockControl1.Id: evaluationtest.MockControl1,
						// evaluationtest.MockSubcontrol11.Id: evaluationtest.MockSubcontrol11,
						// evaluationtest.MockSubcontrol12.Id: evaluationtest.MockSubcontrol12,
						evaluationtest.MockControl2.Id: evaluationtest.MockControl2,
						// evaluationtest.MockSubcontrol21.Id: evaluationtest.MockSubcontrol21,
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
						evaluationtest.MockControl1.Id:     evaluationtest.MockControl1,
						evaluationtest.MockSubcontrol11.Id: evaluationtest.MockSubcontrol11,
						evaluationtest.MockSubcontrol12.Id: evaluationtest.MockSubcontrol12,
						evaluationtest.MockControl2.Id:     evaluationtest.MockControl2,
						evaluationtest.MockSubcontrol21.Id: evaluationtest.MockSubcontrol21,
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
				assert.Contains(t, controlIds, evaluationtest.MockControl2SubcontrolID21)

				return true
			},
			wantErr: assert.NoError,
		},
		{
			// Several audit scopes can share the same target of evaluation and catalog
			name: "manual result of a control in another audit scope of the same target of evaluation: the control is evaluated",
			args: args{
				ctx:        context.Background(),
				auditScope: evaluationtest.MockAuditScope1,
				catalog:    evaluationtest.MockCatalog1,
				interval:   5,
			},
			fields: fields{
				orchestratorClient: newOrchestratorClient(t,
					WithEvaluationResults([]*evaluation.EvaluationResult{evaluationtest.MockManualEvaluationResult3}), // audit scope 2
				),
				catalogControls: catalogControls,
			},
			want: func(t *testing.T, got *Service, msgAndArgs ...any) bool {
				results := automatedResults(t, got, evaluationtest.MockControlId1)

				return assert.Equal(t, 1, len(results)) &&
					assert.Equal(t, evaluationtest.MockAuditScopeId1, results[0].GetAuditScopeId())
			},
			wantErr: assert.NoError,
		},
		{
			name: "manual result of a sub-control in the same audit scope: the sub-control is not evaluated automatically",
			args: args{
				ctx:        context.Background(),
				auditScope: evaluationtest.MockAuditScope1,
				catalog:    evaluationtest.MockCatalog1,
				interval:   5,
			},
			fields: fields{
				orchestratorClient: newOrchestratorClient(t,
					WithEvaluationResults([]*evaluation.EvaluationResult{evaluationtest.MockManualEvaluationResult2}), // audit scope 1
				),
				catalogControls: catalogControls,
			},
			want: func(t *testing.T, got *Service, msgAndArgs ...any) bool {
				return assert.Equal(t, 0, len(automatedResults(t, got, evaluationtest.MockControl1SubcontrolId11)))
			},
			wantErr: assert.NoError,
		},
		{
			name: "manual result of a sub-control in another audit scope: the sub-control is evaluated",
			args: args{
				ctx:        context.Background(),
				auditScope: evaluationtest.MockAuditScope1,
				catalog:    evaluationtest.MockCatalog1,
				interval:   5,
			},
			fields: fields{
				orchestratorClient: newOrchestratorClient(t,
					WithEvaluationResults([]*evaluation.EvaluationResult{manualSubcontrolOtherScope}),
				),
				catalogControls: catalogControls,
			},
			want: func(t *testing.T, got *Service, msgAndArgs ...any) bool {
				results := automatedResults(t, got, evaluationtest.MockControl1SubcontrolId11)

				return assert.Equal(t, 1, len(results)) &&
					assert.Equal(t, evaluationtest.MockAuditScopeId1, results[0].GetAuditScopeId())
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

func TestService_evaluateCatalog_UpdatesCertificateLifecycle(t *testing.T) {
	client, handler := newOrchestratorClientWithHandler(t,
		WithAssessmentResults([]*assessment.AssessmentResult{
			{
				Id:                   evaluationtest.MockAssessmentResultId1,
				MetricId:             evaluationtest.MockMetricId1,
				Compliant:            true,
				ResourceId:           "resource-1",
				TargetOfEvaluationId: evaluationtest.MockToeId1,
			},
		}),
	)

	svc := Service{
		orchestratorClient: client,
		catalogControls: map[string]map[string]*orchestrator.Control{
			evaluationtest.MockCatalog1.Id: {
				evaluationtest.MockControl1.Id:     evaluationtest.MockControl1,
				evaluationtest.MockSubcontrol11.Id: evaluationtest.MockSubcontrol11,
				evaluationtest.MockSubcontrol12.Id: evaluationtest.MockSubcontrol12,
			},
		},
	}

	err := svc.evaluateCatalog(context.Background(), evaluationtest.MockAuditScope1, evaluationtest.MockCatalog1, 5)
	assert.NoError(t, err)

	assert.Equal(t, []string{evaluationtest.MockAuditScope1.Id}, handler.lifecycleUpdates)
}

// TestService_evaluateCatalog_SkipsWhenAlreadyRunning proves the overlap guard: a manual trigger
// (via TriggerEvaluation/RunByTag) racing a scheduled tick for the same audit scope must not run a
// second, concurrent evaluation.
func TestService_evaluateCatalog_SkipsWhenAlreadyRunning(t *testing.T) {
	svc := Service{
		evaluating: map[string]struct{}{
			evaluationtest.MockAuditScopeId1: {},
		},
	}

	// orchestratorClient and catalogControls are intentionally left unset: if the overlap guard
	// did not fire first, evaluateCatalog would try to use them and fail.
	gotErr := svc.evaluateCatalog(context.Background(), evaluationtest.MockAuditScope1, evaluationtest.MockCatalog1, 5)
	assert.NoError(t, gotErr)

	// The in-progress marker set up by the test must still be present: evaluateCatalog's own
	// endEvaluation (deferred after a successful beginEvaluation) must not have run, proving it
	// returned before ever acquiring the guard.
	_, stillRunning := svc.evaluating[evaluationtest.MockAuditScopeId1]
	assert.True(t, stillRunning)
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
					Id:       orchestratortest.MockControlId2,
					Name:     "Mock Control 2",
					Metrics:  nil,
					Controls: nil,
				}

				return fields{
					orchestratorClient: newOrchestratorClientForTest(testSrv),
					catalogControls: map[string]map[string]*orchestrator.Control{
						orchestratortest.MockCatalogId2: {
							ctrl.GetId(): ctrl,
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
					Id:       orchestratortest.MockControlId2,
					Name:     "Mock Control 2",
					Metrics:  nil,
					Controls: nil,
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
					Id:   orchestratortest.MockControl2SubControlId1,
					Name: "Mock Subcontrol 1",
					Metrics: []*assessment.Metric{
						{Id: orchestratortest.MockMetricId1},
					},
				}

				return fields{
					orchestratorClient: newOrchestratorClientForTest(testSrv),
					catalogControls: map[string]map[string]*orchestrator.Control{
						orchestratortest.MockCatalogId1: {
							ctrl.GetId(): ctrl,
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
					Id:   orchestratortest.MockControl2SubControlId1,
					Name: "Mock Subcontrol 1",
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
					ControlId:            orchestratortest.MockControl2SubControlId1,
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
					ControlId:            orchestratortest.MockControl2SubControlId1,
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
								MetricId:             evaluationtest.MockMetricId1,
								Compliant:            true,
								ResourceId:           "resource-1",
								TargetOfEvaluationId: evaluationtest.MockToeId1,
							},
							{
								Id:                   "assessment-result-2",
								MetricId:             evaluationtest.MockMetricId1,
								Compliant:            true,
								ResourceId:           "resource-2",
								TargetOfEvaluationId: evaluationtest.MockToeId1,
							},
						}),
					),
					catalogControls: map[string]map[string]*orchestrator.Control{
						evaluationtest.MockCatalogId1: {
							// orchestratortest.MockSubControl1.GetId(): orchestratortest.MockSubControl1,
							evaluationtest.MockControl1.GetId(): evaluationtest.MockControl1,
						},
					},
				}
			}(),
			args: args{
				ctx: context.Background(),
				auditScope: &orchestrator.AuditScope{
					Id:                   evaluationtest.MockAuditScopeId1,
					TargetOfEvaluationId: evaluationtest.MockToeId1,
					CatalogId:            evaluationtest.MockCatalogId1,
				},
				control: evaluationtest.MockSubcontrol11,
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
					ControlId:            evaluationtest.MockControl1SubcontrolId11,
					ControlCatalogId:     evaluationtest.MockCatalogId1,
					ParentControlId:      new(evaluationtest.MockControlId1),
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
					ControlId:            evaluationtest.MockControl1SubcontrolId11,
					ControlCatalogId:     evaluationtest.MockCatalogId1,
					ParentControlId:      new(evaluationtest.MockControlId1),
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
								MetricId:             evaluationtest.MockMetricId1,
								Compliant:            true,
								ResourceId:           "resource-1",
								TargetOfEvaluationId: evaluationtest.MockToeId1,
							},
							{
								Id:                   "assessment-result-2",
								MetricId:             evaluationtest.MockMetricId1,
								Compliant:            false,
								ResourceId:           "resource-2",
								TargetOfEvaluationId: evaluationtest.MockToeId1,
							},
						}),
					),
					catalogControls: map[string]map[string]*orchestrator.Control{
						evaluationtest.MockCatalogId1: {
							evaluationtest.MockControl1.GetId(): evaluationtest.MockControl1,
						},
					},
				}
			}(),
			args: args{
				ctx: context.Background(),
				auditScope: &orchestrator.AuditScope{
					Id:                   evaluationtest.MockAuditScopeId1,
					TargetOfEvaluationId: evaluationtest.MockToeId1,
					CatalogId:            evaluationtest.MockCatalogId1,
				},
				control: evaluationtest.MockSubcontrol11,
			},
			want: func(t *testing.T, got *evaluation.EvaluationResult, _ ...any) bool {
				assert.NotNil(t, got)
				assert.NotEmpty(t, got.GetId())
				assert.NotEmpty(t, got.GetTimestamp())
				assert.Equal(t, 2, len(got.AssessmentResultIds))

				want := &evaluation.EvaluationResult{
					TargetOfEvaluationId: evaluationtest.MockToeId1,
					AuditScopeId:         evaluationtest.MockAuditScopeId1,
					ControlId:            evaluationtest.MockControl1SubcontrolId11,
					ControlCatalogId:     evaluationtest.MockCatalogId1,
					ParentControlId:      evaluationtest.MockSubcontrol11.ParentControlId,
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
					TargetOfEvaluationId: evaluationtest.MockToeId1,
					AuditScopeId:         evaluationtest.MockAuditScopeId1,
					ControlId:            evaluationtest.MockControl1SubcontrolId11,
					ControlCatalogId:     evaluationtest.MockCatalogId1,
					ParentControlId:      evaluationtest.MockSubcontrol11.ParentControlId,
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

			got, gotErr := svc.evaluateSubcontrol(tt.args.ctx, tt.args.auditScope, tt.args.control, nil)

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
		authz              service.AuthorizationStrategy
	}

	// certHandler is captured by the "certificate is not available" test case below, so its
	// wantSvc check can assert that CreateCertificate was actually invoked.
	var certHandler *mockOrchestratorHandler

	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[*connect.Response[evaluation.StartEvaluationResponse]]
		wantSvc assert.Want[*Service]
		wantErr assert.WantErr
	}{
		{
			name: "err: permission denied",
			args: args{
				ctx: context.Background(),
				req: connect.NewRequest(&evaluation.StartEvaluationRequest{
					AuditScopeId: evaluationtest.MockAuditScopeId1,
				}),
			},
			fields: fields{
				authz:     &denyAuthorizationStrategy{},
				scheduler: gocron.NewScheduler(time.Local),
			},
			want:    assert.Nil[*connect.Response[evaluation.StartEvaluationResponse]],
			wantSvc: assert.NotNil[*Service],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodePermissionDenied)
			},
		},
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
						evaluationtest.MockControl1.Id:     evaluationtest.MockControl1,
						evaluationtest.MockSubcontrol11.Id: evaluationtest.MockSubcontrol11,
						evaluationtest.MockSubcontrol12.Id: evaluationtest.MockSubcontrol12,
						evaluationtest.MockControl2.Id:     evaluationtest.MockControl2,
						evaluationtest.MockSubcontrol21.Id: evaluationtest.MockSubcontrol21,
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
			name: "err: getting catalog error",
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
						evaluationtest.MockControl1.Id:     evaluationtest.MockControl1,
						evaluationtest.MockSubcontrol11.Id: evaluationtest.MockSubcontrol11,
						evaluationtest.MockSubcontrol12.Id: evaluationtest.MockSubcontrol12,
						evaluationtest.MockControl2.Id:     evaluationtest.MockControl2,
						evaluationtest.MockSubcontrol21.Id: evaluationtest.MockSubcontrol21,
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
						evaluationtest.MockControl1.Id:     evaluationtest.MockControl1,
						evaluationtest.MockSubcontrol11.Id: evaluationtest.MockSubcontrol11,
						evaluationtest.MockSubcontrol12.Id: evaluationtest.MockSubcontrol12,
						evaluationtest.MockControl2.Id:     evaluationtest.MockControl2,
						evaluationtest.MockSubcontrol21.Id: evaluationtest.MockSubcontrol21,
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
						[]*orchestrator.Control{evaluationtest.MockControl1, evaluationtest.MockControl2},
					),
					WithCatalog(evaluationtest.MockCatalog1),
				),
				scheduler: gocron.NewScheduler(time.Local),
				catalogControls: map[string]map[string]*orchestrator.Control{
					evaluationtest.MockCatalog1.Id: {
						evaluationtest.MockControl1.Id: evaluationtest.MockControl1,
						evaluationtest.MockControl2.Id: evaluationtest.MockControl2,
					},
				},
			},
			want: func(t *testing.T, got *connect.Response[evaluation.StartEvaluationResponse], _ ...any) bool {
				assert.NotNil(t, got)
				return assert.True(t, got.Msg.GetSuccessful())
			},
			wantSvc: func(t *testing.T, got *Service, msgAndArgs ...any) bool {
				assert.Equal(t, 2, len(got.catalogControls[evaluationtest.MockCatalogId1]))
				return assert.Equal(t, 1, len(got.scheduler.Jobs()))
			},
			wantErr: assert.NoError,
		},
		{
			name: "happy path: certificate is not available for the audit scope, create a new one",
			args: args{
				ctx: context.Background(),
				req: connect.NewRequest(&evaluation.StartEvaluationRequest{
					AuditScopeId: evaluationtest.MockAuditScopeId2,
				}),
			},
			fields: fields{
				orchestratorClient: func() orchestratorconnect.OrchestratorClient {
					var client orchestratorconnect.OrchestratorClient
					client, certHandler = newOrchestratorClientWithHandler(t,
						WithAuditScope(evaluationtest.MockAuditScope2),
						WithControls(
							[]*orchestrator.Control{evaluationtest.MockControl1, evaluationtest.MockControl2},
						),
						WithCatalog(evaluationtest.MockCatalog1),
					)
					return client
				}(),
				scheduler: gocron.NewScheduler(time.Local),
				catalogControls: map[string]map[string]*orchestrator.Control{
					evaluationtest.MockCatalog1.Id: {
						evaluationtest.MockControl1.Id: evaluationtest.MockControl1,
						evaluationtest.MockControl2.Id: evaluationtest.MockControl2,
					},
				},
			},
			want: func(t *testing.T, got *connect.Response[evaluation.StartEvaluationResponse], _ ...any) bool {
				assert.NotNil(t, got)
				return assert.True(t, got.Msg.GetSuccessful())
			},
			wantErr: assert.NoError,
			wantSvc: func(t *testing.T, _ *Service, msgAndArgs ...any) bool {
				return assert.Equal(t, []string{evaluationtest.MockAuditScopeId2}, certHandler.createCertificateCalls)
			},
		},
		{
			name: "happy path: certificate already exists, evaluation still starts",
			args: args{
				ctx: context.Background(),
				req: connect.NewRequest(&evaluation.StartEvaluationRequest{
					AuditScopeId: evaluationtest.MockAuditScopeId2,
				}),
			},
			fields: fields{
				orchestratorClient: newOrchestratorClient(t,
					WithAuditScope(evaluationtest.MockAuditScope2),
					WithControls(
						[]*orchestrator.Control{evaluationtest.MockControl1, evaluationtest.MockControl2},
					),
					WithCatalog(evaluationtest.MockCatalog1),
					WithCreateCertificateError(connect.NewError(connect.CodeAlreadyExists, errors.New("certificate already exists"))),
				),
				scheduler: gocron.NewScheduler(time.Local),
				catalogControls: map[string]map[string]*orchestrator.Control{
					evaluationtest.MockCatalog1.Id: {
						evaluationtest.MockControl1.Id: evaluationtest.MockControl1,
						evaluationtest.MockControl2.Id: evaluationtest.MockControl2,
					},
				},
			},
			want: func(t *testing.T, got *connect.Response[evaluation.StartEvaluationResponse], _ ...any) bool {
				assert.NotNil(t, got)
				return assert.True(t, got.Msg.GetSuccessful())
			},
			wantErr: assert.NoError,
			wantSvc: assert.NotNil[*Service],
		},
		{
			name: "err: CreateCertificate fails",
			args: args{
				ctx: context.Background(),
				req: connect.NewRequest(&evaluation.StartEvaluationRequest{
					AuditScopeId: evaluationtest.MockAuditScopeId2,
				}),
			},
			fields: fields{
				orchestratorClient: newOrchestratorClient(t,
					WithAuditScope(evaluationtest.MockAuditScope2),
					WithControls(
						[]*orchestrator.Control{evaluationtest.MockControl1, evaluationtest.MockControl2},
					),
					WithCatalog(evaluationtest.MockCatalog1),
					WithCreateCertificateError(connect.NewError(connect.CodeInternal, errors.New("database error"))),
				),
				scheduler: gocron.NewScheduler(time.Local),
				catalogControls: map[string]map[string]*orchestrator.Control{
					evaluationtest.MockCatalog1.Id: {
						evaluationtest.MockControl1.Id: evaluationtest.MockControl1,
						evaluationtest.MockControl2.Id: evaluationtest.MockControl2,
					},
				},
			},
			want:    assert.Nil[*connect.Response[evaluation.StartEvaluationResponse]],
			wantSvc: assert.NotNil[*Service],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInternal) &&
					assert.ErrorContains(t, err, "database error")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			svc := Service{
				orchestratorClient: tt.fields.orchestratorClient,
				catalogControls:    tt.fields.catalogControls,
				scheduler:          tt.fields.scheduler,
				authz:              tt.fields.authz,
			}
			got, gotErr := svc.StartEvaluation(tt.args.ctx, tt.args.req)

			tt.want(t, got)
			tt.wantErr(t, gotErr)
			tt.wantSvc(t, &svc)
		})
	}
}

func TestService_ListEvaluationJobs(t *testing.T) {
	type fields struct {
		orchestratorClient orchestratorconnect.OrchestratorClient
		scheduler          *gocron.Scheduler
		authz              service.AuthorizationStrategy
	}
	tests := []struct {
		name    string
		fields  fields
		req     *connect.Request[evaluation.ListEvaluationJobsRequest]
		want    assert.Want[*connect.Response[evaluation.ListEvaluationJobsResponse]]
		wantErr assert.WantErr
	}{
		{
			name: "err: validation error - invalid audit scope id",
			fields: fields{
				orchestratorClient: nil,
				scheduler:          gocron.NewScheduler(time.Local),
			},
			req: connect.NewRequest(&evaluation.ListEvaluationJobsRequest{
				Filter: &evaluation.ListEvaluationJobsRequest_Filter{
					AuditScopeId: new("invalid-uuid"),
				},
			}),
			want: assert.Nil[*connect.Response[evaluation.ListEvaluationJobsResponse]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument) &&
					assert.ErrorContains(t, err, "filter.audit_scope_id: must be a valid UUID")
			},
		},
		{
			name: "authorization failure returns empty list",
			fields: fields{
				orchestratorClient: nil,
				scheduler:          gocron.NewScheduler(time.Local),
				authz:              &denyAuthorizationStrategy{},
			},
			req: connect.NewRequest(&evaluation.ListEvaluationJobsRequest{}),
			want: func(t *testing.T, got *connect.Response[evaluation.ListEvaluationJobsResponse], _ ...any) bool {
				return assert.NotNil(t, got) && assert.Equal(t, 0, len(got.Msg.EvaluationJobs))
			},
			wantErr: assert.NoError,
		},
		{
			name: "happy path: filter by audit scope id",
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
						s := gocron.NewScheduler(time.Local)
						_, err := s.Every(1).Day().Tag("00000000-0000-0000-0000-000000000001").Do(func() { fmt.Println("Job 1") })
						assert.NoError(t, err)
						_, err = s.Every(2).Day().Tag("00000000-0000-0000-0000-000000000002").Do(func() { fmt.Println("Job 2") })
						assert.NoError(t, err)
						return s
					}(),
					authz: &service.AuthorizationStrategyAllowAll{},
				}
			}(),
			req: connect.NewRequest(&evaluation.ListEvaluationJobsRequest{
				Filter: &evaluation.ListEvaluationJobsRequest_Filter{
					AuditScopeId: new("00000000-0000-0000-0000-000000000002"),
				},
			}),
			want: func(t *testing.T, got *connect.Response[evaluation.ListEvaluationJobsResponse], _ ...any) bool {
				want1 := &evaluation.EvaluationJob{
					AuditScopeId: "00000000-0000-0000-0000-000000000002",
					Interval:     2,
				}

				assert.NotNil(t, got)
				assert.Equal(t, 1, len(got.Msg.GetEvaluationJobs()))
				assert.NotEmpty(t, got.Msg.GetEvaluationJobs()[0].GetLastRun())
				assert.Equal(t, 0, got.Msg.GetEvaluationJobs()[0].GetRunCount())
				assert.NotEmpty(t, got.Msg.GetEvaluationJobs()[0].GetStartedAt())
				return assert.Equal(t, want1, got.Msg.GetEvaluationJobs()[0], protocmp.IgnoreFields(&evaluation.EvaluationJob{}, "last_run", "run_count", "started_at"))
			},
			wantErr: assert.NoError,
		},
		{
			name: "happy path: only jobs for allowed audit scopes are listed",
			fields: fields{
				scheduler: func() *gocron.Scheduler {
					s := gocron.NewScheduler(time.Local)
					_, err := s.Every(1).Day().Tag("00000000-0000-0000-0000-000000000001").Do(func() { fmt.Println("Job 1") })
					assert.NoError(t, err)
					_, err = s.Every(2).Day().Tag("00000000-0000-0000-0000-000000000002").Do(func() { fmt.Println("Job 2") })
					assert.NoError(t, err)
					return s
				}(),
				authz: &partialScopeAuthorizationStrategy{
					scopeIds: []string{"00000000-0000-0000-0000-000000000002"},
				},
			},
			req: connect.NewRequest(&evaluation.ListEvaluationJobsRequest{}),
			want: func(t *testing.T, got *connect.Response[evaluation.ListEvaluationJobsResponse], _ ...any) bool {
				assert.NotNil(t, got)
				assert.Equal(t, 1, len(got.Msg.GetEvaluationJobs()))
				return assert.Equal(t, "00000000-0000-0000-0000-000000000002", got.Msg.GetEvaluationJobs()[0].GetAuditScopeId())
			},
			wantErr: assert.NoError,
		},
		{
			name: "happy path",
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
						s := gocron.NewScheduler(time.Local)
						_, err := s.Every(1).Day().Tag("00000000-0000-0000-0000-000000000001").Do(func() { fmt.Println("Job 1") })
						assert.NoError(t, err)
						_, err = s.Every(2).Day().Tag("00000000-0000-0000-0000-000000000002").Do(func() { fmt.Println("Job 2") })
						assert.NoError(t, err)
						return s
					}(),
				}
			}(),
			req: connect.NewRequest(&evaluation.ListEvaluationJobsRequest{}),
			want: func(t *testing.T, got *connect.Response[evaluation.ListEvaluationJobsResponse], _ ...any) bool {
				want1 := &evaluation.EvaluationJob{
					AuditScopeId: "00000000-0000-0000-0000-000000000001",
					Interval:     1,
				}

				// Sort the jobs by AuditScopeId to ensure consistent ordering for the test
				sort.SliceStable(got.Msg.GetEvaluationJobs(), func(i, j int) bool {
					return got.Msg.GetEvaluationJobs()[i].GetAuditScopeId() < got.Msg.GetEvaluationJobs()[j].GetAuditScopeId()
				})

				assert.NotNil(t, got)
				assert.Equal(t, 2, len(got.Msg.GetEvaluationJobs()))
				assert.NotEmpty(t, got.Msg.GetEvaluationJobs()[0].GetLastRun())
				assert.Equal(t, 0, got.Msg.GetEvaluationJobs()[0].GetRunCount())
				assert.NotEmpty(t, got.Msg.GetEvaluationJobs()[0].GetStartedAt())
				return assert.Equal(t, want1, got.Msg.GetEvaluationJobs()[0], protocmp.IgnoreFields(&evaluation.EvaluationJob{}, "last_run", "run_count", "started_at"))
			},
			wantErr: assert.NoError,
		},
		{
			name: "happy path: only jobs for allowed audit scopes are listed",
			fields: fields{
				scheduler: func() *gocron.Scheduler {
					s := gocron.NewScheduler(time.Local)
					_, err := s.Every(1).Day().Tag("00000000-0000-0000-0000-000000000001").Do(func() { fmt.Println("Job 1") })
					assert.NoError(t, err)
					_, err = s.Every(2).Day().Tag("00000000-0000-0000-0000-000000000002").Do(func() { fmt.Println("Job 2") })
					assert.NoError(t, err)
					return s
				}(),
				authz: &partialScopeAuthorizationStrategy{
					scopeIds: []string{"00000000-0000-0000-0000-000000000002"},
				},
			},
			req: connect.NewRequest(&evaluation.ListEvaluationJobsRequest{}),
			want: func(t *testing.T, got *connect.Response[evaluation.ListEvaluationJobsResponse], _ ...any) bool {
				assert.NotNil(t, got)
				assert.Equal(t, 1, len(got.Msg.GetEvaluationJobs()))
				return assert.Equal(t, "00000000-0000-0000-0000-000000000002", got.Msg.GetEvaluationJobs()[0].GetAuditScopeId())
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := Service{
				orchestratorClient: tt.fields.orchestratorClient,
				scheduler:          tt.fields.scheduler,
				authz:              tt.fields.authz,
			}
			got, gotErr := svc.ListEvaluationJobs(context.Background(), tt.req)

			tt.want(t, got)
			tt.wantErr(t, gotErr)
		})
	}
}

func TestService_TriggerEvaluation(t *testing.T) {
	type fields struct {
		orchestratorClient orchestratorconnect.OrchestratorClient
		catalogControls    map[string]map[string]*orchestrator.Control
		scheduler          *gocron.Scheduler
		authz              service.AuthorizationStrategy
	}
	tests := []struct {
		name    string
		fields  fields
		req     *connect.Request[evaluation.TriggerEvaluationRequest]
		want    assert.Want[*connect.Response[evaluation.TriggerEvaluationResponse]]
		wantErr assert.WantErr
	}{
		{
			name: "err: invalid request - invalid audit scope id",
			fields: fields{
				scheduler: gocron.NewScheduler(time.Local),
			},
			req: connect.NewRequest(&evaluation.TriggerEvaluationRequest{
				AuditScopeId: "invalid-uuid",
			}),
			want: assert.Nil[*connect.Response[evaluation.TriggerEvaluationResponse]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument) &&
					assert.ErrorContains(t, err, "audit_scope_id: must be a valid UUID")
			},
		},
		{
			name: "err: permission denied",
			fields: fields{
				scheduler: gocron.NewScheduler(time.Local),
				authz:     &denyAuthorizationStrategy{},
			},
			req: connect.NewRequest(&evaluation.TriggerEvaluationRequest{
				AuditScopeId: evaluationtest.MockAuditScopeId1,
			}),
			want: assert.Nil[*connect.Response[evaluation.TriggerEvaluationResponse]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodePermissionDenied)
			},
		},
		{
			name: "err: no job and audit scope not found",
			fields: fields{
				orchestratorClient: newOrchestratorClient(t,
					WithGetAuditScopeNotFoundError(connect.NewError(connect.CodeNotFound, fmt.Errorf("audit scope not found"))),
				),
				scheduler: gocron.NewScheduler(time.Local),
			},
			req: connect.NewRequest(&evaluation.TriggerEvaluationRequest{
				AuditScopeId: evaluationtest.MockAuditScopeId1,
			}),
			want: assert.Nil[*connect.Response[evaluation.TriggerEvaluationResponse]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeNotFound) &&
					assert.ErrorContains(t, err, "could not get audit scope from orchestrator")
			},
		},
		{
			name: "err: no job and audit scope lookup fails internally (not misreported as not found)",
			fields: fields{
				orchestratorClient: newOrchestratorClient(t,
					WithGetAuditScopeError(connect.NewError(connect.CodeInternal, fmt.Errorf("database is down"))),
				),
				scheduler: gocron.NewScheduler(time.Local),
			},
			req: connect.NewRequest(&evaluation.TriggerEvaluationRequest{
				AuditScopeId: evaluationtest.MockAuditScopeId1,
			}),
			want: assert.Nil[*connect.Response[evaluation.TriggerEvaluationResponse]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInternal) &&
					assert.ErrorContains(t, err, "could not get audit scope from orchestrator")
			},
		},
		{
			name: "err: no job and catalog not found",
			fields: fields{
				orchestratorClient: newOrchestratorClient(t,
					WithAuditScope(evaluationtest.MockAuditScope1),
					WithGetCatalogNotFoundError(fmt.Errorf("catalog not found")),
				),
				scheduler: gocron.NewScheduler(time.Local),
			},
			req: connect.NewRequest(&evaluation.TriggerEvaluationRequest{
				AuditScopeId: evaluationtest.MockAuditScopeId1,
			}),
			want: assert.Nil[*connect.Response[evaluation.TriggerEvaluationResponse]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInternal) &&
					assert.ErrorContains(t, err, "could not get catalog from the orchestrator")
			},
		},
		{
			name: "happy path: one-shot evaluation without existing job",
			fields: fields{
				orchestratorClient: newOrchestratorClient(t,
					WithAuditScope(evaluationtest.MockAuditScope1),
					WithControls(
						[]*orchestrator.Control{evaluationtest.MockControl1, evaluationtest.MockControl2},
					),
					WithCatalog(evaluationtest.MockCatalog1),
					// Only control 1 is in scope; control 2 must be skipped during evaluation
					WithControlsInScope([]*orchestrator.ControlInScope{
						{
							AuditScopeId: evaluationtest.MockAuditScopeId1,
							ControlId:    evaluationtest.MockControl1.Id,
						},
					}),
				),
				scheduler: gocron.NewScheduler(time.Local),
				catalogControls: map[string]map[string]*orchestrator.Control{
					evaluationtest.MockCatalog1.Id: {
						evaluationtest.MockControl1.Id: evaluationtest.MockControl1,
						evaluationtest.MockControl2.Id: evaluationtest.MockControl2,
					},
				},
			},
			req: connect.NewRequest(&evaluation.TriggerEvaluationRequest{
				AuditScopeId: evaluationtest.MockAuditScopeId1,
			}),
			want: func(t *testing.T, got *connect.Response[evaluation.TriggerEvaluationResponse], _ ...any) bool {
				assert.NotNil(t, got)
				return assert.True(t, got.Msg.GetSuccessful())
			},
			wantErr: assert.NoError,
		},
		{
			name: "err: no job and the controls cannot be cached",
			fields: fields{
				orchestratorClient: newOrchestratorClient(t,
					WithAuditScope(evaluationtest.MockAuditScope1),
					WithCatalog(evaluationtest.MockCatalog1),
					func(m *mockOrchestratorHandler) { m.listError = errors.New("boom") },
				),
				scheduler: gocron.NewScheduler(time.Local),
			},
			req: connect.NewRequest(&evaluation.TriggerEvaluationRequest{
				AuditScopeId: evaluationtest.MockAuditScopeId1,
			}),
			want: assert.Nil[*connect.Response[evaluation.TriggerEvaluationResponse]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInternal) &&
					assert.ErrorContains(t, err, "could not cache controls")
			},
		},
		{
			name: "err: no job and the evaluation fails",
			fields: fields{
				orchestratorClient: newOrchestratorClient(t,
					WithAuditScope(evaluationtest.MockAuditScope1),
					WithCatalog(evaluationtest.MockCatalog1),
					WithAdditionalControls(evaluationtest.MockControl1),
					func(m *mockOrchestratorHandler) { m.listEvalError = errors.New("boom") },
				),
				scheduler:       gocron.NewScheduler(time.Local),
				catalogControls: make(map[string]map[string]*orchestrator.Control),
			},
			req: connect.NewRequest(&evaluation.TriggerEvaluationRequest{
				AuditScopeId: evaluationtest.MockAuditScopeId1,
			}),
			want: assert.Nil[*connect.Response[evaluation.TriggerEvaluationResponse]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInternal) &&
					assert.ErrorContains(t, err, "evaluation failed")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := Service{
				orchestratorClient: tt.fields.orchestratorClient,
				catalogControls:    tt.fields.catalogControls,
				scheduler:          tt.fields.scheduler,
				authz:              tt.fields.authz,
			}
			got, gotErr := svc.TriggerEvaluation(context.Background(), tt.req)

			tt.want(t, got)
			tt.wantErr(t, gotErr)
		})
	}
}

func TestService_TriggerEvaluation_ExistingJob(t *testing.T) {
	var (
		ran       = make(chan struct{}, 1)
		scheduler = gocron.NewScheduler(time.Local)
	)

	// Schedule a job for the audit scope, as StartEvaluation would
	_, err := scheduler.Every(1).Day().Tag(evaluationtest.MockAuditScopeId1).Do(func() {
		ran <- struct{}{}
	})
	assert.NoError(t, err)
	scheduler.StartAsync()
	t.Cleanup(scheduler.Stop)

	// Drain the initial run that StartAsync triggers for a fresh job
	select {
	case <-ran:
	case <-time.After(5 * time.Second):
		t.Fatal("initial job run did not happen")
	}

	svc := Service{scheduler: scheduler}
	got, gotErr := svc.TriggerEvaluation(context.Background(), connect.NewRequest(&evaluation.TriggerEvaluationRequest{
		AuditScopeId: evaluationtest.MockAuditScopeId1,
	}))

	assert.NoError(t, gotErr)
	assert.NotNil(t, got)
	assert.True(t, got.Msg.GetSuccessful())

	// The existing job must have been run immediately
	select {
	case <-ran:
	case <-time.After(5 * time.Second):
		t.Fatal("triggered job run did not happen")
	}

	// The scope-change callback must trigger the job in the same way
	assert.NoError(t, svc.OnScopeChanged()(context.Background(), evaluationtest.MockAuditScopeId1))
	select {
	case <-ran:
	case <-time.After(5 * time.Second):
		t.Fatal("scope-change triggered job run did not happen")
	}
}
