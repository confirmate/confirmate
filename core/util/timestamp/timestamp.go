package timestamp

import (
	"time"

	"confirmate.io/core/api/assessment"
	"confirmate.io/core/api/evidence"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Now returns a new timestamp for the current time, truncating to microsecond precision.
func Now() *timestamppb.Timestamp {
	return New(time.Now().Truncate(time.Microsecond))
}

// New returns a new timestamp from the given time, truncating to microsecond precision.
func New(t time.Time) *timestamppb.Timestamp {
	return timestamppb.New(t.Truncate(time.Microsecond))
}

// NormalizeTimestampEvidence normalizes the timestamp in the given evidence to microsecond precision.
func NormalizeTimestampEvidence(ev *evidence.Evidence) {
	if ev.Timestamp != nil {
		ev.Timestamp = New(ev.Timestamp.AsTime())
	}
}

// NormalizeTimestampAssessmentResult normalizes the timestamp in the given assessment result to microsecond precision.
func NormalizeTimestampAssessmentResult(ar *assessment.AssessmentResult) {
	if ar.CreatedAt != nil {
		ar.CreatedAt = New(ar.CreatedAt.AsTime())
	}
	if ar.HistoryUpdatedAt != nil {
		ar.HistoryUpdatedAt = New(ar.HistoryUpdatedAt.AsTime())
	}
	if ar.History != nil {
		for _, h := range ar.History {
			if h.EvidenceRecordedAt != nil {
				h.EvidenceRecordedAt = New(h.EvidenceRecordedAt.AsTime())
			}
		}
	}
	if ar.ValidUntil != nil {
		ar.ValidUntil = New(ar.ValidUntil.AsTime())
	}
}
