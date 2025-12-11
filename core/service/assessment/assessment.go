// Copyright 2025 Fraunhofer AISEC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package assessment

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"confirmate.io/core/api"
	"confirmate.io/core/api/assessment"
	"confirmate.io/core/api/assessment/assessmentconnect"
	"confirmate.io/core/api/evidence"
	"confirmate.io/core/api/ontology"
	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/api/orchestrator/orchestratorconnect"
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

const DefaultOrchestratorURL = "localhost:9090"

type orchestratorConfig struct {
	targetAddress string
	client        *http.Client
}

// Service is an implementation of the Clouditor Assessment service. It should not be used directly,
// but rather the NewService constructor should be used. It implements the AssessmentServer interface.
type Service struct {
	// Embedded for FWD compatibility
	assessmentconnect.UnimplementedAssessmentHandler

	authz service.AuthorizationStrategy

	orchestratorClient orchestratorconnect.OrchestratorClient
	orchestratorStream *stream.RestartableBidiStream[orchestrator.StoreAssessmentResultRequest, orchestrator.StoreAssessmentResultResponse]
	streamMutex        sync.Mutex
	orchestratorConfig orchestratorConfig

	// resultHooks is a list of hook functions that can be used if one wants to be
	// informed about each assessment result
	resultHooks []assessment.ResultHookFunc
	// hookMutex is used for (un)locking result hook calls
	hookMutex sync.RWMutex

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
}

// NewService creates a new assessment service with default values.
func NewService(opts ...service.Option[*Service]) *Service {
	svc := &Service{
		orchestratorConfig: orchestratorConfig{
			targetAddress: DefaultOrchestratorURL,
			client:        http.DefaultClient,
		},
	}

	for _, o := range opts {
		o(svc)
	}

	svc.orchestratorClient = orchestratorconnect.NewOrchestratorClient(svc.orchestratorConfig.client, svc.orchestratorConfig.targetAddress)

	return svc
}

func (svc *Service) Init() {
	config := stream.DefaultRestartConfig()
	config.MaxRetries = 5
	config.InitialBackoff = 100 * time.Millisecond
	config.MaxBackoff = 5 * time.Second
	config.OnRestart = func(attempt int, err error) {
		slog.Warn("Restarting orchestrator stream", "attempt", attempt, "error", err)
	}
	config.OnRestartFailure = func(err error) {
		slog.Error("Failed to restart orchestrator stream after max retries", "error", err)
	}

	// Context für den Stream (sollte die Lebensdauer des Service haben)
	ctx := context.Background()

	// Factory erstellen
	factory := svc.createOrchestratorStreamFactory()

	// RestartableBidiStream erstellen
	rs, err := stream.NewRestartableBidiStream(ctx, factory, config)
	if err != nil {
		slog.Error("Failed to create orchestrator stream", "error", err)
		// Hier könntest du auch paniken oder den Fehler anders behandeln
		return
	}

	svc.orchestratorStream = rs
}

func (svc *Service) createOrchestratorStreamFactory() stream.StreamFactory[orchestrator.StoreAssessmentResultRequest, orchestrator.StoreAssessmentResultsResponse] {
	return func(ctx context.Context) *connect.BidiStreamForClient[orchestrator.StoreAssessmentResultRequest, orchestrator.StoreAssessmentResultsResponse] {
		return svc.orchestratorClient.StoreAssessmentResults(ctx)
	}
}

// AssessEvidence is a method implementation of the assessment interface: It assesses a single evidence
func (svc *Service) AssessEvidence(ctx context.Context, req *assessment.AssessEvidenceRequest) (res *assessment.AssessEvidenceResponse, err error) {

	var (
		resource ontology.IsResource
	)

	// Validate request
	err = api.Validate(req)
	if err != nil {
		slog.Error("AssessEvidence: invalid request", "error", err)
		return nil, err
	}

	// Check if target_of_evaluation_id in the service is within allowed or one can access *all* the target of evaluations
	if !svc.authz.CheckAccess(ctx, service.AccessUpdate, req) {
		slog.Error("AssessEvidence: ", service.ErrPermissionDenied)
		return nil, service.ErrPermissionDenied
	}

	// Retrieve the ontology resource
	resource = req.Evidence.GetOntologyResource()
	if resource == nil {
		err = ontology.ErrNotOntologyResource
		slog.Error("AssessEvidence: Not an ontology resource:", "error", err)
		return nil, err
	}

	// Check, if we can immediately handle this evidence; we assume so at first
	var (
		canHandle                                 = true
		waitingFor map[string]bool                = make(map[string]bool)
		related    map[string]ontology.IsResource = make(map[string]ontology.IsResource)
	)

	svc.em.Lock()

	// We need to check, if by any chance the related resource evidences have already arrived
	//
	// TODO(oxisto): We should also check if they are "recent" enough (which is probably determined by the metric)
	for _, r := range req.Evidence.ExperimentalRelatedResourceIds {
		// If any of the related resource is not available, we cannot handle them immediately, but we need to add it to
		// our waitingFor slice
		if _, ok := svc.evidenceResourceMap[r]; ok {
			ev := svc.evidenceResourceMap[r]

			related[r] = ev.GetOntologyResource()
		} else {
			canHandle = false
			waitingFor[r] = true
		}
	}

	// Update our resourceID to evidence cache
	svc.evidenceResourceMap[resource.GetId()] = req.Evidence
	svc.em.Unlock()

	// Inform any other left over evidences that might be waiting
	go svc.informWaitingRequests(resource.GetId())

	if canHandle {
		// Assess evidence. This also validates the embedded resource and returns a gRPC error if validation fails.
		_, err = svc.handleEvidence(ctx, req.Evidence, resource, related)
		if err != nil {
			slog.Error("AssessEvidence: could not handle evidence:", "error", err)
			return nil, err
		}

		res = &assessment.AssessEvidenceResponse{
			Status: assessment.AssessmentStatus_ASSESSMENT_STATUS_ASSESSED,
		}
	} else {
		slog.Debug("Evidence %s needs to wait for %d more resource(s) to assess evidence", req.Evidence.Id, len(waitingFor))

		// Create a left-over request with all the necessary information
		l := waitingRequest{
			started:      time.Now(),
			waitingFor:   waitingFor,
			resourceId:   resource.GetId(),
			Evidence:     req.Evidence,
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
		svc.requests[req.Evidence.Id] = l
		// Unlock writing
		svc.rm.Unlock()

		res = &assessment.AssessEvidenceResponse{
			Status: assessment.AssessmentStatus_ASSESSMENT_STATUS_WAITING_FOR_RELATED,
		}
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
		types []string
	)

	if resource == nil {
		return nil, status.Errorf(codes.Internal, "invalid embedded resource: %v", ontology.ErrNotOntologyResource)
	}

	slog.Debug("Evaluating evidence %s (%s) collected by %s at %s", ev.Id, resource.GetId(), ev.ToolId, ev.Timestamp.AsTime())
	slog.Debug("Evidence: %+v", ev)

	evaluations, err := svc.pe.Eval(ev, resource, related, svc)
	if err != nil {
		newError := fmt.Errorf("could not evaluate evidence: %w", err)

		go svc.informHooks(ctx, nil, newError)

		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	if err != nil {
		err = fmt.Errorf("could not get stream to orchestrator (%s): %w", svc.orchestrator.Target, err)

		go svc.informHooks(ctx, nil, err)

		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	if len(evaluations) == 0 {
		slog.Debug("No policy evaluation for evidence %s (%s) collected by %s.", ev.Id, resource.GetId(), ev.ToolId)
		return results, nil
	}

	for _, data := range evaluations {
		// That there is an empty (nil) evaluation should be caught beforehand, but you never know.
		if data == nil {
			slog.Error("One empty policy evaluation detected for evidence '%s'. That should not happen.",
				ev.GetId())
			continue
		}
		metricID := data.MetricID

		slog.Debug("Evaluated evidence %v with metric '%v' as %v", ev.Id, metricID, data.Compliant)

		types = ontology.ResourceTypes(resource)

		result := &assessment.AssessmentResult{
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
			slog.Error("Failed to send assessment result to orchestrator", "error", err)
			go svc.informHooks(ctx, nil, fmt.Errorf("failed to send result: %w", err))
		}

		results = append(results, result)
	}

	return results, nil
}

// informHooks informs the registered hook functions
func (svc *Service) informHooks(ctx context.Context, result *assessment.AssessmentResult, err error) {
	svc.hookMutex.RLock()
	hooks := svc.resultHooks
	defer svc.hookMutex.RUnlock()

	// Inform our hook, if we have any
	if len(hooks) > 0 {
		for _, hook := range hooks {
			// We could do hook concurrent again (assuming different hooks don't interfere with each other)
			hook(ctx, result, err)
		}
	}
}
