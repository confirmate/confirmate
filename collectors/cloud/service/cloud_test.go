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

package cloud

import (
	"context"
	"errors"
	"io"
	"sync"
	"testing"
	"time"

	"net/http"
	"net/http/httptest"

	collector "confirmate.io/collectors/cloud/internal/collector"
	"confirmate.io/collectors/cloud/internal/collectortest"
	"confirmate.io/collectors/cloud/internal/config"
	"confirmate.io/collectors/cloud/internal/testdata"
	"confirmate.io/core/api/evidence"
	"confirmate.io/core/api/evidence/evidenceconnect"
	"confirmate.io/core/api/ontology"
	"confirmate.io/core/server"
	"confirmate.io/core/server/servertest"
	"confirmate.io/core/service"
	"confirmate.io/core/service/collection"
	"confirmate.io/core/util/assert"
	"connectrpc.com/connect"
	"github.com/go-co-op/gocron"
	"github.com/urfave/cli/v3"
	"golang.org/x/oauth2/clientcredentials"
)

func TestNewService(t *testing.T) {
	type args struct {
		opts []service.Option[Service]
	}
	tests := []struct {
		name string
		args args
		want assert.Want[*Service]
	}{
		{
			name: "Create service with option 'WithEvidenceStoreAddress'",
			args: args{
				opts: []service.Option[Service]{
					WithEvidenceStoreAddress("localhost:9091", service.DefaultHTTPClient),
				},
			},
			want: func(t *testing.T, got *Service, msgAndArgs ...any) bool {
				if got.cloudConfig.evStreamConfig.client != service.DefaultHTTPClient {
					t.Errorf("expected shared default HTTP client to be used")
					return false
				}

				return assert.Equal(t, "localhost:9091", got.cloudConfig.evStreamConfig.targetAddress)
			},
		},
		{
			name: "Create service with option 'WithTargetOfEvaluationID'",
			args: args{
				opts: []service.Option[Service]{
					WithTargetOfEvaluationID(testdata.MockTargetOfEvaluationID1),
				},
			},
			want: func(t *testing.T, got *Service, msgAndArgs ...any) bool {
				return assert.Equal(t, testdata.MockTargetOfEvaluationID1, got.cloudConfig.targetOfEvaluationID)
			},
		},
		{
			name: "Create service with option 'WithEvidenceCollectorToolID'",
			args: args{
				opts: []service.Option[Service]{
					WithCollectorToolID(testdata.MockEvidenceToolID1),
				},
			},
			want: func(t *testing.T, got *Service, msgAndArgs ...any) bool {
				return assert.Equal(t, testdata.MockEvidenceToolID1, got.cloudConfig.collectorToolID)
			},
		},
		{
			name: "Create service with option 'WithProvider' and one provider given",
			args: args{
				opts: []service.Option[Service]{
					WithProvider("azure"),
				},
			},
			want: func(t *testing.T, got *Service, msgAndArgs ...any) bool {
				return assert.Equal(t, "azure", got.cloudConfig.provider)
			},
		},
		{
			name: "Create service with option 'WithProvider' and no provider given",
			args: args{
				opts: []service.Option[Service]{
					WithProvider(""),
				},
			},
			want: func(t *testing.T, got *Service, msgAndArgs ...any) bool {
				return assert.Equal(t, "", got.cloudConfig.provider)
			},
		},
		{
			name: "Create service with option 'WithAdditionalCollectors'",
			args: args{
				opts: []service.Option[Service]{
					WithAdditionalCollectors([]collector.Collector{&collectortest.TestCollector{ServiceId: config.DefaultTargetOfEvaluationID}}),
				},
			},
			want: func(t *testing.T, got *Service, msgAndArgs ...any) bool {
				return assert.Contains(t, got.collectors, &collectortest.TestCollector{ServiceId: config.DefaultTargetOfEvaluationID})
			},
		},
		{
			name: "Create service with option 'WithCollectorInterval'",
			args: args{
				opts: []service.Option[Service]{
					WithCollectorInterval(time.Duration(8)),
				},
			},
			want: func(t *testing.T, got *Service, msgAndArgs ...any) bool {
				return assert.Equal(t, time.Duration(8), got.cloudConfig.collectorInterval)
			},
		},
		{
			name: "Create service without any option",
			args: args{},
			want: func(t *testing.T, got *Service, msgAndArgs ...any) bool {
				assert.NotNil(t, got.evidenceStoreStream)
				return assert.NotNil(t, got.evidenceStoreClient)
			},
		},
		{
			name: "Create service with option 'WithServiceOAuth2Config'",
			args: args{
				opts: []service.Option[Service]{
					WithServiceOAuth2Config(&clientcredentials.Config{
						ClientID:     "test-client",
						ClientSecret: "test-secret",
						TokenURL:     "http://localhost:8080/v1/auth/token",
					}),
				},
			},
			want: func(t *testing.T, got *Service, msgAndArgs ...any) bool {
				assert.NotNil(t, got.cloudConfig.serviceOAuth2Config)
				// The default HTTP client must have been wrapped with an OAuth2-authenticating
				// transport rather than left untouched.
				return assert.True(t, got.cloudConfig.evStreamConfig.client != service.DefaultHTTPClient)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewService(tt.args.opts...)
			tt.want(t, got)
		})
	}
}

func TestBuildCollectors(t *testing.T) {
	var (
		extraCollector *collectortest.TestCollector
		collectors     []collection.Collector
		err            error
	)

	extraCollector = &collectortest.TestCollector{ServiceId: config.DefaultTargetOfEvaluationID}
	collectors, err = buildCollectors(
		&cli.Command{},
		WithProvider(""),
		WithAdditionalCollectors([]collector.Collector{extraCollector}),
	)

	assert.NoError(t, err)
	assert.Equal(t, 1, len(collectors))
	assert.Equal(t, extraCollector.ID(), collectors[0].ID())
}

func TestBuildCollectors_ReturnsErrorForUnknownProvider(t *testing.T) {
	var err error

	_, err = buildCollectors(
		&cli.Command{},
		WithProvider("unknown"),
	)

	assert.Error(t, err)
}

type mockEvidenceStoreHandler struct {
	evidenceconnect.UnimplementedEvidenceStoreHandler

	mu       sync.Mutex
	requests []*evidence.StoreEvidenceRequest

	responseFunc func(*evidence.StoreEvidenceRequest) (*evidence.StoreEvidencesResponse, error)
}

func (h *mockEvidenceStoreHandler) StoreEvidences(_ context.Context, stream *connect.BidiStream[evidence.StoreEvidenceRequest, evidence.StoreEvidencesResponse]) (err error) {
	var (
		req *evidence.StoreEvidenceRequest
		res *evidence.StoreEvidencesResponse
	)

	for {
		req, err = stream.Receive()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}

		h.mu.Lock()
		h.requests = append(h.requests, req)
		h.mu.Unlock()

		if h.responseFunc != nil {
			res, err = h.responseFunc(req)
			if err != nil {
				return err
			}
		} else {
			res = &evidence.StoreEvidencesResponse{Status: evidence.EvidenceStatus_EVIDENCE_STATUS_OK}
		}

		err = stream.Send(res)
		if err != nil {
			return err
		}
	}
}

func (h *mockEvidenceStoreHandler) Requests() (requests []*evidence.StoreEvidenceRequest) {
	h.mu.Lock()
	defer h.mu.Unlock()

	requests = append([]*evidence.StoreEvidenceRequest(nil), h.requests...)
	return requests
}

// requireBearerInterceptor rejects every streaming request that does not carry the expected
// "Authorization: Bearer <token>" header, mimicking [server.AuthInterceptor]'s behavior for
// unauthenticated requests without needing a real JWT/JWKS setup in tests.
type requireBearerInterceptor struct {
	token string
}

func (i *requireBearerInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return next
}

func (i *requireBearerInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return next
}

func (i *requireBearerInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		if conn.RequestHeader().Get("Authorization") != "Bearer "+i.token {
			return connect.NewError(connect.CodeUnauthenticated, errors.New("invalid auth token"))
		}
		return next(ctx, conn)
	}
}

// newTestTokenServer spins up a minimal OAuth 2.0 client credentials token endpoint that always
// issues the given access token, regardless of the presented client credentials.
func newTestTokenServer(t *testing.T, accessToken string) *httptest.Server {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"` + accessToken + `","token_type":"bearer","expires_in":3600}`))
	}))
	t.Cleanup(srv.Close)

	return srv
}

type startCollectorTestCollector struct {
	name                 string
	id                   string
	targetOfEvaluationID string
	resources            []ontology.IsResource
	collectErr           error
}

func (c *startCollectorTestCollector) Name() string { return c.name }

func (c *startCollectorTestCollector) ID() string { return c.id }

func (c *startCollectorTestCollector) Collect() ([]ontology.IsResource, error) {
	if c.collectErr != nil {
		return nil, c.collectErr
	}

	return c.resources, nil
}

func (c *startCollectorTestCollector) List() ([]ontology.IsResource, error) {
	return c.Collect()
}

func (c *startCollectorTestCollector) TargetOfEvaluationID() string {
	return c.targetOfEvaluationID
}

func TestService_StartCollector(t *testing.T) {
	type fields struct {
		opts      []service.Option[Service]
		collector collector.Collector
	}

	tests := []struct {
		name         string
		fields       fields
		want         assert.Want[*Service]
		wantEvent    []CollectorEventType
		wantCount    int
		responseFunc func(*evidence.StoreEvidenceRequest) (*evidence.StoreEvidencesResponse, error)
	}{
		{
			name: "collector error emits start event without evidence",
			fields: fields{
				opts: []service.Option[Service]{
					WithTargetOfEvaluationID(testdata.MockTargetOfEvaluationID1),
					WithCollectorToolID(testdata.MockEvidenceToolID1),
				},
				collector: &startCollectorTestCollector{
					name:                 "failing-collector",
					id:                   "failing-collector-id",
					targetOfEvaluationID: testdata.MockTargetOfEvaluationID1,
					collectErr:           errors.New("boom"),
				},
			},
			want: func(t *testing.T, got *Service, msgAndArgs ...any) bool {
				return assert.False(t, got.dead)
			},
			wantEvent: []CollectorEventType{CloudCollectorStart},
			wantCount: 0,
		},
		{
			name: "successful collection forwards evidence and emits start and finish events",
			fields: fields{
				opts: []service.Option[Service]{
					WithTargetOfEvaluationID(testdata.MockTargetOfEvaluationID1),
					WithCollectorToolID(testdata.MockEvidenceToolID1),
				},
				collector: &startCollectorTestCollector{
					name:                 "successful-collector",
					id:                   "successful-collector-id",
					targetOfEvaluationID: testdata.MockTargetOfEvaluationID1,
					resources: []ontology.IsResource{
						&ontology.VirtualMachine{Id: new("vm-1")},
						&ontology.ObjectStorage{Id: new("storage-1")},
					},
				},
			},
			want: func(t *testing.T, got *Service, msgAndArgs ...any) bool {
				return assert.False(t, got.dead)
			},
			wantEvent: []CollectorEventType{CloudCollectorStart, CloudCollectorFinished},
			wantCount: 2,
		},
		{
			name: "evidence-store rejection marks stream dead and later sends reopen it",
			fields: fields{
				opts: []service.Option[Service]{
					WithTargetOfEvaluationID(testdata.MockTargetOfEvaluationID1),
					WithCollectorToolID(testdata.MockEvidenceToolID1),
				},
				collector: &startCollectorTestCollector{
					name:                 "rejected-collector",
					id:                   "rejected-collector-id",
					targetOfEvaluationID: testdata.MockTargetOfEvaluationID1,
					resources: []ontology.IsResource{
						&ontology.VirtualMachine{Id: new("vm-1")},
						&ontology.ObjectStorage{Id: new("storage-1")},
					},
				},
			},
			want: func(t *testing.T, got *Service, msgAndArgs ...any) bool {
				return assert.True(t, got.dead)
			},
			wantEvent: []CollectorEventType{CloudCollectorStart, CloudCollectorFinished},
			wantCount: 2,
			responseFunc: func(*evidence.StoreEvidenceRequest) (*evidence.StoreEvidencesResponse, error) {
				return &evidence.StoreEvidencesResponse{
					Status:        evidence.EvidenceStatus_EVIDENCE_STATUS_ERROR,
					StatusMessage: "rejected by test evidence store",
				}, nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				handler     *mockEvidenceStoreHandler
				svc         *Service
				requests    []*evidence.StoreEvidenceRequest
				received    []*CollectorEvent
				collectorEv *CollectorEvent
			)

			handler = &mockEvidenceStoreHandler{}
			handler.responseFunc = tt.responseFunc
			_, testSrv := servertest.NewTestConnectServer(t,
				server.WithHandler(evidenceconnect.NewEvidenceStoreHandler(handler)),
			)
			defer testSrv.Close()

			tt.fields.opts = append(tt.fields.opts, WithEvidenceStoreAddress(testSrv.URL, testSrv.Client()))
			svc = NewService(tt.fields.opts...)
			defer svc.Shutdown()
			svc.Events = make(chan *CollectorEvent, len(tt.wantEvent))

			svc.StartCollector(tt.fields.collector)

			for range tt.wantEvent {
				select {
				case collectorEv = <-svc.Events:
					received = append(received, collectorEv)
				case <-time.After(2 * time.Second):
					t.Fatal("timed out waiting for collector event")
				}
			}

			requests = waitForStoredRequests(t, handler, tt.wantCount)
			tt.want(t, svc)
			assert.Equal(t, len(tt.wantEvent), len(received))
			for index, wantEvent := range tt.wantEvent {
				assert.Equal(t, wantEvent, received[index].Type)
			}

			if len(requests) == 0 {
				assert.Equal(t, 1, len(received))
				return
			}

			assert.Equal(t, tt.wantCount, len(requests))
			assert.Equal(t, testdata.MockTargetOfEvaluationID1, requests[0].GetEvidence().GetTargetOfEvaluationId())
			assert.Equal(t, testdata.MockEvidenceToolID1, requests[0].GetEvidence().GetToolId())
			assert.Equal(t, "vm-1", requests[0].GetEvidence().GetResource().GetVirtualMachine().GetId())
			if len(requests) > 1 {
				assert.Equal(t, "storage-1", requests[1].GetEvidence().GetResource().GetObjectStorage().GetId())
			}
			assert.Equal(t, 2, received[1].CollectedItems)
		})
	}
}

func waitForStoredRequests(t *testing.T, handler *mockEvidenceStoreHandler, wantCount int) (requests []*evidence.StoreEvidenceRequest) {
	t.Helper()

	deadline := time.After(2 * time.Second)
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		requests = handler.Requests()
		if len(requests) == wantCount {
			return requests
		}

		select {
		case <-deadline:
			return requests
		case <-ticker.C:
		}
	}
}

func TestService_StartCollector_AuthRequiredByEvidenceStore(t *testing.T) {
	tests := []struct {
		name           string
		withOAuth2     bool
		wantDead       bool
		wantStoredReqs int
	}{
		{
			name:           "no ServiceOAuth2Config: evidence store rejects the connection",
			withOAuth2:     false,
			wantDead:       true,
			wantStoredReqs: 0,
		},
		{
			name:           "matching ServiceOAuth2Config: evidence store accepts the connection",
			withOAuth2:     true,
			wantDead:       false,
			wantStoredReqs: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := &mockEvidenceStoreHandler{}
			_, testSrv := servertest.NewTestConnectServer(t,
				server.WithHandler(evidenceconnect.NewEvidenceStoreHandler(handler,
					connect.WithInterceptors(&requireBearerInterceptor{token: "test-access-token"}))),
			)
			defer testSrv.Close()

			opts := []service.Option[Service]{
				WithTargetOfEvaluationID(testdata.MockTargetOfEvaluationID1),
				WithCollectorToolID(testdata.MockEvidenceToolID1),
				WithEvidenceStoreAddress(testSrv.URL, testSrv.Client()),
			}
			if tt.withOAuth2 {
				tokenServer := newTestTokenServer(t, "test-access-token")
				opts = append(opts, WithServiceOAuth2Config(&clientcredentials.Config{
					ClientID:     "test-client",
					ClientSecret: "test-secret",
					TokenURL:     tokenServer.URL,
				}))
			}

			svc := NewService(opts...)
			defer svc.Shutdown()

			svc.StartCollector(&startCollectorTestCollector{
				name:                 "collector",
				id:                   "collector-id",
				targetOfEvaluationID: testdata.MockTargetOfEvaluationID1,
				resources:            []ontology.IsResource{&ontology.VirtualMachine{Id: new("vm-1")}},
			})

			requests := waitForStoredRequests(t, handler, tt.wantStoredReqs)
			assert.Equal(t, tt.wantStoredReqs, len(requests))
			assert.Equal(t, tt.wantDead, svc.dead)
		})
	}
}

func TestService_Shutdown(t *testing.T) {
	service := NewService()
	service.Shutdown()

	assert.False(t, service.scheduler.IsRunning())

}

func TestService_Start(t *testing.T) {
	type envVariable struct {
		hasEnvVariable   bool
		envVariableKey   string
		envVariableValue string
	}
	type fields struct {
		evidenceStoreClient evidenceconnect.EvidenceStoreClient
		evidenceStoreStream *connect.BidiStreamForClient[evidence.StoreEvidenceRequest, evidence.StoreEvidencesResponse]
		dead                bool
		scheduler           *gocron.Scheduler
		Events              chan *CollectorEvent
		envVariables        []envVariable
		cloudConfig         CloudCollectorConfig
	}
	type args struct {
		cmd *cli.Command
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    assert.Want[*Service]
		wantErr assert.WantErr
	}{
		{
			name: "Request with wrong provider name",
			fields: fields{
				scheduler: gocron.NewScheduler(time.UTC),
				cloudConfig: CloudCollectorConfig{
					provider: "falseProvider",
				},
			},
			args: args{
				cmd: &cli.Command{},
			},
			want: func(t *testing.T, got *Service, msgAndArgs ...any) bool {
				return assert.False(t, got.scheduler.IsRunning())
			},
			wantErr: func(t *testing.T, gotErr error, msgAndArgs ...any) bool {
				return assert.ErrorContains(t, gotErr, "'falseProvider' not known")
			},
		},
		{
			name: "collector interval error",
			fields: fields{
				scheduler: gocron.NewScheduler(time.UTC),
				cloudConfig: CloudCollectorConfig{
					provider:          ProviderAzure,
					collectorInterval: time.Duration(-5 * time.Minute),
				},
				envVariables: []envVariable{
					{
						hasEnvVariable:   true,
						envVariableKey:   "AZURE_TENANT_ID",
						envVariableValue: "tenant-id-123",
					},
					{
						hasEnvVariable:   true,
						envVariableKey:   "AZURE_CLIENT_ID",
						envVariableValue: "client-id-123",
					},
					{
						hasEnvVariable:   true,
						envVariableKey:   "AZURE_CLIENT_SECRET",
						envVariableValue: "client-secret-456",
					},
				},
			},
			args: args{
				cmd: &cli.Command{},
			},
			want: func(t *testing.T, got *Service, msgAndArgs ...any) bool {
				assert.Equal(t, ProviderAzure, got.cloudConfig.provider)
				return assert.False(t, got.scheduler.IsRunning())
			},
			wantErr: func(t *testing.T, gotErr error, msgAndArgs ...any) bool {
				return assert.ErrorContains(t, gotErr, "could not schedule job for ", ".Every() interval must be greater than 0")
			},
		},
		{
			name: "K8S authorizer error",
			fields: fields{
				scheduler: gocron.NewScheduler(time.UTC),
				cloudConfig: CloudCollectorConfig{
					provider:          ProviderK8S,
					collectorInterval: time.Duration(5 * time.Minute),
				},
				envVariables: []envVariable{
					// We must set HOME to a wrong path so that the K8S authorizer fails
					{
						hasEnvVariable:   true,
						envVariableKey:   "HOME",
						envVariableValue: "",
					},
				},
			},
			args: args{
				cmd: &cli.Command{},
			},
			want: func(t *testing.T, got *Service, msgAndArgs ...any) bool {
				assert.Equal(t, ProviderK8S, got.cloudConfig.provider)
				return assert.False(t, got.scheduler.IsRunning())
			},
			wantErr: func(t *testing.T, gotErr error, msgAndArgs ...any) bool {
				return assert.ErrorContains(t, gotErr, ErrK8sAuth.Error())
			},
		},
		{
			name: "Azure authorizer error",
			fields: fields{
				scheduler: gocron.NewScheduler(time.UTC),
				cloudConfig: CloudCollectorConfig{
					targetOfEvaluationID: config.DefaultTargetOfEvaluationID,
					provider:             ProviderAzure,
					collectorInterval:    time.Duration(5 * time.Minute),
				},
				envVariables: []envVariable{
					{
						hasEnvVariable:   true,
						envVariableKey:   "AZURE_TOKEN_CREDENTIALS",
						envVariableValue: "fail",
					},
				},
			},
			args: args{
				cmd: &cli.Command{},
			},
			want: func(t *testing.T, got *Service, msgAndArgs ...any) bool {
				assert.Equal(t, ProviderAzure, got.cloudConfig.provider)
				assert.Equal(t, config.DefaultTargetOfEvaluationID, got.cloudConfig.targetOfEvaluationID)
				return assert.False(t, got.scheduler.IsRunning())
			},
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.ErrorContains(t, err, ErrAzureAuth.Error())
			},
		},
		{
			name: "AWS authorizer error",
			fields: fields{
				scheduler: gocron.NewScheduler(time.UTC),
				cloudConfig: CloudCollectorConfig{
					targetOfEvaluationID: config.DefaultTargetOfEvaluationID,
					provider:             ProviderAWS,
					collectorInterval:    time.Duration(5 * time.Minute),
				},
			},
			args: args{
				cmd: &cli.Command{},
			},
			want: func(t *testing.T, got *Service, msgAndArgs ...any) bool {
				assert.Equal(t, ProviderAWS, got.cloudConfig.provider)
				assert.Equal(t, config.DefaultTargetOfEvaluationID, got.cloudConfig.targetOfEvaluationID)
				return assert.False(t, got.scheduler.IsRunning())
			},
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.ErrorContains(t, err, ErrAWSAuth.Error())
			},
		},
		{
			name: "OpenStack authorizer error",
			fields: fields{
				scheduler: gocron.NewScheduler(time.UTC),
				cloudConfig: CloudCollectorConfig{
					provider:          ProviderOpenstack,
					collectorInterval: time.Duration(5 * time.Minute),
				},
			},
			args: args{
				cmd: &cli.Command{},
			},
			want: func(t *testing.T, got *Service, msgAndArgs ...any) bool {
				assert.Equal(t, ProviderOpenstack, got.cloudConfig.provider)
				return assert.False(t, got.scheduler.IsRunning())
			},
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.ErrorContains(t, err, ErrOpenstackAuth.Error())
			},
		},
		{
			name: "Happy path: no collector interval error",
			fields: fields{
				scheduler: gocron.NewScheduler(time.UTC),
				envVariables: []envVariable{
					{
						hasEnvVariable:   true,
						envVariableKey:   "AZURE_TENANT_ID",
						envVariableValue: "tenant-id-123",
					},
					{
						hasEnvVariable:   true,
						envVariableKey:   "AZURE_CLIENT_ID",
						envVariableValue: "client-id-123",
					},
					{
						hasEnvVariable:   true,
						envVariableKey:   "AZURE_CLIENT_SECRET",
						envVariableValue: "client-secret-456",
					},
				},
				cloudConfig: CloudCollectorConfig{
					provider:          ProviderAzure,
					collectorInterval: time.Duration(5 * time.Minute),
				},
			},
			args: args{
				cmd: &cli.Command{},
			},
			want: func(t *testing.T, got *Service, msgAndArgs ...any) bool {
				assert.Equal(t, ProviderAzure, got.cloudConfig.provider)
				return assert.True(t, got.scheduler.IsRunning())
			},
			wantErr: assert.Nil[error],
		},
		{
			name: "Happy path: Azure authorizer from ENV",
			fields: fields{
				scheduler: gocron.NewScheduler(time.UTC),
				cloudConfig: CloudCollectorConfig{
					provider:          ProviderAzure,
					collectorInterval: time.Duration(5 * time.Minute),
				},
				envVariables: []envVariable{
					{
						hasEnvVariable:   true,
						envVariableKey:   "AZURE_TENANT_ID",
						envVariableValue: "tenant-id-123",
					},
					{
						hasEnvVariable:   true,
						envVariableKey:   "AZURE_CLIENT_ID",
						envVariableValue: "client-id-123",
					},
					{
						hasEnvVariable:   true,
						envVariableKey:   "AZURE_CLIENT_SECRET",
						envVariableValue: "client-secret-456",
					},
				},
			},
			args: args{
				cmd: &cli.Command{},
			},
			want: func(t *testing.T, got *Service, msgAndArgs ...any) bool {
				assert.Equal(t, ProviderAzure, got.cloudConfig.provider)
				return assert.True(t, got.scheduler.IsRunning())
			},
			wantErr: assert.Nil[error],
		},
		{
			name: "Happy path: Azure with resource group",
			fields: fields{
				scheduler: gocron.NewScheduler(time.UTC),
				cloudConfig: CloudCollectorConfig{
					provider:          ProviderAzure,
					collectorInterval: time.Duration(5 * time.Minute),
				},
				envVariables: []envVariable{
					{
						hasEnvVariable:   true,
						envVariableKey:   "AZURE_TENANT_ID",
						envVariableValue: "tenant-id-123",
					},
					{
						hasEnvVariable:   true,
						envVariableKey:   "AZURE_CLIENT_ID",
						envVariableValue: "client-id-123",
					},
					{
						hasEnvVariable:   true,
						envVariableKey:   "AZURE_CLIENT_SECRET",
						envVariableValue: "client-secret-456",
					},
				},
			},
			args: args{
				cmd: &cli.Command{
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:  "collector-resource-group",
							Value: "my-resource-group",
						},
					},
				},
			},
			want: func(t *testing.T, got *Service, msgAndArgs ...any) bool {
				// We are not able to check if the resource group was set, but at least we can check if the provider is correct and the scheduler is running.
				assert.Equal(t, ProviderAzure, got.cloudConfig.provider)
				assert.NotEmpty(t, got.collectors)
				return assert.True(t, got.scheduler.IsRunning())
			},
			wantErr: assert.Nil[error],
		},
		{
			name: "Happy path: CSAF with domain",
			fields: fields{
				scheduler: gocron.NewScheduler(time.UTC),
				cloudConfig: CloudCollectorConfig{
					provider:          ProviderCSAF,
					collectorInterval: time.Duration(5 * time.Minute),
				},
			},
			args: args{
				cmd: &cli.Command{
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:  "collector-csaf-domain",
							Value: "example.com",
						},
					},
				},
			},
			want: func(t *testing.T, got *Service, msgAndArgs ...any) bool {
				// We are not able to check if the CSAF domain was set, but at least we can check if the provider is correct and the scheduler is running.
				assert.Equal(t, ProviderCSAF, got.cloudConfig.provider)
				assert.NotEmpty(t, got.collectors)
				return assert.True(t, got.scheduler.IsRunning())
			},
			wantErr: assert.Nil[error],
		},
		{
			name: "Happy path: CSAF without domain",
			fields: fields{
				scheduler: gocron.NewScheduler(time.UTC),
				cloudConfig: CloudCollectorConfig{
					provider:          ProviderCSAF,
					collectorInterval: time.Duration(5 * time.Minute),
				},
			},
			args: args{
				cmd: &cli.Command{},
			},
			want: func(t *testing.T, got *Service, msgAndArgs ...any) bool {
				// We are not able to check if the CSAF domain was set, but at least we can check if the provider is correct and the scheduler is running.
				assert.Equal(t, ProviderCSAF, got.cloudConfig.provider)
				assert.NotEmpty(t, got.collectors)
				return assert.True(t, got.scheduler.IsRunning())
			},
			wantErr: assert.Nil[error],
		},
		// Note: Currently not possible to test K8S, because it requires a kubconfig file to be present in the environment.
		// {
		// 	name: "Happy path: K8S",
		// 	fields: fields{
		// 		scheduler: gocron.NewScheduler(time.UTC),
		// 		cloudConfig: CloudCollectorConfig{
		// 			provider:          ProviderK8S,
		// 			collectorInterval: time.Duration(5 * time.Minute),
		// 		},
		// 	},
		// 	args: args{
		// 		cmd: &cli.Command{},
		// 	},
		// 	want: func(t *testing.T, got *Service, msgAndArgs ...any) bool {
		// 		assert.Equal(t, ProviderK8S, got.cloudConfig.provider)
		// 		return assert.True(t, got.scheduler.IsRunning())
		// 	},
		// 	wantErr: assert.Nil[error],
		// },
		{
			name: "Happy path: OpenStack",
			fields: fields{
				scheduler: gocron.NewScheduler(time.UTC),
				cloudConfig: CloudCollectorConfig{
					provider:          ProviderOpenstack,
					collectorInterval: time.Duration(5 * time.Minute),
				},
				envVariables: []envVariable{
					{
						hasEnvVariable:   true,
						envVariableKey:   "OS_AUTH_URL",
						envVariableValue: "project-id-123",
					},
					{
						hasEnvVariable:   true,
						envVariableKey:   "OS_USERID",
						envVariableValue: "client-id-123",
					},
					{
						hasEnvVariable:   true,
						envVariableKey:   "OS_PASSWORD",
						envVariableValue: "client-secret-456",
					},
				},
			},
			args: args{
				cmd: &cli.Command{},
			},
			want: func(t *testing.T, got *Service, msgAndArgs ...any) bool {
				assert.Equal(t, ProviderOpenstack, got.cloudConfig.provider)
				return assert.True(t, got.scheduler.IsRunning())
			},
			wantErr: assert.Nil[error],
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{
				evidenceStoreClient: tt.fields.evidenceStoreClient,
				evidenceStoreStream: tt.fields.evidenceStoreStream,
				dead:                tt.fields.dead,
				scheduler:           tt.fields.scheduler,
				Events:              tt.fields.Events,
				cloudConfig:         tt.fields.cloudConfig,
			}

			// Set env variables
			for _, env := range tt.fields.envVariables {
				if env.hasEnvVariable {
					t.Setenv(env.envVariableKey, env.envVariableValue)
				}
			}

			err := svc.Start(tt.args.cmd)

			tt.want(t, svc)
			tt.wantErr(t, err)
		})
	}
}

func TestService_GetTargetOfEvaluationId(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for receiver constructor.
		opts service.Option[Service]
		want assert.Want[string]
	}{
		{
			name: "Happy path",
			opts: WithTargetOfEvaluationID("test-target-of-eval-id"),
			want: func(t *testing.T, got string, msgAndArgs ...any) bool {
				return assert.Equal(t, "test-target-of-eval-id", got)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewService(tt.opts)
			got := svc.GetTargetOfEvaluationId()

			tt.want(t, got)
		})
	}
}
