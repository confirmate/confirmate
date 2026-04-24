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

package collection

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"confirmate.io/core/api/evidence"
	"confirmate.io/core/api/evidence/evidenceconnect"
	"confirmate.io/core/api/ontology"
	"confirmate.io/core/service"
	"confirmate.io/core/stream"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const DefaultEvidenceStoreAddress = ""

// Collector is the interface that all collectors must implement. A collector is responsible for
// collecting evidence and translating them to ontology resources.
type Collector interface {
	// Name returns a human-readable collector name.
	Name() string

	// ID returns a stable collector identifier.
	ID() string

	// Collect executes the collection process and returns a list of collected resources or an error
	// if the collection failed.
	Collect() (list []ontology.IsResource, err error)
}

// Service is the service implementation for the collection service. It has one or more collectors,
// which do the actual work of collecting evidence. The service itself is responsible for
// orchestrating the collectors, including scheduling their execution at the configured interval.
type Service struct {
	cfg                 Config
	evidenceStoreClient evidenceconnect.EvidenceStoreClient
	evidenceStoreStream *stream.RestartableBidiStream[evidence.StoreEvidenceRequest, evidence.StoreEvidencesResponse]
}

// DefaultConfig is the default configuration for the collection service.
var DefaultConfig = Config{
	Interval:                5 * time.Minute,
	EvidenceStoreAddress:    DefaultEvidenceStoreAddress,
	EvidenceStoreHTTPClient: service.DefaultHTTPClient,
}

// Config is the configuration for the collection service.
type Config struct {
	// Interval defines how often collectors are executed.
	Interval time.Duration

	// Collectors is a list of collectors to use for collecting evidence. At least one collector
	// must be provided.
	Collectors []Collector

	// EvidenceStoreAddress defines the evidence store base URL. If empty, forwarding collected
	// evidence is disabled.
	EvidenceStoreAddress string

	// EvidenceStoreHTTPClient is used for evidence store communication. If nil,
	// [service.DefaultHTTPClient] is used.
	EvidenceStoreHTTPClient *http.Client

	// TargetOfEvaluationID is used when creating evidence records from collected resources.
	TargetOfEvaluationID string

	// ToolID overrides the collector ID when creating evidence records. If empty, the collector's
	// own ID is used.
	ToolID string
}

// WithConfig sets the service configuration, overriding the default configuration.
func WithConfig(cfg Config) service.Option[Service] {
	return func(svc *Service) {
		svc.cfg = cfg
	}
}

// NewService creates a new collection service with default values.
func NewService(opts ...service.Option[Service]) (svc *Service, err error) {
	var (
		cfg        Config
		httpClient *http.Client
	)

	svc = &Service{
		cfg: DefaultConfig,
	}

	for _, o := range opts {
		o(svc)
	}

	cfg = svc.cfg

	if cfg.Interval <= 0 {
		cfg.Interval = DefaultConfig.Interval
	}

	if len(cfg.Collectors) == 0 {
		return nil, fmt.Errorf("at least one collector must be provided")
	}

	if cfg.TargetOfEvaluationID != "" {
		if err = uuid.Validate(cfg.TargetOfEvaluationID); err != nil {
			return nil, fmt.Errorf("invalid target of evaluation id: %w", err)
		}
	}

	if cfg.EvidenceStoreAddress != "" {
		if cfg.TargetOfEvaluationID == "" {
			return nil, fmt.Errorf("target of evaluation id must be set when evidence store forwarding is enabled")
		}
	}

	svc.cfg = cfg

	if cfg.EvidenceStoreAddress != "" {
		httpClient = cfg.EvidenceStoreHTTPClient
		if httpClient == nil {
			httpClient = service.DefaultHTTPClient
		}

		svc.evidenceStoreClient = evidenceconnect.NewEvidenceStoreClient(httpClient, cfg.EvidenceStoreAddress)
	}

	if svc.evidenceStoreClient != nil {
		err = svc.initEvidenceStoreStream()
		if err != nil {
			return nil, err
		}
	}

	return svc, nil
}

func (svc *Service) initEvidenceStoreStream() (err error) {
	if svc.evidenceStoreClient == nil {
		return nil
	}

	factory := func(ctx context.Context) *connect.BidiStreamForClient[evidence.StoreEvidenceRequest, evidence.StoreEvidencesResponse] {
		return svc.evidenceStoreClient.StoreEvidences(ctx)
	}

	svc.evidenceStoreStream, err = stream.NewRestartableBidiStream(context.Background(), factory, stream.DefaultRestartConfig(), "StoreEvidences")
	if err != nil {
		return err
	}

	return nil
}

// Close releases resources owned by the collection service.
func (svc *Service) Close() (err error) {
	if svc.evidenceStoreStream == nil {
		return nil
	}

	err = svc.evidenceStoreStream.Close()
	if err != nil {
		return err
	}

	return nil
}

// sendResourcesToEvidenceStore sends the given resources to the evidence store, associating them
// with the configured target of evaluation ID and the collector as the tool ID.
func (svc *Service) sendResourcesToEvidenceStore(ctx context.Context, collector Collector, resources []ontology.IsResource) (err error) {
	var (
		storeErr error
		res      *evidence.StoreEvidencesResponse
		req      *evidence.StoreEvidenceRequest
		toolID   string
	)

	if svc.evidenceStoreStream == nil {
		return nil
	}

	toolID = collector.ID()
	if svc.cfg.ToolID != "" {
		toolID = svc.cfg.ToolID
	}

	for _, resource := range resources {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		req = &evidence.StoreEvidenceRequest{
			Evidence: &evidence.Evidence{
				Id:                   uuid.NewString(),
				Timestamp:            timestamppb.Now(),
				TargetOfEvaluationId: svc.cfg.TargetOfEvaluationID,
				ToolId:               toolID,
				Resource:             ontology.ProtoResource(resource),
			},
		}

		storeErr = svc.evidenceStoreStream.Send(req)
		if storeErr != nil {
			err = fmt.Errorf("failed to send evidence for collector %q: %w", collector.Name(), storeErr)
			return err
		}

		res, storeErr = svc.evidenceStoreStream.Receive()
		if storeErr != nil {
			if storeErr == io.EOF {
				err = fmt.Errorf("evidence store stream closed while sending evidence for collector %q", collector.Name())
				return err
			}

			err = fmt.Errorf("failed to receive evidence-store status for collector %q: %w", collector.Name(), storeErr)
			return err
		}

		if res.GetStatus() != evidence.EvidenceStatus_EVIDENCE_STATUS_OK {
			err = fmt.Errorf("evidence-store rejected evidence for collector %q: %s", collector.Name(), res.GetStatusMessage())
			return err
		}
	}

	return nil
}

// Start runs all collectors immediately and then repeatedly at the configured interval.
// The returned channel is closed when ctx is canceled.
func (svc *Service) Start(ctx context.Context) (resultCh <-chan CollectionResult) {
	var (
		results chan CollectionResult
	)

	results = make(chan CollectionResult)

	go func() {
		svc.runLoop(ctx, results)
	}()

	return results
}
