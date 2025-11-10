package cloud

import (
	"context"
	"fmt"
	"io"
	"sync"
	"testing"
	"time"

	"confirmate.io/collectors/cloud/internal/config"
	"confirmate.io/core/api/evidence"
	"confirmate.io/core/api/evidence/evidenceconnect"
	"confirmate.io/core/util/assert"
	"connectrpc.com/connect"
	"github.com/go-co-op/gocron"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
)

// func TestNewService(t *testing.T) {
// 	type args struct {
// 		opts []service.Option[*Service]
// 	}
// 	tests := []struct {
// 		name string
// 		args args
// 		want assert.Want[*Service]
// 	}{
// 		{
// 			name: "Create service with option 'WithEvidenceStoreAddress'",
// 			args: args{
// 				opts: []service.Option[*Service]{
// 					WithEvidenceStoreAddress("localhost:9091"),
// 				},
// 			},
// 			want: func(t *testing.T, got *Service) bool {
// 				return assert.Equal(t, "localhost:9091", got.evidenceStore.Target)
// 			},
// 		},
// 		{
// 			name: "Create service with option 'WithDefaultTargetOfEvaluationID'",
// 			args: args{
// 				opts: []service.Option[*Service]{
// 					WithTargetOfEvaluationID(testdata.MockTargetOfEvaluationID1),
// 				},
// 			},
// 			want: func(t *testing.T, got *Service) bool {
// 				return assert.Equal(t, testdata.MockTargetOfEvaluationID1, got.ctID)
// 			},
// 		},
// 		{
// 			name: "Create service with option 'WithCollectorToolID'",
// 			args: args{
// 				opts: []service.Option[*Service]{
// 					WithEvidenceCollectorToolID(testdata.MockEvidenceToolID1),
// 				},
// 			},
// 			want: func(t *testing.T, got *Service) bool {
// 				return assert.Equal(t, testdata.MockEvidenceToolID1, got.collectorID)
// 			},
// 		},
// 		{
// 			name: "Create service with option 'WithAuthorizationStrategy'",
// 			args: args{
// 				opts: []service.Option[*Service]{
// 					WithAuthorizationStrategy(&service.AuthorizationStrategyJWT{AllowAllKey: "test"}),
// 				},
// 			},
// 			want: func(t *testing.T, got *Service) bool {
// 				return assert.Equal[service.AuthorizationStrategy](t, &service.AuthorizationStrategyJWT{AllowAllKey: "test"}, got.authz)
// 			},
// 		},
// 		{
// 			name: "Create service with option 'WithProviders' and one provider given",
// 			args: args{
// 				opts: []service.Option[*Service]{
// 					WithProviders([]string{"azure"}),
// 				},
// 			},
// 			want: func(t *testing.T, got *Service) bool {
// 				return assert.Equal(t, []string{"azure"}, got.providers)
// 			},
// 		},
// 		{
// 			name: "Create service with option 'WithProviders' and no provider given",
// 			args: args{
// 				opts: []service.Option[*Service]{
// 					WithProviders([]string{}),
// 				},
// 			},
// 			want: func(t *testing.T, got *Service) bool {
// 				return assert.Equal(t, []string{}, got.providers)
// 			},
// 		},
// 		{
// 			name: "Create service with option 'WithAdditionalDiscoverers'",
// 			args: args{
// 				opts: []service.Option[*Service]{
// 					WithAdditionalDiscoverers([]cloud.Collector{&discoverytest.TestDiscoverer{ServiceId: config.DefaultTargetOfEvaluationID}}),
// 				},
// 			},
// 			want: func(t *testing.T, got *Service) bool {
// 				return assert.Contains(t, got.discoverers, &discoverytest.TestDiscoverer{ServiceId: config.DefaultTargetOfEvaluationID})
// 			},
// 		},
// 		{
// 			name: "Create service with option 'WithDiscoveryInterval'",
// 			args: args{
// 				opts: []service.Option[*Service]{
// 					WithDiscoveryInterval(time.Duration(8)),
// 				},
// 			},
// 			want: func(t *testing.T, got *Service) bool {
// 				return assert.Equal(t, time.Duration(8), got.discoveryInterval)
// 			},
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got := NewService(tt.args.opts...)
// 			tt.want(t, got)
// 		})
// 	}
// }

// func TestService_StartDiscovery(t *testing.T) {
// 	type fields struct {
// 		discoverer  cloud.Collector
// 		ctID        string
// 		collectorID string
// 	}

// 	tests := []struct {
// 		name          string
// 		fields        fields
// 		checkEvidence bool
// 	}{
// 		{
// 			name: "Err in discoverer",
// 			fields: fields{
// 				discoverer: &discoverytest.TestDiscoverer{TestCase: 0, ServiceId: config.DefaultTargetOfEvaluationID},
// 				ctID:       config.DefaultTargetOfEvaluationID,
// 			},
// 		},
// 		{
// 			name: "No err with default target of evaluation ID",
// 			fields: fields{
// 				discoverer:  &discoverytest.TestDiscoverer{TestCase: 2, ServiceId: config.DefaultTargetOfEvaluationID},
// 				ctID:        config.DefaultTargetOfEvaluationID,
// 				collectorID: config.DefaultEvidenceCollectorToolID,
// 			},
// 			checkEvidence: true,
// 		},
// 		{
// 			name: "No err with custom target of evaluation ID",
// 			fields: fields{
// 				discoverer:  &discoverytest.TestDiscoverer{TestCase: 2, ServiceId: testdata.MockTargetOfEvaluationID1},
// 				ctID:        testdata.MockTargetOfEvaluationID1,
// 				collectorID: config.DefaultEvidenceCollectorToolID,
// 			},
// 			checkEvidence: true,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			mockStream := &mockEvidenceStoreStream{connectionEstablished: true, expected: 2}
// 			mockStream.Prepare()

// 			svc := NewService()
// 			svc.ctID = tt.fields.ctID
// 			svc.collectorID = tt.fields.collectorID
// 			svc.evidenceStoreStreams = api.NewStreamsOf[evidence.EvidenceStore_StoreEvidencesClient, *evidence.StoreEvidenceRequest]()
// 			_, _ = svc.evidenceStoreStreams.GetStream("mock", "Evidence Store", func(target string, additionalOpts ...grpc.DialOption) (stream evidence.EvidenceStore_StoreEvidencesClient, err error) {
// 				return mockStream, nil
// 			})
// 			svc.evidenceStore = &api.RPCConnection[evidence.EvidenceStoreClient]{Target: "mock"}
// 			go svc.StartDiscovery(tt.fields.discoverer)

// 			if tt.checkEvidence {
// 				mockStream.Wait()
// 				want, _ := tt.fields.discoverer.List()

// 				got := mockStream.sentEvidences
// 				assert.Equal(t, len(want), len(got))

// 				// Retrieve the last one
// 				eWant := want[len(want)-1]
// 				eGot := got[len(got)-1]
// 				err := api.Validate(eGot)
// 				assert.NotNil(t, eGot)
// 				assert.NoError(t, err)

// 				or := eGot.GetOntologyResource()

// 				// Only the last element sent can be checked
// 				// The TestDiscoverer adds a random number to the ID, so we have to delete the last 3 characters as we do not know which random number will be added.
// 				assert.Equal(t, eWant.GetId()[:len(eWant.GetId())-3], or.GetId()[:len(or.GetId())-3])

// 				// Assert target of evaluation ID
// 				assert.Equal(t, tt.fields.ctID, eGot.TargetOfEvaluationId)
// 			}
// 		})
// 	}
// }

func TestService_Shutdown(t *testing.T) {
	service := NewService()
	service.Shutdown()

	assert.False(t, service.scheduler.IsRunning())

}

// mockEvidenceStoreStream implements Evidence_StoreEvidenceClient interface
type mockEvidenceStoreStream struct {
	// We add sentEvidence field to test the evidence that would be sent over gRPC
	sentEvidences []*evidence.Evidence
	// We add connectionEstablished to differentiate between the case where evidences can be sent and not
	connectionEstablished bool
	counter               int
	expected              int
	wg                    sync.WaitGroup
}

func (m *mockEvidenceStoreStream) Prepare() {
	m.wg.Add(m.expected)
}

func (m *mockEvidenceStoreStream) Wait() {
	m.wg.Wait()
}

func (m *mockEvidenceStoreStream) Recv() (*evidence.StoreEvidencesResponse, error) {
	if m.counter == 0 {
		m.counter++
		return &evidence.StoreEvidencesResponse{
			Status:        evidence.EvidenceStatus_EVIDENCE_STATUS_ERROR,
			StatusMessage: "mockError1",
		}, nil
	} else if m.counter == 1 {
		m.counter++
		return &evidence.StoreEvidencesResponse{
			Status: evidence.EvidenceStatus_EVIDENCE_STATUS_OK,
		}, nil
	} else {
		return nil, io.EOF
	}
}

func (m *mockEvidenceStoreStream) Send(req *evidence.StoreEvidenceRequest) (err error) {
	return m.SendMsg(req)
}

func (*mockEvidenceStoreStream) CloseAndRecv() (*emptypb.Empty, error) {
	return nil, nil
}

func (*mockEvidenceStoreStream) Header() (metadata.MD, error) {
	return nil, nil
}

func (*mockEvidenceStoreStream) Trailer() metadata.MD {
	return nil
}

func (*mockEvidenceStoreStream) CloseSend() error {
	return nil
}

func (*mockEvidenceStoreStream) Context() context.Context {
	return nil
}

func (m *mockEvidenceStoreStream) SendMsg(req interface{}) (err error) {
	e := req.(*evidence.StoreEvidenceRequest).Evidence
	if m.connectionEstablished {
		m.sentEvidences = append(m.sentEvidences, e)
	} else {
		err = fmt.Errorf("mock send error")
	}

	m.wg.Done()

	return
}

func (*mockEvidenceStoreStream) RecvMsg(_ interface{}) error {
	return nil
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
		providers           []string
		discoveryInterval   time.Duration
		Events              chan *DiscoveryEvent
		ctID                string
		envVariables        []envVariable
		cloudConfig         CloudCollectorConfig
	}
	tests := []struct {
		name    string
		fields  fields
		want    assert.Want[*Service]
		wantErr assert.WantErr
	}{
		// TODO(all): How to test for AWS and K8S authorizer without failure?
		{
			name: "Request with wrong provider name",
			fields: fields{
				scheduler: gocron.NewScheduler(time.UTC),
				providers: []string{"falseProvider"},
			},
			want: func(t *testing.T, got *Service) bool {
				return assert.False(t, got.scheduler.IsRunning())
			},
			wantErr: func(t *testing.T, gotErr error) bool {
				return assert.ErrorContains(t, gotErr, "provider falseProvider not known")
			},
		},
		// {
		// 	name: "Wrong permission",
		// 	fields: fields{
		// 		scheduler: gocron.NewScheduler(time.UTC),
		// 		providers: []string{},
		// 	},
		// 	want: assert.Nil[*Service],
		// 	wantErr: func(t *testing.T, gotErr error) bool {
		// 		return assert.ErrorContains(t, gotErr, "access denied")
		// 	},
		// },
		{
			name: "discovery interval error",
			fields: fields{
				scheduler:         gocron.NewScheduler(time.UTC),
				providers:         []string{ProviderAzure},
				discoveryInterval: time.Duration(-5 * time.Minute),
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
			want: func(t *testing.T, got *Service) bool {
				assert.Equal(t, []string{ProviderAzure}, got.providers)
				return assert.False(t, got.scheduler.IsRunning())
			},
			wantErr: func(t *testing.T, gotErr error) bool {
				return assert.ErrorContains(t, gotErr, "could not schedule job for ", ".Every() interval must be greater than 0")
			},
		},
		{
			name: "K8S authorizer error",
			fields: fields{
				scheduler:         gocron.NewScheduler(time.UTC),
				providers:         []string{ProviderK8S},
				discoveryInterval: time.Duration(5 * time.Minute),
				envVariables: []envVariable{
					// We must set HOME to a wrong path so that the K8S authorizer fails
					{
						hasEnvVariable:   true,
						envVariableKey:   "HOME",
						envVariableValue: "",
					},
				},
			},
			want: func(t *testing.T, got *Service) bool {
				assert.Equal(t, []string{ProviderK8S}, got.providers)
				return assert.False(t, got.scheduler.IsRunning())
			},
			wantErr: func(t *testing.T, gotErr error) bool {
				return assert.ErrorContains(t, gotErr, ErrK8sAuth.Error())
			},
		},
		{
			name: "Azure authorizer error",
			fields: fields{
				scheduler:         gocron.NewScheduler(time.UTC),
				providers:         []string{ProviderAzure},
				discoveryInterval: time.Duration(5 * time.Minute),
				cloudConfig: CloudCollectorConfig{
					TargetOfEvaluationID: config.DefaultTargetOfEvaluationID,
				},
				envVariables: []envVariable{
					{
						hasEnvVariable:   true,
						envVariableKey:   "AZURE_TOKEN_CREDENTIALS",
						envVariableValue: "fail",
					},
				},
			},
			want: func(t *testing.T, got *Service) bool {
				assert.Equal(t, []string{ProviderAzure}, got.providers)
				assert.Equal(t, config.DefaultTargetOfEvaluationID, got.cloudConfig.TargetOfEvaluationID)
				return assert.False(t, got.scheduler.IsRunning())
			},
			wantErr: func(t *testing.T, err error) bool {
				return assert.ErrorContains(t, err, ErrAzureAuth.Error())
			},
		},
		{
			name: "AWS authorizer error",
			fields: fields{
				scheduler:         gocron.NewScheduler(time.UTC),
				providers:         []string{ProviderAWS},
				discoveryInterval: time.Duration(5 * time.Minute),
				cloudConfig: CloudCollectorConfig{
					TargetOfEvaluationID: config.DefaultTargetOfEvaluationID,
				},
			},
			want: func(t *testing.T, got *Service) bool {
				assert.Equal(t, []string{ProviderAWS}, got.providers)
				assert.Equal(t, config.DefaultTargetOfEvaluationID, got.cloudConfig.TargetOfEvaluationID)
				return assert.False(t, got.scheduler.IsRunning())
			},
			wantErr: func(t *testing.T, err error) bool {
				return assert.ErrorContains(t, err, ErrAWSAuth.Error())
			},
		},
		{
			name: "OpenStack authorizer error",
			fields: fields{
				scheduler:         gocron.NewScheduler(time.UTC),
				providers:         []string{ProviderOpenstack},
				discoveryInterval: time.Duration(5 * time.Minute),
			},
			want: func(t *testing.T, got *Service) bool {
				assert.Equal(t, []string{ProviderOpenstack}, got.providers)
				return assert.False(t, got.scheduler.IsRunning())
			},
			wantErr: func(t *testing.T, err error) bool {
				return assert.ErrorContains(t, err, ErrOpenstackAuth.Error())
			},
		},
		{
			name: "Happy path: no discovery interval error",
			fields: fields{
				scheduler:         gocron.NewScheduler(time.UTC),
				providers:         []string{ProviderAzure},
				discoveryInterval: time.Duration(5 * time.Minute),
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
			want: func(t *testing.T, got *Service) bool {
				assert.Equal(t, []string{ProviderAzure}, got.providers)
				return assert.True(t, got.scheduler.IsRunning())
			},
			wantErr: assert.Nil[error],
		},
		{
			name: "Happy path: Azure authorizer from ENV",
			fields: fields{
				scheduler:         gocron.NewScheduler(time.UTC),
				providers:         []string{ProviderAzure},
				discoveryInterval: time.Duration(5 * time.Minute),
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
			want: func(t *testing.T, got *Service) bool {
				assert.Equal(t, []string{ProviderAzure}, got.providers)
				return assert.True(t, got.scheduler.IsRunning())
			},
			wantErr: assert.Nil[error],
		},
		{
			name: "Happy path: Azure with resource group",
			fields: fields{
				scheduler: gocron.NewScheduler(time.UTC),
				providers: []string{ProviderAzure},
				cloudConfig: CloudCollectorConfig{
					DiscoveryResourceGroup: "my-resource-group",
				},
				discoveryInterval: time.Duration(5 * time.Minute),
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
			want: func(t *testing.T, got *Service) bool {
				assert.Equal(t, []string{ProviderAzure}, got.providers)
				assert.Equal(t, "my-resource-group", got.cloudConfig.DiscoveryResourceGroup)
				return assert.True(t, got.scheduler.IsRunning())
			},
			wantErr: assert.Nil[error],
		},
		{
			name: "Happy path: CSAF with domain",
			fields: fields{
				scheduler:         gocron.NewScheduler(time.UTC),
				providers:         []string{ProviderCSAF},
				discoveryInterval: time.Duration(5 * time.Minute),
				cloudConfig: CloudCollectorConfig{
					DiscoveryCSAFDomain: "example.com",
				},
			},
			want: func(t *testing.T, got *Service) bool {
				assert.Equal(t, "example.com", got.cloudConfig.DiscoveryCSAFDomain)
				assert.Equal(t, []string{ProviderCSAF}, got.providers)
				return assert.True(t, got.scheduler.IsRunning())
			},
			wantErr: assert.Nil[error],
		},
		{
			name: "Happy path: CSAF without domain",
			fields: fields{
				scheduler:         gocron.NewScheduler(time.UTC),
				providers:         []string{ProviderCSAF},
				discoveryInterval: time.Duration(5 * time.Minute),
			},
			want: func(t *testing.T, got *Service) bool {
				assert.Equal(t, "", got.cloudConfig.DiscoveryCSAFDomain)
				assert.Equal(t, []string{ProviderCSAF}, got.providers)
				return assert.True(t, got.scheduler.IsRunning())
			},
			wantErr: assert.Nil[error],
		},
		{
			name: "Happy path: K8S",
			fields: fields{
				scheduler:         gocron.NewScheduler(time.UTC),
				providers:         []string{ProviderK8S},
				discoveryInterval: time.Duration(5 * time.Minute),
			},
			want: func(t *testing.T, got *Service) bool {
				assert.Equal(t, []string{ProviderK8S}, got.providers)
				return assert.True(t, got.scheduler.IsRunning())
			},
			wantErr: assert.Nil[error],
		},
		{
			name: "Happy path: OpenStack",
			fields: fields{
				scheduler:         gocron.NewScheduler(time.UTC),
				providers:         []string{ProviderOpenstack},
				discoveryInterval: time.Duration(5 * time.Minute),
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
			want: func(t *testing.T, got *Service) bool {
				assert.Equal(t, []string{ProviderOpenstack}, got.providers)
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
				providers:           tt.fields.providers,
				discoveryInterval:   tt.fields.discoveryInterval,
				Events:              tt.fields.Events,
				ctID:                tt.fields.ctID,
				cloudConfig:         tt.fields.cloudConfig,
			}

			// Set env variables
			for _, env := range tt.fields.envVariables {
				if env.hasEnvVariable {
					t.Setenv(env.envVariableKey, env.envVariableValue)
				}
			}

			err := svc.Start()

			tt.want(t, svc)
			tt.wantErr(t, err)
		})
	}
}

// TODO(all): Delete or fix?
// func TestDefaultServiceSpec(t *testing.T) {
// 	tests := []struct {
// 		name      string
// 		prepViper func()
// 		want      assert.Want[launcher.ServiceSpec]
// 	}{
// 		{
// 			name: "Happy path: providers given",
// 			prepViper: func() {
// 				viper.Set(config.DiscoveryProviderFlag, "azure")

// 			},
// 			want: func(t *testing.T, got launcher.ServiceSpec) bool {
// 				return assert.NotNil(t, got)
// 			},
// 		},
// 		{
// 			name:      "Happy path: no providers given",
// 			prepViper: func() {},
// 			want: func(t *testing.T, got launcher.ServiceSpec) bool {
// 				return assert.NotNil(t, got)

// 			},
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			viper.Reset()
// 			tt.prepViper()

// 			got := DefaultServiceSpec()

// 			tt.want(t, got)
// 		})
// 	}
// }
