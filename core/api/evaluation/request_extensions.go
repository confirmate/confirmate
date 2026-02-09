package evaluation

// GetTargetOfEvaluationId is a shortcut to implement TargetOfEvaluationRequest. It returns
// the target of evaluation ID of the inner object.
func (req *CreateEvaluationResultRequest) GetTargetOfEvaluationId() string {
	return req.GetResult().GetTargetOfEvaluationId()
}
