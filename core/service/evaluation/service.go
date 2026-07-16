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

package evaluation

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"

	"confirmate.io/core/api"
	"confirmate.io/core/api/assessment"
	"confirmate.io/core/api/evaluation"
	"confirmate.io/core/api/evaluation/evaluationconnect"
	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/api/orchestrator/orchestratorconnect"
	"confirmate.io/core/log"
	"confirmate.io/core/service"

	"connectrpc.com/connect"
	"github.com/go-co-op/gocron"
	"github.com/google/uuid"
	"golang.org/x/oauth2/clientcredentials"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	DefaultOrchestratorURL = "http://localhost:8080"

	// defaultInterval is the default interval time for the scheduler. If no interval is set in the StartEvaluationRequest, the default value is taken.
	defaultInterval int = 5
)

// Service implements the Evaluation Service handler (see
// [evaluationconnect.EvaluationHandler]).
type Service struct {
	evaluationconnect.UnimplementedEvaluationHandler
	cfg   Config
	authz service.AuthorizationStrategy

	orchestratorClient orchestratorconnect.OrchestratorClient

	scheduler *gocron.Scheduler

	// catalogControls stores the catalog controls so that they do not always have to be retrieved from Orchestrators getControl endpoint.
	// map[catalog_id][control_id]*orchestrator.Control
	catalogControls map[string]map[string]*orchestrator.Control
	catalogsMutex   sync.RWMutex
}

// DefaultConfig is the default configuration for the evaluation [Service].
var DefaultConfig = Config{
	OrchestratorAddress: DefaultOrchestratorURL,
	OrchestratorClient:  service.DefaultHTTPClient,
}

// Config represents the configuration for the evaluation [Service].
type Config struct {
	// OrchestratorAddress is the address of the Orchestrator service to connect to.
	OrchestratorAddress string
	// OrchestratorClient is the HTTP client to use for connecting to the Orchestrator service.
	OrchestratorClient *http.Client
	// ServiceOAuth2Config is the OAuth2 client credentials configuration used for
	// service-to-service authentication with the orchestrator. When set, all outgoing
	// orchestrator calls use this token.
	ServiceOAuth2Config *clientcredentials.Config
}

// WithConfig sets the service configuration, overriding the default configuration.
func WithConfig(cfg Config) service.Option[Service] {
	return func(svc *Service) {
		svc.cfg = cfg
	}
}

// WithAuthorizationStrategy configures a custom authorization strategy.
func WithAuthorizationStrategy(authz service.AuthorizationStrategy) service.Option[Service] {
	return func(svc *Service) {
		svc.authz = authz
	}
}

// WithAuthorizationStrategyPermissionStore configures permission store-based authorization backed
// by the orchestrator API. The permission store is wired up after the orchestrator client is
// initialized in [NewService].
func WithAuthorizationStrategyPermissionStore() service.Option[Service] {
	return func(svc *Service) {
		svc.authz = &service.AuthorizationStrategyPermissionStore{}
	}
}

// NewService creates a new Evaluation service
func NewService(opts ...service.Option[Service]) (handler evaluationconnect.EvaluationHandler, err error) {
	var (
		svc = &Service{
			cfg:             DefaultConfig,
			scheduler:       gocron.NewScheduler(time.Local),
			catalogControls: make(map[string]map[string]*orchestrator.Control),
		}
	)

	for _, o := range opts {
		o(svc)
	}

	if svc.authz == nil {
		svc.authz = &service.AuthorizationStrategyAllowAll{}
	}

	// If service OAuth2 credentials are configured, wrap the HTTP client so all outgoing
	// orchestrator calls authenticate using the client credentials flow. This also fixes the
	// scheduled-job token expiry issue: auth is handled at the transport level rather than via
	// the original request context.
	orchestratorHTTPClient := svc.cfg.OrchestratorClient
	if svc.cfg.ServiceOAuth2Config != nil {
		orchestratorHTTPClient = api.NewOAuthHTTPClient(
			orchestratorHTTPClient,
			api.NewOAuthAuthorizerFromClientCredentials(svc.cfg.ServiceOAuth2Config),
		)
	}

	// Initialize the orchestrator service client
	svc.orchestratorClient = orchestratorconnect.NewOrchestratorClient(orchestratorHTTPClient, svc.cfg.OrchestratorAddress)

	// If using permission store-based authorization, back it with the orchestrator client so the
	// evaluation service can check permissions without direct database access.
	if permStrat, ok := svc.authz.(*service.AuthorizationStrategyPermissionStore); ok {
		permStrat.Permissions = &service.OrchestratorPermissionStore{Client: svc.orchestratorClient}
	}

	slog.Info("Orchestrator URL is set", slog.String("url", svc.cfg.OrchestratorAddress))

	handler = svc
	return
}

func (svc *Service) Shutdown() {
	svc.scheduler.Stop()
}

// StartEvaluation is a method implementation of the evaluation interface: It periodically starts the evaluation of a
// target of evaluation and the given catalog in the audit_scope. If no interval time is given, the default value is
// used.
func (svc *Service) StartEvaluation(ctx context.Context, req *connect.Request[evaluation.StartEvaluationRequest]) (res *connect.Response[evaluation.StartEvaluationResponse], err error) {
	var (
		interval      int
		auditScope    *orchestrator.AuditScope
		auditScopeRes *connect.Response[orchestrator.AuditScope]
		catalog       *orchestrator.Catalog
		catalogRes    *connect.Response[orchestrator.Catalog]
		jobs          []*gocron.Job
	)

	// Validate the request
	if err = service.Validate(req); err != nil {
		return nil, err
	}

	// Check access via the configured auth strategy
	var allowed bool
	allowed, _, err = checkAccess(ctx, svc.authz, orchestrator.RequestType_REQUEST_TYPE_UPDATED, req.Msg.GetAuditScopeId(), orchestrator.ObjectType_OBJECT_TYPE_AUDIT_SCOPE)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if !allowed {
		return nil, service.ErrPermissionDenied
	}

	// Get Audit Scope
	auditScopeRes, err = svc.orchestratorClient.GetAuditScope(ctx, connect.NewRequest(&orchestrator.GetAuditScopeRequest{
		AuditScopeId: req.Msg.GetAuditScopeId(),
	}))
	if err != nil {
		slog.Error("Could not get audit scope from orchestrator", log.Err(err))
		return nil, connect.NewError(connect.CodeNotFound, errors.New("could not get audit scope from orchestrator"))
	}
	auditScope = auditScopeRes.Msg

	// Make sure that the scheduler is already running
	svc.scheduler.StartAsync()

	// Set the interval to the default value if not set. If the interval is set to 0, the default interval is used.
	if req.Msg.GetInterval() == 0 {
		interval = defaultInterval
	} else {
		interval = int(req.Msg.GetInterval())
	}

	// Get all Controls from Orchestrator for the evaluation
	err = svc.cacheControls(auditScope.GetCatalogId())
	if err != nil {
		slog.Error("Could not cache controls", log.Err(err))
		return nil, connect.NewError(connect.CodeInternal, errors.New("could not cache controls"))
	}

	// Retrieve the catalog
	catalogRes, err = svc.orchestratorClient.GetCatalog(ctx, connect.NewRequest(&orchestrator.GetCatalogRequest{
		CatalogId: auditScope.GetCatalogId(),
	}))
	if err != nil {
		slog.Error("Could not get catalog from the orchestrator", log.Err(err))
		return nil, connect.NewError(connect.CodeInternal, errors.New("could not get catalog from the orchestrator"))
	}
	catalog = catalogRes.Msg

	// Check, if a previous job exists and/or is running
	jobs, err = svc.scheduler.FindJobsByTag(auditScope.GetId())
	if err != nil && !errors.Is(err, gocron.ErrJobNotFoundWithTag) {
		slog.Error("Could not find existing scheduler job", log.Err(err))
		return nil, connect.NewError(connect.CodeInternal, errors.New("no scheduler job found"))
	} else if len(jobs) > 0 {
		slog.Error("Evaluation already started for Audit scope", slog.String("audit scope", auditScope.GetId()), slog.String("target of evaluation", auditScope.GetTargetOfEvaluationId()), slog.String("catalog id", auditScope.GetCatalogId()))
		return nil, connect.NewError(connect.CodeAlreadyExists, fmt.Errorf("evaluation already started for the given audit scope '%s'", auditScope.GetId()))
	}

	slog.Info("Starting evaluation ...")

	// Add job to scheduler
	err = svc.addJobToScheduler(ctx, auditScope, catalog, interval)
	// We can return the error as it is
	if err != nil {
		return nil, err
	}

	slog.Info("Scheduled to evaluate audit scope",
		slog.String("audit scope", auditScope.GetId()),
		slog.Int("interval (in minutes)", interval),
	)

	res = connect.NewResponse(&evaluation.StartEvaluationResponse{
		Successful: true,
	})

	return res, nil
}

// StopEvaluation is a method implementation of the evaluation interface: It stops the evaluation for a
// AuditScope.
func (svc *Service) StopEvaluation(ctx context.Context, req *connect.Request[evaluation.StopEvaluationRequest]) (res *connect.Response[evaluation.StopEvaluationResponse], err error) {
	// Validate the request
	if err = service.Validate(req); err != nil {
		return nil, err
	}

	// Check access via the configured auth strategy
	var allowed bool
	allowed, _, err = checkAccess(ctx, svc.authz, orchestrator.RequestType_REQUEST_TYPE_UPDATED, req.Msg.GetAuditScopeId(), orchestrator.ObjectType_OBJECT_TYPE_AUDIT_SCOPE)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if !allowed {
		return nil, service.ErrPermissionDenied
	}

	auditScopeId := req.Msg.GetAuditScopeId()

	// Stop jobs(s) for given audit scope
	err = svc.scheduler.RemoveByTags(auditScopeId)
	if err != nil && errors.Is(err, gocron.ErrJobNotFoundWithTag) {
		slog.Error("Job for audit scope is not running", slog.String("audit scope", auditScopeId), log.Err(err))
		return nil, connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("job for audit scope '%s' is not running", auditScopeId))
	} else if err != nil {
		slog.Error("Could not remove jobs for audit scope '%s': %w", auditScopeId, log.Err(err))
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("could not remove jobs for audit scope '%s'", auditScopeId))
	}

	res = &connect.Response[evaluation.StopEvaluationResponse]{}

	return
}

// ListEvaluationJobs lists all running evaluation jobs.
func (svc *Service) ListEvaluationJobs(ctx context.Context, req *connect.Request[evaluation.ListEvaluationJobsRequest]) (res *connect.Response[evaluation.ListEvaluationJobsResponse], err error) {
	var (
		jobs           []*gocron.Job
		allowed        bool
		scopeIds       []string
		evaluationJobs = make([]*evaluation.EvaluationJob, 0)
	)

	// Validate the request
	if err = service.Validate(req); err != nil {
		return nil, err
	}

	// Check access via the configured auth strategy
	allowed, scopeIds, err = checkAccess(ctx, svc.authz, orchestrator.RequestType_REQUEST_TYPE_LIST, "", orchestrator.ObjectType_OBJECT_TYPE_AUDIT_SCOPE)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if !allowed && len(scopeIds) == 0 {
		return connect.NewResponse(&evaluation.ListEvaluationJobsResponse{
			EvaluationJobs: []*evaluation.EvaluationJob{},
		}), nil
	}

	// Build a set of allowed scope IDs for filtering
	scopeIdSet := make(map[string]struct{}, len(scopeIds))
	for _, id := range scopeIds {
		scopeIdSet[id] = struct{}{}
	}

	// Get all jobs from the scheduler
	jobs = svc.scheduler.Jobs()

	for _, job := range jobs {
		jobScopeId := job.Tags()[0]
		// Filter by audit scope ID if provided
		if req.Msg.GetFilter().GetAuditScopeId() != "" && jobScopeId != req.Msg.GetFilter().GetAuditScopeId() {
			continue
		}
		// Filter by permission — if not allowed to see all scopes, only show
		// jobs for scopes the user has access to
		if !allowed {
			if _, ok := scopeIdSet[jobScopeId]; !ok {
				continue
			}
		}
		evaluationJobs = append(evaluationJobs, &evaluation.EvaluationJob{
			AuditScopeId: jobScopeId,
			RunCount:     int32(job.FinishedRunCount()),
			LastRun:      timestamppb.New(job.LastRun()),
			Interval:     int32(job.ScheduledInterval()),
			StartedAt:    timestamppb.New(job.LastRun()),
		})
	}

	return connect.NewResponse(&evaluation.ListEvaluationJobsResponse{
		EvaluationJobs: evaluationJobs,
	}), nil
}

// addJobToScheduler adds a job for the given control to the scheduler and sets the scheduler interval to the given
// interval. It returns an buf connect error that can be used directly by the caller
func (svc *Service) addJobToScheduler(ctx context.Context, auditScope *orchestrator.AuditScope, catalog *orchestrator.Catalog, interval int) (err error) {
	// Check inputs and log error
	if auditScope == nil {
		err = errors.New("audit scope is invalid")
	}
	if interval == 0 {
		err = errors.New("interval is invalid")
	}
	if err != nil {
		slog.Error("Evaluation cannot be scheduled", log.Err(err))
		return connect.NewError(connect.CodeInternal, errors.New("evaluation cannot be scheduled due to invalid input"))
	}

	// Use context.Background() rather than the original request context: auth for outgoing
	// orchestrator calls is handled by the OAuth2 HTTP transport, so the scheduled job does not
	// need (or want) to inherit the caller's token, which would eventually expire.
	_, err = svc.scheduler.
		Every(interval).
		Minute().
		Tag(auditScope.GetId()).
		Do(svc.evaluateCatalog, context.Background(), auditScope, catalog, interval)
	if err != nil {
		slog.Error("Evaluation cannot be scheduled", slog.String("audit scope", auditScope.GetId()), log.Err(err))
		return connect.NewError(connect.CodeInternal, errors.New("evaluation cannot be scheduled"))
	}

	slog.Debug("Audit scope added to scheduler",
		slog.String("audit scope id", auditScope.GetId()))

	return
}

// evaluateCatalog evaluates all [orchestrator.Control] items in the catalog whether their associated metrics are
// fulfilled or not.
func (svc *Service) evaluateCatalog(ctx context.Context, auditScope *orchestrator.AuditScope, catalog *orchestrator.Catalog, interval int) error {
	var (
		controls   []*orchestrator.Control
		relevant   []*orchestrator.Control
		ignored    []string
		manual     map[string][]*evaluation.EvaluationResult
		inScopeIds map[string]struct{}
		err        error
		cancel     context.CancelFunc
	)

	// Retrieve all controls that match our assurance level, sorted by the control ID for easier debugging
	controls = slices.Collect(maps.Values(svc.catalogControls[auditScope.CatalogId]))
	slices.SortFunc(controls, func(a *orchestrator.Control, b *orchestrator.Control) int {
		return strings.Compare(a.Id, b.Id)
	})

	// Fetch ControlInScope records for this audit scope so we can skip
	// controls that have been explicitly removed from scope.
	inScopeIds, err = svc.fetchInScopeControlIds(ctx, auditScope.Id)
	if err != nil {
		slog.Warn("Could not fetch controls in scope, evaluating all controls", log.Err(err))
		// Fall back to evaluating all controls — treat every control as in scope
		inScopeIds = nil
	}

	// First, look for any manual evaluation results that are still within their validity period, to see whether we need to ignore some of the automated ones
	results, err := api.ListAllPaginated(ctx, &orchestrator.ListEvaluationResultsRequest{
		Filter: &orchestrator.ListEvaluationResultsRequest_Filter{
			TargetOfEvaluationId: &auditScope.TargetOfEvaluationId,
			CatalogId:            &auditScope.CatalogId,
			ValidManualOnly:      new(true),
		},
		LatestByControlId: new(true),
	},
		func(ctx context.Context, req *orchestrator.ListEvaluationResultsRequest) (*orchestrator.ListEvaluationResultsResponse, error) {
			res, err := svc.orchestratorClient.ListEvaluationResults(ctx, connect.NewRequest(req))
			if err != nil {
				return nil, err
			}
			return res.Msg, nil
		}, func(res *orchestrator.ListEvaluationResultsResponse) []*evaluation.EvaluationResult {
			return res.Results
		})
	if err != nil {
		err = fmt.Errorf("could not retrieve existing manual evaluation results: %w", err)
		return err
	}

	manual = make(map[string][]*evaluation.EvaluationResult)

	// Gather a list of controls, we are ignoring
	ignored = make([]string, 0, len(results))
	for _, result := range results {
		if result.GetParentControlId() != "" {
			manual[*result.ParentControlId] = append(manual[*result.ParentControlId], result)
		} else {
			ignored = append(ignored, result.ControlId)
		}
	}

	// Filter relevant controls (only parent controls)
	for _, c := range controls {
		// Only parent controls
		if c.ParentControlId != nil {
			continue
		}

		// If we ignore the control, we can skip it
		if slices.Contains(ignored, c.Id) {
			continue
		}

		// Skip controls that are not in scope for this audit scope
		if inScopeIds != nil {
			if _, ok := inScopeIds[c.Id]; !ok {
				continue
			}
		}

		if c.IsRelevantFor(auditScope, catalog) {
			relevant = append(relevant, c)
		}
	}

	slog.Info("Starting catalog evaluation",
		slog.String("target of evaluation id", auditScope.GetTargetOfEvaluationId()),
		slog.String("catalog id", auditScope.GetCatalogId()),
		slog.Int("number of relevant controls", len(relevant)),
		slog.Int("number of ignored controls", len(ignored)),
	)

	// We are using a timeout equal to the interval, so that we reduce premature cancellations
	// while still aiming to avoid overlapping executions.
	ctx, cancel = context.WithTimeout(context.Background(), time.Duration(interval)*time.Minute)
	defer cancel()

	g, gctx := errgroup.WithContext(ctx)
	for _, control := range relevant {
		g.Go(func() error {
			err := svc.evaluateControl(gctx, auditScope, catalog, control, manual[control.Id])
			if err != nil {
				return err
			}

			return nil
		})
	}

	// Wait until all sub-controls are evaluated
	err = g.Wait()
	if err != nil {
		slog.Error("Wait group error", log.Err(err))
		return err
	}

	return nil
}

// fetchInScopeControlIds returns a set of control IDs that are currently in
// scope for the given audit scope. Controls that have been removed from scope
// (no ControlInScope record) are excluded.
func (svc *Service) fetchInScopeControlIds(ctx context.Context, auditScopeId string) (map[string]struct{}, error) {
	cisList, err := api.ListAllPaginated(ctx, &orchestrator.ListControlsInScopeRequest{
		Filter: &orchestrator.ListControlsInScopeRequest_Filter{
			AuditScopeId: &auditScopeId,
		},
	}, func(ctx context.Context, req *orchestrator.ListControlsInScopeRequest) (*orchestrator.ListControlsInScopeResponse, error) {
		res, err := svc.orchestratorClient.ListControlsInScope(ctx, connect.NewRequest(req))
		if err != nil {
			return nil, err
		}
		return res.Msg, nil
	}, func(res *orchestrator.ListControlsInScopeResponse) []*orchestrator.ControlInScope {
		return res.ControlsInScope
	})
	if err != nil {
		return nil, err
	}

	// No ControlInScope records means there is no scoping information for this
	// audit scope (e.g. it was created before scoping existed) — return nil so
	// the caller evaluates all controls instead of none.
	if len(cisList) == 0 {
		return nil, nil
	}

	ids := make(map[string]struct{}, len(cisList))
	for _, cis := range cisList {
		ids[cis.ControlId] = struct{}{}
	}
	return ids, nil
}

// evaluateControl evaluates a control, e.g., OPS-13. Therefore, the method needs to wait till all sub-controls (e.g.,
// OPS-13.1) are evaluated.
func (svc *Service) evaluateControl(ctx context.Context, auditScope *orchestrator.AuditScope, catalog *orchestrator.Catalog, control *orchestrator.Control, manual []*evaluation.EvaluationResult) (err error) {
	var (
		status              = evaluation.EvaluationStatus_EVALUATION_STATUS_PENDING
		result              *evaluation.EvaluationResult
		evaluationResults   []*evaluation.EvaluationResult
		assessmentResultIds = []string{}
		relevantSubcontrol  []*orchestrator.Control
		ignored             []string
	)

	// TODO(lebogg): Don't think this is 100% correct. 1st) if all sub controls are manually evaluated we would ignore all of them and status would be still pending according to our logic below and 2nd) In theory, we could also have manual NON-complaint results. These would be then ignored but shouldn't be.
	// Gather a list of sub control IDs that we have manual results for and thus we are ignoring
	ignored = make([]string, 0, len(manual))
	for _, result := range manual {
		ignored = append(ignored, result.ControlId)
	}

	// Filter relevant controls
	for _, subControl := range control.Controls {
		// If we ignore the control, we can skip it
		if slices.Contains(ignored, subControl.Id) {
			continue
		}

		if subControl.IsRelevantFor(auditScope, catalog) {
			relevantSubcontrol = append(relevantSubcontrol, subControl)
		}
	}

	slog.Info("Starting control evaluation",
		slog.String("target of evaluation id", auditScope.TargetOfEvaluationId),
		slog.String("catalog id", auditScope.CatalogId),
		slog.String("control id", control.Id),
		slog.Int("number of relevant controls for the audit scope", len(relevantSubcontrol)))

	// Prepare the results slice
	evaluationResults = make([]*evaluation.EvaluationResult, len(relevantSubcontrol)+len(manual))

	// evaluate all subcontrols in parallel
	g, gctx := errgroup.WithContext(ctx)
	for i, sub := range relevantSubcontrol {
		g.Go(func() error {
			r, err := svc.evaluateSubcontrol(gctx, auditScope, sub)
			if err != nil {
				return err
			}
			evaluationResults[i] = r
			return nil
		})
	}

	// Wait until all sub-controls are evaluated
	err = g.Wait()
	if err != nil {
		slog.Error("Wait group error", log.Err(err))
		return
	}

	// Copy the manual results
	copy(evaluationResults[len(relevantSubcontrol):], manual)

	for _, r := range evaluationResults {
		// Special case: If the evaluation result of the parent control was set to "COMPLIANT MANUALLY", the whole
		// control will be evaluated as compliant, regardless of the subcontrol results.
		// Note: Depending on the ordering of the (sub)controls, we might lose some resultIds. Because manual results
		// are appended to the end (see above), it should be good, though. Also you could argue it doesn't matter with
		// a manual result.
		// If we have a manual compliant result for the parent control, we can skip all sub-controls and set the status to compliant manually. We can do this because the parent control is evaluated as compliant manually regardless of the sub-control results.
		// TODO(lebogg): This only works for two layered controls where we only have one parent control. For more than 1 sub controls we would need a more sophisticated approach (maybe add all sub controls of a manual result to the ignored list)
		if r.Status == evaluation.EvaluationStatus_EVALUATION_STATUS_COMPLIANT_MANUALLY && r.ParentControlId == nil {
			status = evaluation.EvaluationStatus_EVALUATION_STATUS_COMPLIANT_MANUALLY
			continue
		}

		// status is the current evaluation status, r.Status is the status of the evaluation result of the subcontrol
		// Note: Status should only contain the evaluation status without _MANUALLY!
		switch status {
		case evaluation.EvaluationStatus_EVALUATION_STATUS_PENDING:
			// check the given evaluation result for the current evaluation status PENDING
			status = handlePending(r)
		case evaluation.EvaluationStatus_EVALUATION_STATUS_COMPLIANT, evaluation.EvaluationStatus_EVALUATION_STATUS_COMPLIANT_MANUALLY:
			// check the given evaluation results for the current evaluation status COMPLIANT
			status = handleCompliant(r)
		case evaluation.EvaluationStatus_EVALUATION_STATUS_NOT_COMPLIANT, evaluation.EvaluationStatus_EVALUATION_STATUS_NOT_COMPLIANT_MANUALLY:
			// Evaluation status does not change if it is already not_compliant
		}

		// We are interested in all result IDs in order to provide a trace back from evaluation result back to assessment (and evidence).
		assessmentResultIds = append(assessmentResultIds, r.AssessmentResultIds...)
	}

	// Create evaluation result
	// slices.Compact only removes adjacent duplicates, so sort first to ensure full deduplication.
	slices.Sort(assessmentResultIds)

	result = &evaluation.EvaluationResult{
		Id:                   uuid.NewString(),
		Timestamp:            timestamppb.Now(),
		ControlCatalogId:     auditScope.CatalogId,
		ControlId:            control.Id,
		TargetOfEvaluationId: auditScope.TargetOfEvaluationId,
		AuditScopeId:         auditScope.Id,
		Status:               status,
		AssessmentResultIds:  slices.Compact(assessmentResultIds),
	}

	_, err = svc.orchestratorClient.StoreEvaluationResult(ctx, connect.NewRequest(&orchestrator.StoreEvaluationResultRequest{
		Result: result,
	}))
	if err != nil {
		slog.Error("Failed to send evaluation result to orchestrator", log.Err(err))
		return errors.New("failed to send evaluation result to orchestrator")
	}

	slog.Info("Evaluation result created",
		slog.String("control id", control.Id),
		slog.String("target of evaluation id", auditScope.TargetOfEvaluationId),
		slog.String("status", result.Status.String()))

	return
}

// evaluateSubcontrol evaluates the sub-controls, e.g., OPS-13.2
func (svc *Service) evaluateSubcontrol(ctx context.Context, auditScope *orchestrator.AuditScope, control *orchestrator.Control) (eval *evaluation.EvaluationResult, err error) {
	var (
		assessments []*assessment.AssessmentResult
		status      evaluation.EvaluationStatus
		resultIds   []string
	)

	// TODO(lebogg): Why we don't return an error here?
	if auditScope == nil || control == nil {
		slog.Error("Audit_scope and/or control is missing")
		return
	}

	// Get metrics from control and sub-controls
	metrics := getMetricsFromControl(control)
	slog.Debug("Evaluate subcontrol",
		slog.String("control_name", control.GetName()),
		slog.String("audit_scope_id", auditScope.GetId()),
		slog.Int("number of metrics", len(metrics)))
	if len(metrics) == 0 {
		slog.Error("Could not get metrics for",
			slog.String("control id", control.Id),
			slog.String("target of evaluation id", auditScope.GetTargetOfEvaluationId()),
			log.Err(err))
	}

	if len(metrics) != 0 {
		// Get latest assessment_results by resource_id filtered by
		// * target of evaluation id
		// * metric ids
		assessments, err = api.ListAllPaginated(ctx, &orchestrator.ListAssessmentResultsRequest{
			Filter: &orchestrator.ListAssessmentResultsRequest_Filter{
				TargetOfEvaluationId: &auditScope.TargetOfEvaluationId,
				MetricIds:            getMetricIds(metrics),
			},
			LatestByResourceId: new(true),
		}, func(ctx context.Context, req *orchestrator.ListAssessmentResultsRequest) (*orchestrator.ListAssessmentResultsResponse, error) {
			res, err := svc.orchestratorClient.ListAssessmentResults(ctx, connect.NewRequest(req))
			if err != nil {
				return nil, err
			}
			return res.Msg, nil
		}, func(res *orchestrator.ListAssessmentResultsResponse) []*assessment.AssessmentResult {
			return res.Results
		})

		if err != nil {
			// We let the scheduler running if we do not get the assessment results from the orchestrator, maybe it is
			// only a temporary network problem
			slog.Error("Could not get assessment results",
				slog.String("target of evaluation id", auditScope.GetTargetOfEvaluationId()),
				slog.Any("metric ids", getMetricIds(metrics)),
				log.Err(err))
		} else if len(assessments) == 0 {
			// We let the scheduler running if we do not get the assessment results from the orchestrator, maybe it is
			// only a temporary network problem
			slog.Debug("No assessment results available",
				slog.String("audit_scope_id", auditScope.GetId()),
				slog.Any("metric_ids", getMetricIds(metrics)))
		}
	} else {
		slog.Debug("No metrics available for the given control",
			slog.String("control_name", control.GetName()),
			slog.String("audit_scope_id", auditScope.GetId()))
	}

	// If no assessment_results are available we are stuck at pending
	if len(assessments) == 0 {
		status = evaluation.EvaluationStatus_EVALUATION_STATUS_PENDING
	} else {
		// Otherwise, there are some results and first we assume that everything is compliant, unless someone proves it
		// otherwise
		status = evaluation.EvaluationStatus_EVALUATION_STATUS_COMPLIANT
	}

	// Here the actual evaluation takes place. We check if the assessment results are compliant.
	for _, r := range assessments {
		if !r.Compliant {
			status = evaluation.EvaluationStatus_EVALUATION_STATUS_NOT_COMPLIANT
		}
		resultIds = append(resultIds, r.GetId())
	}

	// Create evaluation result
	eval = &evaluation.EvaluationResult{
		Id:                   uuid.NewString(),
		Timestamp:            timestamppb.Now(),
		ControlCatalogId:     auditScope.CatalogId,
		ControlId:            control.Id,
		ParentControlId:      control.ParentControlId,
		TargetOfEvaluationId: auditScope.TargetOfEvaluationId,
		AuditScopeId:         auditScope.Id,
		Status:               status,
		AssessmentResultIds:  resultIds,
	}

	_, err = svc.orchestratorClient.StoreEvaluationResult(ctx, connect.NewRequest(&orchestrator.StoreEvaluationResultRequest{
		Result: eval,
	}))
	if err != nil {
		slog.Error("Failed to send evaluation result to orchestrator", log.Err(err))
		return nil, errors.New("failed to send evaluation result to orchestrator")
	}

	slog.Info("Evaluation result created",
		slog.String("control id", control.Id),
		slog.String("target of evaluation id", auditScope.GetTargetOfEvaluationId()),
		slog.String("status", eval.Status.String()))

	return
}

// getMetricsFromControl returns all metrics from a given control. If the control has sub-controls, get also all metrics from the sub-controls.
func getMetricsFromControl(control *orchestrator.Control) (metrics []*assessment.Metric) {
	// Add metric of control to the metrics list
	metrics = append(metrics, control.Metrics...)

	// Add metric of sub-controls to the metrics list if exist
	for _, subControl := range control.Controls {
		metrics = append(metrics, subControl.Metrics...)
	}

	return metrics
}

// cacheControls caches the catalog controls for the given catalog.
func (svc *Service) cacheControls(catalogId string) error {
	var (
		err      error
		tag      string
		controls []*orchestrator.Control
	)

	if catalogId == "" {
		return errors.New("catalog ID is missing")
	}

	// Get controls for given catalog
	// TODO(anatheka): Shouldn´t we use the ListControlsInScope endpoint?
	controls, err = api.ListAllPaginated(context.Background(), &orchestrator.ListControlsRequest{
		Filter: &orchestrator.ListControlsRequest_Filter{
			CatalogId: &catalogId,
			Full:      new(true),
		},
	}, func(ctx context.Context, req *orchestrator.ListControlsRequest) (*orchestrator.ListControlsResponse, error) {
		res, err := svc.orchestratorClient.ListControls(ctx, connect.NewRequest(req))
		if err != nil {
			return nil, err
		}
		return res.Msg, nil
	}, func(res *orchestrator.ListControlsResponse) []*orchestrator.Control {
		return res.Controls
	})
	if err != nil {
		return err
	}

	if len(controls) == 0 {
		return fmt.Errorf("no controls for catalog '%s' available", catalogId)
	}

	// Store controls in map
	svc.catalogsMutex.Lock()
	svc.catalogControls[catalogId] = make(map[string]*orchestrator.Control)
	for _, control := range controls {
		tag = control.GetId()
		svc.catalogControls[catalogId][tag] = control
	}
	svc.catalogsMutex.Unlock()

	return nil
}

// handlePending evaluates the given evaluation result when the current control evaluation status is PENDING
func handlePending(er *evaluation.EvaluationResult) evaluation.EvaluationStatus {
	var (
		evalStatus = evaluation.EvaluationStatus_EVALUATION_STATUS_PENDING
	)

	switch er.Status {
	case evaluation.EvaluationStatus_EVALUATION_STATUS_PENDING:
		// Evaluation status does not change
	case evaluation.EvaluationStatus_EVALUATION_STATUS_COMPLIANT,
		evaluation.EvaluationStatus_EVALUATION_STATUS_COMPLIANT_MANUALLY:
		evalStatus = evaluation.EvaluationStatus_EVALUATION_STATUS_COMPLIANT
	case evaluation.EvaluationStatus_EVALUATION_STATUS_NOT_COMPLIANT,
		evaluation.EvaluationStatus_EVALUATION_STATUS_NOT_COMPLIANT_MANUALLY:
		evalStatus = evaluation.EvaluationStatus_EVALUATION_STATUS_NOT_COMPLIANT
	}

	return evalStatus
}

// handleCompliant evaluates the given evaluation result when the current control evaluation status is COMPLIANT
func handleCompliant(er *evaluation.EvaluationResult) evaluation.EvaluationStatus {
	var (
		evalStatus = evaluation.EvaluationStatus_EVALUATION_STATUS_COMPLIANT
	)

	switch er.Status {
	case evaluation.EvaluationStatus_EVALUATION_STATUS_PENDING, evaluation.EvaluationStatus_EVALUATION_STATUS_COMPLIANT, evaluation.EvaluationStatus_EVALUATION_STATUS_COMPLIANT_MANUALLY:
		// valuation status does not change
	case evaluation.EvaluationStatus_EVALUATION_STATUS_NOT_COMPLIANT, evaluation.EvaluationStatus_EVALUATION_STATUS_NOT_COMPLIANT_MANUALLY:
		evalStatus = evaluation.EvaluationStatus_EVALUATION_STATUS_NOT_COMPLIANT
	}

	return evalStatus
}

// getMetricIds returns the metric Ids for the given metrics
func getMetricIds(metrics []*assessment.Metric) []string {
	var metricIds []string

	for m := range slices.Values(metrics) {
		metricIds = append(metricIds, m.GetId())
	}

	return metricIds
}
