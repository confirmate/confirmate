package orchestrator

import (
	"context"

	"confirmate.io/core/api/orchestrator"
	"connectrpc.com/connect"
)

// GetCurrentUser retrieves the current authenticated user based on the context of the request.
func (svc *Service) GetCurrentUser(
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
func (svc *Service) GetUser(
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
func (svc *Service) ListUsers(
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
func (svc *Service) ListUserRoles(
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
