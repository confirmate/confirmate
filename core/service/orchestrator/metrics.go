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
)

// CreateMetric creates a new metric.
func (svc *Service) CreateMetric(
	ctx context.Context,
	req *connect.Request[orchestrator.CreateMetricRequest],
) (res *connect.Response[assessment.Metric], err error) {
	var (
		metric = req.Msg.Metric
	)

	// Persist the new metric in the database
	err = svc.db.Create(metric)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("could not add metric to the database: %w", err))
	}

	res = connect.NewResponse(metric)
	return
}

// GetMetric retrieves a metric by ID.
func (svc *Service) GetMetric(
	ctx context.Context,
	req *connect.Request[orchestrator.GetMetricRequest],
) (res *connect.Response[assessment.Metric], err error) {
	var (
		metric assessment.Metric
	)

	err = svc.db.Get(&metric, "id = ?", req.Msg.MetricId)
	if errors.Is(err, persistence.ErrRecordNotFound) {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("metric not found"))
	} else if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("database error: %w", err))
	}

	res = connect.NewResponse(&metric)
	return
}

// ListMetrics lists all metrics.
func (svc *Service) ListMetrics(
	ctx context.Context,
	req *connect.Request[orchestrator.ListMetricsRequest],
) (res *connect.Response[orchestrator.ListMetricsResponse], err error) {
	var (
		metrics []*assessment.Metric
	)

	err = svc.db.List(&metrics, "id", true, 0, -1, nil)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("could not list metrics: %w", err))
	}

	res = connect.NewResponse(&orchestrator.ListMetricsResponse{
		Metrics: metrics,
	})
	return
}

// UpdateMetric updates an existing metric.
func (svc *Service) UpdateMetric(
	ctx context.Context,
	req *connect.Request[orchestrator.UpdateMetricRequest],
) (res *connect.Response[assessment.Metric], err error) {
	var (
		count  int64
		metric = req.Msg.Metric
	)

	// Check if the metric exists
	count, err = svc.db.Count(metric, "id = ?", metric.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("database error: %w", err))
	}

	if count == 0 {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("metric not found"))
	}

	// Save the updated metric
	err = svc.db.Save(metric, "id = ?", metric.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("database error: %w", err))
	}

	res = connect.NewResponse(metric)
	return
}

// RemoveMetric removes a metric by ID.
func (svc *Service) RemoveMetric(
	ctx context.Context,
	req *connect.Request[orchestrator.RemoveMetricRequest],
) (res *connect.Response[emptypb.Empty], err error) {
	var (
		metric assessment.Metric
	)

	// Delete the metric
	err = svc.db.Delete(&metric, "id = ?", req.Msg.MetricId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("database error: %w", err))
	}

	res = connect.NewResponse(&emptypb.Empty{})
	return
}

// GetMetricImplementation retrieves a metric implementation by metric ID.
func (svc *Service) GetMetricImplementation(
	ctx context.Context,
	req *connect.Request[orchestrator.GetMetricImplementationRequest],
) (res *connect.Response[assessment.MetricImplementation], err error) {
	var (
		impl assessment.MetricImplementation
	)

	err = svc.db.Get(&impl, "metric_id = ?", req.Msg.MetricId)
	if errors.Is(err, persistence.ErrRecordNotFound) {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("metric implementation not found"))
	} else if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("database error: %w", err))
	}

	res = connect.NewResponse(&impl)
	return
}

// UpdateMetricImplementation updates an existing metric implementation.
func (svc *Service) UpdateMetricImplementation(
	ctx context.Context,
	req *connect.Request[orchestrator.UpdateMetricImplementationRequest],
) (res *connect.Response[assessment.MetricImplementation], err error) {
	var (
		count int64
		impl  = req.Msg.Implementation
	)

	// Check if the metric implementation exists
	count, err = svc.db.Count(impl, "metric_id = ?", impl.MetricId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("database error: %w", err))
	}

	if count == 0 {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("metric implementation not found"))
	}

	// Save the updated metric implementation
	err = svc.db.Save(impl, "metric_id = ?", impl.MetricId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("database error: %w", err))
	}

	res = connect.NewResponse(impl)
	return
}

// GetMetricConfiguration retrieves a metric configuration for a specific TOE and metric.
func (svc *Service) GetMetricConfiguration(
	ctx context.Context,
	req *connect.Request[orchestrator.GetMetricConfigurationRequest],
) (res *connect.Response[assessment.MetricConfiguration], err error) {
	var (
		config assessment.MetricConfiguration
	)

	// Use WithoutPreload because MetricConfiguration contains structpb.Value which has unexported fields
	err = svc.db.Get(&config, persistence.WithoutPreload(), "target_of_evaluation_id = ? AND metric_id = ?",
		req.Msg.TargetOfEvaluationId, req.Msg.MetricId)
	if errors.Is(err, persistence.ErrRecordNotFound) {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("metric configuration not found"))
	} else if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("database error: %w", err))
	}

	res = connect.NewResponse(&config)
	return
}

// ListMetricConfigurations lists all metric configurations for a specific TOE.
func (svc *Service) ListMetricConfigurations(
	ctx context.Context,
	req *connect.Request[orchestrator.ListMetricConfigurationRequest],
) (res *connect.Response[orchestrator.ListMetricConfigurationResponse], err error) {
	var (
		configs   []*assessment.MetricConfiguration
		configMap = make(map[string]*assessment.MetricConfiguration)
	)

	// Use WithoutPreload because MetricConfiguration contains structpb.Value which has unexported fields
	err = svc.db.List(&configs, "metric_id", true, 0, -1,
		persistence.WithoutPreload(), "target_of_evaluation_id = ?", req.Msg.TargetOfEvaluationId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("could not list metric configurations: %w", err))
	}

	// Convert slice to map indexed by metric ID
	for _, config := range configs {
		configMap[config.MetricId] = config
	}

	res = connect.NewResponse(&orchestrator.ListMetricConfigurationResponse{
		Configurations: configMap,
	})
	return
}

// UpdateMetricConfiguration updates a metric configuration for a specific TOE and metric.
func (svc *Service) UpdateMetricConfiguration(
	ctx context.Context,
	req *connect.Request[orchestrator.UpdateMetricConfigurationRequest],
) (res *connect.Response[assessment.MetricConfiguration], err error) {
	var (
		count  int64
		config = req.Msg.Configuration
	)

	// Check if the metric configuration exists
	count, err = svc.db.Count(config, "target_of_evaluation_id = ? AND metric_id = ?",
		req.Msg.TargetOfEvaluationId, req.Msg.MetricId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("database error: %w", err))
	}

	if count == 0 {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("metric configuration not found"))
	}

	// Ensure IDs match
	config.TargetOfEvaluationId = req.Msg.TargetOfEvaluationId
	config.MetricId = req.Msg.MetricId

	// Save the updated metric configuration
	err = svc.db.Save(config, "target_of_evaluation_id = ? AND metric_id = ?",
		req.Msg.TargetOfEvaluationId, req.Msg.MetricId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("database error: %w", err))
	}

	res = connect.NewResponse(config)
	return
}
