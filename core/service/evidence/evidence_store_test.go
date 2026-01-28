package evidence

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"confirmate.io/core/api/assessment"
	"confirmate.io/core/api/assessment/assessmentconnect"
	"confirmate.io/core/api/evidence"
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
// - Error when a new DB has to be created (because of the way the DB is initialized in the evidence service)
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
				return assert.Equal(t, "localhost:9091", got.assessmentConfig.targetAddress)
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewService(tt.args.opts...)
			assert.Nil(t, err)
			tt.want(t, got)
			tt.wantErr(t, err)
		})
	}
}

type assessmentStreamRecorder struct {
	assessmentconnect.UnimplementedAssessmentHandler
	received chan *assessment.AssessEvidenceRequest
}

func (r *assessmentStreamRecorder) AssessEvidences(
	_ context.Context,
	stream *connect.BidiStream[assessment.AssessEvidenceRequest, assessment.AssessEvidencesResponse],
) error {
	for {
		msg, err := stream.Receive()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		r.received <- msg
		if err = stream.Send(&assessment.AssessEvidencesResponse{
			Status: assessment.AssessmentStatus_ASSESSMENT_STATUS_ASSESSED,
		}); err != nil {
			return err
		}
	}
}

func newAssessmentTestServer(t *testing.T) (*assessmentStreamRecorder, *server.Server, *httptest.Server) {
	t.Helper()
	recorder := &assessmentStreamRecorder{
		received: make(chan *assessment.AssessEvidenceRequest, 10),
	}
	srv, testSrv := servertest.NewTestConnectServer(
		t,
		server.WithHandler(assessmentconnect.NewAssessmentHandler(recorder)),
	)
	return recorder, srv, testSrv
}

func awaitAssessmentRequest(t *testing.T, ch <-chan *assessment.AssessEvidenceRequest, wantID string) {
	t.Helper()
	select {
	case msg := <-ch:
		assert.Equal(t, wantID, msg.GetEvidence().GetId())
	case <-time.After(2 * time.Second):
		assert.Fail(t, "timed out waiting for assessment request")
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

// TestService_initAssessmentStream verifies creation and idempotency.
func TestService_initAssessmentStream(t *testing.T) {
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

	// Force re-init for coverage of the creation path.
	svc.assessmentStream = nil
	err = svc.initAssessmentStream()
	assert.NoError(t, err)
	assert.NotNil(t, svc.assessmentStream)

	streamBefore := svc.assessmentStream
	err = svc.initAssessmentStream()
	assert.NoError(t, err)
	assert.Same(t, streamBefore, svc.assessmentStream)
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
