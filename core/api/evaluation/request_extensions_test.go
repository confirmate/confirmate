package evaluation_test

import (
	"testing"

	"confirmate.io/core/api/evaluation"
	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/service/evidence/evidencetest"
)

func TestStoreEvaluationResultRequest_GetTargetOfEvaluationId(t *testing.T) {
	type args struct {
		req *orchestrator.StoreEvaluationResultRequest
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Happy path",
			want: evidencetest.MockTargetOfEvaluationID1,
			args: args{
				req: &orchestrator.StoreEvaluationResultRequest{
					Result: &evaluation.EvaluationResult{
						TargetOfEvaluationId: evidencetest.MockTargetOfEvaluationID1,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.args.req.GetTargetOfEvaluationId()

			if got != tt.want {
				t.Errorf("StoreEvaluationResultRequest.GetTargetOfEvaluationId() = %v, want %v", got, tt.want)
			}
		})
	}
}
