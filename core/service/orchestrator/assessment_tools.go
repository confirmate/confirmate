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
)

// RegisterAssessmentTool registers a new assessment tool.
func (svc *Service) RegisterAssessmentTool(
	ctx context.Context,
	req *connect.Request[orchestrator.RegisterAssessmentToolRequest],
) (res *connect.Response[orchestrator.AssessmentTool], err error) {
	var (
		tool = req.Msg.Tool
	)

	// Persist the new assessment tool in the database
	err = svc.db.Create(tool)
	if err = service.HandleDatabaseError(err); err != nil {
		return nil, err
	}

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
	)

	err = svc.db.List(&tools, "id", true, 0, -1, nil)
	if err = service.HandleDatabaseError(err); err != nil {
		return nil, err
	}

	res = connect.NewResponse(&orchestrator.ListAssessmentToolsResponse{
		Tools: tools,
	})
	return
}

// UpdateAssessmentTool updates an existing assessment tool.
func (svc *Service) UpdateAssessmentTool(
	ctx context.Context,
	req *connect.Request[orchestrator.UpdateAssessmentToolRequest],
) (res *connect.Response[orchestrator.AssessmentTool], err error) {
	var (
		count int64
		tool  = req.Msg.Tool
	)

	// Check if the assessment tool exists
	count, err = svc.db.Count(tool, "id = ?", tool.Id)
	if err = service.HandleDatabaseError(err); err != nil {
		return nil, err
	}

	if count == 0 {
		return nil, service.ErrNotFound("assessment tool")
	}

	// Save the updated assessment tool
	err = svc.db.Save(tool, "id = ?", tool.Id)
	if err = service.HandleDatabaseError(err); err != nil {
		return nil, err
	}

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

	// Delete the assessment tool
	err = svc.db.Delete(&tool, "id = ?", req.Msg.ToolId)
	if err = service.HandleDatabaseError(err, service.ErrNotFound("assessment tool")); err != nil {
		return nil, err
	}

	res = connect.NewResponse(&emptypb.Empty{})
	return
}
