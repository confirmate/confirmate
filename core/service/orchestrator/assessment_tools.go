// Copyright 2016-2025 Fraunhofer AISEC
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

	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/service"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// RegisterAssessmentTool registers a new assessment tool.
func (svc *Service) RegisterAssessmentTool(
	ctx context.Context,
	req *connect.Request[orchestrator.RegisterAssessmentToolRequest],
) (res *connect.Response[orchestrator.AssessmentTool], err error) {
	var (
		tool *orchestrator.AssessmentTool
	)

	// Validate the request
	if err = service.Validate(req); err != nil {
		return nil, err
	}

	tool = req.Msg.Tool

	// Persist the new assessment tool in the database
	err = svc.db.Create(tool)
	if err = service.HandleDatabaseError(err); err != nil {
		return nil, err
	}

	// Notify subscribers
	go svc.publishEvent(&orchestrator.ChangeEvent{
		Timestamp:  timestamppb.Now(),
		Category:   orchestrator.EventCategory_EVENT_CATEGORY_ASSESSMENT_TOOL,
		ChangeType: orchestrator.ChangeType_CHANGE_TYPE_CREATED,
		EntityId:   tool.Id,
		Entity: &orchestrator.ChangeEvent_AssessmentTool{
			AssessmentTool: tool,
		},
	})

	res = connect.NewResponse(tool)
	return
}

// GetAssessmentTool retrieves an assessment tool by ID.
func (svc *Service) GetAssessmentTool(
	ctx context.Context,
	req *connect.Request[orchestrator.GetAssessmentToolRequest],
) (res *connect.Response[orchestrator.AssessmentTool], err error) {
	var (
		tool orchestrator.AssessmentTool
	)

	// Validate the request
	if err = service.Validate(req); err != nil {
		return nil, err
	}

	err = svc.db.Get(&tool, "id = ?", req.Msg.ToolId)
	if err = service.HandleDatabaseError(err, service.ErrNotFound("assessment tool")); err != nil {
		return nil, err
	}

	res = connect.NewResponse(&tool)
	return
}

// ListAssessmentTools lists all assessment tools.
func (svc *Service) ListAssessmentTools(
	ctx context.Context,
	req *connect.Request[orchestrator.ListAssessmentToolsRequest],
) (res *connect.Response[orchestrator.ListAssessmentToolsResponse], err error) {
	var (
		tools []*orchestrator.AssessmentTool
		npt   string
	)

	err = service.Validate(req)
	if err != nil {
		return nil, err
	}

	// Set default ordering
	if req.Msg.OrderBy == "" {
		req.Msg.OrderBy = "id"
		req.Msg.Asc = true
	}

	tools, npt, err = service.PaginateStorage[*orchestrator.AssessmentTool](req.Msg, svc.db, service.DefaultPaginationOpts)
	if err = service.HandleDatabaseError(err); err != nil {
		return nil, err
	}

	res = connect.NewResponse(&orchestrator.ListAssessmentToolsResponse{
		Tools:         tools,
		NextPageToken: npt,
	})
	return
}

// UpdateAssessmentTool updates an existing assessment tool.
func (svc *Service) UpdateAssessmentTool(
	ctx context.Context,
	req *connect.Request[orchestrator.UpdateAssessmentToolRequest],
) (res *connect.Response[orchestrator.AssessmentTool], err error) {
	var tool = req.Msg.Tool

	// Validate the request
	if err = service.Validate(req); err != nil {
		return nil, err
	}

	// Update the assessment tool
	err = svc.db.Update(tool, "id = ?", tool.Id)
	if err = service.HandleDatabaseError(err, service.ErrNotFound("assessment tool")); err != nil {
		return nil, err
	}

	// Notify subscribers
	go svc.publishEvent(&orchestrator.ChangeEvent{
		Timestamp:  timestamppb.Now(),
		Category:   orchestrator.EventCategory_EVENT_CATEGORY_ASSESSMENT_TOOL,
		ChangeType: orchestrator.ChangeType_CHANGE_TYPE_UPDATED,
		EntityId:   tool.Id,
		Entity: &orchestrator.ChangeEvent_AssessmentTool{
			AssessmentTool: tool,
		},
	})

	res = connect.NewResponse(tool)
	return
}

// DeregisterAssessmentTool removes an assessment tool by ID.
func (svc *Service) DeregisterAssessmentTool(
	ctx context.Context,
	req *connect.Request[orchestrator.DeregisterAssessmentToolRequest],
) (res *connect.Response[emptypb.Empty], err error) {
	var (
		tool orchestrator.AssessmentTool
	)

	// Validate the request
	if err = service.Validate(req); err != nil {
		return nil, err
	}

	// Delete the assessment tool
	err = svc.db.Delete(&tool, "id = ?", req.Msg.ToolId)
	if err = service.HandleDatabaseError(err, service.ErrNotFound("assessment tool")); err != nil {
		return nil, err
	}

	// Notify subscribers
	go svc.publishEvent(&orchestrator.ChangeEvent{
		Timestamp:  timestamppb.Now(),
		Category:   orchestrator.EventCategory_EVENT_CATEGORY_ASSESSMENT_TOOL,
		ChangeType: orchestrator.ChangeType_CHANGE_TYPE_DELETED,
		EntityId:   req.Msg.ToolId,
	})

	res = connect.NewResponse(&emptypb.Empty{})
	return
}
