package evidence

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"

	"confirmate.io/core/api/assessment/assessmentconnect"
	"confirmate.io/core/api/evidence"
	"confirmate.io/core/internal/testutil/servicetest/evidencetest"
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
func TestNewService(t *testing.T) {
	type args struct {
		opts []service.Option[*Service]
	}
	tests := []struct {
		name    string
		args    args
		want    assert.Want[*Service]
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "EvidenceStoreServer created without options",
			want: func(t *testing.T, got *Service) bool {
				// Storage should be default (in-memory storage). Hard to check since its type is not exported
				assert.NotNil(t, got.db)
				return true
			},
			wantErr: assert.NoError,
		},
		{
			name: "Happy path: EvidenceStoreServer created with option 'WithDB'",
			args: args{opts: []service.Option[*Service]{
				WithDB(persistencetest.NewInMemoryDB(t, types, nil, evidencetest.InitDBWithEvidence))}},
			want: func(t *testing.T, got *Service) bool {
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
		// TODO(lebogg): Currently, forcing error on DB creation is not possible because of the way the DB is initialized in the evidence service.
		//{
		//	name: "Error: EvidenceStoreServer created with option 'WithDB'",
		//	args: args{opts: []service.Option[*Service]{}},
		//	want: assert.Nil[*Service],
		//	wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
		//		return assert.ErrorContains(t, err, "could not create db")
		//	},
		//},
		{
			name: "EvidenceStoreServer created with option 'WithAssessmentConfig' - no client provided",
			args: args{opts: []service.Option[*Service]{
				WithAssessmentConfig(assessmentConfig{
					targetAddress: "localhost:9091",
					client:        nil,
				})}},
			want: func(t *testing.T, got *Service) bool {
				// We didn't provide a client, so it should be the default (timeout is zero value)
				assert.Equal(t, 0, got.assessmentConfig.client.Timeout)
				return assert.Equal(t, "localhost:9091", got.assessmentConfig.targetAddress)
			},
			wantErr: assert.NoError,
		},
		{
			name: "EvidenceStoreServer created with option 'WithAssessmentConfig' - with client",
			args: args{opts: []service.Option[*Service]{
				WithAssessmentConfig(assessmentConfig{
					targetAddress: "localhost:9091",
					client:        &http.Client{Timeout: time.Duration(1)},
				})}},
			want: func(t *testing.T, got *Service) bool {
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
	svc, err := NewService(WithAssessmentConfig(assessmentConfig{
		targetAddress: testSrv.URL,
		client:        testSrv.Client(),
	}))
	assert.NoError(t, err)

	// handle Evidence (pass)
	err = svc.handleEvidence(&evidence.Evidence{
		Id: uuid.NewString(),
	},
		0)
	assert.NoError(t, err)

	// handle another Evidence (pass)
	err = svc.handleEvidence(&evidence.Evidence{
		Id: uuid.NewString(),
	},
		0)
	assert.NoError(t, err)

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
	err = svc.handleEvidence(&evidence.Evidence{
		Id: uuid.NewString(),
	},
		0)
	assert.NoError(t, err)
}
