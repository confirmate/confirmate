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
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"confirmate.io/core/api/assessment"
	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/log"
	"confirmate.io/core/persistence"
	"confirmate.io/core/service"
	"confirmate.io/core/util"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gopkg.in/yaml.v3"
)

var (
	// defaultMetricConfigurations holds the default configurations loaded from data.json files
	defaultMetricConfigurations = make(map[string]*assessment.MetricConfiguration)
)

// CreateMetric creates a new metric.
func (svc *Service) CreateMetric(
	ctx context.Context,
	req *connect.Request[orchestrator.CreateMetricRequest],
) (res *connect.Response[assessment.Metric], err error) {
	var (
		metric *assessment.Metric
	)

	// Validate the request
	if err = service.Validate(req); err != nil {
		return nil, err
	}

	metric = req.Msg.Metric

	// Persist the new metric in the database
	err = svc.db.Create(metric)
	if err = service.HandleDatabaseError(err); err != nil {
		return nil, err
	}

	// Notify subscribers
	go svc.publishEvent(&orchestrator.ChangeEvent{
		Timestamp:   timestamppb.Now(),
		Category:    orchestrator.EventCategory_EVENT_CATEGORY_METRIC,
		RequestType: orchestrator.RequestType_REQUEST_TYPE_CREATED,
		EntityId:    metric.Id,
		Entity: &orchestrator.ChangeEvent_Metric{
			Metric: metric,
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
	if err = service.Validate(req); err != nil {
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
		npt     string
	)

	// Validate the request
	if err = service.Validate(req); err != nil {
		return nil, err
	}

	// Set default ordering
	if req.Msg.OrderBy == "" {
		req.Msg.OrderBy = "id"
		req.Msg.Asc = true
	}

	metrics, npt, err = service.PaginateStorage[*assessment.Metric](req.Msg, svc.db, service.DefaultPaginationOpts)
	if err = service.HandleDatabaseError(err); err != nil {
		return nil, err
	}

	res = connect.NewResponse(&orchestrator.ListMetricsResponse{
		Metrics:       metrics,
		NextPageToken: npt,
	})
	return
}

// UpdateMetric updates an existing metric.
func (svc *Service) UpdateMetric(
	ctx context.Context,
	req *connect.Request[orchestrator.UpdateMetricRequest],
) (res *connect.Response[assessment.Metric], err error) {
	var metric *assessment.Metric

	// Validate the request
	if err = service.Validate(req); err != nil {
		return nil, err
	}

	metric = req.Msg.Metric

	// Update the metric
	err = svc.db.Update(metric, "id = ?", metric.Id)
	if err = service.HandleDatabaseError(err, service.ErrNotFound("metric")); err != nil {
		return nil, err
	}

	// Notify subscribers
	go svc.publishEvent(&orchestrator.ChangeEvent{
		Timestamp:   timestamppb.Now(),
		Category:    orchestrator.EventCategory_EVENT_CATEGORY_METRIC,
		RequestType: orchestrator.RequestType_REQUEST_TYPE_UPDATED,
		EntityId:    metric.Id,
		Entity: &orchestrator.ChangeEvent_Metric{
			Metric: metric,
		},
	})

	res = connect.NewResponse(metric)
	return
}

// RemoveMetric removes a metric by ID. The metric is not deleted for backward compatibility,
// but the deprecated_since field is set to the current timestamp.
func (svc *Service) RemoveMetric(
	ctx context.Context,
	req *connect.Request[orchestrator.RemoveMetricRequest],
) (res *connect.Response[emptypb.Empty], err error) {
	var (
		metric *assessment.Metric
	)

	// Validate the request
	if err = service.Validate(req); err != nil {
		return nil, err
	}

	// Check if metric exists
	err = svc.db.Get(&metric, "id = ?", req.Msg.MetricId)
	if err = service.HandleDatabaseError(err, service.ErrNotFound("metric")); err != nil {
		return nil, err
	}

	// Set timestamp if not already set (soft delete)
	if metric.DeprecatedSince == nil {
		metric.DeprecatedSince = timestamppb.Now()
	}

	// Update the metric with the deprecated timestamp
	err = svc.db.Update(metric, "id = ?", req.Msg.MetricId)
	if err = service.HandleDatabaseError(err); err != nil {
		return nil, err
	}

	// Notify subscribers
	go svc.publishEvent(&orchestrator.ChangeEvent{
		Timestamp:   timestamppb.Now(),
		Category:    orchestrator.EventCategory_EVENT_CATEGORY_METRIC,
		RequestType: orchestrator.RequestType_REQUEST_TYPE_DELETED,
		EntityId:    req.Msg.MetricId,
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
	if err = service.Validate(req); err != nil {
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
	var impl *assessment.MetricImplementation

	// Validate the request
	if err = service.Validate(req); err != nil {
		return nil, err
	}

	impl = req.Msg.Implementation

	// Update the metric implementation
	err = svc.db.Update(impl, "metric_id = ?", impl.MetricId)
	if err = service.HandleDatabaseError(err, service.ErrNotFound("metric implementation")); err != nil {
		return nil, err
	}

	// Notify subscribers
	go svc.publishEvent(&orchestrator.ChangeEvent{
		Timestamp:   timestamppb.Now(),
		Category:    orchestrator.EventCategory_EVENT_CATEGORY_METRIC_IMPLEMENTATION,
		RequestType: orchestrator.RequestType_REQUEST_TYPE_UPDATED,
		EntityId:    impl.MetricId,
		Entity: &orchestrator.ChangeEvent_MetricImplementation{
			MetricImplementation: impl,
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
	if err = service.Validate(req); err != nil {
		return nil, err
	}

	// Use WithoutPreload because MetricConfiguration contains structpb.Value which has unexported fields
	err = svc.db.Get(&config, persistence.WithoutPreload(), "target_of_evaluation_id = ? AND metric_id = ?",
		req.Msg.TargetOfEvaluationId, req.Msg.MetricId)
	if err != nil {
		// If not found in DB, fall back to default configuration
		if errors.Is(err, persistence.ErrRecordNotFound) {
			if defaultConfig, ok := defaultMetricConfigurations[req.Msg.MetricId]; ok {
				// Copy the default configuration and set the target of evaluation ID
				config = assessment.MetricConfiguration{
					Operator:             defaultConfig.Operator,
					TargetValue:          defaultConfig.TargetValue,
					IsDefault:            true,
					MetricId:             req.Msg.MetricId,
					TargetOfEvaluationId: req.Msg.TargetOfEvaluationId,
				}
				res = connect.NewResponse(&config)
				return res, nil
			}
		}

		return nil, service.HandleDatabaseError(err, service.ErrNotFound("metric configuration"))
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
		npt       string
	)

	// Validate the request
	if err = service.Validate(req); err != nil {
		return nil, err
	}

	// Set default ordering
	if req.Msg.OrderBy == "" {
		req.Msg.OrderBy = "metric_id"
		req.Msg.Asc = true
	}

	// Use WithoutPreload because MetricConfiguration contains structpb.Value which has unexported fields
	configs, npt, err = service.PaginateStorage[*assessment.MetricConfiguration](req.Msg, svc.db, service.DefaultPaginationOpts,
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
		NextPageToken:  npt,
	})
	return
}

// UpdateMetricConfiguration updates a metric configuration for a specific TOE and metric.
func (svc *Service) UpdateMetricConfiguration(
	ctx context.Context,
	req *connect.Request[orchestrator.UpdateMetricConfigurationRequest],
) (res *connect.Response[assessment.MetricConfiguration], err error) {
	var (
		config *assessment.MetricConfiguration
	)

	// Validate the request
	if err = service.Validate(req); err != nil {
		return nil, err
	}

	config = req.Msg.Configuration

	// Save the updated metric configuration
	err = svc.db.Save(config)
	if err = service.HandleDatabaseError(err); err != nil {
		return nil, err
	}

	// Notify subscribers
	go svc.publishEvent(&orchestrator.ChangeEvent{
		Timestamp:            timestamppb.Now(),
		Category:             orchestrator.EventCategory_EVENT_CATEGORY_METRIC_CONFIGURATION,
		RequestType:          orchestrator.RequestType_REQUEST_TYPE_UPDATED,
		EntityId:             config.MetricId,
		TargetOfEvaluationId: util.Ref(config.TargetOfEvaluationId),
		Entity: &orchestrator.ChangeEvent_MetricConfiguration{
			MetricConfiguration: config,
		},
	})

	res = connect.NewResponse(config)
	return
}

// loadMetrics loads metric definitions from configured sources.
// It loads metrics from:
// 1. DefaultMetricsPath (if LoadDefaultMetrics is true) - typically the security-metrics repository
// 2. LoadMetricsFunc (if provided) for additional custom metrics
func (svc *Service) loadMetrics() (err error) {
	var metrics []*assessment.Metric

	// Load default metrics from repository if enabled
	if svc.cfg.LoadDefaultMetrics {
		defaultMetrics, err := svc.loadMetricsFromRepository()
		if err != nil {
			return fmt.Errorf("could not load default metrics: %w", err)
		}
		metrics = append(metrics, defaultMetrics...)
	}

	// Load additional metrics from custom function if provided
	if svc.cfg.LoadMetricsFunc != nil {
		additionalMetrics, err := svc.cfg.LoadMetricsFunc(svc)
		if err != nil {
			return fmt.Errorf("could not load additional metrics: %w", err)
		}
		metrics = append(metrics, additionalMetrics...)
	}

	// Save all metrics to DB (only if we have any)
	if len(metrics) > 0 {
		return svc.db.Save(metrics)
	}

	return nil
}

// loadMetricsFromRepository loads metric definitions from the security-metrics submodule
// by walking through YAML files in the policies/security-metrics/metrics directory.
func (svc *Service) loadMetricsFromRepository() (metrics []*assessment.Metric, err error) {
	metrics = make([]*assessment.Metric, 0)

	// Check if the directory exists (it might not in test environments)
	if _, err := os.Stat(svc.cfg.DefaultMetricsPath); os.IsNotExist(err) {
		// Return empty metrics list if directory doesn't exist (e.g., in tests)
		return metrics, nil
	}

	// Walk through the security-metrics repository and import metrics
	err = filepath.Walk(svc.cfg.DefaultMetricsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error accessing path %s: %w", path, err)
		}

		// Skip directories and non-yaml files
		if info.IsDir() || (!strings.HasSuffix(info.Name(), ".yaml") && !strings.HasSuffix(info.Name(), ".yml")) {
			return nil
		}

		// Read the YAML file
		b, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("error reading file %s: %w", path, err)
		}

		var metric assessment.Metric

		dec := yaml.NewDecoder(bytes.NewReader(b))
		err = dec.Decode(&metric)
		if err != nil {
			return fmt.Errorf("error decoding metric %s: %w", path, err)
		}

		metrics = append(metrics, &metric)

		// Load default configuration from data.json if it exists
		if err := prepareMetric(&metric, path); err != nil {
			slog.Warn("Could not prepare metric", "metric", metric.Id, log.Err(err))
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking through metrics directory: %w", err)
	}

	return metrics, nil
}

// prepareMetric loads the default configuration for a metric from its data.json file.
func prepareMetric(m *assessment.Metric, metricPath string) (err error) {
	var (
		config *assessment.MetricConfiguration
	)

	// Get the directory of the metric file
	metricDir := filepath.Dir(metricPath)

	// Construct path to data.json
	dataJsonPath := filepath.Join(metricDir, "data.json")

	// Check if data.json exists
	if _, err := os.Stat(dataJsonPath); os.IsNotExist(err) {
		// No default configuration available, which is fine
		return nil
	}

	// Load the default configuration file
	b, err := os.ReadFile(dataJsonPath)
	if err != nil {
		return fmt.Errorf("could not retrieve default configuration for metric %s: %w", m.Id, err)
	}

	config = &assessment.MetricConfiguration{}
	err = json.Unmarshal(b, config)
	if err != nil {
		return fmt.Errorf("error in reading default configuration for metric %s: %w", m.Id, err)
	}

	config.IsDefault = true
	config.MetricId = m.Id

	defaultMetricConfigurations[m.Id] = config

	return nil
}
