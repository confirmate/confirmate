package evidence

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"

	"confirmate.io/core/api/assessment/assessmentconnect"
	"confirmate.io/core/api/evidence"
	"confirmate.io/core/internal/testutil/servicetest/evidencetest"
	"confirmate.io/core/persistence"
	"confirmate.io/core/persistence/persistencetest"
	"confirmate.io/core/server"
	"confirmate.io/core/server/servertest"
	"confirmate.io/core/service"
	"confirmate.io/core/util/assert"
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

// TODO(lebogg): Continue. Fix test. Check Stream stuff again (also other PRs)
func TestService_handleEvidence(t *testing.T) {
	// Create Assessment Service + Server
	assessmentService := assessmentconnect.UnimplementedAssessmentHandler{}
	srv, testSrv := servertest.NewTestConnectServer(
		t,
		server.WithHandler(assessmentconnect.NewAssessmentHandler(assessmentService)),
	)
	defer testSrv.Close()
	assert.NotNil(t, srv)
	assert.NotNil(t, testSrv)

	// Create Evidence Service
	svc, err := NewService(
		WithDB(persistencetest.NewInMemoryDB(t, types, nil)),
		WithAssessmentConfig(assessmentConfig{
			targetAddress: testSrv.URL,
			client:        testSrv.Client(),
		}))
	assert.NoError(t, err)

	// handle Evidence (pass)
	e := &evidence.Evidence{
		Id: uuid.NewString(),
	}
	err = svc.handleEvidence(e,
		1)
	assert.NoError(t, err)
	slog.Info("Sent evidence", slog.Any("id", e.Id))

	// handle another Evidence (pass)
	e = &evidence.Evidence{
		Id: uuid.NewString(),
	}
	err = svc.handleEvidence(e,
		1)
	assert.NoError(t, err)
	slog.Info("Sent evidence", slog.Any("id", e.Id))

	// Break up stream from the assessment side
	testSrv.Close()
	go func() {
		// Restart server. In production, we will have a fixed URL but here we have to adapt to the test server
		time.Sleep(15 * time.Second)
		srv, testSrv = servertest.NewTestConnectServer(
			t,
			server.WithHandler(assessmentconnect.NewAssessmentHandler(assessmentService)),
		)
		// Since we have new server, we need to update the config
		svc.assessmentConfig = assessmentConfig{
			targetAddress: testSrv.URL,
			client:        testSrv.Client(),
		}
		svc.assessmentClient = assessmentconnect.NewAssessmentClient(
			svc.assessmentConfig.client, svc.assessmentConfig.targetAddress)
	}()

	// handle another Evidence (automatically recreate stream, pass)
	e = &evidence.Evidence{
		Id: uuid.NewString(),
	}
	err = svc.handleEvidence(e,
		1)
	assert.NoError(t, err)
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
