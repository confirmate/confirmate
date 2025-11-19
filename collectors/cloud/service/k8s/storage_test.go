package k8s

import (
	"context"
	"testing"
	"time"

	cloud "confirmate.io/collectors/cloud/api"
	"confirmate.io/collectors/cloud/internal/testdata"
	"confirmate.io/core/api/ontology"
	"confirmate.io/core/util/assert"

	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/timestamppb"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

func TestNewKubernetesStorageCollector(t *testing.T) {
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
			want: &k8sStorageCollector{
				k8sCollector: k8sCollector{},
			},
		},
		{
			name: "Happy path",
			args: args{
				intf:                 &fake.Clientset{},
				TargetOfEvaluationID: testdata.MockTargetOfEvaluationID1,
			},
			want: &k8sStorageCollector{
				k8sCollector: k8sCollector{
					intf: &fake.Clientset{},
					ctID: testdata.MockTargetOfEvaluationID1,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewKubernetesStorageCollector(tt.args.intf, tt.args.TargetOfEvaluationID)
			assert.Equal(t, tt.want, got, assert.CompareAllUnexported())
			assert.Equal(t, "Kubernetes Storage", got.Name())
		})
	}
}

func Test_k8sStorageCollector_List(t *testing.T) {

	var (
		volumeName              = "my-volume"
		volumeUID               = "00000000-0000-0000-0000-000000000000"
		volumeCreationTime      = metav1.Now()
		volumeLabel             = map[string]string{"my": "label"}
		persistenVolumeDiskName = "my-disk"
	)

	client := fake.NewSimpleClientset()

	// Create persistent volumes
	v := &corev1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name:              volumeName,
			UID:               types.UID(volumeUID),
			CreationTimestamp: volumeCreationTime,
			Labels:            volumeLabel,
		},
		Spec: corev1.PersistentVolumeSpec{
			PersistentVolumeSource: corev1.PersistentVolumeSource{
				AzureDisk: &corev1.AzureDiskVolumeSource{
					DiskName: persistenVolumeDiskName,
				},
			},
		},
	}

	_, err := client.CoreV1().PersistentVolumes().Create(context.TODO(), v, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error injecting volume add: %v", err)
	}

	d := NewKubernetesStorageCollector(client, testdata.MockTargetOfEvaluationID1)

	list, err := d.List()
	assert.NoError(t, err)
	assert.NotNil(t, list)

	// Check persistentVolume
	volume, ok := list[0].(*ontology.BlockStorage)

	// Create expected ontology.BlockStorage
	expectedVolume := &ontology.BlockStorage{
		Id:               volumeUID,
		Name:             volumeName,
		CreationTime:     volume.CreationTime,
		Labels:           volumeLabel,
		AtRestEncryption: &ontology.AtRestEncryption{},
	}

	assert.True(t, ok)
	assert.Equal(t, expectedVolume, volume, protocmp.IgnoreFields(&ontology.BlockStorage{}, "raw"))
}

func Test_k8sStorageCollector_handlePV(t *testing.T) {
	type fields struct {
		k8sCollector k8sCollector
	}
	type args struct {
		pv *corev1.PersistentVolume
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   ontology.IsResource
	}{
		{
			name:   "file-based",
			fields: fields{},
			args: args{
				pv: &corev1.PersistentVolume{
					ObjectMeta: metav1.ObjectMeta{
						UID:               "my-id",
						Name:              "test",
						CreationTimestamp: metav1.NewTime(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
					},
					Spec: corev1.PersistentVolumeSpec{
						PersistentVolumeSource: corev1.PersistentVolumeSource{
							HostPath: &corev1.HostPathVolumeSource{
								Path: "/tmp",
							},
						},
					},
				},
			},
			want: &ontology.FileStorage{
				Id:               "my-id",
				Name:             "test",
				AtRestEncryption: &ontology.AtRestEncryption{},
				CreationTime:     timestamppb.New(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
				Raw:              `{"*v1.PersistentVolume":[{"metadata":{"name":"test","uid":"my-id","creationTimestamp":"2024-01-01T00:00:00Z"},"spec":{"hostPath":{"path":"/tmp"}},"status":{}}]}`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &k8sStorageCollector{
				k8sCollector: tt.fields.k8sCollector,
			}
			got := d.handlePV(tt.args.pv)
			assert.Equal(t, tt.want, got)
		})
	}
}
