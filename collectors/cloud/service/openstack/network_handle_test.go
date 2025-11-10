package openstack

import (
	"testing"
	"time"

	"confirmate.io/collectors/cloud/internal/testdata"
	"confirmate.io/core/api/ontology"
	"confirmate.io/core/util"
	"confirmate.io/core/util/assert"
	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/networks"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func Test_openstackDiscovery_handleNetworkInterfaces(t *testing.T) {
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
		network *networks.Network
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    assert.Want[ontology.IsResource]
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "Happy path",
			fields: fields{
				region: "test region",
			},
			args: args{
				network: &networks.Network{
					ID:        testdata.MockOpenstackNetworkID1,
					Name:      testdata.MockOpenstackNetworkName1,
					ProjectID: testdata.MockOpenstackServerTenantID,
					CreatedAt: testTime,
				},
			},
			want: func(t *testing.T, got ontology.IsResource) bool {
				want := &ontology.NetworkInterface{
					Id:           testdata.MockOpenstackNetworkID1,
					Name:         testdata.MockOpenstackNetworkName1,
					CreationTime: timestamppb.New(testTime),
					GeoLocation: &ontology.GeoLocation{
						Region: "test region",
					},
					ParentId: util.Ref(testdata.MockOpenstackServerTenantID),
				}

				gotNew := got.(*ontology.NetworkInterface)

				assert.NotEmpty(t, gotNew.GetRaw())
				gotNew.Raw = ""
				return assert.Equal(t, want, gotNew)
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &openstackDiscovery{
				ctID:     tt.fields.ctID,
				clients:  tt.fields.clients,
				authOpts: tt.fields.authOpts,
				region:   tt.fields.region,
				domain:   tt.fields.domain,
				project:  tt.fields.project,
			}
			got, err := d.handleNetworkInterfaces(tt.args.network)

			tt.want(t, got)
			tt.wantErr(t, err)
		})
	}
}
