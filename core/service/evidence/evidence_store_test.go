package evidence

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"confirmate.io/core/api/evidence"
	"confirmate.io/core/api/evidence/evidenceconnect"
	"confirmate.io/core/api/ontology"
	"confirmate.io/core/internal/testutil/servicetest/evidencetest"
	"confirmate.io/core/persistence"
	"confirmate.io/core/persistence/persistencetest"
	"confirmate.io/core/server"
	"confirmate.io/core/server/servertest"
	"confirmate.io/core/service"
	"confirmate.io/core/util"
	"confirmate.io/core/util/assert"
	"connectrpc.com/connect"
	"github.com/google/uuid"
)

func TestMain(m *testing.M) {
	// Start the Evidence Store server

	// Start the Assessment server

	code := m.Run()
	os.Exit(code)
}

// TestNewService provides simple tests for NewService
// What we could not test:
// - Error when a new DB has to be created (NewService doesn't expose a way to inject failing DB creation)
func TestNewService(t *testing.T) {
	type args struct {
		opts []service.Option[Service]
	}
	tests := []struct {
		name    string
		args    args
		want    assert.Want[*Service]
		wantErr assert.WantErr
	}{
		{
			name: "EvidenceStoreServer created with in-memory DB",
			args: args{opts: []service.Option[Service]{
				WithDB(persistencetest.NewInMemoryDB(t, types, nil)),
			}},
			want: func(t *testing.T, got *Service, msgAndArgs ...any) bool {
				// Storage should be in-memory storage
				assert.NotNil(t, got.db)
				assert.NotNil(t, got.assessmentStream)
				return true
			},
			wantErr: assert.NoError,
		},
		{
			name: "Happy path: EvidenceStoreServer created with option 'WithDB'",
			args: args{opts: []service.Option[Service]{
				WithDB(persistencetest.NewInMemoryDB(t, types, nil, evidencetest.InitDBWithEvidence))}},
			want: func(t *testing.T, got *Service, msgAndArgs ...any) bool {
				// Storage should be gorm (in-memory storage). Hard to check since its type is not exported
				assert.NotNil(t, got.db)
				assert.NotNil(t, got.assessmentStream)
				// But we can check if we can get the evidence we inserted into the custom DB
				gotEvidence, err := got.GetEvidence(context.Background(), &connect.Request[evidence.GetEvidenceRequest]{
					Msg: &evidence.GetEvidenceRequest{EvidenceId: evidencetest.MockEvidence1.Id}})
				assert.NoError(t, err)
				assert.NotNil(t, gotEvidence)
				assert.Equal(t, evidencetest.MockEvidence1.Id, gotEvidence.Msg.Id)
				return true
			},
			wantErr: assert.NoError,
		},
		{
			name: "EvidenceStoreServer created with option 'WithConfig' - no client provided",
			args: args{opts: []service.Option[Service]{
				WithDB(persistencetest.NewInMemoryDB(t, types, nil)),
				WithConfig(Config{
					AssessmentAddress: "localhost:9091",
					AssessmentClient:  nil,
					PersistenceConfig: persistence.DefaultConfig,
					EvidenceQueueSize: DefaultConfig.EvidenceQueueSize,
				}),
			}},
			want: func(t *testing.T, got *Service, msgAndArgs ...any) bool {
				// We didn't provide a client, so it should use http.DefaultClient
				assert.NotNil(t, got.assessmentStream)
				return assert.Equal(t, "localhost:9091", got.cfg.AssessmentAddress)
			},
			wantErr: assert.NoError,
		},
		{
			name: "EvidenceStoreServer created with option 'WithConfig' - with client",
			args: args{opts: []service.Option[Service]{
				WithDB(persistencetest.NewInMemoryDB(t, types, nil)),
				WithConfig(Config{
					AssessmentAddress: "localhost:9091",
					AssessmentClient:  &http.Client{Timeout: time.Duration(1)},
					PersistenceConfig: persistence.DefaultConfig,
					EvidenceQueueSize: DefaultConfig.EvidenceQueueSize,
				}),
			}},
			want: func(t *testing.T, got *Service, msgAndArgs ...any) bool {
				// We provided a client with timeout set to 1 nanosecond
				assert.Equal(t, 1, got.cfg.AssessmentClient.Timeout)
				assert.NotNil(t, got.assessmentStream)
				return assert.Equal(t, "localhost:9091", got.cfg.AssessmentAddress)
			},
			wantErr: assert.NoError,
		},
		{
			name: "Error - assessment stream init fails",
			args: args{opts: []service.Option[Service]{
				WithDB(persistencetest.NewInMemoryDB(t, types, nil)),
				WithAssessmentClient(nilAssessmentClient{}),
			}},
			want: assert.Nil[*Service],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.ErrorContains(t, err, "factory returned nil stream")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewService(tt.args.opts...)
			tt.want(t, got)
			tt.wantErr(t, err)
		})
	}
}

// TestNewService_WithConfig ensures Config is honored during initialization.
func TestNewService_WithConfig(t *testing.T) {
	var (
		cfg Config
		svc *Service
		err error
	)

	cfg = DefaultConfig
	cfg.PersistenceConfig.InMemoryDB = true

	svc, err = NewService(WithConfig(cfg))
	assert.NoError(t, err)
	assert.NotNil(t, svc)
	assert.NotNil(t, svc.db)
}

// TestService_sendToAssessment validates the evidence->assessment wiring.
// NOTE: We intentionally do not validate restart behavior here; that is covered
// in core/stream tests and would make this integration test flaky.
func TestService_sendToAssessment(t *testing.T) {
	// Step 1: Start assessment server.
	recorder, srv, testSrv := newAssessmentTestServer(t)
	defer testSrv.Close()
	assert.NotNil(t, srv)
	assert.NotNil(t, testSrv)

	// Step 2: Create service with assessment client.
	svc, err := NewService(
		WithDB(persistencetest.NewInMemoryDB(t, types, nil)),
		WithConfig(Config{
			AssessmentAddress: testSrv.URL,
			AssessmentClient:        testSrv.Client(),
		}),
	)
	assert.NoError(t, err)
	assert.NotNil(t, svc.assessmentStream)

	// Step 3: Happy path send.
	evidenceOne := &evidence.Evidence{Id: uuid.NewString()}
	err = svc.sendToAssessment(evidenceOne)
	assert.NoError(t, err)
	awaitAssessmentRequest(t, recorder.received, evidenceOne.Id)

	// Step 4: Close stream and verify send error.
	assert.NoError(t, svc.assessmentStream.Close())
	evidenceTwo := &evidence.Evidence{Id: uuid.NewString()}
	err = svc.sendToAssessment(evidenceTwo)
	assert.Error(t, err)
	assert.ErrorContains(t, err, "failed to send evidence")

	// Step 5: Nil stream guard.
	svc.assessmentStream = nil
	evidenceThree := &evidence.Evidence{Id: uuid.NewString()}
	err = svc.sendToAssessment(evidenceThree)
	assert.Error(t, err)
	assert.ErrorContains(t, err, "assessment stream is not initialized")
}

// TestService_initAssessmentStream verifies creation, idempotency, and error handling.
func TestService_initAssessmentStream(t *testing.T) {
	// Start an assessment server for the stream factory.
	_, _, testSrv := newAssessmentTestServer(t)
	defer testSrv.Close()

	svc, err := NewService(
		WithDB(persistencetest.NewInMemoryDB(t, types, nil)),
		WithConfig(Config{
			AssessmentAddress: testSrv.URL,
			AssessmentClient:        testSrv.Client(),
		}),
	)
	assert.NoError(t, err)

	tests := []struct {
		name             string
		resetStream      bool
		overrideClient   bool
		expectErr        bool
		expectSameStream bool
	}{
		{
			name:        "happy path: create stream when nil",
			resetStream: true,
		},
		{
			name:             "happy path: idempotent when stream exists",
			expectSameStream: true,
		},
		{
			name:           "error: factory returns nil",
			resetStream:    true,
			overrideClient: true,
			expectErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			streamBefore := svc.assessmentStream
			if tt.resetStream {
				svc.assessmentStream = nil
			}
			if tt.overrideClient {
				svc.assessmentClient = nilAssessmentClient{}
			}

			err = svc.initAssessmentStream()
			if tt.expectErr {
				assert.Error(t, err)
				assert.ErrorContains(t, err, "factory returned nil stream")
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, svc.assessmentStream)
			if tt.expectSameStream {
				assert.Same(t, streamBefore, svc.assessmentStream)
			}
		})
	}
}

// TestService_StoreEvidence tests the StoreEvidence method of the Service implementation
func TestService_StoreEvidence(t *testing.T) {
	type args struct {
		ctx context.Context
		req *connect.Request[evidence.StoreEvidenceRequest]
	}
	type fields struct {
		svc *Service
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[*connect.Response[evidence.StoreEvidenceResponse]]
		wantErr assert.WantErr
	}{
		{
			name: "Error - Validation fails",
			args: args{
				req: &connect.Request[evidence.StoreEvidenceRequest]{Msg: &evidence.StoreEvidenceRequest{Evidence: nil}},
			},
			fields: fields{
				svc: nil, // service isn't needed; validating arg errors only
			},
			want: assert.Nil[*connect.Response[evidence.StoreEvidenceResponse]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument)
			},
		},
		{
			name: "Error creating evidence - already exists",
			args: args{
				ctx: context.Background(),
				req: &connect.Request[evidence.StoreEvidenceRequest]{Msg: &evidence.StoreEvidenceRequest{
					Evidence: evidencetest.MockEvidence1,
				}},
			},
			fields: fields{svc: func() *Service {
				svc, err := NewService(WithDB(persistencetest.NewInMemoryDB(t, types, nil, func(db persistence.DB) {
					// Create evidence
					err := db.Create(evidencetest.MockEvidence1)
					assert.NoError(t, err)
				})))
				assert.NoError(t, err)
				return svc
			}()},
			want: assert.Nil[*connect.Response[evidence.StoreEvidenceResponse]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeAlreadyExists)
			},
		},
		{
			name: "Error creating evidence - internal",
			args: args{
				ctx: context.Background(),
				req: &connect.Request[evidence.StoreEvidenceRequest]{Msg: &evidence.StoreEvidenceRequest{
					Evidence: evidencetest.MockEvidence2SameResourceAs1,
				}},
			},
			fields: fields{svc: func() *Service {
				svc, err := NewService(WithDB(persistencetest.CreateErrorDB(t, persistence.ErrDatabase, types, nil)))
				assert.NoError(t, err)
				return svc
			}()},
			want: assert.Nil[*connect.Response[evidence.StoreEvidenceResponse]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInternal)
			},
		},
		{
			name: "Error saving resource - internal",
			args: args{
				ctx: context.Background(),
				req: &connect.Request[evidence.StoreEvidenceRequest]{Msg: &evidence.StoreEvidenceRequest{
					Evidence: evidencetest.MockEvidence2SameResourceAs1,
				}},
			},
			fields: fields{svc: func() *Service {
				svc, err := NewService(WithDB(persistencetest.SaveErrorDB(t, persistence.ErrDatabase, types, nil)))
				assert.NoError(t, err)
				return svc
			}()},
			want: assert.Nil[*connect.Response[evidence.StoreEvidenceResponse]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInternal)
			},
		},
		{
			name: "Happy path - create new resource",
			args: args{
				ctx: context.Background(),
				req: &connect.Request[evidence.StoreEvidenceRequest]{Msg: &evidence.StoreEvidenceRequest{
					Evidence: evidencetest.MockEvidence2SameResourceAs1,
				}},
			},
			fields: fields{svc: func() *Service {
				svc, err := NewService(WithDB(persistencetest.NewInMemoryDB(t, types, nil)))
				assert.NoError(t, err)
				return svc
			}()},
			want:    assert.NotNil[*connect.Response[evidence.StoreEvidenceResponse]],
			wantErr: assert.NoError,
		},
		{
			name: "error - nil resource",
			args: args{
				ctx: context.Background(),
				req: &connect.Request[evidence.StoreEvidenceRequest]{Msg: &evidence.StoreEvidenceRequest{
					Evidence: evidencetest.MockEvidenceNoResource,
				}},
			},
			fields: fields{svc: func() *Service {
				svc, err := NewService(WithDB(persistencetest.NewInMemoryDB(t, types, nil)))
				assert.NoError(t, err)
				return svc
			}()},
			want: assert.Nil[*connect.Response[evidence.StoreEvidenceResponse]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInternal)
			},
		},
		{
			name: "Happy path - updating resource",
			args: args{
				ctx: context.Background(),
				req: &connect.Request[evidence.StoreEvidenceRequest]{Msg: &evidence.StoreEvidenceRequest{
					Evidence: evidencetest.MockEvidence2SameResourceAs1,
				}},
			},
			fields: fields{svc: func() *Service {
				svc, err := NewService(WithDB(persistencetest.NewInMemoryDB(t, types, nil, func(db persistence.DB) {
					// Create a resource already such that `save` will update it instead of creating a new entry
					r, err := evidence.ToEvidenceResource(evidencetest.MockEvidence1.GetOntologyResource(), evidencetest.MockEvidence1.GetTargetOfEvaluationId(), evidencetest.MockEvidence1.GetToolId())
					assert.NoError(t, err)
					err = db.Create(r)
					assert.NoError(t, err)
				})))
				assert.NoError(t, err)
				return svc
			}()},
			want:    assert.NotNil[*connect.Response[evidence.StoreEvidenceResponse]],
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := tt.fields.svc.StoreEvidence(tt.args.ctx, tt.args.req)
			tt.wantErr(t, err)
			tt.want(t, res)
		})
	}
}

// TestService_StoreEvidences tests the streaming StoreEvidences RPC.
// It focuses on happy-path Send/Receive cycles and per-message status handling.
func TestService_StoreEvidences(t *testing.T) {
	type fields struct {
		db persistence.DB
	}
	tests := []struct {
		name         string
		fields       fields
		evidences    []*evidence.Evidence
		wantStatuses assert.Want[[]evidence.EvidenceStatus]
		wantErr      assert.WantErr
	}{
		{
			name: "stream - single evidence",
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, nil),
			},
			evidences: []*evidence.Evidence{
				evidencetest.MockEvidence1,
			},
			wantStatuses: func(t *testing.T, got []evidence.EvidenceStatus, args ...any) bool {
				return assert.Equal(t, []evidence.EvidenceStatus{evidence.EvidenceStatus_EVIDENCE_STATUS_OK}, got)
			},
			wantErr: assert.NoError,
		},
		{
			name: "stream - multiple evidences in sequence",
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, nil),
			},
			evidences: []*evidence.Evidence{
				evidencetest.MockEvidence1,
				evidencetest.MockEvidence2SameResourceAs1,
			},
			wantStatuses: func(t *testing.T, got []evidence.EvidenceStatus, args ...any) bool {
				return assert.Equal(t, []evidence.EvidenceStatus{
					evidence.EvidenceStatus_EVIDENCE_STATUS_OK,
					evidence.EvidenceStatus_EVIDENCE_STATUS_OK,
				}, got)
			},
			wantErr: assert.NoError,
		},
		{
			name: "stream - resilience with partial failures",
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, nil, func(db persistence.DB) {
					assert.NoError(t, db.Create(evidencetest.MockEvidence1))
				}),
			},
			evidences: []*evidence.Evidence{
				evidencetest.MockEvidence1,                // duplicate - should fail
				evidencetest.MockEvidence2SameResourceAs1, // should succeed
			},
			wantStatuses: func(t *testing.T, got []evidence.EvidenceStatus, args ...any) bool {
				return assert.Equal(t, []evidence.EvidenceStatus{
					evidence.EvidenceStatus_EVIDENCE_STATUS_ERROR,
					evidence.EvidenceStatus_EVIDENCE_STATUS_OK,
				}, got)
			},
			wantErr: assert.NoError,
		},
		{
			name: "stream - empty (no messages)",
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, nil),
			},
			evidences: []*evidence.Evidence{},
			wantStatuses: func(t *testing.T, got []evidence.EvidenceStatus, args ...any) bool {
				return len(got) == 0
			},
			wantErr: assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				err      error
				statuses []evidence.EvidenceStatus
			)

			_, _, assessmentSrv := newAssessmentTestServer(t)
			defer assessmentSrv.Close()

			svc, svcErr := NewService(
				WithDB(tt.fields.db),
				WithConfig(Config{
					AssessmentAddress: assessmentSrv.URL,
					AssessmentClient:        assessmentSrv.Client(),
				}),
			)
			assert.NoError(t, svcErr)
			streamHandle := svc.assessmentStream
			defer func() {
				if streamHandle != nil {
					_ = streamHandle.Close()
				}
			}()

			_, storeSrv := servertest.NewTestConnectServer(t,
				server.WithHandler(evidenceconnect.NewEvidenceStoreHandler(svc)),
			)
			defer storeSrv.Close()

			client := evidenceconnect.NewEvidenceStoreClient(storeSrv.Client(), storeSrv.URL)
			stream := client.StoreEvidences(context.Background())

			for _, ev := range tt.evidences {
				sendErr := stream.Send(&evidence.StoreEvidenceRequest{Evidence: ev})
				assert.NoError(t, sendErr)
				if sendErr != nil {
					err = sendErr
					break
				}

				res, recvErr := stream.Receive()
				if recvErr != nil {
					err = recvErr
					break
				}
				statuses = append(statuses, res.Status)
			}

			_ = stream.CloseRequest()

			tt.wantStatuses(t, statuses)
			tt.wantErr(t, err)
		})
	}
}

// TestService_StoreEvidences_ReceiveError isolates transport-level receive failures.
func TestService_StoreEvidences_ReceiveError(t *testing.T) {
	_, _, assessmentSrv := newAssessmentTestServer(t)
	defer assessmentSrv.Close()

	svc, err := NewService(
		WithDB(persistencetest.NewInMemoryDB(t, types, nil)),
		WithConfig(Config{
			AssessmentAddress: assessmentSrv.URL,
			AssessmentClient:        assessmentSrv.Client(),
		}),
	)
	assert.NoError(t, err)
	streamHandle := svc.assessmentStream
	defer func() {
		if streamHandle != nil {
			_ = streamHandle.Close()
		}
	}()

	_, storeSrv := servertest.NewTestConnectServer(t,
		server.WithHandler(evidenceconnect.NewEvidenceStoreHandler(svc)),
	)
	defer storeSrv.Close()

	client := evidenceconnect.NewEvidenceStoreClient(storeSrv.Client(), storeSrv.URL)
	stream := client.StoreEvidences(context.Background())
	assert.NoError(t, stream.Send(&evidence.StoreEvidenceRequest{Evidence: evidencetest.MockEvidence1}))
	_, recvErr := stream.Receive()
	assert.NoError(t, recvErr)

	storeSrv.CloseClientConnections()
	_, recvErr = stream.Receive()
	assert.Error(t, recvErr)
}

// TestService_StoreEvidences_SendErrors uses a fake stream to deterministically trigger send EOF/error paths.
func TestService_StoreEvidences_SendErrors(t *testing.T) {
	type fields struct {
		db persistence.DB
	}
	tests := []struct {
		name    string
		fields  fields
		sendErr error
		wantErr assert.WantErr
	}{
		{
			name: "send EOF returns nil",
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, nil),
			},
			sendErr: io.EOF,
			wantErr: assert.NoError,
		},
		{
			name: "send error returns CodeUnknown",
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, nil),
			},
			sendErr: errors.New("send failed"),
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeUnknown)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{
				db:              tt.fields.db,
				channelEvidence: make(chan *evidence.Evidence, 1),
			}

			stream := &fakeEvidenceStream{
				receives: []fakeReceive{{req: &evidence.StoreEvidenceRequest{Evidence: evidencetest.MockEvidence1}}},
				sendErr:  tt.sendErr,
			}

			err := svc.storeEvidencesStream(context.Background(), stream)
			tt.wantErr(t, err)
		})
	}
}

// TestService_ListEvidences uses table tests to cover filters and pagination behaviors.
func TestService_ListEvidences(t *testing.T) {
	ev1 := evidencetest.MockEvidenceListA
	ev2 := evidencetest.MockEvidenceListB
	ev3 := evidencetest.MockEvidenceListC

	type fields struct {
		db persistence.DB
	}
	tests := []struct {
		name        string
		fields      fields
		req         *connect.Request[evidence.ListEvidencesRequest]
		want        assert.Want[*connect.Response[evidence.ListEvidencesResponse]]
		wantErr     assert.WantErr
		nextReq     func(res *connect.Response[evidence.ListEvidencesResponse]) *connect.Request[evidence.ListEvidencesRequest]
		wantNext    assert.Want[*connect.Response[evidence.ListEvidencesResponse]]
		wantNextErr assert.WantErr
	}{
		{
			name:   "error - nil request",
			fields: fields{db: persistencetest.NewInMemoryDB(t, types, nil)},
			req:    nil,
			want:   assert.Nil[*connect.Response[evidence.ListEvidencesResponse]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument)
			},
		},
		{
			name:   "error - list failure",
			fields: fields{db: persistencetest.ListErrorDB(t, errors.New("list failed"), types, nil)},
			req:    &connect.Request[evidence.ListEvidencesRequest]{Msg: &evidence.ListEvidencesRequest{}},
			want:   assert.Nil[*connect.Response[evidence.ListEvidencesResponse]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInternal)
			},
		},
		{
			name: "happy path - no filter",
			fields: fields{db: persistencetest.NewInMemoryDB(t, types, nil, func(db persistence.DB) {
				assert.NoError(t, db.Create(ev1))
				assert.NoError(t, db.Create(ev2))
				assert.NoError(t, db.Create(ev3))
			})},
			req: &connect.Request[evidence.ListEvidencesRequest]{Msg: &evidence.ListEvidencesRequest{}},
			want: func(t *testing.T, got *connect.Response[evidence.ListEvidencesResponse], msgAndArgs ...any) bool {
				assert.NotNil(t, got)
				if !assert.Equal(t, 3, len(got.Msg.Evidences)) {
					return false
				}
				ids := make([]string, 0, len(got.Msg.Evidences))
				for _, ev := range got.Msg.Evidences {
					ids = append(ids, ev.Id)
				}
				assert.Contains(t, ids, ev1.Id)
				assert.Contains(t, ids, ev2.Id)
				assert.Contains(t, ids, ev3.Id)
				return true
			},
			wantErr: assert.NoError,
		},
		{
			name: "happy path - filter by target of evaluation",
			fields: fields{db: persistencetest.NewInMemoryDB(t, types, nil, func(db persistence.DB) {
				assert.NoError(t, db.Create(ev1))
				assert.NoError(t, db.Create(ev2))
				assert.NoError(t, db.Create(ev3))
			})},
			req: &connect.Request[evidence.ListEvidencesRequest]{Msg: &evidence.ListEvidencesRequest{
				Filter: &evidence.Filter{TargetOfEvaluationId: util.Ref(ev1.TargetOfEvaluationId)},
			}},
			want: func(t *testing.T, got *connect.Response[evidence.ListEvidencesResponse], msgAndArgs ...any) bool {
				assert.NotNil(t, got)
				if !assert.Equal(t, 2, len(got.Msg.Evidences)) {
					return false
				}
				ids := make([]string, 0, len(got.Msg.Evidences))
				for _, ev := range got.Msg.Evidences {
					ids = append(ids, ev.Id)
				}
				assert.Contains(t, ids, ev1.Id)
				assert.Contains(t, ids, ev3.Id)
				return true
			},
			wantErr: assert.NoError,
		},
		{
			name: "happy path - filter by tool",
			fields: fields{db: persistencetest.NewInMemoryDB(t, types, nil, func(db persistence.DB) {
				assert.NoError(t, db.Create(ev1))
				assert.NoError(t, db.Create(ev2))
				assert.NoError(t, db.Create(ev3))
			})},
			req: &connect.Request[evidence.ListEvidencesRequest]{Msg: &evidence.ListEvidencesRequest{
				Filter: &evidence.Filter{ToolId: util.Ref(ev1.ToolId)},
			}},
			want: func(t *testing.T, got *connect.Response[evidence.ListEvidencesResponse], msgAndArgs ...any) bool {
				assert.NotNil(t, got)
				if !assert.Equal(t, 2, len(got.Msg.Evidences)) {
					return false
				}
				ids := make([]string, 0, len(got.Msg.Evidences))
				for _, ev := range got.Msg.Evidences {
					ids = append(ids, ev.Id)
				}
				assert.Contains(t, ids, ev1.Id)
				assert.Contains(t, ids, ev2.Id)
				return true
			},
			wantErr: assert.NoError,
		},
		{
			name: "happy path - pagination",
			fields: fields{db: persistencetest.NewInMemoryDB(t, types, nil, func(db persistence.DB) {
				assert.NoError(t, db.Create(ev1))
				assert.NoError(t, db.Create(ev2))
				assert.NoError(t, db.Create(ev3))
			})},
			req: &connect.Request[evidence.ListEvidencesRequest]{Msg: &evidence.ListEvidencesRequest{
				PageSize: 1,
				OrderBy:  "id",
				Asc:      true,
			}},
			want: func(t *testing.T, got *connect.Response[evidence.ListEvidencesResponse], msgAndArgs ...any) bool {
				assert.NotNil(t, got)
				assert.Equal(t, 1, len(got.Msg.Evidences))
				return assert.NotEmpty(t, got.Msg.NextPageToken)
			},
			wantErr: assert.NoError,
			nextReq: func(res *connect.Response[evidence.ListEvidencesResponse]) *connect.Request[evidence.ListEvidencesRequest] {
				return &connect.Request[evidence.ListEvidencesRequest]{Msg: &evidence.ListEvidencesRequest{
					PageSize:  1,
					OrderBy:   "id",
					Asc:       true,
					PageToken: res.Msg.NextPageToken,
				}}
			},
			wantNext: func(t *testing.T, got *connect.Response[evidence.ListEvidencesResponse], msgAndArgs ...any) bool {
				firstID, _ := msgAndArgs[0].(string)
				assert.NotNil(t, got)
				if !assert.Equal(t, 1, len(got.Msg.Evidences)) {
					return false
				}
				return assert.NotEqual(t, firstID, got.Msg.Evidences[0].Id)
			},
			wantNextErr: assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{db: tt.fields.db}

			res, err := svc.ListEvidences(context.Background(), tt.req)
			tt.wantErr(t, err)
			tt.want(t, res)

			if tt.nextReq != nil {
				firstID := ""
				if res != nil && len(res.Msg.Evidences) > 0 {
					firstID = res.Msg.Evidences[0].Id
				}
				nextRes, nextErr := svc.ListEvidences(context.Background(), tt.nextReq(res))
				tt.wantNextErr(t, nextErr)
				tt.wantNext(t, nextRes, firstID)
			}
		})
	}
}

// TestService_GetEvidence uses table tests to cover validation, not found, and success paths.
func TestService_GetEvidence(t *testing.T) {
	type fields struct {
		db persistence.DB
	}
	tests := []struct {
		name    string
		fields  fields
		req     *connect.Request[evidence.GetEvidenceRequest]
		want    assert.Want[*connect.Response[evidence.Evidence]]
		wantErr assert.WantErr
	}{
		{
			name:   "error - nil request",
			fields: fields{db: persistencetest.NewInMemoryDB(t, types, nil)},
			req:    &connect.Request[evidence.GetEvidenceRequest]{Msg: nil},
			want:   assert.Nil[*connect.Response[evidence.Evidence]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument)
			},
		},
		{
			name:   "error - invalid evidence id",
			fields: fields{db: persistencetest.NewInMemoryDB(t, types, nil)},
			req: &connect.Request[evidence.GetEvidenceRequest]{Msg: &evidence.GetEvidenceRequest{
				EvidenceId: "not-a-uuid",
			}},
			want: assert.Nil[*connect.Response[evidence.Evidence]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument)
			},
		},
		{
			name:   "error - not found",
			fields: fields{db: persistencetest.NewInMemoryDB(t, types, nil)},
			req: &connect.Request[evidence.GetEvidenceRequest]{Msg: &evidence.GetEvidenceRequest{
				EvidenceId: "00000000-0000-0000-0000-000000000000",
			}},
			want: assert.Nil[*connect.Response[evidence.Evidence]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeNotFound)
			},
		},
		{
			name:   "error - database failure",
			fields: fields{db: persistencetest.GetErrorDB(t, errors.New("get failed"), types, nil)},
			req: &connect.Request[evidence.GetEvidenceRequest]{Msg: &evidence.GetEvidenceRequest{
				EvidenceId: evidencetest.MockEvidence1.Id,
			}},
			want: assert.Nil[*connect.Response[evidence.Evidence]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInternal)
			},
		},
		{
			name: "happy path - returns evidence",
			fields: fields{db: persistencetest.NewInMemoryDB(t, types, nil, func(db persistence.DB) {
				assert.NoError(t, db.Create(evidencetest.MockEvidence1))
			})},
			req: &connect.Request[evidence.GetEvidenceRequest]{Msg: &evidence.GetEvidenceRequest{
				EvidenceId: evidencetest.MockEvidence1.Id,
			}},
			want: func(t *testing.T, got *connect.Response[evidence.Evidence], msgAndArgs ...any) bool {
				assert.NotNil(t, got)
				return assert.Equal(t, evidencetest.MockEvidence1.Id, got.Msg.Id)
			},
			wantErr: assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{db: tt.fields.db}
			res, err := svc.GetEvidence(context.Background(), tt.req)
			tt.wantErr(t, err)
			tt.want(t, res)
		})
	}
}

// TestService_ListSupportedResourceTypes covers validation and happy-path responses.
func TestService_ListSupportedResourceTypes(t *testing.T) {
	tests := []struct {
		name    string
		req     *connect.Request[evidence.ListSupportedResourceTypesRequest]
		want    assert.Want[*connect.Response[evidence.ListSupportedResourceTypesResponse]]
		wantErr assert.WantErr
	}{
		{
			name: "error - nil request",
			req:  nil,
			want: assert.Nil[*connect.Response[evidence.ListSupportedResourceTypesResponse]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument)
			},
		},
		{
			name: "happy path - returns resource types",
			req:  &connect.Request[evidence.ListSupportedResourceTypesRequest]{Msg: &evidence.ListSupportedResourceTypesRequest{}},
			want: func(t *testing.T, got *connect.Response[evidence.ListSupportedResourceTypesResponse], msgAndArgs ...any) bool {
				assert.NotNil(t, got)
				return assert.Equal(t, ontology.ListResourceTypes(), got.Msg.ResourceType)
			},
			wantErr: assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{}
			res, err := svc.ListSupportedResourceTypes(context.Background(), tt.req)
			tt.wantErr(t, err)
			tt.want(t, res)
		})
	}
}

// TestService_ListResources uses table tests to cover filters, pagination, and error handling.
func TestService_ListResources(t *testing.T) {
	res1 := evidencetest.MockResourceListA
	res2 := evidencetest.MockResourceListB
	res3 := evidencetest.MockResourceListC

	type fields struct {
		db persistence.DB
	}
	type args struct {
		ctx context.Context
		req *connect.Request[evidence.ListResourcesRequest]
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		wantRes     assert.Want[*connect.Response[evidence.ListResourcesResponse]]
		wantErr     assert.WantErr
		nextReq     func(res *connect.Response[evidence.ListResourcesResponse]) *connect.Request[evidence.ListResourcesRequest]
		wantNext    assert.Want[*connect.Response[evidence.ListResourcesResponse]]
		wantNextErr assert.WantErr
	}{
		{
			name:   "error - nil request",
			fields: fields{db: persistencetest.NewInMemoryDB(t, types, nil)},
			args: args{
				ctx: context.Background(),
				req: nil,
			},
			wantRes: assert.Nil[*connect.Response[evidence.ListResourcesResponse]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument)
			},
		},
		{
			name:   "error - list failure",
			fields: fields{db: persistencetest.ListErrorDB(t, errors.New("list failed"), types, nil)},
			args: args{
				ctx: context.Background(),
				req: &connect.Request[evidence.ListResourcesRequest]{Msg: &evidence.ListResourcesRequest{}},
			},
			wantRes: assert.Nil[*connect.Response[evidence.ListResourcesResponse]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInternal)
			},
		},
		{
			name: "happy path - no filter",
			fields: fields{db: persistencetest.NewInMemoryDB(t, types, nil, func(db persistence.DB) {
				assert.NoError(t, db.Create(res1))
				assert.NoError(t, db.Create(res2))
				assert.NoError(t, db.Create(res3))
			})},
			args: args{
				ctx: context.Background(),
				req: &connect.Request[evidence.ListResourcesRequest]{Msg: &evidence.ListResourcesRequest{}},
			},
			wantRes: func(t *testing.T, got *connect.Response[evidence.ListResourcesResponse], msgAndArgs ...any) bool {
				assert.NotNil(t, got)
				return assert.Equal(t, 3, len(got.Msg.Results))
			},
			wantErr: assert.NoError,
		},
		{
			name: "happy path - filter by target of evaluation",
			fields: fields{db: persistencetest.NewInMemoryDB(t, types, nil, func(db persistence.DB) {
				assert.NoError(t, db.Create(res1))
				assert.NoError(t, db.Create(res2))
				assert.NoError(t, db.Create(res3))
			})},
			args: args{
				ctx: context.Background(),
				req: &connect.Request[evidence.ListResourcesRequest]{Msg: &evidence.ListResourcesRequest{
					Filter: &evidence.ListResourcesRequest_Filter{TargetOfEvaluationId: util.Ref(res1.TargetOfEvaluationId)},
				}},
			},
			wantRes: func(t *testing.T, got *connect.Response[evidence.ListResourcesResponse], msgAndArgs ...any) bool {
				assert.NotNil(t, got)
				if !assert.Equal(t, 2, len(got.Msg.Results)) {
					return false
				}
				ids := []string{got.Msg.Results[0].Id, got.Msg.Results[1].Id}
				assert.Contains(t, ids, res1.Id)
				assert.Contains(t, ids, res3.Id)
				return true
			},
			wantErr: assert.NoError,
		},
		{
			name: "happy path - filter by tool",
			fields: fields{db: persistencetest.NewInMemoryDB(t, types, nil, func(db persistence.DB) {
				assert.NoError(t, db.Create(res1))
				assert.NoError(t, db.Create(res2))
				assert.NoError(t, db.Create(res3))
			})},
			args: args{
				ctx: context.Background(),
				req: &connect.Request[evidence.ListResourcesRequest]{Msg: &evidence.ListResourcesRequest{
					Filter: &evidence.ListResourcesRequest_Filter{ToolId: util.Ref(res1.ToolId)},
				}},
			},
			wantRes: func(t *testing.T, got *connect.Response[evidence.ListResourcesResponse], msgAndArgs ...any) bool {
				assert.NotNil(t, got)
				if !assert.Equal(t, 2, len(got.Msg.Results)) {
					return false
				}
				ids := []string{got.Msg.Results[0].Id, got.Msg.Results[1].Id}
				assert.Contains(t, ids, res1.Id)
				assert.Contains(t, ids, res2.Id)
				return true
			},
			wantErr: assert.NoError,
		},
		{
			name: "happy path - filter by type (Currently fails due to ramsql LIKE limitation)",
			fields: fields{db: persistencetest.NewInMemoryDB(t, types, nil, func(db persistence.DB) {
				assert.NoError(t, db.Create(res1))
				assert.NoError(t, db.Create(res2))
				assert.NoError(t, db.Create(res3))
			})},
			args: args{
				ctx: context.Background(),
				req: &connect.Request[evidence.ListResourcesRequest]{Msg: &evidence.ListResourcesRequest{
					Filter: &evidence.ListResourcesRequest_Filter{Type: util.Ref(res1.ResourceType)},
				}},
			},
			wantRes: func(t *testing.T, got *connect.Response[evidence.ListResourcesResponse], msgAndArgs ...any) bool {
				assert.NotNil(t, got)
				if !assert.Equal(t, 2, len(got.Msg.Results)) {
					return false
				}
				ids := []string{got.Msg.Results[0].Id, got.Msg.Results[1].Id}
				assert.Contains(t, ids, res1.Id)
				assert.Contains(t, ids, res3.Id)
				return true
			},
			wantErr: assert.NoError,
		},
		{
			name: "happy path - pagination",
			fields: fields{db: persistencetest.NewInMemoryDB(t, types, nil, func(db persistence.DB) {
				assert.NoError(t, db.Create(res1))
				assert.NoError(t, db.Create(res2))
				assert.NoError(t, db.Create(res3))
			})},
			args: args{
				ctx: context.Background(),
				req: &connect.Request[evidence.ListResourcesRequest]{Msg: &evidence.ListResourcesRequest{
					PageSize: 1,
					OrderBy:  "id",
					Asc:      true,
				}},
			},
			wantRes: func(t *testing.T, got *connect.Response[evidence.ListResourcesResponse], msgAndArgs ...any) bool {
				assert.NotNil(t, got)
				assert.Equal(t, 1, len(got.Msg.Results))
				return assert.NotEmpty(t, got.Msg.NextPageToken)
			},
			wantErr: assert.NoError,
			nextReq: func(res *connect.Response[evidence.ListResourcesResponse]) *connect.Request[evidence.ListResourcesRequest] {
				return &connect.Request[evidence.ListResourcesRequest]{Msg: &evidence.ListResourcesRequest{
					PageSize:  1,
					OrderBy:   "id",
					Asc:       true,
					PageToken: res.Msg.NextPageToken,
				}}
			},
			wantNext: func(t *testing.T, got *connect.Response[evidence.ListResourcesResponse], msgAndArgs ...any) bool {
				firstID, _ := msgAndArgs[0].(string)
				assert.NotNil(t, got)
				if !assert.Equal(t, 1, len(got.Msg.Results)) {
					return false
				}
				return assert.NotEqual(t, firstID, got.Msg.Results[0].Id)
			},
			wantNextErr: assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{db: tt.fields.db}

			res, err := svc.ListResources(tt.args.ctx, tt.args.req)
			tt.wantErr(t, err)
			tt.wantRes(t, res)

			if tt.nextReq != nil {
				firstID := ""
				if res != nil && len(res.Msg.Results) > 0 {
					firstID = res.Msg.Results[0].Id
				}
				nextRes, nextErr := svc.ListResources(context.Background(), tt.nextReq(res))
				tt.wantNextErr(t, nextErr)
				tt.wantNext(t, nextRes, firstID)
			}
		})
	}
}

// TestService_RegisterEvidenceHook verifies hook registration.
func TestService_RegisterEvidenceHook(t *testing.T) {
	svc := &Service{}
	assert.Equal(t, 0, len(svc.evidenceHooks))

	svc.RegisterEvidenceHook(func(context.Context, *evidence.Evidence, error) {})
	assert.Equal(t, 1, len(svc.evidenceHooks))
}

// TestService_informHooks verifies all registered hooks are invoked.
func TestService_informHooks(t *testing.T) {
	svc := &Service{}
	count := 0

	svc.RegisterEvidenceHook(func(context.Context, *evidence.Evidence, error) {
		count++
	})
	svc.RegisterEvidenceHook(func(context.Context, *evidence.Evidence, error) {
		count++
	})

	svc.informHooks(context.Background(), evidencetest.MockEvidence1, nil)
	assert.Equal(t, 2, count)
}

func TestService_initEvidenceChannel(t *testing.T) {
	assessmentRecorder, _, testSrv := newAssessmentTestServer(t)
	defer testSrv.Close()

	svc, err := NewService(
		WithDB(persistencetest.NewInMemoryDB(t, types, nil)),
		WithConfig(Config{
			AssessmentAddress: testSrv.URL,
			AssessmentClient:        testSrv.Client(),
		}),
	)
	assert.NoError(t, err)
	// Keep a handle for shutdown even if svc.assessmentStream is set to nil later.
	stream := svc.assessmentStream
	defer func() {
		if stream != nil {
			_ = stream.Close()
		}
	}()

	// Nil evidence should be ignored.
	svc.channelEvidence <- nil

	ev := &evidence.Evidence{Id: uuid.NewString()}
	svc.channelEvidence <- ev
	awaitAssessmentRequest(t, assessmentRecorder.received, ev.Id)

	// Error path: no assessment stream available.
	svc.assessmentStream = nil
	svc.channelEvidence <- &evidence.Evidence{Id: uuid.NewString()}
	close(svc.channelEvidence)
	select {
	case <-assessmentRecorder.received:
		assert.Fail(t, "unexpected assessment request after stream was removed")
	case <-time.After(200 * time.Millisecond):
	}
}
