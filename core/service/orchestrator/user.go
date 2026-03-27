package orchestrator

import (
	"context"
	"fmt"
	"time"

	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/auth"
	"confirmate.io/core/service"
	"confirmate.io/core/util"
	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// UpsertCurrentUser ensures that the calling user exists in the DB (create or update) and returns the corresponding user record.
func (svc *Service) UpsertCurrentUser(
	ctx context.Context,
	req *connect.Request[orchestrator.UpsertCurrentUserRequest],
) (res *connect.Response[orchestrator.UpsertCurrentUserResponse], err error) {
	// Validate the request
	if err = service.Validate(req); err != nil {
		return nil, err
	}

	err = svc.db.Save(req.Msg.User)
	if err = service.HandleDatabaseError(err); err != nil {
		return nil, err
	}

	// Notify subscribers
	go svc.publishEvent(&orchestrator.ChangeEvent{
		Timestamp:   timestamppb.Now(),
		Category:    orchestrator.EventCategory_EVENT_CATEGORY_USER,
		RequestType: orchestrator.RequestType_REQUEST_TYPE_STORED,
		EntityId:    req.Msg.User.Id,
		Entity: &orchestrator.ChangeEvent_User{
			User: req.Msg.User,
		},
	})

	// Add/update time fields
	if req.Msg.User.ExpirationDate == nil {
		claims, _ := auth.ClaimsFromContext(ctx)
		req.Msg.User.ExpirationDate = timestamppb.New(time.Unix(getClaimInt64(claims, "exp"), 0)) // Set expiration date from JWT claim if not provided
	}
	req.Msg.User.LastAccess = timestamppb.Now() // Set last access to now

	res = connect.NewResponse(&orchestrator.UpsertCurrentUserResponse{
		User: req.Msg.User,
	})
	return
}

// UpsertCurrentUserPermission ensures that the calling user has the specified permission for the given resource (create or update).
func (svc *Service) UpsertUserPermission(
	ctx context.Context,
	req *connect.Request[orchestrator.UpsertUserPermissionRequest],
) (res *connect.Response[orchestrator.UpsertUserPermissionResponse], err error) {
	// Validate the request
	if err = service.Validate(req); err != nil {
		return nil, err
	}

	// TODO (anatheka): Add authorization check to ensure the user has the right to upsert this permission

	err = svc.db.Save(req.Msg.UserPermission)
	if err = service.HandleDatabaseError(err); err != nil {
		return nil, err
	}

	res = connect.NewResponse(&orchestrator.UpsertUserPermissionResponse{})
	return
}

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
		users []*orchestrator.User
		conds []any
		npt   string
	)

	// Validate request
	err = service.Validate(req)
	if err != nil {
		return nil, err
	}

	// Set default ordering
	if req.Msg.OrderBy == "" {
		req.Msg.OrderBy = "id"
		req.Msg.Asc = true
	}

	// TODO (anatheka): Implement
	// First implementation for testing purposes
	users, npt, err = service.PaginateStorage[*orchestrator.User](req.Msg, svc.db, service.DefaultPaginationOpts, conds...)
	if err = service.HandleDatabaseError(err); err != nil {
		return nil, err
	}

	res = connect.NewResponse(&orchestrator.ListUsersResponse{
		Users:         users,
		NextPageToken: npt,
	})
	return
}

// ListUserPermissions lists all user permissions for a given userID.
func (svc *Service) ListUserPermissions(
	ctx context.Context,
	req *connect.Request[orchestrator.ListUserPermissionsRequest],
) (res *connect.Response[orchestrator.ListUserPermissionsResponse], err error) {
	var (
		permissions []*orchestrator.UserPermission
		conds       []any
		npt         string
	)

	// Validate request
	err = service.Validate(req)
	if err != nil {
		return nil, err
	}

	// TODO(anatheka): Add auth check

	// Set default ordering
	if req.Msg.OrderBy == "" {
		req.Msg.OrderBy = "user_id"
		req.Msg.Asc = true
	}

	// TODO (anatheka): Implement
	// First implementation for testing purposes
	permissions, npt, err = service.PaginateStorage[*orchestrator.UserPermission](req.Msg, svc.db, service.DefaultPaginationOpts, conds...)
	if err = service.HandleDatabaseError(err); err != nil {
		return nil, err
	}

	res = connect.NewResponse(&orchestrator.ListUserPermissionsResponse{
		UserPermissions: permissions,
		NextPageToken:   npt,
	})
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

// RemoveUser deletes a user from the system based on their unique identifier.
func (svc *Service) RemoveUser(
	ctx context.Context,
	req *connect.Request[orchestrator.RemoveUserRequest],
) (res *connect.Response[emptypb.Empty], err error) {
	// Validate the request
	if err = service.Validate(req); err != nil {
		return nil, err
	}

	// TODO (anatheka): Implement

	res = connect.NewResponse(&emptypb.Empty{})
	return
}

// CheckAccess is a helper function to check if the user associated with the given context has access to perform the specified request type and request. It extracts user information from the JWT claims, ensures the user exists in the database, and then checks access using the provided authorization strategy.
func CheckAccess[T any](ctx context.Context, authz service.AuthorizationStrategy, svc *Service, typ orchestrator.RequestType, req *connect.Request[T]) error {
	var (
		user *orchestrator.User
		err  error
	)
	claims, _ := auth.ClaimsFromContext(ctx)

	// Extract user information from claims
	user = &orchestrator.User{
		Id:             getClaim(claims, "sub"),
		Username:       util.Ref(getClaim(claims, "preferred_username")),
		FirstName:      util.Ref(getClaim(claims, "given_name")),
		LastName:       util.Ref(getClaim(claims, "family_name")),
		Enabled:        true,
		Email:          util.Ref(getClaim(claims, "email")),
		ExpirationDate: timestamppb.New(time.Unix(getClaimInt64(claims, "exp"), 0)),
		LastAccess:     timestamppb.Now(),
	}

	// Ensure the user exists in the DB and is up to date with the information from the claims.
	// We do not need the response, we only want to know if an error occurred.
	_, err = svc.UpsertCurrentUser(ctx, &connect.Request[orchestrator.UpsertCurrentUserRequest]{
		Msg: &orchestrator.UpsertCurrentUserRequest{
			User: user,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to ensure current user: %w", err)
	}

	auth := service.CheckAccess(authz, ctx, typ, req)
	if !auth {
		return fmt.Errorf("access denied for user %s", "userID")
	}

	return nil
}

// getClaim is a helper function to extract a specific claim from the claims map, returning an empty string if the claim is not present or not a string.
func getClaim(claims map[string]any, key string) string {
	if val, ok := claims[key]; ok && val != nil {
		return val.(string)
	}
	return ""
}

// getClaimInt64 extracts a numeric claim as int64, returning 0 if missing or not numeric.
func getClaimInt64(claims map[string]any, key string) int64 {
	val, ok := claims[key]
	if !ok || val == nil {
		return 0
	}

	switch v := val.(type) {
	case int64:
		return v
	case int:
		return int64(v)
	case int32:
		return int64(v)
	case float64:
		return int64(v)
	case float32:
		return int64(v)
	case uint64:
		return int64(v)
	case uint:
		return int64(v)
	case uint32:
		return int64(v)
	default:
		return 0
	}
}
