// Copyright 2016-2025 Fraunhofer AISEC
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

package evidence

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"sync"

	"confirmate.io/core/api/assessment"
	"confirmate.io/core/api/assessment/assessmentconnect"
	"confirmate.io/core/api/evidence"
	"confirmate.io/core/api/evidence/evidenceconnect"
	"confirmate.io/core/api/ontology"
	"confirmate.io/core/persistence"
	"confirmate.io/core/service"
	"connectrpc.com/connect"

	"buf.build/go/protovalidate"
	"github.com/lmittmann/tint"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	logger *slog.Logger
)

const DefaultAssessmentURL = "localhost:9090"

type assessmentConfig struct {
	targetAddress string
	client        *http.Client
}

// Service is an implementation of the Confirmate req service (evidenceServer)
type Service struct {
	db *persistence.DB

	// TODO(lebogg): Test
	assessmentClient assessmentconnect.AssessmentClient
	assessmentStream *connect.BidiStreamForClient[assessment.AssessEvidenceRequest, assessment.AssessEvidencesResponse]
	streamMu         sync.Mutex
	assessmentConfig assessmentConfig

	// channel that is used to send evidences from the StoreEvidence method to the worker threat to process the evidence
	channelEvidence chan *evidence.Evidence

	// evidenceHooks is a list of hook functions that can be used if one wants to be
	// informed about each evidence
	evidenceHooks []evidence.EvidenceHookFunc
	// mu is used for (un)locking result hook calls
	mu sync.Mutex

	// TODO(all): Add authorization strategy

	evidenceconnect.EvidenceStoreHandler
}

func init() {
	logger = slog.New(tint.NewHandler(os.Stdout, nil))

	slog.SetDefault(logger)
}

func WithDB(db *persistence.DB) service.Option[*Service] {
	return func(svc *Service) {
		svc.db = db
	}
}

// WithAssessmentConfig is an option to configure the assessment service gRPC address.
func WithAssessmentConfig(target string, client *http.Client) service.Option[*Service] {

	return func(s *Service) {
		slog.Info("Assessment URL is set to %s", target)
		s.assessmentConfig.targetAddress = target
		s.assessmentConfig.client = client
	}
}

func NewService(opts ...service.Option[*Service]) (svc *Service) {
	var (
		err error
	)
	svc = &Service{
		assessmentConfig: assessmentConfig{
			targetAddress: DefaultAssessmentURL,
			client:        http.DefaultClient,
		},
	}

	for _, o := range opts {
		o(svc)
	}

	svc.assessmentClient = assessmentconnect.NewAssessmentClient(
		svc.assessmentConfig.client, svc.assessmentConfig.targetAddress)

	if svc.db == nil {
		svc.db, err = persistence.NewDB(persistence.WithAutoMigration(types))
		if err != nil {
			slog.Error("Could not initialize the storage: %v", err)
		}
	}

	// Create a channel to send evidence to the worker thread
	svc.initEvidenceChannel()

	return
}

// handleEvidence processes an evidence, sending it to the assessment service via a persistent stream.
func (svc *Service) handleEvidence(evidence *evidence.Evidence) error {
	// TODO(anatheka): It must be checked if the evidence changed since the last time and then send to the assessment service. Add in separate PR
	// Get or create the stream (lazy initialization)
	stream, err := svc.getOrCreateStream()
	if err != nil {
		return fmt.Errorf("could not get assessment stream: %w", err)
	}

	// Send evidence to the assessment service using the persistent stream
	// TODO(lebogg): Check EOF response of server
	err = stream.Send(&assessment.AssessEvidenceRequest{Evidence: evidence})
	if err != nil {
		// Invalidate the stream so it will be recreated on the next attempt
		svc.streamMu.Lock()
		svc.assessmentStream = nil
		svc.streamMu.Unlock()

		return fmt.Errorf("failed to send evidence %s: %w", evidence.Id, err)
	}

	return nil
}

// initEvidenceChannel initializes the channel used to send evidences from the StoreEvidence method to the worker threat
// to process the evidence.
func (svc *Service) initEvidenceChannel() {
	// Start a worker thread to process the evidence that is being passed to the StoreEvidence function to use the fire-and-forget strategy.
	go func() {
		for {
			// Wait for a new evidence to be passed to the channel
			e := <-svc.channelEvidence

			// Process the evidence
			err := svc.handleEvidence(e)
			if err != nil {
				slog.Error("Error while processing evidence: %v", err)
			}
		}
	}()
}

// StoreEvidence receives an evidence and stores it into the database
// This implements the [evidenceconnect.EvidenceStoreHandler.StoreEvidence] RPC method.
func (svc *Service) StoreEvidence(ctx context.Context, req *connect.Request[evidence.StoreEvidenceRequest]) (res *connect.Response[evidence.StoreEvidenceResponse], err error) {
	// Validate request
	if protovalidate.Validate(req.Msg) != nil {
		err = connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid request: %w", err))
		return
	}

	// Store evidence
	err = svc.db.Create(req.Msg.Evidence)
	if err != nil && errors.Is(err, persistence.ErrUniqueConstraintFailed) {
		return nil, status.Error(codes.AlreadyExists, persistence.ErrEntryAlreadyExists.Error())
	} else if err != nil {
		return nil, status.Errorf(codes.Internal, "%v: %v", persistence.ErrDatabase, err)
	}

	// Store Resource:
	// Build a resource struct. This will hold the latest sync state of the
	// resource for our storage layer. This is needed to store the resource in our DBs
	r, err := evidence.ToEvidenceResource(req.Msg.Evidence.GetOntologyResource(), req.Msg.GetTargetOfEvaluationId(), req.Msg.Evidence.GetToolId())
	if err != nil {
		slog.Error("Could not convert resource: %v", err)
		return nil, status.Errorf(codes.Internal, "could not convert resource: %v", err)
	}
	// Persist the latest state of the resource
	err = svc.db.Save(&r, "id = ?", r.Id)
	if err != nil {
		slog.Error("Could not save resource with ID '%s' to storage: %v", r.Id, err)
		return nil, status.Errorf(codes.Internal, "%v: %v", persistence.ErrDatabase, err)
	}

	go svc.informHooks(ctx, req.Msg.Evidence, nil)

	// Send evidence to the channel for further processing and acknowledge receipt, without waiting for the processing to finish. This allows the sender to continue
	// without waiting for the evidence to be processed.
	svc.channelEvidence <- req.Msg.Evidence

	slog.Debug("received and handled store evidence request: %v", req)
	res = connect.NewResponse(&evidence.StoreEvidenceResponse{})
	return
}

// StoreEvidences receives a stream of evidences and stores them to the evidence database.
// It processes each evidence individually and returns a response for each one indicating
// success or failure. This implements the [evidenceconnect.EvidenceStoreHandler.StoreEvidences] RPC method.
func (svc *Service) StoreEvidences(ctx context.Context,
	stream *connect.BidiStream[evidence.StoreEvidenceRequest, evidence.StoreEvidencesResponse]) error {
	var (
		req *evidence.StoreEvidenceRequest
		res *evidence.StoreEvidencesResponse
		err error
	)

	for {
		req, err = stream.Receive()
		// If no more input of the stream is available, return
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			err = fmt.Errorf("cannot receive stream request: %w", err)
			slog.Error(err.Error())
			return connect.NewError(connect.CodeUnknown, err)
		}

		// Call StoreEvidence() for storing a single evidence
		evidenceRequest := connect.NewRequest(&evidence.StoreEvidenceRequest{
			Evidence: req.Evidence,
		})
		_, err = svc.StoreEvidence(ctx, evidenceRequest)
		if err != nil {
			slog.Error("Error storing evidence", slog.Any("error", err))
			// Create a response message. The StoreEvidence method does not need that message, so we have to create it here for the stream response.
			res = &evidence.StoreEvidencesResponse{
				Status:        evidence.EvidenceStatus_EVIDENCE_STATUS_ERROR,
				StatusMessage: err.Error(),
			}
		} else {
			res = &evidence.StoreEvidencesResponse{
				Status: evidence.EvidenceStatus_EVIDENCE_STATUS_OK,
			}
		}

		// Send the response back to the client
		err = stream.Send(res)

		// Check for send errors
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			err = fmt.Errorf("cannot send response to the client: %w", err)
			slog.Error("failed to send response to client", slog.Any("error", err))
			return connect.NewError(connect.CodeUnknown, err)
		}
	}
}

// ListEvidences returns all evidence.
// This implements the [evidenceconnect.EvidenceStoreHandler.ListEvidences] RPC method.
// TODO(all): Add authorization (we currently just list all evidence, i.e. evidence for all TOEs
func (svc *Service) ListEvidences(_ context.Context, req *connect.Request[evidence.ListEvidencesRequest]) (
	res *connect.Response[evidence.ListEvidencesResponse], err error) {

	var (
		query []string
		args  []any
	)
	res = connect.NewResponse(&evidence.ListEvidencesResponse{})

	// Validate request
	err = protovalidate.Validate(req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid request: %w", err))
	}

	// Apply filter options
	if filter := req.Msg.GetFilter(); filter != nil {
		if TargetOfEvaluationId := filter.GetTargetOfEvaluationId(); TargetOfEvaluationId != "" {
			query = append(query, "target_of_evaluation_id = ?")
			args = append(args, TargetOfEvaluationId)
		}
		if toolId := filter.GetToolId(); toolId != "" {
			query = append(query, "tool_id = ?")
			args = append(args, toolId)
		}
	}

	// Paginate the evidences according to the request
	res.Msg.Evidences, res.Msg.NextPageToken, err = service.PaginateStorage[*evidence.Evidence](req.Msg, svc.db,
		service.DefaultPaginationOpts, persistence.BuildConds(query, args)...)

	if err != nil {
		err = connect.NewError(connect.CodeInternal, fmt.Errorf("could not paginate results: %w", err))
		return
	}

	return
}

// GetEvidence receives an evidenc ID and returns the corresponding evidence of the storage
// This implements the [evidenceconnect.EvidenceStoreHandler.GetEvidence] RPC method.
// TODO(all): Add authorization
func (svc *Service) GetEvidence(_ context.Context, req *connect.Request[evidence.GetEvidenceRequest]) (
	res *connect.Response[evidence.Evidence], err error) {

	var conds []any
	res = connect.NewResponse(&evidence.Evidence{})

	// Validate request
	err = protovalidate.Validate(req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid request: %w", err))
	}

	err = svc.db.Get(res.Msg, conds...)
	if errors.Is(err, persistence.ErrRecordNotFound) {
		return nil, status.Errorf(codes.NotFound, "evidence not found")
	} else if err != nil {
		return nil, status.Errorf(codes.Internal, "database error: %v", err)
	}

	return
}

// ListSupportedResourceTypes returns the resource types that are supported by this service
// This implements the [evidenceconnect.EvidenceStoreHandler.ListSupportedResourceTypes] RPC method.
func (svc *Service) ListSupportedResourceTypes(_ context.Context, req *connect.Request[evidence.ListSupportedResourceTypesRequest]) (
	res *connect.Response[evidence.ListSupportedResourceTypesResponse], err error) {

	res = connect.NewResponse(&evidence.ListSupportedResourceTypesResponse{})

	// Validate request
	err = protovalidate.Validate(req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid request: %w", err))
	}

	// Get the supported resource types
	res.Msg = &evidence.ListSupportedResourceTypesResponse{
		ResourceType: ontology.ListResourceTypes(),
	}

	return
}

// ListResources returns the list of resources, a pagination token, or an error if the operation fails.
// This implements the [evidenceconnect.EvidenceStoreHandler.ListResources] RPC method.
func (svc *Service) ListResources(_ context.Context, req *connect.Request[evidence.ListResourcesRequest]) (
	res *connect.Response[evidence.ListResourcesResponse], err error) {
	var (
		query []string
		args  []any
	)
	res = connect.NewResponse(&evidence.ListResourcesResponse{})

	// Validate request
	err = protovalidate.Validate(req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid request: %w", err))
	}

	// Filtering the resources by
	// * target of evaluation ID
	// * resource type
	// * tool ID
	if f := req.Msg.Filter; f != nil {
		if f.TargetOfEvaluationId != nil {
			query = append(query, "target_of_evaluation_id = ?")
			args = append(args, f.GetTargetOfEvaluationId())
		}
		if f.Type != nil {
			query = append(query, "(resource_type LIKE ? OR resource_type LIKE ? OR resource_type LIKE ?)")
			args = append(args, f.GetType()+",%", "%,"+f.GetType()+",%", "%,"+f.GetType())
		}
		if f.ToolId != nil {
			query = append(query, "tool_id = ?")
			args = append(args, f.GetToolId())
		}
	}

	res.Msg = new(evidence.ListResourcesResponse)

	// Join query with AND and prepend the query
	args = append([]any{strings.Join(query, " AND ")}, args...)

	res.Msg.Results, res.Msg.NextPageToken, err = service.PaginateStorage[*evidence.Resource](req.Msg, svc.db, service.DefaultPaginationOpts, args...)

	return
}

func (svc *Service) RegisterEvidenceHook(evidenceHook evidence.EvidenceHookFunc) {
	svc.mu.Lock()
	defer svc.mu.Unlock()
	svc.evidenceHooks = append(svc.evidenceHooks, evidenceHook)
}

func (svc *Service) informHooks(ctx context.Context, result *evidence.Evidence, err error) {
	svc.mu.Lock()
	defer svc.mu.Unlock()

	// Inform our hook if we have any
	if svc.evidenceHooks != nil {
		for _, hook := range svc.evidenceHooks {
			// TODO(all): We could do hook concurrent again (assuming different hooks don't interfere with each other)
			hook(ctx, result, err)
		}
	}
}
