package collector

import (
	"context"
	"fmt"
	"net/http"

	"confirmate.io/core/api/evidence"
	"confirmate.io/core/api/evidence/evidenceconnect"
	"connectrpc.com/connect"
)

// EvidenceStoreSink sends evidence records to the evidence store service.
type EvidenceStoreSink struct {
	client evidenceconnect.EvidenceStoreClient
}

// NewEvidenceStoreSink creates a sink backed by the EvidenceStore connect client.
func NewEvidenceStoreSink(httpClient *http.Client, addr string) *EvidenceStoreSink {
	return &EvidenceStoreSink{
		client: evidenceconnect.NewEvidenceStoreClient(httpClient, addr),
	}
}

// StoreEvidence stores one evidence object in the evidence store service.
func (s *EvidenceStoreSink) StoreEvidence(ctx context.Context, ev *evidence.Evidence) error {
	if ev == nil {
		return fmt.Errorf("evidence must not be nil")
	}

	_, err := s.client.StoreEvidence(ctx, connect.NewRequest(&evidence.StoreEvidenceRequest{Evidence: ev}))
	if err != nil {
		return fmt.Errorf("storing evidence %q: %w", ev.GetId(), err)
	}

	return nil
}
