package k8s

import (
	"context"
	"testing"

	cloud "confirmate.io/collectors/cloud/api"
	"confirmate.io/collectors/cloud/internal/testdata"
	"confirmate.io/core/api/ontology"
	"confirmate.io/core/util/assert"

	"google.golang.org/protobuf/testing/protocmp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

func TestNewKubernetesComputeCollector(t *testing.T) {
	type args struct {
		intf                 kubernetes.Interface
		TargetOfEvaluationID string
	}
	tests := []struct {
		name string
		args args
		want cloud.Collector
	}{
		{
			name: "empty input",
			want: &k8sComputeCollector{
				k8sCollector: k8sCollector{},
			},
		},
		{
			name: "Happy path",
			args: args{
				intf:                 &fake.Clientset{},
				TargetOfEvaluationID: testdata.MockTargetOfEvaluationID1,
			},
			want: &k8sComputeCollector{
				k8sCollector: k8sCollector{
					intf: &fake.Clientset{},
					ctID: testdata.MockTargetOfEvaluationID1,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewKubernetesComputeCollector(tt.args.intf, tt.args.TargetOfEvaluationID)
			assert.Equal(t, tt.want, got, assert.CompareAllUnexported())
			assert.Equal(t, "Kubernetes Compute", got.Name())
		})
	}
}

func Test_k8sComputeCollector_List(t *testing.T) {
	var (
		volumeName      = "my-volume"
		diskName        = "my-disk"
		podCreationTime = metav1.Now()
		podName         = "my-pod"
		podID           = "/namespaces/my-namespace/containers/my-pod"
		podNamespace    = "my-namespace"
		podLabel        = map[string]string{"my": "label"}
	)

	client := fake.NewSimpleClientset()

	// Create an Pod with name, creationTimestamp and a AzureDisk volume
	p := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:              podName,
			CreationTimestamp: podCreationTime,
			Labels:            podLabel,
		},
		Spec: corev1.PodSpec{
			Volumes: []corev1.Volume{
				{
					Name: volumeName,
					VolumeSource: corev1.VolumeSource{
						AzureDisk: &corev1.AzureDiskVolumeSource{
							DiskName: diskName,
						},
					},
				},
			},
		},
	}
	_, err := client.CoreV1().Pods(podNamespace).Create(context.TODO(), p, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error injecting pod add: %v", err)
	}

	type fields struct {
		collector cloud.Collector
	}
	tests := []struct {
		name    string
		fields  fields
		want    assert.Want[[]ontology.IsResource]
		wantErr assert.WantErr
	}{
		{
			name: "Happy path",
			fields: fields{
				NewKubernetesComputeCollector(client, testdata.MockTargetOfEvaluationID1),
			},
			want: func(t *testing.T, got []ontology.IsResource, msgAndargs ...any) bool {
				container, ok := got[0].(*ontology.Container)
				if !assert.True(t, ok) {
					return false
				}
				// Create expected ontology.Container
				expectedContainer := &ontology.Container{
					Id:     podID,
					Name:   podName,
					Labels: podLabel,
					NetworkInterfaceIds: []string{
						podNamespace,
					},
				}

				// We need to ignore creation_time in the comparison because it is random and raw because it includes the creation_time
				assert.NotNil(t, container.CreationTime)
				assert.NotEmpty(t, container.Raw)
				assert.Equal(t, expectedContainer, container, protocmp.IgnoreFields(&ontology.Container{}, "creation_time", "raw"))

				// Check volume
				volume, ok := got[1].(*ontology.BlockStorage)
				assert.True(t, ok)

				// Create expected ontology.BlockStorage
				expectedVolume := &ontology.BlockStorage{
					Id:               volumeName,
					Name:             volumeName,
					CreationTime:     nil,
					AtRestEncryption: &ontology.AtRestEncryption{},
				}

				// We need to ignore raw because it contains the random creation time of the pod
				return assert.NotEmpty(t, container.Raw) && assert.Equal(t, expectedVolume, volume, protocmp.IgnoreFields(&ontology.BlockStorage{}, "raw"))
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := tt.fields.collector

			got, err := d.List()
			tt.wantErr(t, err)

			if tt.want != nil {
				tt.want(t, got)
			}
		})
	}
}

func Test_k8sComputeCollector_handlePodVolume(t *testing.T) {
	type fields struct {
		k8sCollector k8sCollector
	}
	type args struct {
		pod *corev1.Pod
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []ontology.IsResource
	}{
		{
			name:   "file storage",
			fields: fields{},
			args: args{
				pod: &corev1.Pod{
					Spec: corev1.PodSpec{
						Volumes: []corev1.Volume{
							{
								Name: "test",
								VolumeSource: corev1.VolumeSource{
									HostPath: &corev1.HostPathVolumeSource{
										Path: "/tmp",
									},
								},
							},
						},
					},
				},
			},
			want: []ontology.IsResource{
				&ontology.FileStorage{
					Id:               "test",
					Name:             "test",
					AtRestEncryption: &ontology.AtRestEncryption{},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &k8sComputeCollector{
				k8sCollector: tt.fields.k8sCollector,
			}

			got := d.handlePodVolume(tt.args.pod)

			// Check if raw field is set and then remove it for comparison
			for _, res := range got {
				switch r := res.(type) {
				case *ontology.FileStorage:
					assert.NotEmpty(t, r.Raw)
					r.Raw = ""
				}
			}

			assert.Equal(t, tt.want, got)
		})
	}
}
