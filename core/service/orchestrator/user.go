package orchestrator

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/auth"
	"confirmate.io/core/persistence"
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
		req.Msg.User.ExpirationDate = timestamppb.New(time.Unix(service.GetClaimInt64(claims, "exp"), 0)) // Set expiration date from JWT claim if not provided
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
		user    orchestrator.User
		allowed bool
	)

	// Validate the request
	if err = service.Validate(req); err != nil {
		return nil, err
	}

	// Check access via the configured strategy, which may include JIT provisioning of the user in the context for JWT-based authz strategies
	allowed, _, err = CheckAccess(ctx, svc.authz, svc, orchestrator.RequestType_REQUEST_TYPE_UNSPECIFIED, orchestrator.UserPermission_RESOURCE_TYPE_USER, req.Msg.UserId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("%w: %w", service.ErrDatabaseError, err))
	}

	if !allowed {
		slog.Debug("access denied", slog.String("userID", req.Msg.UserId), slog.String("requestType", orchestrator.RequestType_REQUEST_TYPE_UNSPECIFIED.String()))
		return nil, connect.NewError(connect.CodePermissionDenied, service.ErrPermissionDenied)
	}

	err = svc.db.Get(&user, "id = ?", req.Msg.UserId)
	if err = service.HandleDatabaseError(err, service.ErrNotFound("assessment result")); err != nil {
		return nil, err
	}

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
func CheckAccess(ctx context.Context, authz service.AuthorizationStrategy, svc *Service, reqType orchestrator.RequestType, resourceType orchestrator.UserPermission_ResourceType, resourceId string) (bool, []string, error) {
	var (
		user *orchestrator.User
		err  error
	)

	if svc == nil {
		return false, nil, fmt.Errorf("service is nil")
	}
	if svc.db == nil {
		return false, nil, fmt.Errorf("database is not initialized")
	}
	if authz == nil {
		return false, nil, fmt.Errorf("authorization strategy is not configured")
	}

	// Get claims from context
	claims, _ := auth.ClaimsFromContext(ctx)

	// TODO(anatheka): Should we check if the user is already in the DB?
	// Extract user information from claims
	user = &orchestrator.User{
		Id:             service.GetClaim(claims, "iss") + "|" + service.GetClaim(claims, "sub"),
		Username:       util.Ref(service.GetClaim(claims, "preferred_username")),
		FirstName:      util.Ref(service.GetClaim(claims, "given_name")),
		LastName:       util.Ref(service.GetClaim(claims, "family_name")),
		Enabled:        true,
		Email:          util.Ref(service.GetClaim(claims, "email")),
		ExpirationDate: timestamppb.New(time.Unix(service.GetClaimInt64(claims, "exp"), 0)),
		LastAccess:     timestamppb.Now(),
	}

	// Ensure the user exists in the DB.
	// We do not need the response, we only want to know if an error occurred.
	err = svc.db.Save(user)
	if err != nil {
		return false, nil, fmt.Errorf("failed to ensure current user: %w", err)
	}

	allowed, resourceIDs := authz.CheckAccess(ctx, user.Id, reqType, resourceType, orchestrator.UserPermission_PERMISSION_READER, resourceId)

	return allowed, resourceIDs, nil
}

type permissionStore struct {
	db persistence.DB
}

// HasPermission checks if the given user has the specified permission for the resource.
func (ps permissionStore) HasPermission(ctx context.Context, userId string, resourceType orchestrator.UserPermission_ResourceType, resourceId string, permission orchestrator.UserPermission_Permission) (bool, error) {
	var (
		count          int64
		err            error
		userPermission orchestrator.UserPermission
	)

	// Check if the user has the required permission for the resource by querying the database for matching user permissions.
	// If a lower permission is requested, also accept higher permissions (ADMIN > CONTRIBUTOR > READER).
	allowed := []orchestrator.UserPermission_Permission{permission}
	switch permission {
	case orchestrator.UserPermission_PERMISSION_READER:
		allowed = []orchestrator.UserPermission_Permission{
			orchestrator.UserPermission_PERMISSION_READER,
			orchestrator.UserPermission_PERMISSION_CONTRIBUTOR,
			orchestrator.UserPermission_PERMISSION_ADMIN,
		}
	case orchestrator.UserPermission_PERMISSION_CONTRIBUTOR:
		allowed = []orchestrator.UserPermission_Permission{
			orchestrator.UserPermission_PERMISSION_CONTRIBUTOR,
			orchestrator.UserPermission_PERMISSION_ADMIN,
		}
	case orchestrator.UserPermission_PERMISSION_ADMIN:
		allowed = []orchestrator.UserPermission_Permission{
			orchestrator.UserPermission_PERMISSION_ADMIN,
		}
	}

	count, err = ps.db.Count(
		&userPermission,
		"user_id = ? AND resource_type = ? AND resource_id = ? AND permission IN (?)",
		userId, resourceType, resourceId, allowed,
	)
	if err != nil {
		return false, fmt.Errorf("failed to check permissions: %w", err)
	}

	return count > 0, nil
}

// PermissionForResource returns a list of resource IDs for which the given user has at least the specified permission.
func (ps permissionStore) PermissionForResources(ctx context.Context, userID string, resourceType orchestrator.UserPermission_ResourceType, permission orchestrator.UserPermission_Permission) ([]string, error) {
	var (
		userPermissions []orchestrator.UserPermission
		err             error
	)

	allowed := []orchestrator.UserPermission_Permission{permission}
	switch permission {
	case orchestrator.UserPermission_PERMISSION_READER:
		allowed = []orchestrator.UserPermission_Permission{
			orchestrator.UserPermission_PERMISSION_READER,
			orchestrator.UserPermission_PERMISSION_CONTRIBUTOR,
			orchestrator.UserPermission_PERMISSION_ADMIN,
		}
	case orchestrator.UserPermission_PERMISSION_CONTRIBUTOR:
		allowed = []orchestrator.UserPermission_Permission{
			orchestrator.UserPermission_PERMISSION_CONTRIBUTOR,
			orchestrator.UserPermission_PERMISSION_ADMIN,
		}
	case orchestrator.UserPermission_PERMISSION_ADMIN:
		allowed = []orchestrator.UserPermission_Permission{
			orchestrator.UserPermission_PERMISSION_ADMIN,
		}
	}

	// Use shared pagination helper; add explicit condition args
	userPermissions, _, err = service.PaginateStorage[orchestrator.UserPermission](
		&orchestrator.ListUserPermissionsRequest{
			OrderBy: "resource_id",
			Asc:     true,
		},
		ps.db,
		service.DefaultPaginationOpts,
		"user_id = ? AND resource_type = ? AND permission IN (?)",
		userID,
		resourceType,
		allowed,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve permissions: %w", err)
	}

	resourceIDs := make([]string, len(userPermissions))
	for i := range userPermissions {
		resourceIDs[i] = userPermissions[i].ResourceId
	}

	return resourceIDs, nil
}
