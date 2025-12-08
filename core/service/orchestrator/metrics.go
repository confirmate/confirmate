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
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"confirmate.io/core/api/assessment"
	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/persistence"
	"confirmate.io/core/service"

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

	// Validate the request
	if err = service.Validate(req.Msg); err != nil {
		return nil, err
	}

	// Persist the new metric in the database
	err = svc.db.Create(metric)
	if err = service.HandleDatabaseError(err); err != nil {
		return nil, err
	}

	// Notify subscribers
	go svc.publishEvent(&orchestrator.ChangeEvent{
		Type: orchestrator.ChangeEvent_TYPE_METRIC_CHANGE,
		Event: &orchestrator.ChangeEvent_MetricChange{
			MetricChange: &orchestrator.MetricChangeEvent{
				Type:     orchestrator.MetricChangeEvent_TYPE_METADATA_CHANGED,
				MetricId: metric.Id,
			},
		},
	})

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

	// Validate the request
	if err = service.Validate(req.Msg); err != nil {
		return nil, err
	}

	err = svc.db.Get(&metric, "id = ?", req.Msg.MetricId)
	if err = service.HandleDatabaseError(err, service.ErrNotFound("metric")); err != nil {
		return nil, err
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

	// Validate the request
	if err = service.Validate(req.Msg); err != nil {
		return nil, err
	}

	err = svc.db.List(&metrics, "id", true, 0, -1, nil)
	if err = service.HandleDatabaseError(err); err != nil {
		return nil, err
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

	// Validate the request
	if err = service.Validate(req.Msg); err != nil {
		return nil, err
	}

	// Check if the metric exists
	count, err = svc.db.Count(metric, "id = ?", metric.Id)
	if err = service.HandleDatabaseError(err); err != nil {
		return nil, err
	}

	if count == 0 {
		return nil, service.ErrNotFound("metric")
	}

	// Save the updated metric
	err = svc.db.Save(metric, "id = ?", metric.Id)
	if err = service.HandleDatabaseError(err); err != nil {
		return nil, err
	}

	// Notify subscribers
	go svc.publishEvent(&orchestrator.ChangeEvent{
		Type: orchestrator.ChangeEvent_TYPE_METRIC_CHANGE,
		Event: &orchestrator.ChangeEvent_MetricChange{
			MetricChange: &orchestrator.MetricChangeEvent{
				Type:     orchestrator.MetricChangeEvent_TYPE_METADATA_CHANGED,
				MetricId: metric.Id,
			},
		},
	})

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

	// Validate the request
	if err = service.Validate(req.Msg); err != nil {
		return nil, err
	}

	// Delete the metric
	err = svc.db.Delete(&metric, "id = ?", req.Msg.MetricId)
	if err = service.HandleDatabaseError(err, service.ErrNotFound("metric")); err != nil {
		return nil, err
	}

	// Notify subscribers
	go svc.publishEvent(&orchestrator.ChangeEvent{
		Type: orchestrator.ChangeEvent_TYPE_METRIC_CHANGE,
		Event: &orchestrator.ChangeEvent_MetricChange{
			MetricChange: &orchestrator.MetricChangeEvent{
				Type:     orchestrator.MetricChangeEvent_TYPE_METADATA_CHANGED,
				MetricId: req.Msg.MetricId,
			},
		},
	})

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

	// Validate the request
	if err = service.Validate(req.Msg); err != nil {
		return nil, err
	}

	err = svc.db.Get(&impl, "metric_id = ?", req.Msg.MetricId)
	if err = service.HandleDatabaseError(err, service.ErrNotFound("metric implementation")); err != nil {
		return nil, err
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

	// Validate the request
	if err = service.Validate(req.Msg); err != nil {
		return nil, err
	}

	// Check if the metric implementation exists
	count, err = svc.db.Count(impl, "metric_id = ?", impl.MetricId)
	if err = service.HandleDatabaseError(err); err != nil {
		return nil, err
	}

	if count == 0 {
		return nil, service.ErrNotFound("metric implementation")
	}

	// Save the updated metric implementation
	err = svc.db.Save(impl, "metric_id = ?", impl.MetricId)
	if err = service.HandleDatabaseError(err); err != nil {
		return nil, err
	}

	// Notify subscribers
	go svc.publishEvent(&orchestrator.ChangeEvent{
		Type: orchestrator.ChangeEvent_TYPE_METRIC_CHANGE,
		Event: &orchestrator.ChangeEvent_MetricChange{
			MetricChange: &orchestrator.MetricChangeEvent{
				Type:     orchestrator.MetricChangeEvent_TYPE_IMPLEMENTATION_CHANGED,
				MetricId: impl.MetricId,
			},
		},
	})

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

	// Validate the request
	if err = service.Validate(req.Msg); err != nil {
		return nil, err
	}

	// Use WithoutPreload because MetricConfiguration contains structpb.Value which has unexported fields
	err = svc.db.Get(&config, persistence.WithoutPreload(), "target_of_evaluation_id = ? AND metric_id = ?",
		req.Msg.TargetOfEvaluationId, req.Msg.MetricId)
	if err = service.HandleDatabaseError(err, service.ErrNotFound("metric configuration")); err != nil {
		return nil, err
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

	// Validate the request
	if err = service.Validate(req.Msg); err != nil {
		return nil, err
	}

	// Use WithoutPreload because MetricConfiguration contains structpb.Value which has unexported fields
	err = svc.db.List(&configs, "metric_id", true, 0, -1,
		persistence.WithoutPreload(), "target_of_evaluation_id = ?", req.Msg.TargetOfEvaluationId)
	if err = service.HandleDatabaseError(err); err != nil {
		return nil, err
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
		config = req.Msg.Configuration
	)

	// Validate the request
	if err = service.Validate(req.Msg); err != nil {
		return nil, err
	}

	// Ensure IDs match
	config.TargetOfEvaluationId = req.Msg.TargetOfEvaluationId
	config.MetricId = req.Msg.MetricId

	// Save the updated metric configuration
	err = svc.db.Save(config)
	if err = service.HandleDatabaseError(err); err != nil {
		return nil, err
	}

	// Notify subscribers
	go svc.publishEvent(&orchestrator.ChangeEvent{
		Type: orchestrator.ChangeEvent_TYPE_METRIC_CHANGE,
		Event: &orchestrator.ChangeEvent_MetricChange{
			MetricChange: &orchestrator.MetricChangeEvent{
				Type:                 orchestrator.MetricChangeEvent_TYPE_CONFIG_CHANGED,
				MetricId:             config.MetricId,
				TargetOfEvaluationId: config.TargetOfEvaluationId,
			},
		},
	})

	res = connect.NewResponse(config)
	return
}

// loadMetrics loads metric definitions from a JSON file.
func (svc *Service) loadMetrics() (err error) {
	var metrics []*assessment.Metric

	if svc.metricsFolder == "" {
		return nil
	}

	// Get all filenames
	files, err := os.ReadDir(svc.metricsFolder)
	if err != nil {
		return fmt.Errorf("could not read metrics folder: %w", err)
	}

	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
			continue
		}

		var metricsFromFile []*assessment.Metric
		b, err := os.ReadFile(filepath.Join(svc.metricsFolder, file.Name()))
		if err != nil {
			// log error?
			continue
		}

		err = json.Unmarshal(b, &metricsFromFile)
		if err != nil {
			// log error?
			continue
		}

		metrics = append(metrics, metricsFromFile...)
	}

	// Save to DB
	return svc.db.Save(metrics)
}

