package evidence

import (
	"context"
	"log/slog"

	"confirmate.io/core/api/assessment"
	"connectrpc.com/connect"
)

// getOrCreateStream returns a stream to the assessment service. If a stream already exists, it is returned.
// Otherwise, a new stream is created and returned.
func (svc *Service) getOrCreateStream() (*connect.BidiStreamForClient[assessment.AssessEvidenceRequest, assessment.AssessEvidencesResponse], error) {
	svc.streamMu.Lock()
	defer svc.streamMu.Unlock()

	// If we already have a stream, return it
	if svc.assessmentStream != nil {
		return svc.assessmentStream, nil
	}

	// Create new stream and
	slog.Info("Creating new stream to assessment service", slog.Any("target address", svc.assessmentConfig.targetAddress))
	svc.assessmentStream = svc.assessmentClient.AssessEvidences(context.Background())

	return svc.assessmentStream, nil
}
