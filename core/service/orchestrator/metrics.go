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

	"confirmate.io/core/api/assessment"
	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/persistence"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
)

// CreateMetric creates a new metric.
func (svc *service) CreateMetric(
	ctx context.Context,
	req *connect.Request[orchestrator.CreateMetricRequest],
) (*connect.Response[assessment.Metric], error) {
	// Persist the new metric in the database
	err := svc.db.Create(req.Msg.Metric)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("could not add metric to the database: %w", err))
	}

	return connect.NewResponse(req.Msg.Metric), nil
}

// GetMetric retrieves a metric by ID.
func (svc *service) GetMetric(
	ctx context.Context,
	req *connect.Request[orchestrator.GetMetricRequest],
) (*connect.Response[assessment.Metric], error) {
	var res assessment.Metric

	err := svc.db.Get(&res, "id = ?", req.Msg.MetricId)
	if errors.Is(err, persistence.ErrRecordNotFound) {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("metric not found"))
	} else if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("database error: %w", err))
	}

	return connect.NewResponse(&res), nil
}

// ListMetrics lists all metrics.
func (svc *service) ListMetrics(
	ctx context.Context,
	req *connect.Request[orchestrator.ListMetricsRequest],
) (*connect.Response[orchestrator.ListMetricsResponse], error) {
	var metrics []*assessment.Metric

	err := svc.db.List(&metrics, "id", true, 0, -1, nil)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("could not list metrics: %w", err))
	}

	return connect.NewResponse(&orchestrator.ListMetricsResponse{
		Metrics: metrics,
	}), nil
}

// UpdateMetric updates an existing metric.
func (svc *service) UpdateMetric(
	ctx context.Context,
	req *connect.Request[orchestrator.UpdateMetricRequest],
) (*connect.Response[assessment.Metric], error) {
	// Check if the metric exists
	count, err := svc.db.Count(req.Msg.Metric, "id = ?", req.Msg.Metric.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("database error: %w", err))
	}

	if count == 0 {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("metric not found"))
	}

	// Save the updated metric
	err = svc.db.Save(req.Msg.Metric, "id = ?", req.Msg.Metric.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("database error: %w", err))
	}

	return connect.NewResponse(req.Msg.Metric), nil
}

// RemoveMetric removes a metric by ID.
func (svc *service) RemoveMetric(
	ctx context.Context,
	req *connect.Request[orchestrator.RemoveMetricRequest],
) (*connect.Response[emptypb.Empty], error) {
	var metric assessment.Metric

	// Delete the metric
	err := svc.db.Delete(&metric, "id = ?", req.Msg.MetricId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("database error: %w", err))
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}

// GetMetricImplementation retrieves a metric implementation by metric ID.
func (svc *service) GetMetricImplementation(
	ctx context.Context,
	req *connect.Request[orchestrator.GetMetricImplementationRequest],
) (*connect.Response[assessment.MetricImplementation], error) {
	var res assessment.MetricImplementation

	err := svc.db.Get(&res, "metric_id = ?", req.Msg.MetricId)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("metric implementation not found"))
	} else if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("database error: %w", err))
	}

	return connect.NewResponse(&res), nil
}

// UpdateMetricImplementation updates an existing metric implementation.
func (svc *service) UpdateMetricImplementation(
	ctx context.Context,
	req *connect.Request[orchestrator.UpdateMetricImplementationRequest],
) (*connect.Response[assessment.MetricImplementation], error) {
	// Check if the metric implementation exists
	count, err := svc.db.Count(req.Msg.Implementation, "metric_id = ?", req.Msg.Implementation.MetricId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("database error: %w", err))
	}

	if count == 0 {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("metric implementation not found"))
	}

	// Save the updated metric implementation
	err = svc.db.Save(req.Msg.Implementation, "metric_id = ?", req.Msg.Implementation.MetricId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("database error: %w", err))
	}

	return connect.NewResponse(req.Msg.Implementation), nil
}

// GetMetricConfiguration retrieves a metric configuration for a specific TOE and metric.
func (svc *service) GetMetricConfiguration(
	ctx context.Context,
	req *connect.Request[orchestrator.GetMetricConfigurationRequest],
) (*connect.Response[assessment.MetricConfiguration], error) {
	var res assessment.MetricConfiguration

	err := svc.db.Get(&res, "target_of_evaluation_id = ? AND metric_id = ?",
		req.Msg.TargetOfEvaluationId, req.Msg.MetricId)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("metric configuration not found"))
	} else if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("database error: %w", err))
	}

	return connect.NewResponse(&res), nil
}

// ListMetricConfigurations lists all metric configurations for a specific TOE.
func (svc *service) ListMetricConfigurations(
	ctx context.Context,
	req *connect.Request[orchestrator.ListMetricConfigurationRequest],
) (*connect.Response[orchestrator.ListMetricConfigurationResponse], error) {
	var configs []*assessment.MetricConfiguration

	err := svc.db.List(&configs, "metric_id", true, 0, -1,
		"target_of_evaluation_id = ?", req.Msg.TargetOfEvaluationId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("could not list metric configurations: %w", err))
	}

	// Convert slice to map indexed by metric ID
	configMap := make(map[string]*assessment.MetricConfiguration)
	for _, config := range configs {
		configMap[config.MetricId] = config
	}

	return connect.NewResponse(&orchestrator.ListMetricConfigurationResponse{
		Configurations: configMap,
	}), nil
}

// UpdateMetricConfiguration updates a metric configuration for a specific TOE and metric.
func (svc *service) UpdateMetricConfiguration(
	ctx context.Context,
	req *connect.Request[orchestrator.UpdateMetricConfigurationRequest],
) (*connect.Response[assessment.MetricConfiguration], error) {
	// Check if the metric configuration exists
	count, err := svc.db.Count(req.Msg.Configuration, "target_of_evaluation_id = ? AND metric_id = ?",
		req.Msg.TargetOfEvaluationId, req.Msg.MetricId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("database error: %w", err))
	}

	if count == 0 {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("metric configuration not found"))
	}

	// Ensure IDs match
	req.Msg.Configuration.TargetOfEvaluationId = req.Msg.TargetOfEvaluationId
	req.Msg.Configuration.MetricId = req.Msg.MetricId

	// Save the updated metric configuration
	err = svc.db.Save(req.Msg.Configuration, "target_of_evaluation_id = ? AND metric_id = ?",
		req.Msg.TargetOfEvaluationId, req.Msg.MetricId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("database error: %w", err))
	}

	return connect.NewResponse(req.Msg.Configuration), nil
}
