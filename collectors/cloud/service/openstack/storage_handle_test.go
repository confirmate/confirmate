package openstack

import (
	"testing"
	"time"

	"confirmate.io/collectors/cloud/internal/testdata"
	"confirmate.io/core/api/ontology"
	"confirmate.io/core/util"
	"confirmate.io/core/util/assert"
	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack/blockstorage/v3/volumes"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func Test_openstackCollector_handleBlockStorage(t *testing.T) {
	testTime := time.Date(2000, 01, 20, 9, 20, 12, 123, time.UTC)

	type fields struct {
		ctID     string
		clients  clients
		authOpts *gophercloud.AuthOptions
		region   string
		domain   *domain
		project  *project
	}
	type args struct {
		volume *volumes.Volume
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    assert.Want[ontology.IsResource]
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "Happy path: volume name missing",
			fields: fields{
				region: "test region",
			},
			args: args{
				volume: &volumes.Volume{
					ID: testdata.MockOpenstackVolumeID1,
					// Name:      testdata.MockOpenstackVolumeName1,
					TenantID:  testdata.MockOpenstackVolumeTenantID,
					CreatedAt: testTime,
				},
			},
			want: func(t *testing.T, got ontology.IsResource) bool {
				want := &ontology.BlockStorage{
					Id:           testdata.MockOpenstackVolumeID1,
					Name:         testdata.MockOpenstackVolumeID1,
					CreationTime: timestamppb.New(testTime),
					GeoLocation: &ontology.GeoLocation{
						Region: "test region",
					},
					ParentId: util.Ref(testdata.MockOpenstackVolumeTenantID),
				}

				gotNew := got.(*ontology.BlockStorage)

				assert.NotEmpty(t, gotNew.GetRaw())
				gotNew.Raw = ""
				return assert.Equal(t, want, gotNew)
			},
			wantErr: assert.NoError,
		},
		{
			name: "Happy path: volume name available",
			fields: fields{
				region: "test region",
			},
			args: args{
				volume: &volumes.Volume{
					ID:        testdata.MockOpenstackVolumeID1,
					Name:      testdata.MockOpenstackVolumeName1,
					TenantID:  testdata.MockOpenstackVolumeTenantID,
					CreatedAt: testTime,
				},
			},
			want: func(t *testing.T, got ontology.IsResource) bool {
				want := &ontology.BlockStorage{
					Id:           testdata.MockOpenstackVolumeID1,
					Name:         testdata.MockOpenstackVolumeName1,
					CreationTime: timestamppb.New(testTime),
					GeoLocation: &ontology.GeoLocation{
						Region: "test region",
					},
					ParentId: util.Ref(testdata.MockOpenstackVolumeTenantID),
				}

				gotNew := got.(*ontology.BlockStorage)

				assert.NotEmpty(t, gotNew.GetRaw())
				gotNew.Raw = ""
				return assert.Equal(t, want, gotNew)
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &openstackCollector{
				ctID:     tt.fields.ctID,
				clients:  tt.fields.clients,
				authOpts: tt.fields.authOpts,
				region:   tt.fields.region,
				domain:   tt.fields.domain,
				project:  tt.fields.project,
			}
			got, err := d.handleBlockStorage(tt.args.volume)

			tt.want(t, got)
			tt.wantErr(t, err)
		})
	}
}
