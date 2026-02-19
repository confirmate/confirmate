package evaluation

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"

	"connectrpc.com/connect"
	"golang.org/x/sync/errgroup"

	"confirmate.io/core/api"
	"confirmate.io/core/api/assessment"
	"confirmate.io/core/api/evaluation"
	"confirmate.io/core/api/evaluation/evaluationconnect"
	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/api/orchestrator/orchestratorconnect"
	"confirmate.io/core/log"
	"confirmate.io/core/persistence"
	"confirmate.io/core/service"
	"confirmate.io/core/util"

	"github.com/go-co-op/gocron"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	DefaultOrchestratorURL = "http://localhost:9090"

	// defaultInterval is the default interval time for the scheduler. If no interval is set in the StartEvaluationRequest, the default value is taken.
	defaultInterval int = 5
)

// Service implements the Evaluation Service handler (see
// [evaluationconnect.EvaluationHandler]).
type Service struct {
	evaluationconnect.UnimplementedEvaluationHandler
	db  persistence.DB
	cfg Config

	orchestratorClient orchestratorconnect.OrchestratorClient

	streamMutex sync.Mutex

	scheduler *gocron.Scheduler

	// TODO(lebogg): Try to use  map[catalogId]map[controlKey]*orchestrator.Control where catalogId is typed string and controlKey is struct and has has categoryName and controlId fields.
	// controls stores the catalog controls so that they do not always have to be retrieved from Orchestrators getControl endpoint
	// map[catalog_id][category_name-control_id]*orchestrator.Control
	catalogControls map[string]map[string]*orchestrator.Control
	catalogsMutex   sync.RWMutex
}

// DefaultConfig is the default configuration for the evaluation [Service].
var DefaultConfig = Config{
	OrchestratorAddress: DefaultOrchestratorURL,
	OrchestratorClient:  http.DefaultClient,
	PersistenceConfig:   persistence.DefaultConfig,
}

// Config represents the configuration for the evaluation [Service].
type Config struct {
	// OrchestratorAddress is the address of the Orchestrator service to connect to.
	OrchestratorAddress string
	// OrchestratorClient is the HTTP client to use for connecting to the Orchestrator service.
	OrchestratorClient *http.Client

	// PersistenceConfig is the configuration for the persistence layer. If not set, defaults will be used.
	PersistenceConfig persistence.Config
}

// WithConfig sets the service configuration, overriding the default configuration.
func WithConfig(cfg Config) service.Option[Service] {
	return func(svc *Service) {
		svc.cfg = cfg
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

	// Initialize the database with the defined auto-migration types and join tables
	pcfg := svc.cfg.PersistenceConfig
	pcfg.Types = types
	svc.db, err = persistence.NewDB(persistence.WithConfig(pcfg))
	if err != nil {
		return nil, fmt.Errorf("could not create db: %w", err)
	}

	// Initialize the Orchestrator client
	svc.orchestratorClient = orchestratorconnect.NewOrchestratorClient(svc.cfg.OrchestratorClient, svc.cfg.OrchestratorAddress)

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

	// Get Audit Scope
	auditScopeRes, err = svc.orchestratorClient.GetAuditScope(ctx, connect.NewRequest(&orchestrator.GetAuditScopeRequest{
		AuditScopeId: req.Msg.GetAuditScopeId(),
	}))
	if err != nil {
		newErr := fmt.Errorf("could not start evaluation: %w", service.ErrNotFound("audit scope"))
		slog.Error("%w: %w", log.Err(newErr), log.Err(err))
		return nil, status.Errorf(codes.NotFound, "%s", newErr)
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
		statusErr := fmt.Errorf("could not cache controls")
		slog.Error("%w: %w", log.Err(statusErr), log.Err(err))
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	// Retrieve the catalog
	catalogRes, err = svc.orchestratorClient.GetCatalog(ctx, connect.NewRequest(&orchestrator.GetCatalogRequest{
		CatalogId: auditScope.GetCatalogId(),
	}))
	if err != nil {
		statusErr := fmt.Errorf("could not get catalog: %w", service.ErrNotFound("catalog"))
		slog.Error("%w: %w", log.Err(statusErr), log.Err(err))
		return nil, status.Errorf(codes.Internal, "%s", err)
	}
	catalog = catalogRes.Msg

	// Check, if a previous job exists and/or is running
	jobs, err = svc.scheduler.FindJobsByTag(auditScope.GetId())
	if err != nil && !errors.Is(err, gocron.ErrJobNotFoundWithTag) {
		err = fmt.Errorf("error while retrieving existing scheduler job: %w", err)
		slog.Error("%w", log.Err(err))
		return nil, status.Errorf(codes.Internal, "internal error")
	} else if len(jobs) > 0 {
		err = fmt.Errorf("evaluation for Audit Scope '%s' (target of evaluation '%s' and catalog ID '%s') already started", auditScope.GetId(), auditScope.GetTargetOfEvaluationId(), auditScope.GetCatalogId())
		slog.Error("%w", log.Err(err))
		return nil, status.Errorf(codes.AlreadyExists, "%s", err)
	}

	slog.Info("Starting evaluation ...")

	// Add job to scheduler
	err = svc.addJobToScheduler(ctx, auditScope, catalog, interval)
	// We can return the error as it is
	if err != nil {
		return nil, err
	}

	slog.Info("Scheduled to evaluate audit scope '%s' every %d minutes...",
		slog.String("audit scope", auditScope.GetId()),
		slog.Int("interval", interval),
	)

	res = connect.NewResponse(&evaluation.StartEvaluationResponse{
		Successful: true,
	})

	return res, nil
}

// StopEvaluation is a method implementation of the evaluation interface: It stops the evaluation for a
// AuditScope.
func (svc *Service) StopEvaluation(ctx context.Context, req *connect.Request[evaluation.StopEvaluationRequest]) (res *connect.Response[evaluation.StopEvaluationResponse], err error) {
	var auditScopeResponse *connect.Response[orchestrator.AuditScope]

	// Validate the request
	if err = service.Validate(req); err != nil {
		return nil, err
	}

	// Get audit scope
	// auditScope, err = svc.orchestratorClient.GetAuditScope(context.Background(), &orchestrator.GetAuditScopeRequest{
	// 	AuditScopeId: req.GetAuditScopeId(),
	// })
	auditScopeResponse, err = svc.orchestratorClient.GetAuditScope(ctx, connect.NewRequest(&orchestrator.GetAuditScopeRequest{
		AuditScopeId: req.Msg.GetAuditScopeId(),
	}))
	if err != nil {
		newErr := fmt.Errorf("could not stop evaluation: %w", service.ErrNotFound("audit scope"))
		slog.Error("%w: %w", log.Err(newErr), log.Err(err))
		return nil, status.Errorf(codes.Internal, "%s", newErr)
	}

	auditScope := auditScopeResponse.Msg

	// Stop jobs(s) for given audit scope
	err = svc.scheduler.RemoveByTags(auditScope.GetId())
	if err != nil && errors.Is(err, gocron.ErrJobNotFoundWithTag) {
		return nil, status.Errorf(codes.FailedPrecondition, "job for audit scope '%s' not running", auditScope.GetId())
	} else if err != nil {
		err = fmt.Errorf("error while removing jobs for audit scope '%s': %w", auditScope.GetId(), err)
		slog.Error("%w", log.Err(err))
		return nil, status.Errorf(codes.Internal, "%s", err)
	}

	res = &connect.Response[evaluation.StopEvaluationResponse]{}

	return
}

// ListEvaluationResults is a method implementation of the assessment interface
func (svc *Service) ListEvaluationResults(_ context.Context,
	req *connect.Request[evaluation.ListEvaluationResultsRequest],
) (res *connect.Response[evaluation.ListEvaluationResultsResponse], err error) {
	var (
		query     []string
		partition []string
		args      []any
	)

	// Validate the request
	if err = service.Validate(req); err != nil {
		return nil, err
	}

	// Filtering evaluation results by
	// * target of evaluation ID
	// * control ID
	// * sub-controls
	if req.Msg.Filter != nil {
		if req.Msg.Filter.TargetOfEvaluationId != nil {
			query = append(query, "target_of_evaluation_id = ?")
			args = append(args, req.Msg.Filter.GetTargetOfEvaluationId())
		}

		if req.Msg.Filter.CatalogId != nil {
			query = append(query, "control_catalog_id = ?")
			args = append(args, req.Msg.Filter.GetCatalogId())
		}

		if req.Msg.Filter.ControlId != nil {
			query = append(query, "control_id = ?")
			args = append(args, req.Msg.Filter.GetControlId())
		}

		// TODO(anatheka): change that, in other catalogs maybe it's not that easy to get the sub-control by name
		if req.Msg.Filter.SubControls != nil {
			partition = append(partition, "control_id")
			query = append(query, "control_id LIKE ?")
			args = append(args, fmt.Sprintf("%s%%", req.Msg.Filter.GetSubControls()))
		}

		if util.Deref(req.Msg.Filter.ParentsOnly) {
			query = append(query, "parent_control_id IS NULL")
		}

		if util.Deref(req.Msg.Filter.ValidManualOnly) {
			query = append(query, "status IN ?")
			args = append(args, []any{
				evaluation.EvaluationStatus_EVALUATION_STATUS_COMPLIANT_MANUALLY,
				evaluation.EvaluationStatus_EVALUATION_STATUS_NOT_COMPLIANT_MANUALLY,
			})

			// Use parameterized query instead of CURRENT_TIMESTAMP SQL function for compatibility with in-memory test
			// database (ramsql)
			query = append(query, "valid_until IS NULL OR valid_until >= ?")
			args = append(args, time.Now())
		}
	}

	res = &connect.Response[evaluation.ListEvaluationResultsResponse]{Msg: &evaluation.ListEvaluationResultsResponse{Results: make([]*evaluation.EvaluationResult, 0)}}

	// If we want to have it grouped by resource ID, we need to do a raw query
	if req.Msg.GetLatestByControlId() {
		// In the raw SQL, we need to build the whole WHERE statement
		var (
			where string
			p     = ""
		)

		if len(query) > 0 {
			where = "WHERE " + strings.Join(query, " AND ")
		}

		if len(partition) > 0 {
			p = ", " + strings.Join(partition, ",")
		}

		// Execute the raw SQL statement
		err = svc.db.Raw(&res.Msg.Results,
			fmt.Sprintf(`WITH sorted_results AS (
				SELECT *, ROW_NUMBER() OVER (PARTITION BY control_id %s ORDER BY timestamp DESC) AS row_number
				FROM evaluation_results
				%s
		  	)
		  	SELECT * FROM sorted_results WHERE row_number = 1 ORDER BY control_catalog_id, control_id;`, p, where), args...)
		if err = service.HandleDatabaseError(err); err != nil {
			return nil, err
		}
	} else {
		// join query with AND and prepend the query
		args = append([]any{strings.Join(query, " AND ")}, args...)

		// Paginate the results according to the request
		res.Msg.Results, res.Msg.NextPageToken, err = service.PaginateStorage[*evaluation.EvaluationResult](req.Msg, svc.db, service.DefaultPaginationOpts, args...)
		if err = service.HandleDatabaseError(err); err != nil {
			return nil, err
		}
	}

	return
}

// CreateEvaluationResult is a method implementation of the assessment interface to store only manually created Evaluation Results
func (svc *Service) CreateEvaluationResult(ctx context.Context, req *connect.Request[evaluation.CreateEvaluationResultRequest]) (res *connect.Response[evaluation.EvaluationResult], err error) {
	var (
		eval *evaluation.EvaluationResult
	)

	// Validate the request
	if err = validateCreateEvaluationResultRequest(req); err != nil {
		return nil, err
	}

	eval = req.Msg.Result
	err = svc.db.Create(eval)
	// TODO(lebogg): Add Test
	if err = service.HandleDatabaseError(err); err != nil {
		return nil, err
	}

	res = connect.NewResponse(eval)

	return res, nil
}

// validateCreateEvaluationResultRequest validates the CreateEvaluationResultRequest and prepares it for processing.
func validateCreateEvaluationResultRequest(req *connect.Request[evaluation.CreateEvaluationResultRequest]) error {
	// Validate the request with a preparation function
	if err := service.ValidateWithPrep(req, func() {
		// Check if Result is nil before accessing it to avoid nil pointer dereference. ValidateWithPrep will then
		// return the `invalid request` error
		if req.Msg.Result == nil {
			return
		}
		// A manually created evaluation result typically does not contain a UUID; therefore, we will add one here. This must be done before the validation check to prevent validation failure.
		req.Msg.Result.Id = uuid.NewString()
	}); err != nil {
		return err
	}

	// We only allow manually created statuses
	if req.Msg.Result.Status != evaluation.EvaluationStatus_EVALUATION_STATUS_COMPLIANT_MANUALLY &&
		req.Msg.Result.Status != evaluation.EvaluationStatus_EVALUATION_STATUS_NOT_COMPLIANT_MANUALLY {
		return connect.NewError(connect.CodeInvalidArgument, errors.New("only manually set statuses are allowed"))
	}

	// The ValidUntil field must be checked separately as it is an optional field and not checked by the request
	// validation. It is only mandatory when manually creating a result.
	if req.Msg.Result.ValidUntil == nil {
		return connect.NewError(connect.CodeInvalidArgument, errors.New("validity must be set"))
	}

	return nil
}

// addJobToScheduler adds a job for the given control to the scheduler and sets the scheduler interval to the given interval
func (svc *Service) addJobToScheduler(ctx context.Context, auditScope *orchestrator.AuditScope, catalog *orchestrator.Catalog, interval int) (err error) {
	// Check inputs and log error
	if auditScope == nil {
		err = errors.New("audit scope is invalid")
	}
	if interval == 0 {
		err = errors.New("interval is invalid")
	}
	if err != nil {
		statusErr := fmt.Errorf("evaluation cannot be scheduled")
		slog.Error("%w: %w", log.Err(statusErr), log.Err(err))
		return status.Errorf(codes.Internal, "%s", statusErr)
	}

	_, err = svc.scheduler.
		Every(interval).
		Minute().
		Tag(auditScope.GetId()).
		Do(svc.evaluateCatalog, ctx, auditScope, catalog, interval)
	if err != nil {
		statusErr := fmt.Errorf("evaluation for audit scope '%s' cannot be scheduled", auditScope.GetId())
		slog.Error("%w: %w", log.Err(statusErr), log.Err(err))
		return status.Errorf(codes.Internal, "%s", statusErr)
	}

	slog.Debug("Audit scope '%s' added to scheduler", slog.String("audit scope", auditScope.GetId()))

	return
}

// evaluateCatalog evaluates all [orchestrator.Control] items in the catalog whether their associated metrics are
// fulfilled or not.
func (svc *Service) evaluateCatalog(ctx context.Context, auditScope *orchestrator.AuditScope, catalog *orchestrator.Catalog, interval int) error {
	var (
		controls []*orchestrator.Control
		relevant []*orchestrator.Control
		ignored  []string
		manual   map[string][]*evaluation.EvaluationResult
		err      error
		g        *errgroup.Group
		cancel   context.CancelFunc
	)

	// TODO(lebogg) Where is assurance level matched?
	// Retrieve all controls that match our assurance level, sorted by the control ID for easier debugging
	controls = values(svc.catalogControls[auditScope.CatalogId])
	slices.SortFunc(controls, func(a *orchestrator.Control, b *orchestrator.Control) int {
		return strings.Compare(a.Id, b.Id)
	})

	// First, look for any manual evaluation results that are still within their validity period, to see whether we need
	// to ignore some of the automated ones
	//
	// TODO(oxisto): Its problematic to use the context from the original StartEvaluation request, since this token
	// might time out at some point
	results, err := api.ListAllPaginated(ctx, &evaluation.ListEvaluationResultsRequest{
		Filter: &evaluation.ListEvaluationResultsRequest_Filter{
			TargetOfEvaluationId: &auditScope.TargetOfEvaluationId,
			CatalogId:            &auditScope.CatalogId,
			ValidManualOnly:      util.Ref(true),
		},
		LatestByControlId: util.Ref(true),
	},
		func(ctx context.Context, req *evaluation.ListEvaluationResultsRequest) (*evaluation.ListEvaluationResultsResponse, error) {
			res, err := svc.ListEvaluationResults(ctx, connect.NewRequest(req))
			if err != nil {
				return nil, err
			}
			return res.Msg, nil
		}, func(res *evaluation.ListEvaluationResultsResponse) []*evaluation.EvaluationResult {
			return res.Results
		})
	if err != nil {
		slog.Error("could not retrieve existing manual evaluation results: %w", log.Err(err))
		return err
	}

	manual = make(map[string][]*evaluation.EvaluationResult)

	// Gather a list of controls, we are ignoring
	ignored = make([]string, 0, len(results))
	for _, result := range results {
		if result.ParentControlId != nil {
			manual[*result.ParentControlId] = append(manual[*result.ParentControlId], result)
		} else {
			ignored = append(ignored, result.ControlId)
		}
	}

	// Filter relevant controls
	for _, c := range controls {
		// Only parent controls
		if c.ParentControlId != nil {
			continue
		}

		// If we ignore the control, we can skip it
		if slices.Contains(ignored, c.Id) {
			continue
		}

		if c.IsRelevantFor(auditScope, catalog) {
			relevant = append(relevant, c)
		}
	}

	slog.Info("Starting catalog evaluation for Target of Evaluation '%s', Catalog ID '%s'. Waiting for the evaluation of %d control(s)",
		slog.String("target of evaluation id", auditScope.GetTargetOfEvaluationId()),
		slog.String("catalog id", auditScope.GetCatalogId()),
		slog.Int("number of relevant controls for the audit scope", len(relevant)),
	)

	// We are using a timeout of half the interval, so that we are not running into overlapping executions
	ctx, cancel = context.WithTimeout(ctx, time.Duration(interval)*time.Minute/2)
	defer cancel()

	g, ctx = errgroup.WithContext(ctx)
	for _, control := range relevant {
		control := control // https://golang.org/doc/faq#closures_and_goroutines needed until Go 1.22 (loopvar)
		g.Go(func() error {
			err := svc.evaluateControl(ctx, auditScope, catalog, control, manual[control.Id])
			if err != nil {
				return err
			}

			return nil
		})
	}

	// Wait until all sub-controls are evaluated
	err = g.Wait()
	if err != nil {
		slog.Error("wait group error", log.Err(err))
		return err
	}

	return nil
}

// evaluateControl evaluates a control, e.g., OPS-13. Therefore, the method needs to wait till all sub-controls (e.g.,
// OPS-13.1) are evaluated.
func (svc *Service) evaluateControl(ctx context.Context, auditScope *orchestrator.AuditScope, catalog *orchestrator.Catalog, control *orchestrator.Control, manual []*evaluation.EvaluationResult) (err error) {
	var (
		status   = evaluation.EvaluationStatus_EVALUATION_STATUS_PENDING
		result   *evaluation.EvaluationResult
		results  []*evaluation.EvaluationResult
		relevant []*orchestrator.Control
		ignored  []string
		g        *errgroup.Group
	)

	// Gather a list of sub control IDs that we have manual results for and thus we are ignoring
	ignored = make([]string, 0, len(manual))
	for _, result := range manual {
		ignored = append(ignored, result.ControlId)
	}

	// Filter relevant controls
	for _, sub := range control.Controls {
		// If we ignore the control, we can skip it
		if slices.Contains(ignored, sub.Id) {
			continue
		}

		if sub.IsRelevantFor(auditScope, catalog) {
			relevant = append(relevant, sub)
		}
	}

	slog.Info("Starting control evaluation for Target of Evaluation '%s', Catalog ID '%s' and Control '%s'. Waiting for the evaluation of %d sub-control(s)",
		slog.String("target of evaluation id", auditScope.TargetOfEvaluationId),
		slog.String("catalog id", auditScope.CatalogId),
		slog.String("control id", control.Id),
		slog.Int("number of relevant controls for the audit scope", len(relevant)),
	)

	// Prepare the results slice
	results = make([]*evaluation.EvaluationResult, len(relevant)+len(manual))

	g, ctx = errgroup.WithContext(ctx)
	for i, sub := range relevant {
		i, sub := i, sub // https://golang.org/doc/faq#closures_and_goroutines needed until Go 1.22 (loopvar)
		g.Go(func() error {
			result, err := svc.evaluateSubcontrol(ctx, auditScope, sub)
			if err != nil {
				return err
			}

			results[i] = result
			return nil
		})
	}

	// Wait until all sub-controls are evaluated
	err = g.Wait()
	if err != nil {
		slog.Error("wait group error", log.Err(err))
		return
	}

	// Copy the manual results
	copy(results[len(relevant):], manual)

	var resultIds = []string{}

	for _, r := range results {
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
		resultIds = append(resultIds, r.AssessmentResultIds...)
	}

	// Create evaluation result
	result = &evaluation.EvaluationResult{
		Id:                   uuid.NewString(),
		Timestamp:            timestamppb.Now(),
		ControlCategoryName:  control.CategoryName,
		ControlCatalogId:     control.CategoryCatalogId,
		ControlId:            control.Id,
		TargetOfEvaluationId: auditScope.TargetOfEvaluationId,
		AuditScopeId:         auditScope.Id,
		Status:               status,
		AssessmentResultIds:  resultIds,
	}

	err = svc.db.Create(result)
	if err = service.HandleDatabaseError(err); err != nil {
		return err
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

	if auditScope == nil || control == nil {
		slog.Error("input is missing")
		return
	}

	// Get metrics from control and sub-controls
	metrics, err := svc.getAllMetricsFromControl(auditScope.GetCatalogId(), control.CategoryName, control.Id)
	if err != nil {
		slog.Error("could not get metrics for",
			slog.String("control id", control.Id),
			slog.String("target of evaluation id", auditScope.GetTargetOfEvaluationId()),
			log.Err(err))
		return
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
			LatestByResourceId: util.Ref(true),
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
			slog.Error("could not get assessment results",
				slog.String("target of evaluation id", auditScope.GetTargetOfEvaluationId()),
				slog.Any("metric ids", getMetricIds(metrics)),
				log.Err(err))
		} else if len(assessments) == 0 {
			// We let the scheduler running if we do not get the assessment results from the orchestrator, maybe it is
			// only a temporary network problem
			slog.Debug("no assessment results available",
				slog.String("target of evaluation id", auditScope.GetTargetOfEvaluationId()),
				slog.Any("metric ids", getMetricIds(metrics)))
		}
	} else {
		slog.Debug("no metrics available for the given control")
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
		ControlCategoryName:  control.CategoryName,
		ControlCatalogId:     control.CategoryCatalogId,
		ControlId:            control.Id,
		ParentControlId:      control.ParentControlId,
		TargetOfEvaluationId: auditScope.TargetOfEvaluationId,
		AuditScopeId:         auditScope.Id,
		Status:               status,
		AssessmentResultIds:  resultIds,
	}

	err = svc.db.Create(eval)
	if err = service.HandleDatabaseError(err); err != nil {
		return nil, err
	}

	slog.Info("Evaluation result created",
		slog.String("control id", control.Id),
		slog.String("target of evaluation id", auditScope.GetTargetOfEvaluationId()),
		slog.String("status", eval.Status.String()))

	return
}

// getAllMetricsFromControl returns all metrics from a given controlId.
//
// For now a control has either sub-controls or metrics. If the control has sub-controls, get also all metrics from the
// sub-controls.
func (svc *Service) getAllMetricsFromControl(catalogId, categoryName, controlId string) (metrics []*assessment.Metric, err error) {
	var subControlMetrics []*assessment.Metric

	control, err := svc.getControl(catalogId, categoryName, controlId)
	if err != nil {
		err = fmt.Errorf("could not get control for control id {%s}: %w", controlId, err)
		return
	}

	// Add metric of control to the metrics list
	metrics = append(metrics, control.Metrics...)

	// Add sub-control metrics to the metric list if exist
	if len(control.Controls) != 0 {
		// Get the metrics from the next sub-control
		subControlMetrics, err = svc.getMetricsFromSubcontrols(control)
		if err != nil {
			err = fmt.Errorf("error getting metrics from sub-controls: %w", err)
			return
		}

		metrics = append(metrics, subControlMetrics...)
	}

	return
}

// getMetricsFromSubcontrols returns a list of metrics from the sub-controls.
func (svc *Service) getMetricsFromSubcontrols(control *orchestrator.Control) (metrics []*assessment.Metric, err error) {
	var subcontrol *orchestrator.Control

	if control == nil {
		return nil, errors.New("control is missing")
	}

	for _, control := range control.Controls {
		subcontrol, err = svc.getControl(control.CategoryCatalogId, control.CategoryName, control.Id)
		if err != nil {
			return
		}

		metrics = append(metrics, subcontrol.Metrics...)
	}

	return
}

// cacheControls caches the catalog controls for the given catalog.
func (svc *Service) cacheControls(catalogId string) error {
	var (
		err      error
		tag      string
		controls []*orchestrator.Control
	)

	if catalogId == "" {
		return service.ErrIsMissing("catalog ID")
	}

	// Get controls for given catalog
	controls, err = api.ListAllPaginated(context.Background(), &orchestrator.ListControlsRequest{
		CatalogId: catalogId,
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
	svc.catalogControls[catalogId] = make(map[string]*orchestrator.Control)
	for _, control := range controls {
		tag = fmt.Sprintf("%s-%s", control.GetCategoryName(), control.GetId())
		svc.catalogControls[catalogId][tag] = control
	}

	return nil
}

// getControl returns the control for the given catalogID, CategoryName and controlID.
func (svc *Service) getControl(catalogId, categoryName, controlId string) (control *orchestrator.Control, err error) {
	if catalogId == "" {
		return nil, service.ErrIsMissing("catalog id")
	} else if categoryName == "" {
		return nil, service.ErrIsMissing("category name")
	} else if controlId == "" {
		return nil, service.ErrIsMissing("control id")
	}

	tag := fmt.Sprintf("%s-%s", categoryName, controlId)

	control, ok := svc.catalogControls[catalogId][tag]
	if !ok {
		return nil, service.ErrControlNotAvailable
	}

	return
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

// TODO(lebogg): Try it out
// TODO(oxisto): We can remove it with maps.Values in Go 1.22+
func values[M ~map[K]V, K comparable, V any](m M) []V {
	rr := make([]V, 0, len(m))

	for _, v := range m {
		rr = append(rr, v)
	}

	return rr
}
