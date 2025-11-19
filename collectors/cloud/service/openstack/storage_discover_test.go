package openstack

import (
	"testing"
	"time"

	"confirmate.io/collectors/cloud/internal/collectortest/openstacktest"
	"confirmate.io/collectors/cloud/internal/testdata"
	"confirmate.io/core/api/ontology"
	"confirmate.io/core/util"
	"confirmate.io/core/util/assert"
	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/testhelper"
	"github.com/gophercloud/gophercloud/v2/testhelper/client"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func Test_openstackCollector_collectBlockStorage(t *testing.T) {
	testhelper.SetupHTTP()
	defer testhelper.TeardownHTTP()

	openstacktest.MockStorageListResponse(t)

	type fields struct {
		ctID     string
		clients  clients
		authOpts *gophercloud.AuthOptions
		region   string
		domain   *domain
		project  *project
	}
	tests := []struct {
		name     string
		fields   fields
		wantList assert.Want[[]ontology.IsResource]
		wantErr  assert.ErrorAssertionFunc
	}{
		{
			name: "Happy path",
			fields: fields{
				authOpts: &gophercloud.AuthOptions{
					IdentityEndpoint: testdata.MockOpenstackIdentityEndpoint,
					Username:         testdata.MockOpenstackUsername,
					Password:         testdata.MockOpenstackPassword,
					TenantName:       testdata.MockOpenstackTenantName,
				},
				clients: clients{
					provider: &gophercloud.ProviderClient{
						TokenID: client.TokenID,
						EndpointLocator: func(eo gophercloud.EndpointOpts) (string, error) {
							return testhelper.Endpoint(), nil
						},
					},
					storageClient: client.ServiceClient(),
				},
				region:  "test region",
				domain:  &domain{},
				project: &project{},
			},
			wantList: func(t *testing.T, got []ontology.IsResource) bool {
				assert.Equal(t, 2, len(got))

				t1, err := time.Parse("2006-01-02T15:04:05.000000", "2015-09-17T03:35:03.000000")
				assert.NoError(t, err)

				want := &ontology.BlockStorage{
					Id:           "289da7f8-6440-407c-9fb4-7db01ec49164",
					Name:         "vol-001",
					Description:  "",
					CreationTime: timestamppb.New(t1),
					GeoLocation: &ontology.GeoLocation{
						Region: "test region",
					},
					ParentId: util.Ref("83ec2e3b-4321-422b-8706-a84185f52a0a"),
					Labels:   map[string]string{},
				}

				got0 := got[0].(*ontology.BlockStorage)

				assert.NotEmpty(t, got0.GetRaw())
				got0.Raw = ""
				return assert.Equal(t, want, got0)
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
			gotList, err := d.collectBlockStorage()

			tt.wantList(t, gotList)
			tt.wantErr(t, err)
		})
	}
}
