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
	"errors"
	"fmt"

	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/persistence"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/emptypb"
)

// RegisterAssessmentTool registers a new assessment tool.
func (svc *service) RegisterAssessmentTool(
	ctx context.Context,
	req *connect.Request[orchestrator.RegisterAssessmentToolRequest],
) (*connect.Response[orchestrator.AssessmentTool], error) {
	// Persist the new assessment tool in the database
	err := svc.db.Create(req.Msg.Tool)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("could not register assessment tool: %w", err))
	}

	return connect.NewResponse(req.Msg.Tool), nil
}

// GetAssessmentTool retrieves an assessment tool by ID.
func (svc *service) GetAssessmentTool(
	ctx context.Context,
	req *connect.Request[orchestrator.GetAssessmentToolRequest],
) (*connect.Response[orchestrator.AssessmentTool], error) {
	var res orchestrator.AssessmentTool

	err := svc.db.Get(&res, "id = ?", req.Msg.ToolId)
	if errors.Is(err, persistence.ErrRecordNotFound) {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("assessment tool not found"))
	} else if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("database error: %w", err))
	}

	return connect.NewResponse(&res), nil
}

// ListAssessmentTools lists all assessment tools.
func (svc *service) ListAssessmentTools(
	ctx context.Context,
	req *connect.Request[orchestrator.ListAssessmentToolsRequest],
) (*connect.Response[orchestrator.ListAssessmentToolsResponse], error) {
	var tools []*orchestrator.AssessmentTool

	err := svc.db.List(&tools, "id", true, 0, -1, nil)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("could not list assessment tools: %w", err))
	}

	return connect.NewResponse(&orchestrator.ListAssessmentToolsResponse{
		Tools: tools,
	}), nil
}

// UpdateAssessmentTool updates an existing assessment tool.
func (svc *service) UpdateAssessmentTool(
	ctx context.Context,
	req *connect.Request[orchestrator.UpdateAssessmentToolRequest],
) (*connect.Response[orchestrator.AssessmentTool], error) {
	// Check if the assessment tool exists
	count, err := svc.db.Count(req.Msg.Tool, "id = ?", req.Msg.Tool.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("database error: %w", err))
	}

	if count == 0 {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("assessment tool not found"))
	}

	// Save the updated assessment tool
	err = svc.db.Save(req.Msg.Tool, "id = ?", req.Msg.Tool.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("database error: %w", err))
	}

	return connect.NewResponse(req.Msg.Tool), nil
}

// DeregisterAssessmentTool removes an assessment tool by ID.
func (svc *service) DeregisterAssessmentTool(
	ctx context.Context,
	req *connect.Request[orchestrator.DeregisterAssessmentToolRequest],
) (*connect.Response[emptypb.Empty], error) {
	var tool orchestrator.AssessmentTool

	// Delete the assessment tool
	err := svc.db.Delete(&tool, "id = ?", req.Msg.ToolId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("database error: %w", err))
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}
