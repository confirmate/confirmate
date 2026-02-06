package k8s

import (
	"testing"

	"confirmate.io/collectors/cloud/internal/testdata"
	"k8s.io/client-go/kubernetes"
)

func Test_k8sCollector_TargetOfEvaluationID(t *testing.T) {
	type fields struct {
		intf kubernetes.Interface
		ctID string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "Happy path",
			fields: fields{
				ctID: testdata.MockTargetOfEvaluationID1,
			},
			want: testdata.MockTargetOfEvaluationID1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &k8sCollector{
				intf: tt.fields.intf,
				ctID: tt.fields.ctID,
			}
			if got := d.TargetOfEvaluationID(); got != tt.want {
				t.Errorf("k8sCollector.TargetOfEvaluationID() = %v, want %v", got, tt.want)
			}
		})
	}
}
