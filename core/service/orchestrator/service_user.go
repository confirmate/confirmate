package orchestrator

import (
	"context"
	"sync"

	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/api/orchestrator/orchestratorconnect"
	"confirmate.io/core/service"
	"connectrpc.com/connect"
)

// Service implements the User Management Service handler (see
// [orchestratorconnect.UserManegementHandler]).
type UserManagementService struct {
	orchestratorconnect.UnimplementedUserManagementHandler
	cfg UserManagementConfig

	// authz defines our authorization strategy for target-of-evaluation scoped access.
	authz service.AuthorizationStrategy

	// subscribers is a map of subscribers for change events
	subscribers      map[int64]*subscriber
	subscribersMutex sync.RWMutex

	nextSubscriberId int64
}

// DefaultConfig is the default configuration for the user management [Service].
var DefaultUserManagementConfig = UserManagementConfig{}

type UserManagementConfig struct {
}

// WithUserManagementConfig sets the service configuration, overriding the default configuration.
func WithUserManagementConfig(cfg UserManagementConfig) service.Option[UserManagementService] {
	return func(svc *UserManagementService) {
		svc.cfg = cfg
	}
}

func NewUserManagementService(opts ...service.Option[UserManagementService]) (handler orchestratorconnect.UserManagementHandler, err error) {
	var (
		svc = &UserManagementService{
			cfg:         DefaultUserManagementConfig,
			authz:       nil,
			subscribers: make(map[int64]*subscriber),
		}
	)

	for _, opt := range opts {
		opt(svc)
	}

	handler = svc

	return
}

// GetCurrentUser retrieves the current authenticated user based on the context of the request.
func (svc *UserManagementService) GetCurrentUser(
	ctx context.Context,
	req *connect.Request[orchestrator.GetCurrentUserRequest],
) (res *connect.Response[orchestrator.User], err error) {
	var (
		user orchestrator.User
	)

	// TODO (anatheka): Implement

	res = connect.NewResponse(&user)
	return
}

// GetUser retrieves a user by their unique identifier.
func (svc *UserManagementService) GetUser(
	ctx context.Context,
	req *connect.Request[orchestrator.GetUserRequest],
) (res *connect.Response[orchestrator.User], err error) {
	var (
		user orchestrator.User
	)

	// TODO (anatheka): Implement

	res = connect.NewResponse(&user)
	return
}

// ListUsers lists all users in the system, optionally filtered by role or other criteria.
func (svc *UserManagementService) ListUsers(
	ctx context.Context,
	req *connect.Request[orchestrator.ListUsersRequest],
) (res *connect.Response[orchestrator.ListUsersResponse], err error) {
	var (
		users orchestrator.ListUsersResponse
	)

	// TODO (anatheka): Implement

	res = connect.NewResponse(&users)
	return
}

// ListUserRoles lists all predefined user roles available in the system.
func (svc *UserManagementService) ListUserRoles(
	ctx context.Context,
	req *connect.Request[orchestrator.ListUserRolesRequest],
) (res *connect.Response[orchestrator.ListUserRolesResponse], err error) {
	var (
		roles orchestrator.ListUserRolesResponse
	)

	// TODO (anatheka): Implement

	res = connect.NewResponse(&roles)
	return
}
