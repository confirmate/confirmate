package internal

const (
	// DefaultTargetOfEvaluationID is the default target of evaluation ID. Currently, our discoverers have no way to differentiate between different
	// targets, but we need this feature in the future. This serves as a default to already prepare the necessary
	// structures for this feature.
	DefaultTargetOfEvaluationID = "00000000-0000-0000-0000-000000000000"

	// DefaultEvidenceCollectorToolID is the default evidence collector tool ID.
	DefaultEvidenceCollectorToolID = "Confirmate Evidence Collector"
)
