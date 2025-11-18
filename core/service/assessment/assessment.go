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
	"log/slog"
	"net/http"

	"confirmate.io/core/api"
	"confirmate.io/core/api/assessment"
	"confirmate.io/core/api/assessment/assessmentconnect"
	"confirmate.io/core/api/orchestrator/orchestratorconnect"
	"confirmate.io/core/service"
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
	orchestratorConfig orchestratorConfig
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

func (svc *Service) Init() {}

// AssessEvidence is a method implementation of the assessment interface: It assesses a single evidence
func (svc *Service) AssessEvidence(ctx context.Context, req *assessment.AssessEvidenceRequest) (res *assessment.AssessEvidenceResponse, err error) {

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

	res = &assessment.AssessEvidenceResponse{
		Status: assessment.AssessmentStatus_ASSESSMENT_STATUS_ASSESSED,
	}

	return res, nil
}
