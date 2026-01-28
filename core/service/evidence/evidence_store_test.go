package evidence

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"

	"confirmate.io/core/api/evidence"
	"confirmate.io/core/api/evidence/evidenceconnect"
	"confirmate.io/core/internal/testutil/servicetest/evidencetest"
	"confirmate.io/core/persistence"
	"confirmate.io/core/persistence/persistencetest"
	"confirmate.io/core/server"
	"confirmate.io/core/server/servertest"
	"confirmate.io/core/service"
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
			name: "EvidenceStoreServer created with option 'WithAssessmentConfig' - no client provided",
			args: args{opts: []service.Option[Service]{
				WithDB(persistencetest.NewInMemoryDB(t, types, nil)),
				WithAssessmentConfig(assessmentConfig{
					targetAddress: "localhost:9091",
					client:        nil,
				}),
			}},
			want: func(t *testing.T, got *Service, msgAndArgs ...any) bool {
				// We didn't provide a client, so it should be the default (timeout is zero value)
				assert.Equal(t, 0, got.assessmentConfig.client.Timeout)
				assert.NotNil(t, got.assessmentStream)
				return assert.Equal(t, "localhost:9091", got.assessmentConfig.targetAddress)
			},
			wantErr: assert.NoError,
		},
		{
			name: "EvidenceStoreServer created with option 'WithAssessmentConfig' - with client",
			args: args{opts: []service.Option[Service]{
				WithDB(persistencetest.NewInMemoryDB(t, types, nil)),
				WithAssessmentConfig(assessmentConfig{
					targetAddress: "localhost:9091",
					client:        &http.Client{Timeout: time.Duration(1)},
				}),
			}},
			want: func(t *testing.T, got *Service, msgAndArgs ...any) bool {
				// We provided a client with timeout set to 1 second
				assert.Equal(t, 1, got.assessmentConfig.client.Timeout)
				assert.NotNil(t, got.assessmentStream)
				return assert.Equal(t, "localhost:9091", got.assessmentConfig.targetAddress)
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
		WithAssessmentConfig(assessmentConfig{
			targetAddress: testSrv.URL,
			client:        testSrv.Client(),
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
		WithAssessmentConfig(assessmentConfig{
			targetAddress: testSrv.URL,
			client:        testSrv.Client(),
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
// This focuses on Send/Receive cycles and per-message status handling.
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
				WithAssessmentConfig(assessmentConfig{
					targetAddress: assessmentSrv.URL,
					client:        assessmentSrv.Client(),
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

func TestService_initEvidenceChannel(t *testing.T) {
	assessmentRecorder, _, testSrv := newAssessmentTestServer(t)
	defer testSrv.Close()

	svc, err := NewService(
		WithDB(persistencetest.NewInMemoryDB(t, types, nil)),
		WithAssessmentConfig(assessmentConfig{
			targetAddress: testSrv.URL,
			client:        testSrv.Client(),
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
