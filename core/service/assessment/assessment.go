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

	"confirmate.io/core/api/assessment"
	"confirmate.io/core/api/assessment/assessmentconnect"
)

var (
	logger *slog.Logger
)

// Service is an implementation of the Clouditor Assessment service. It should not be used directly,
// but rather the NewService constructor should be used. It implements the AssessmentServer interface.
type Service struct {
	// Embedded for FWD compatibility
	assessmentconnect.UnimplementedAssessmentHandler
}

// NewService creates a new assessment service with default values.
func NewService() *Service {
	svc := &Service{}

	return svc
}

func (svc *Service) Init() {}

// AssessEvidence is a method implementation of the assessment interface: It assesses a single evidence
func (svc *Service) AssessEvidence(ctx context.Context, req *assessment.AssessEvidenceRequest) (res *assessment.AssessEvidenceResponse, err error) {

	res = &assessment.AssessEvidenceResponse{
		Status: assessment.AssessmentStatus_ASSESSMENT_STATUS_ASSESSED,
	}

	return res, nil
}
