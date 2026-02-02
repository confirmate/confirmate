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

package assessment

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"buf.build/go/protovalidate"
	"confirmate.io/core/api/assessment"
	"confirmate.io/core/api/assessment/assessmentconnect"
	"confirmate.io/core/api/evidence"
	"confirmate.io/core/api/ontology"
	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/api/orchestrator/orchestratorconnect"
	"confirmate.io/core/log"
	"confirmate.io/core/policies"
	"confirmate.io/core/service"
	"confirmate.io/core/stream"
	"confirmate.io/core/util"
	"connectrpc.com/connect"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	logger *slog.Logger
)

const DefaultOrchestratorURL = "http://localhost:9090"

// DefaultConfig is the default configuration for the assessment [Service].
var DefaultConfig = Config{
	OrchestratorAddress: DefaultOrchestratorURL,
	OrchestratorClient:  http.DefaultClient,
	RegoPackage:         policies.DefaultRegoPackage,
}

// Config represents the configuration for the assessment [Service].
type Config struct {
	// OrchestratorAddress is the address of the orchestrator service.
	OrchestratorAddress string
	// OrchestratorClient is the HTTP client to use for orchestrator communication.
	OrchestratorClient *http.Client
	// RegoPackage is the package name to use for Rego policy evaluation.
	RegoPackage string
}

type orchestratorConfig struct {
	targetAddress string
	client        *http.Client
}

const (
	// EvictionTime is the time after which an entry in the metric configuration is invalid
	EvictionTime = time.Hour * 1
)

type cachedConfiguration struct {
	cachedAt time.Time
	*assessment.MetricConfiguration
}

// subscriber represents a subscription to metric change events
type subscriber struct {
	ch     chan *orchestrator.ChangeEvent
	filter *orchestrator.SubscribeRequest_Filter
}

// Service is an implementation of the Clouditor Assessment service. It should not be used directly,
// but rather the NewService constructor should be used. It implements the AssessmentServer interface.
type Service struct {
	// Embedded for FWD compatibility
	assessmentconnect.UnimplementedAssessmentHandler

	// TODO: we can remove the client if we handle everything via the bidistream
	orchestratorClient orchestratorconnect.OrchestratorClient
	orchestratorStream *stream.RestartableBidiStream[orchestrator.StoreAssessmentResultRequest, orchestrator.StoreAssessmentResultsResponse]
	streamMutex        sync.Mutex
	orchestratorConfig orchestratorConfig

	// resultHooks is a list of hook functions that can be used if one wants to be
	// informed about each assessment result
	resultHooks []assessment.ResultHookFunc
	// hookMutex is used for (un)locking result hook calls
	hookMutex sync.RWMutex

	// cachedConfigurations holds cached metric configurations for faster access with key being the corresponding
	// metric name
	cachedConfigurations map[string]cachedConfiguration
	// TODO(oxisto): combine with hookMutex and replace with a generic version of a mutex'd map
	confMutex sync.Mutex

	// evidenceResourceMap is a cache which maps a resource ID (key) to its latest available evidence
	// TODO(oxisto): replace this with storage queries
	evidenceResourceMap map[string]*evidence.Evidence
	em                  sync.RWMutex
	wg                  sync.WaitGroup

	// requests contains a map of our waiting requests
	requests map[string]waitingRequest

	// rm is a RWMutex for the requests property
	rm sync.RWMutex

	// pe contains the actual policy evaluation engine we use
	pe policies.PolicyEval

	// cfg contains the service configuration
	cfg Config

	// subscribers is a map of subscribers for metric change events
	subscribers      map[int64]*subscriber
	subscribersMutex sync.RWMutex
	nextSubscriberId int64
}

// WithConfig sets the service configuration, overriding the default configuration.
func WithConfig(cfg Config) service.Option[Service] {
	return func(svc *Service) {
		svc.cfg = cfg
	}
}

// NewService creates a new assessment service handler with default values.
func NewService(opts ...service.Option[Service]) (handler assessmentconnect.AssessmentHandler, err error) {
	var (
		svc *Service
		o   service.Option[Service]
	)

	svc = &Service{
		cfg: DefaultConfig,
		orchestratorConfig: orchestratorConfig{
			targetAddress: DefaultOrchestratorURL,
			client:        http.DefaultClient,
		},
		evidenceResourceMap:  make(map[string]*evidence.Evidence),
		requests:             make(map[string]waitingRequest),
		cachedConfigurations: make(map[string]cachedConfiguration),
		subscribers:          make(map[int64]*subscriber),
	}

	for _, o = range opts {
		o(svc)
	}

	// Apply configuration to internal fields
	svc.orchestratorConfig.targetAddress = svc.cfg.OrchestratorAddress
	svc.orchestratorConfig.client = svc.cfg.OrchestratorClient

	slog.Info("Orchestrator URL is set", slog.String("url", svc.cfg.OrchestratorAddress))

	// Initialize the policy evaluator with event subscription
	svc.pe = policies.NewRegoEval(
		policies.WithPackageName(svc.cfg.RegoPackage),
		policies.WithEventSubscriber(svc),
	)

	svc.orchestratorClient = orchestratorconnect.NewOrchestratorClient(svc.orchestratorConfig.client, svc.orchestratorConfig.targetAddress)
	err = svc.initOrchestratorStream()
	if err != nil {
		return nil, err
	}

	handler = svc
	return
}

func (svc *Service) initOrchestratorStream() (err error) {
	var (
		factory           stream.StreamFactory[orchestrator.StoreAssessmentResultRequest, orchestrator.StoreAssessmentResultsResponse]
		restartableStream *stream.RestartableBidiStream[orchestrator.StoreAssessmentResultRequest, orchestrator.StoreAssessmentResultsResponse]
	)

	factory = func(ctx context.Context) *connect.BidiStreamForClient[orchestrator.StoreAssessmentResultRequest, orchestrator.StoreAssessmentResultsResponse] {
		return svc.orchestratorClient.StoreAssessmentResults(ctx)
	}
	restartableStream, err = stream.NewRestartableBidiStream(context.Background(), factory, stream.DefaultRestartConfig(), "StoreAssessmentResults")
	if err != nil {
		return err
	}
	svc.orchestratorStream = restartableStream
	return
}

func (svc *Service) createOrchestratorStreamFactory() (factory stream.StreamFactory[orchestrator.StoreAssessmentResultRequest, orchestrator.StoreAssessmentResultsResponse]) {
	factory = func(ctx context.Context) *connect.BidiStreamForClient[orchestrator.StoreAssessmentResultRequest, orchestrator.StoreAssessmentResultsResponse] {
		return svc.orchestratorClient.StoreAssessmentResults(ctx)
	}
	return
}

func (svc *Service) AssessEvidences(ctx context.Context, stream *connect.BidiStream[assessment.AssessEvidenceRequest, assessment.AssessEvidencesResponse]) (err error) {
	var (
		req           *assessment.AssessEvidenceRequest
		res           *assessment.AssessEvidencesResponse
		assessmentReq *connect.Request[assessment.AssessEvidenceRequest]
		assessmentRes *connect.Response[assessment.AssessEvidenceResponse]
	)

	for {
		req, err = stream.Receive()
		// If no more input of the stream is available, return
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			err = fmt.Errorf("cannot receive stream request: %w", err)
			slog.Error("cannot receive stream request", log.Err(err))
			return connect.NewError(connect.CodeUnknown, err)
		}
		assessmentReq = connect.NewRequest(&assessment.AssessEvidenceRequest{
			Evidence: req.Evidence,
		})

		assessmentRes, err = svc.AssessEvidence(ctx, assessmentReq)
		if err != nil {
			slog.Error("AssessEvidenceStream: could not assess evidence:", log.Err(err))
			res = &assessment.AssessEvidencesResponse{
				Status:        assessment.AssessmentStatus_ASSESSMENT_STATUS_FAILED,
				StatusMessage: err.Error(),
			}
		} else {
			res = &assessment.AssessEvidencesResponse{
				Status: assessmentRes.Msg.Status,
			}
		}

		err = stream.Send(res)
		if err != nil {
			slog.Error("AssessEvidenceStream: could not send response:", log.Err(err))
			return connect.NewError(connect.CodeUnknown, fmt.Errorf("could not send stream response: %w", err))
		}
	}
}

// AssessEvidence is a method implementation of the assessment interface: It assesses a single evidence
func (svc *Service) AssessEvidence(ctx context.Context, req *connect.Request[assessment.AssessEvidenceRequest]) (res *connect.Response[assessment.AssessEvidenceResponse], err error) {
	var (
		resource        ontology.IsResource
		ev              *evidence.Evidence
		canHandle       bool
		waitingFor      map[string]bool
		related         map[string]ontology.IsResource
		ok              bool
		relatedEvidence *evidence.Evidence
		l               waitingRequest
	)

	// Validate the request
	if err = service.Validate(req); err != nil {
		return nil, err
	}

	ev = req.Msg.Evidence
	// Validate request
	err = protovalidate.Validate(ev)
	if err != nil {
		slog.Error("AssessEvidence: invalid request", log.Err(err))
		return nil, err
	}

	// // TODO: Check if target_of_evaluation_id in the service is within allowed or one can access *all* the target of evaluations
	// if !svc.authz.CheckAccess(ctx, service.AccessUpdate, req) {
	// 	slog.Error("AssessEvidence: ", log.Err(service.ErrPermissionDenied))
	// 	return nil, service.ErrPermissionDenied
	// }

	// Retrieve the ontology resource
	resource = ev.GetOntologyResource()
	if resource == nil {
		err = ontology.ErrNotOntologyResource
		slog.Error("AssessEvidence: Not an ontology resource:", log.Err(err))
		return nil, err
	}

	// Check, if we can immediately handle this evidence; we assume so at first
	canHandle = true
	waitingFor = make(map[string]bool)
	related = make(map[string]ontology.IsResource)

	svc.em.Lock()

	// We need to check, if by any chance the related resource evidences have already arrived
	//
	// TODO(oxisto): We should also check if they are "recent" enough (which is probably determined by the metric)
	for _, r := range ev.ExperimentalRelatedResourceIds {
		// If any of the related resource is not available, we cannot handle them immediately, but we need to add it to
		// our waitingFor slice
		relatedEvidence, ok = svc.evidenceResourceMap[r]
		if ok {
			related[r] = relatedEvidence.GetOntologyResource()
		} else {
			canHandle = false
			waitingFor[r] = true
		}
	}

	// Update our resourceID to evidence cache
	svc.evidenceResourceMap[resource.GetId()] = ev
	svc.em.Unlock()

	// Inform any other left over evidences that might be waiting
	go svc.informWaitingRequests(resource.GetId())

	if canHandle {
		// Assess evidence. This also validates the embedded resource and returns an error if validation fails.
		_, err = svc.handleEvidence(ctx, ev, resource, related)
		if err != nil {
			return nil, err
		}

		res = connect.NewResponse(&assessment.AssessEvidenceResponse{
			Status: assessment.AssessmentStatus_ASSESSMENT_STATUS_ASSESSED,
		})
	} else {
		slog.Debug("Evidence needs to wait for more resource(s) to assess evidence", slog.Any("evidence", ev), slog.Int("waitingFor", len(waitingFor)))

		// Create a left-over request with all the necessary information
		l = waitingRequest{
			started:      time.Now(),
			waitingFor:   waitingFor,
			resourceId:   resource.GetId(),
			Evidence:     ev,
			s:            svc,
			newResources: make(chan string, 1000),
			ctx:          ctx,
		}

		// Add it to our wait group
		svc.wg.Add(1)

		// Wait for evidences in the background and handle them
		go l.WaitAndHandle()

		// Lock requests for writing
		svc.rm.Lock()
		svc.requests[ev.Id] = l
		// Unlock writing
		svc.rm.Unlock()

		res = connect.NewResponse(&assessment.AssessEvidenceResponse{
			Status: assessment.AssessmentStatus_ASSESSMENT_STATUS_WAITING_FOR_RELATED,
		})
	}

	return res, nil
}

// handleEvidence is the helper method for the actual assessment used by AssessEvidence and AssessEvidences. This will
// also validate the resource embedded into the evidence and return an error if validation fails. In order to
// distinguish between internal errors and validation errors, this function already returns a gRPC error.
func (svc *Service) handleEvidence(
	ctx context.Context,
	ev *evidence.Evidence,
	resource ontology.IsResource,
	related map[string]ontology.IsResource,
) (results []*assessment.AssessmentResult, err error) {
	var (
		types       []string
		evaluations []*policies.CombinedResult
		newError    error
		metricID    string
		result      *assessment.AssessmentResult
	)

	if resource == nil {
		return nil, status.Errorf(codes.Internal, "invalid embedded resource: %v", ontology.ErrNotOntologyResource)
	}

	slog.Debug("Evaluating evidence",
		slog.String("Evidence", ev.Id),
		slog.String("Resource", resource.GetId()),
		slog.String("ToolId", ev.ToolId),
		slog.Any("Timestamp", ev.Timestamp.AsTime()),
	)

	evaluations, err = svc.pe.Eval(ev, resource, related, svc)
	if err != nil {
		newError = fmt.Errorf("could not evaluate evidence: %w", err)

		go svc.informHooks(ctx, nil, newError)

		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	if err != nil {
		err = fmt.Errorf("could not get stream to orchestrator (%s): %w", svc.orchestratorConfig.targetAddress, err)

		go svc.informHooks(ctx, nil, err)

		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	if len(evaluations) == 0 {
		slog.Debug("No policy evaluation for evidence", slog.String("Evidence", ev.Id), slog.String("Resource", resource.GetId()), slog.String("ToolId", ev.ToolId))
		return results, nil
	}

	for _, data := range evaluations {
		// That there is an empty (nil) evaluation should be caught beforehand, but you never know.
		if data == nil {
			slog.Error("One empty policy evaluation detected for evidence. That should not happen.", slog.String("Evidence", ev.GetId()))
			continue
		}
		metricID = data.MetricID

		slog.Debug("Evaluated evidence with metric", slog.String("Evidence", ev.Id), slog.String("MetricID", metricID), slog.Bool("Compliant", data.Compliant))

		types = ontology.ResourceTypes(resource)

		result = &assessment.AssessmentResult{
			Id:                   uuid.NewString(),
			CreatedAt:            timestamppb.Now(),
			TargetOfEvaluationId: ev.GetTargetOfEvaluationId(),
			MetricId:             metricID,
			MetricConfiguration:  data.Config,
			Compliant:            data.Compliant,
			EvidenceId:           ev.GetId(),
			ResourceId:           resource.GetId(),
			ResourceTypes:        types,
			ComplianceComment:    data.Message,
			ComplianceDetails:    data.ComparisonResult,
			ToolId:               util.Ref(assessment.AssessmentToolId),
			HistoryUpdatedAt:     timestamppb.Now(),
			History: []*assessment.Record{{ // TODO(all): Update history in another PR, see Issue #1724
				EvidenceId:         ev.GetId(),
				EvidenceRecordedAt: timestamppb.Now(),
			}},
		}

		// Inform hooks about new assessment result
		go svc.informHooks(ctx, result, nil)

		svc.streamMutex.Lock()
		err = svc.orchestratorStream.Send(&orchestrator.StoreAssessmentResultRequest{
			Result: result,
		})
		svc.streamMutex.Unlock()

		if err != nil {
			slog.Error("Failed to send assessment result to orchestrator", log.Err(err))
			go svc.informHooks(ctx, nil, fmt.Errorf("failed to send result: %w", err))
		}

		results = append(results, result)
	}

	return results, nil
}

// informHooks informs the registered hook functions
func (svc *Service) informHooks(ctx context.Context, result *assessment.AssessmentResult, err error) {
	var (
		hooks []assessment.ResultHookFunc
	)

	svc.hookMutex.RLock()
	hooks = svc.resultHooks
	defer svc.hookMutex.RUnlock()

	// Inform our hook, if we have any
	if len(hooks) > 0 {
		for _, hook := range hooks {
			// We could do hook concurrent again (assuming different hooks don't interfere with each other)
			hook(ctx, result, err)
		}
	}
}

func (svc *Service) RegisterAssessmentResultHook(assessmentResultsHook func(ctx context.Context, result *assessment.AssessmentResult, err error)) {
	svc.hookMutex.Lock()
	defer svc.hookMutex.Unlock()
	svc.resultHooks = append(svc.resultHooks, assessmentResultsHook)
}

// Metrics implements MetricsSource by retrieving the metric list from the orchestrator.
func (svc *Service) Metrics() (metrics []*assessment.Metric, err error) {
	res, err := svc.orchestratorClient.ListMetrics(
		context.Background(),
		connect.NewRequest(
			&orchestrator.ListMetricsRequest{}),
	)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve metric list from orchestrator: %w", err)
	}
	metrics = res.Msg.Metrics

	return metrics, nil
}

// MetricImplementation implements MetricsSource by retrieving the metric implementation
// from the orchestrator.
func (svc *Service) MetricImplementation(lang assessment.MetricImplementation_Language, metric *assessment.Metric) (impl *assessment.MetricImplementation, err error) {
	if lang != assessment.MetricImplementation_LANGUAGE_REGO {
		return nil, errors.New("unsupported language")
	}

	resp, err := svc.orchestratorClient.GetMetricImplementation(
		context.Background(),
		connect.NewRequest(&orchestrator.GetMetricImplementationRequest{
			MetricId: metric.Id,
		}))
	if err != nil {
		return nil, fmt.Errorf("could not retrieve metric implementation for %s from orchestrator: %w", metric.Id, err)
	}

	// Unwrap the response
	return resp.Msg, nil
}

// MetricConfiguration implements MetricsSource by getting the corresponding metric configuration for the
// given target of evaluation
func (svc *Service) MetricConfiguration(TargetOfEvaluationID string, metric *assessment.Metric) (config *assessment.MetricConfiguration, err error) {
	var (
		ok    bool
		cache cachedConfiguration
		key   string
		req   *connect.Request[orchestrator.GetMetricConfigurationRequest]
		resp  *connect.Response[assessment.MetricConfiguration]
	)

	// Calculate the cache key
	key = fmt.Sprintf("%s-%s", TargetOfEvaluationID, metric.Id)

	// Retrieve our cached entry
	svc.confMutex.Lock()
	cache, ok = svc.cachedConfigurations[key]
	svc.confMutex.Unlock()

	// Check if entry is not there or is expired
	if !ok || cache.cachedAt.After(time.Now().Add(EvictionTime)) {
		req = connect.NewRequest(&orchestrator.GetMetricConfigurationRequest{
			TargetOfEvaluationId: TargetOfEvaluationID,
			MetricId:             metric.Id,
		})

		resp, err = svc.orchestratorClient.GetMetricConfiguration(context.Background(), req)
		config = resp.Msg

		if err != nil {
			return nil, fmt.Errorf("could not retrieve metric configuration for %s: %w", metric.Id, err)
		}

		cache = cachedConfiguration{
			cachedAt:            time.Now(),
			MetricConfiguration: config,
		}

		svc.confMutex.Lock()
		// Update the metric configuration
		svc.cachedConfigurations[key] = cache
		defer svc.confMutex.Unlock()
	}

	return cache.MetricConfiguration, nil
}

// RegisterSubscriber registers a new subscriber for metric change events.
// It returns a channel to receive events and a subscriber ID for later unregistration.
func (svc *Service) RegisterSubscriber(filter *orchestrator.SubscribeRequest_Filter) (ch <-chan *orchestrator.ChangeEvent, id int64) {
	var (
		channelBuf chan *orchestrator.ChangeEvent
	)

	svc.subscribersMutex.Lock()
	defer svc.subscribersMutex.Unlock()

	channelBuf = make(chan *orchestrator.ChangeEvent, 100)
	id = svc.nextSubscriberId
	svc.nextSubscriberId++

	svc.subscribers[id] = &subscriber{
		ch:     channelBuf,
		filter: filter,
	}

	ch = channelBuf
	return
}

// UnregisterSubscriber removes a subscriber from receiving metric change events.
func (svc *Service) UnregisterSubscriber(id int64) (err error) {
	var (
		sub *subscriber
		ok  bool
	)

	svc.subscribersMutex.Lock()
	defer svc.subscribersMutex.Unlock()

	sub, ok = svc.subscribers[id]
	if ok {
		delete(svc.subscribers, id)
		close(sub.ch)
	}

	return nil
}

// publishEvent publishes a metric change event to all registered subscribers.
func (svc *Service) publishEvent(event *orchestrator.ChangeEvent) {
	svc.subscribersMutex.RLock()
	defer svc.subscribersMutex.RUnlock()

	for _, sub := range svc.subscribers {
		select {
		case sub.ch <- event:
		default:
			// Channel full, skip (non-blocking send)
		}
	}
}
