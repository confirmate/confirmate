package evidence

import (
	"testing"

	"confirmate.io/core/service"
	"confirmate.io/core/util/assert"
)

func TestMain(m *testing.M) {
	// Start the Evidence Store server

	// Start the Assessment server
}

// TestNewService is a simply test for NewService
func TestNewService(t *testing.T) {
	//db, err := gorm.NewStorage(gorm.WithInMemory())
	//assert.NoError(t, err)

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
		//{
		//	name: "EvidenceStoreServer created with option 'WithDB'",
		//	args: args{opts: []service.Option[*Service]{WithDB(db)}},
		//	want: func(t *testing.T, got *Service) bool {
		//		// Storage should be gorm (in-memory storage). Hard to check since its type is not exported
		//		assert.NotNil(t, got.db)
		//		return true
		//	},
		//},
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
