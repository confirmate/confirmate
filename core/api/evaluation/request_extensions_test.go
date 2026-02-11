package evaluation_test

import (
	"testing"

	"confirmate.io/core/api/evaluation"
	"confirmate.io/core/service/evidence/evidencetest"
)

func TestCreateEvaluationResultRequest_GetTargetOfEvaluationId(t *testing.T) {
	type args struct {
		req *evaluation.CreateEvaluationResultRequest
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
				req: &evaluation.CreateEvaluationResultRequest{
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
				t.Errorf("CreateEvaluationResultRequest.GetTargetOfEvaluationId() = %v, want %v", got, tt.want)
			}
		})
	}
}
