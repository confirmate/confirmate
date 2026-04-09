package orchestrator

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/auth"
	"confirmate.io/core/persistence"
	"confirmate.io/core/service"
	"confirmate.io/core/util"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// UpsertCurrentUserPermission ensures that the calling user has the specified permission for the given resource (create or update).
func (svc *Service) UpsertUserPermission(
	ctx context.Context,
	req *connect.Request[orchestrator.UpsertUserPermissionRequest],
) (res *connect.Response[orchestrator.UpsertUserPermissionResponse], err error) {
	var (
		allowed bool
	)

	// Validate the request
	if err = service.Validate(req); err != nil {
		return nil, err
	}

	// Only admins may grant or revoke permissions.
	allowed, _, err = CheckAccess(ctx, svc.authz, svc, orchestrator.RequestType_REQUEST_TYPE_UPDATED, "", orchestrator.ObjectType_OBJECT_TYPE_USER)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if !allowed {
		return nil, service.ErrPermissionDenied
	}

	err = svc.db.Save(req.Msg.UserPermission)
	if err = service.HandleDatabaseError(err); err != nil {
		return nil, err
	}

	res = connect.NewResponse(&orchestrator.UpsertUserPermissionResponse{
		UserPermission: req.Msg.UserPermission,
	})
	return
}

// GetCurrentUser retrieves the current authenticated user based on the context of the request.
func (svc *Service) GetCurrentUser(
	ctx context.Context,
	req *connect.Request[orchestrator.GetCurrentUserRequest],
) (res *connect.Response[orchestrator.User], err error) {
	var (
		claims *auth.OAuthClaims
		user   orchestrator.User
		ok     bool
	)

	claims, ok = auth.ClaimsFromContext(ctx)
	if !ok || claims == nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("no authentication context"))
	}

	userId := claims.Issuer + "|" + claims.Subject
	err = svc.db.Get(&user, "id = ?", userId)
	if err = service.HandleDatabaseError(err, service.ErrNotFound("user")); err != nil {
		return nil, err
	}

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
	allowed, _, err = CheckAccess(ctx, svc.authz, svc, orchestrator.RequestType_REQUEST_TYPE_UNSPECIFIED, req.Msg.UserId, orchestrator.ObjectType_OBJECT_TYPE_USER)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("%w: %w", service.ErrDatabaseError, err))
	}

	if !allowed {
		slog.Debug("access denied", slog.String("userID", req.Msg.UserId), slog.String("requestType", orchestrator.RequestType_REQUEST_TYPE_UNSPECIFIED.String()))
		return nil, service.ErrNotFound("user")
	}

	err = svc.db.Get(&user, "id = ?", req.Msg.UserId)
	if err = service.HandleDatabaseError(err, service.ErrNotFound("user")); err != nil {
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
		users   []*orchestrator.User
		conds   []any
		npt     string
		allowed bool
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

	// Only admins may list users.
	allowed, _, err = CheckAccess(ctx, svc.authz, svc, orchestrator.RequestType_REQUEST_TYPE_LIST, "", orchestrator.ObjectType_OBJECT_TYPE_USER)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if !allowed {
		return nil, service.ErrNotFound("user")
	}

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
		allowed     bool
	)

	// Validate request
	err = service.Validate(req)
	if err != nil {
		return nil, err
	}

	// Only admins may list permissions.
	allowed, _, err = CheckAccess(ctx, svc.authz, svc, orchestrator.RequestType_REQUEST_TYPE_GET, "", orchestrator.ObjectType_OBJECT_TYPE_USER)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if !allowed {
		return nil, service.ErrPermissionDenied
	}

	// Set default ordering
	if req.Msg.OrderBy == "" {
		req.Msg.OrderBy = "user_id"
		req.Msg.Asc = true
	}

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
	var roles []orchestrator.Role

	for v := range orchestrator.Role_name {
		role := orchestrator.Role(v)
		if role == orchestrator.Role_ROLE_UNSPECIFIED {
			continue
		}
		roles = append(roles, role)
	}

	res = connect.NewResponse(&orchestrator.ListUserRolesResponse{Roles: roles})
	return
}

// RemoveUser disables a user in the system based on their unique identifier.
func (svc *Service) RemoveUser(
	ctx context.Context,
	req *connect.Request[orchestrator.RemoveUserRequest],
) (res *connect.Response[emptypb.Empty], err error) {
	var (
		user    orchestrator.User
		allowed bool
	)

	// Validate the request
	if err = service.Validate(req); err != nil {
		return nil, err
	}

	// Only admins may delete users.
	allowed, _, err = CheckAccess(ctx, svc.authz, svc, orchestrator.RequestType_REQUEST_TYPE_DELETED, "", orchestrator.ObjectType_OBJECT_TYPE_USER)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if !allowed {
		return nil, service.ErrPermissionDenied
	}

	err = svc.db.Get(&user, "id = ?", req.Msg.UserId)
	if err = service.HandleDatabaseError(err, service.ErrNotFound("user")); err != nil {
		return nil, err
	}

	user.Enabled = false
	err = svc.db.Save(&user)
	if err = service.HandleDatabaseError(err); err != nil {
		return nil, err
	}

	res = connect.NewResponse(&emptypb.Empty{})
	return
}

// CheckAccess is a helper function to check if the user associated with the given context has access to perform the specified request type and request. It extracts user information from the JWT claims, ensures the user exists in the database, and then checks access using the provided authorization strategy.
func CheckAccess(ctx context.Context, authz service.AuthorizationStrategy, svc *Service, reqType orchestrator.RequestType, resourceId string, objectType orchestrator.ObjectType) (bool, []string, error) {
	var (
		user   *orchestrator.User
		claims *auth.OAuthClaims
		ok     bool
		err    error
		userId string
	)

	// If we have no authorization strategy, we allow all access by default. This is useful for
	// internal (test) services that do not require authentication/authorization.
	if authz == nil {
		return true, nil, nil
	}

	// If JWT claims are present, provision the user in the DB (JIT provisioning) and use their ID.
	// If no claims are present, use an empty user ID — strategies like AuthorizationStrategyAllowAll
	// don't require a user ID, while AuthorizationStrategyPermissionStore will deny empty IDs.
	claims, ok = auth.ClaimsFromContext(ctx)
	if ok && claims != nil {
		if svc == nil {
			return false, nil, fmt.Errorf("service is nil")
		}

		if svc.db == nil {
			return false, nil, fmt.Errorf("database is not initialized")
		}

		// JIT-provision the user using a read-then-update approach to avoid overwriting existing
		// fields (e.g. enabled status) on every request. The user ID is "iss|sub" as recommended
		// by the OIDC specification, ensuring uniqueness across identity providers.
		userId = claims.Issuer + "|" + claims.Subject
		user = &orchestrator.User{}
		err = svc.db.Get(user, "id = ?", userId)

		if errors.Is(err, persistence.ErrRecordNotFound) {
			// User not found: create them with all identity fields.
			user = &orchestrator.User{
				Id:         userId,
				Username:   util.Ref(claims.PreferredUsername),
				FirstName:  util.Ref(claims.GivenName),
				LastName:   util.Ref(claims.FamilyName),
				Enabled:    true,
				Email:      util.Ref(claims.Email),
				LastAccess: timestamppb.Now(),
			}
			err = svc.db.Create(user)
			if err != nil {
				return false, nil, fmt.Errorf("failed to create user: %w", err)
			}
		} else if err != nil {
			return false, nil, fmt.Errorf("failed to look up user: %w", err)
		} else {
			// User exists: update only identity fields and last_access to preserve other persisted
			// fields such as enabled status.
			user.Username = util.Ref(claims.PreferredUsername)
			user.FirstName = util.Ref(claims.GivenName)
			user.LastName = util.Ref(claims.FamilyName)
			user.Email = util.Ref(claims.Email)
			user.LastAccess = timestamppb.Now()
			err = svc.db.Save(user)
			if err != nil {
				return false, nil, fmt.Errorf("failed to update user: %w", err)
			}
		}
	}

	allowed, resourceIDs := authz.CheckAccess(ctx, userId, reqType, orchestrator.UserPermission_PERMISSION_READER, resourceId, objectType)

	return allowed, resourceIDs, nil
}

type permissionStore struct {
	db persistence.DB
}

// HasPermission checks if the given user has the specified permission for the resource.
func (ps permissionStore) HasPermission(ctx context.Context, userId string, resourceId string, permission orchestrator.UserPermission_Permission, reqType orchestrator.RequestType, objectType orchestrator.ObjectType) (bool, error) {
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
		userId, objectType, resourceId, allowed,
	)
	if err != nil {
		return false, fmt.Errorf("failed to check permissions: %w", err)
	}

	return count > 0, nil
}

// PermissionForResource returns a list of resource IDs for which the given user has at least the specified permission.
func (ps permissionStore) PermissionForResources(ctx context.Context, userID string, permission orchestrator.UserPermission_Permission, reqType orchestrator.RequestType, objectType orchestrator.ObjectType) ([]string, error) {
	var (
		conds           []any
		userPermissions []orchestrator.UserPermission
		err             error
	)

	// Define a list of allowed permissions based on the requested permission.
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

	// Get all permissions for the user and object type that match the allowed permissions, then extract the resource IDs from those permissions.
	conds = []any{
		"user_id = ? AND resource_type = ? AND permission IN (?)",
		userID,
		objectType,
		allowed,
	}
	err = ps.db.List(
		&userPermissions,
		"resource_id",
		true,
		0,
		-1,
		conds...,
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
