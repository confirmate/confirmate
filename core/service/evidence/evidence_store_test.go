package evidence

import (
	"context"
	"os"
	"testing"

	"confirmate.io/core/api/evidence"
	"confirmate.io/core/internal/testutil/servicetest/evidencetest"
	"confirmate.io/core/persistence/persistencetest"
	"confirmate.io/core/service"
	"confirmate.io/core/util/assert"
	"connectrpc.com/connect"
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
		name string
		args args
		want assert.Want[*Service]
	}{
		{
			name: "EvidenceStoreServer created without options",
			want: func(t *testing.T, got *Service) bool {
				// Storage should be default (in-memory storage). Hard to check since its type is not exported
				assert.NotNil(t, got.db)
				return true
			},
		},
		{
			name: "EvidenceStoreServer created with option 'WithDB'",
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
		},
		//{
		//	name: "EvidenceStoreServer created with option 'WithAssessmentConfig'",
		//	args: args{opts: []service.Option[*Service]{WithAssessmentConfig("localhost:9091")}},
		//	want: func(t *testing.T, got *Service) bool {
		//		return assert.Equal(t, "localhost:9091", got.assessment.Target)
		//	},
		//},
		//{
		//	name: "EvidenceStoreServer created with option 'WithOAuth2Authorizer'",
		//	args: args{opts: []service.Option[*Service]{WithOAuth2Authorizer(&clientcredentials.Config{ClientID: "client"})}},
		//	want: func(t *testing.T, got *Service) bool {
		//		return assert.NotNil(t, got.assessment.Authorizer())
		//	},
		//},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewService(tt.args.opts...)
			assert.Nil(t, err)
			tt.want(t, got)
		})
	}
}
