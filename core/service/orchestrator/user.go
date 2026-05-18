// Copyright 2016-2026 Fraunhofer AISEC
//
// SPDX-License-Identifier: Apache-2.0
//
//                                 /$$$$$$  /$$                                     /$$
//                               /$$__  $$|__/                                    | $$
//   /$$$$$$$  /$$$$$$  /$$$$$$$ | $$  \__/ /$$  /$$$$$$  /$$$$$$/$$$$   /$$$$$$  /$$$$$$    /$$$$$$
//  /$$_____/ /$$__  $$| $$__  $$| $$$$    | $$ /$$__  $$| $$_  $$_  $$ |____  $$|_  $$_/   /$$__  $$
// | $$      | $$  \ $$| $$  \ $$| $$_/    | $$| $$  \__/| $$ \ $$ \ $$  /$$$$$$$  | $$    | $$$$$$$$
// | $$      | $$  | $$| $$  | $$| $$      | $$| $$      | $$ | $$ | $$ /$$__  $$  | $$ /$$| $$_____/
// |  $$$$$$$|  $$$$$$/| $$  | $$| $$      | $$| $$      | $$ | $$ | $$|  $$$$$$$  |  $$$$/|  $$$$$$$
// \_______/ \______/ |__/  |__/|__/      |__/|__/      |__/ |__/ |__/ \_______/   \___/   \_______/
//
// This file is part of Confirmate Core.

package orchestrator

import (
	"context"
	"errors"
	"fmt"

	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/auth"
	"confirmate.io/core/persistence"
	"confirmate.io/core/service"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// UpsertUserPermission creates or updates a specific permission entry for a user and resource.
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
	allowed, _, err = CheckAccess(ctx, svc.authz, svc, orchestrator.RequestType_REQUEST_TYPE_UPDATED, "", orchestrator.ObjectType_OBJECT_TYPE_USER_PERMISSION)
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

// RemoveUserPermission removes a specific permission entry for a user and resource.
func (svc *Service) RemoveUserPermission(
	ctx context.Context,
	req *connect.Request[orchestrator.RemoveUserPermissionRequest],
) (res *connect.Response[emptypb.Empty], err error) {
	var (
		permission orchestrator.UserPermission
		allowed    bool
	)

	// Validate the request
	if err = service.Validate(req); err != nil {
		return nil, err
	}

	// Only admins may revoke permissions.
	allowed, _, err = CheckAccess(ctx, svc.authz, svc, orchestrator.RequestType_REQUEST_TYPE_DELETED, "", orchestrator.ObjectType_OBJECT_TYPE_USER_PERMISSION)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if !allowed {
		return nil, service.ErrPermissionDenied
	}

	err = svc.db.Delete(
		&permission,
		"user_id = ? AND resource_id = ? AND resource_type = ?",
		req.Msg.GetUserPermission().GetUserId(),
		req.Msg.GetUserPermission().GetResourceId(),
		req.Msg.GetUserPermission().GetResourceType(),
	)
	if err = service.HandleDatabaseError(err, service.ErrNotFound("user permission")); err != nil {
		return nil, err
	}

	res = connect.NewResponse(&emptypb.Empty{})
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
		user orchestrator.User
	)

	// Validate the request
	if err = service.Validate(req); err != nil {
		return nil, err
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
		users []*orchestrator.User
		conds []any
		npt   string
	)

	// Validate request
	err = service.Validate(req)
	if err != nil {
		return nil, err
	}

	// JIT-provision the caller without enforcing authorization.
	if _, err = provisionCurrentUser(ctx, svc); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Set default ordering
	if req.Msg.OrderBy == "" {
		req.Msg.OrderBy = "id"
		req.Msg.Asc = true
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
	allowed, _, err = CheckAccess(ctx, svc.authz, svc, orchestrator.RequestType_REQUEST_TYPE_LIST, "", orchestrator.ObjectType_OBJECT_TYPE_USER_PERMISSION)
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

	// Filter by user_id if provided
	if userId := req.Msg.GetUserId(); userId != "" {
		conds = []any{"user_id = ?", userId}
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

// provisionCurrentUser extracts JWT claims from ctx and JIT-provisions the user in the database,
// returning the user ID. Returns an empty string when no claims are present.
func provisionCurrentUser(ctx context.Context, svc *Service) (string, error) {
	var (
		claims *auth.OAuthClaims
		user   *orchestrator.User
		userId string
		ok     bool
		err    error
	)

	claims, ok = auth.ClaimsFromContext(ctx)
	if !ok || claims == nil || claims.Issuer == "" || claims.Subject == "" {
		return "", nil
	}

	if svc == nil {
		return "", fmt.Errorf("service is nil")
	}
	if svc.db == nil {
		return "", fmt.Errorf("database is not initialized")
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
			Username:   new(claims.PreferredUsername),
			FirstName:  new(claims.GivenName),
			LastName:   new(claims.FamilyName),
			Enabled:    true,
			Email:      new(claims.Email),
			LastAccess: timestamppb.Now(),
		}
		err = svc.db.Create(user)
		if err != nil {
			return "", fmt.Errorf("failed to create user: %w", err)
		}
	} else if err != nil {
		return "", fmt.Errorf("failed to look up user: %w", err)
	} else {
		// User exists: update only identity fields and last_access to preserve other persisted
		// fields such as enabled status.
		user.Username = new(claims.PreferredUsername)
		user.FirstName = new(claims.GivenName)
		user.LastName = new(claims.FamilyName)
		user.Email = new(claims.Email)
		user.LastAccess = timestamppb.Now()
		err = svc.db.Save(user)
		if err != nil {
			return "", fmt.Errorf("failed to update user: %w", err)
		}
	}

	return userId, nil
}

// CheckAccess is a helper function to check if the user associated with the given context has access to perform the specified request type and request. It extracts user information from the JWT claims, ensures the user exists in the database, and then checks access using the provided authorization strategy.
func CheckAccess(ctx context.Context, authz service.AuthorizationStrategy, svc *Service, reqType orchestrator.RequestType, resourceId string, objectType orchestrator.ObjectType) (bool, []string, error) {
	var (
		userId      string
		err         error
		allowed     bool
		resourceIDs []string
	)

	// If JWT claims are present, provision the user in the DB (JIT provisioning) and use their ID.
	// If no claims are present, use an empty user ID — strategies like AuthorizationStrategyAllowAll
	// don't require a user ID, while AuthorizationStrategyPermissionStore will deny empty IDs.
	userId, err = provisionCurrentUser(ctx, svc)
	if err != nil {
		return false, nil, err
	}

	allowed, resourceIDs = authz.CheckAccess(ctx, userId, reqType, orchestrator.UserPermission_PERMISSION_READER, resourceId, objectType)

	return allowed, resourceIDs, nil
}

// grantCreatorAdminPermission persists an ADMIN [orchestrator.UserPermission] for the user who is
// making the current request (derived from JWT claims in ctx). It is called after a new resource
// has been created so that the creator immediately has full administrative access to that resource
// without requiring a separate permission update call.
//
// If no authenticated user can be determined from ctx (e.g. the context carries no claims, or the
// claims lack issuer/subject), the function is a no-op and returns nil. This preserves the
// existing allow-all behavior when authentication is disabled.
func grantCreatorAdminPermission(ctx context.Context, db persistence.DB, resourceId string, objectType orchestrator.ObjectType) (err error) {
	var (
		claims *auth.OAuthClaims
		ok     bool
		userId string
	)

	if db == nil {
		return fmt.Errorf("database is not initialized")
	}

	claims, ok = auth.ClaimsFromContext(ctx)
	if !ok {
		return nil
	}

	userId = auth.GetConfirmateUserIDFromClaims(claims)
	if userId == "" {
		return nil
	}

	err = db.Create(&orchestrator.UserPermission{
		UserId:       userId,
		ResourceId:   resourceId,
		ResourceType: objectType,
		Permission:   orchestrator.UserPermission_PERMISSION_ADMIN,
	})
	if err = service.HandleDatabaseError(err); err != nil {
		return err
	}

	return nil
}
