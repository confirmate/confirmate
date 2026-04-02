package orchestrator

import (
	"context"
	"testing"

	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/auth"
	"confirmate.io/core/persistence"
	"confirmate.io/core/persistence/persistencetest"
	"confirmate.io/core/service"
	"confirmate.io/core/service/orchestrator/orchestratortest"
	"confirmate.io/core/util/assert"
	"connectrpc.com/connect"
	"github.com/golang-jwt/jwt/v5"
)

func TestService_UpsertCurrentUser(t *testing.T) {
	type args struct {
		ctx context.Context
		req *connect.Request[orchestrator.UpsertCurrentUserRequest]
	}
	type fields struct {
		db    persistence.DB
		authz service.AuthorizationStrategy
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[*connect.Response[orchestrator.UpsertCurrentUserResponse]]
		wantErr assert.WantErr
	}{
		{
			name: "err: db error",
			args: args{
				ctx: context.Background(),
				req: connect.NewRequest(&orchestrator.UpsertCurrentUserRequest{
					User: orchestratortest.MockUser1,
				}),
			},
			fields: fields{
				db: persistencetest.SaveErrorDB(t, persistence.ErrDatabase, types, joinTables),
			},
			want: assert.Nil[*connect.Response[orchestrator.UpsertCurrentUserResponse]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInternal)
			},
		},
		{
			name: "err: invalid request",
			args: args{
				ctx: context.Background(),
				req: connect.NewRequest(&orchestrator.UpsertCurrentUserRequest{}),
			},
			fields: fields{},
			want:   assert.Nil[*connect.Response[orchestrator.UpsertCurrentUserResponse]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument)
			},
		},
		{
			name: "happy path",
			args: args{
				ctx: auth.WithClaims(context.Background(), jwt.MapClaims{
					"exp": float64(9999999999),
					"sub": "cliadmin",
				}),
				req: connect.NewRequest(&orchestrator.UpsertCurrentUserRequest{
					User: orchestratortest.MockUser1,
				}),
			},
			fields: fields{
				db:    persistencetest.NewInMemoryDB(t, types, joinTables),
				authz: &service.AuthorizationStrategyJWT{},
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.UpsertCurrentUserResponse], args ...any) bool {
				want := orchestratortest.MockUser1

				assert.NotEmpty(t, got.Msg.User.GetExpirationDate())
				assert.NotEmpty(t, got.Msg.User.GetLastAccess())
				got.Msg.User.ExpirationDate = nil
				got.Msg.User.LastAccess = nil

				return assert.Equal(t, want, got.Msg.User)
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{
				db:    tt.fields.db,
				authz: tt.fields.authz,
			}

			res, err := svc.UpsertCurrentUser(tt.args.ctx, tt.args.req)
			tt.want(t, res)
			tt.wantErr(t, err)
		})
	}
}
