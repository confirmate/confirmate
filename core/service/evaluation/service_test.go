package evaluation

import (
	"net/http"
	"testing"

	"confirmate.io/core/api/evaluation/evaluationconnect"
	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/api/orchestrator/orchestratorconnect"
	"confirmate.io/core/persistence"
	"confirmate.io/core/service"
	"confirmate.io/core/util/assert"
)

func TestNewService(t *testing.T) {
	type args struct {
		opts []service.Option[Service]
	}
	tests := []struct {
		name string
		// Named input parameters for target function.
		args    args
		want    assert.Want[evaluationconnect.EvaluationHandler]
		wantErr assert.WantErr
	}{
		{
			name: "db error - creating db with invalid config",
			args: args{
				opts: []service.Option[Service]{
					WithConfig(Config{
						PersistenceConfig: persistence.Config{
							Host:             "localhost",
							Port:             5432,
							DBName:           "confirmate",
							User:             "confirmate",
							Password:         "confirmate",
							SSLMode:          "disable",
							MaxConn:          10,
							InMemoryDB:       false,
							Types:            []any{},
							CustomJoinTables: []persistence.CustomJoinTable{},
						}}),
				},
			},
			want: assert.Nil[evaluationconnect.EvaluationHandler],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.ErrorContains(t, err, "could not create db:")
			},
		},
		{
			name: "happy path with WithConfig option",
			args: args{
				opts: []service.Option[Service]{
					WithConfig(Config{
						PersistenceConfig:   persistence.DefaultConfig,
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
				assert.NotEmpty(t, svc.db)
				assert.Equal(t, Config{
					OrchestratorAddress: "http://testhost:8080",
					OrchestratorClient:  http.DefaultClient,
					PersistenceConfig:   persistence.DefaultConfig,
				}, svc.cfg)
				assert.NotEmpty(t, svc.scheduler)
				assert.NotEmpty(t, orchestratorconnect.NewOrchestratorClient(svc.cfg.OrchestratorClient, "http:://testhost:8080"), svc.orchestratorClient)
				assert.Equal(t, make(map[string]map[string]*orchestrator.Control), svc.catalogControls)
				assert.NotNil(t, &svc.streamMutex)
				return assert.NotNil(t, &svc.catalogsMutex)
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
				assert.NotEmpty(t, svc.db)
				assert.Equal(t, DefaultConfig, svc.cfg)
				assert.NotEmpty(t, svc.scheduler)
				assert.NotEmpty(t, orchestratorconnect.NewOrchestratorClient(svc.cfg.OrchestratorClient, svc.cfg.OrchestratorAddress), svc.orchestratorClient)
				assert.Equal(t, make(map[string]map[string]*orchestrator.Control), svc.catalogControls)
				assert.NotNil(t, &svc.streamMutex)
				return assert.NotNil(t, &svc.catalogsMutex)
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
